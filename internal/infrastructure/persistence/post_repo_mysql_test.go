package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"api/internal/domain"
	"api/internal/infrastructure/migrate"
	sqlfiles "api/migrations"
)

func newPostSlugMySQLTestDB(t *testing.T) (*gorm.DB, uint, uint) {
	t.Helper()
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("set MYSQL_TEST_DSN to a disposable MySQL 8.4 server to run integration tests")
	}

	config, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("parse MYSQL_TEST_DSN: %v", err)
	}
	adminDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to MySQL test server: %v", err)
	}
	adminSQL, err := adminDB.DB()
	if err != nil {
		t.Fatalf("get MySQL admin connection: %v", err)
	}
	t.Cleanup(func() { _ = adminSQL.Close() })

	databaseName := fmt.Sprintf("ai_cms_slug_test_%d", time.Now().UnixNano())
	if err := adminDB.Exec("CREATE DATABASE `" + databaseName + "` CHARACTER SET utf8mb4").Error; err != nil {
		t.Fatalf("create isolated integration database: %v", err)
	}
	t.Cleanup(func() {
		_ = adminDB.Exec("DROP DATABASE `" + databaseName + "`").Error
	})

	config.DBName = databaseName
	db, err := gorm.Open(mysql.Open(config.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to isolated integration database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get isolated MySQL connection: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	migrationDSN, err := migrate.WithMultiStatements(config.FormatDSN())
	if err != nil {
		t.Fatalf("build migration DSN: %v", err)
	}
	migrationDB, err := sql.Open("mysql", migrationDSN)
	if err != nil {
		t.Fatalf("open migration connection: %v", err)
	}
	t.Cleanup(func() { _ = migrationDB.Close() })
	migrations, err := migrate.Load(sqlfiles.FS)
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	if err := migrate.Run(context.Background(), migrationDB, migrations, t.Logf); err != nil {
		t.Fatalf("apply migrations to test schema: %v", err)
	}

	user := UserModel{Name: "Test Author", Email: "author@example.com", PasswordHash: "test-hash", Role: "author"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create test author: %v", err)
	}
	category := CategoryModel{Name: "Test", Slug: fmt.Sprintf("test-%d", time.Now().UnixNano())}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create test category: %v", err)
	}
	return db, user.ID, category.ID
}

func newTestPost(t *testing.T, userID, categoryID uint, slug string) *domain.Post {
	t.Helper()
	post, err := domain.NewPost("Test "+slug, slug, "test content", domain.PostStatusDraft, categoryID, userID)
	if err != nil {
		t.Fatalf("create domain post: %v", err)
	}
	return post
}

func TestTranslatePostWriteErrorRecognizesActiveSlugConflict(t *testing.T) {
	err := translatePostWriteError(&mysqlDriver.MySQLError{
		Number:  1062,
		Message: "Duplicate entry 'hello' for key 'posts.idx_posts_active_slug'",
	})
	if !errors.Is(err, domain.ErrPostSlugAlreadyExists) {
		t.Fatalf("expected slug conflict sentinel, got %v", err)
	}
	if err.Error() == domain.ErrPostSlugAlreadyExists.Error() {
		t.Fatal("expected original database cause to remain available for diagnostics")
	}
}

func TestTranslatePostWriteErrorDoesNotMislabelOtherUniqueConflicts(t *testing.T) {
	err := translatePostWriteError(&mysqlDriver.MySQLError{
		Number:  1062,
		Message: "Duplicate entry 'hello' for key 'posts.other_unique_index'",
	})
	if errors.Is(err, domain.ErrPostSlugAlreadyExists) {
		t.Fatal("non-slug unique conflict was reported as a slug conflict")
	}
}

func TestPostRepositoryMySQLEnforcesActiveSlugUniquenessAndAllowsReuseAfterDelete(t *testing.T) {
	db, userID, categoryID := newPostSlugMySQLTestDB(t)
	repository := &PostRepository{db: db}

	first, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "duplicate-slug"), nil)
	if err != nil {
		t.Fatalf("create first post: %v", err)
	}
	if _, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "duplicate-slug"), nil); !errors.Is(err, domain.ErrPostSlugAlreadyExists) {
		t.Fatalf("duplicate active slug error = %v, want ErrPostSlugAlreadyExists", err)
	}

	second, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "other-slug"), nil)
	if err != nil {
		t.Fatalf("create second post: %v", err)
	}
	secondData := second.Data()
	if err := second.Update("Updated title", "duplicate-slug", "updated content", domain.PostStatusDraft, categoryID); err != nil {
		t.Fatalf("update post domain data: %v", err)
	}
	if _, err := repository.Update(context.Background(), second, nil); !errors.Is(err, domain.ErrPostSlugAlreadyExists) {
		t.Fatalf("duplicate slug on update error = %v, want ErrPostSlugAlreadyExists", err)
	}

	second = domain.ReconstitutePost(secondData)
	if err := second.Update("Updated title", "other-slug", "updated content", domain.PostStatusDraft, categoryID); err != nil {
		t.Fatalf("update post without changing slug: %v", err)
	}
	if _, err := repository.Update(context.Background(), second, nil); err != nil {
		t.Fatalf("updating post with its existing slug should succeed: %v", err)
	}

	if err := repository.Delete(context.Background(), first.Data().ID); err != nil {
		t.Fatalf("soft-delete first post: %v", err)
	}
	reused, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "duplicate-slug"), nil)
	if err != nil {
		t.Fatalf("reuse slug after deleting previous post: %v", err)
	}
	if err := repository.Delete(context.Background(), reused.Data().ID); err != nil {
		t.Fatalf("soft-delete reused post: %v", err)
	}
	if _, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "duplicate-slug"), nil); err != nil {
		t.Fatalf("multiple deleted posts may retain the same slug: %v", err)
	}
}

func TestPostRepositoryMySQLConcurrentCreatesOnlyAllowOneActiveSlug(t *testing.T) {
	db, userID, categoryID := newPostSlugMySQLTestDB(t)
	repository := &PostRepository{db: db}

	const attempts = 2
	errorsFound := make(chan error, attempts)
	var waitGroup sync.WaitGroup
	for i := 0; i < attempts; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := repository.Create(context.Background(), newTestPost(t, userID, categoryID, "concurrent-slug"), nil)
			errorsFound <- err
		}()
	}
	waitGroup.Wait()
	close(errorsFound)

	successes, conflicts := 0, 0
	for err := range errorsFound {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrPostSlugAlreadyExists):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent create error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent create results: successes=%d conflicts=%d, want 1 each", successes, conflicts)
	}
}

func TestPostRepositoryMySQLPublishIfDueOnlyOneConcurrentCallerWins(t *testing.T) {
	db, userID, categoryID := newPostSlugMySQLTestDB(t)
	repository := &PostRepository{db: db}

	scheduledAt := time.Now().Add(-time.Minute).Truncate(time.Second)
	model := PostModel{
		Title: "Due", Slug: "due-post", Status: string(domain.PostStatusScheduled),
		AuthorID: userID, CategoryID: categoryID, ScheduledAt: &scheduledAt,
	}
	if err := db.Create(&model).Error; err != nil {
		t.Fatalf("create scheduled post: %v", err)
	}

	const callers = 8
	results := make(chan bool, callers)
	failures := make(chan error, callers)
	var waitGroup sync.WaitGroup
	for i := 0; i < callers; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			won, err := repository.PublishIfDue(context.Background(), model.ID, time.Now())
			if err != nil {
				failures <- err
				return
			}
			results <- won
		}()
	}
	waitGroup.Wait()
	close(results)
	close(failures)

	for err := range failures {
		t.Fatalf("PublishIfDue returned error: %v", err)
	}
	winners := 0
	for won := range results {
		if won {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("winners = %d, want exactly 1", winners)
	}

	var stored PostModel
	if err := db.First(&stored, model.ID).Error; err != nil {
		t.Fatalf("reload post: %v", err)
	}
	if stored.Status != string(domain.PostStatusPublished) || stored.ScheduledAt != nil {
		t.Fatalf("stored status=%q scheduled_at=%v, want published with no schedule", stored.Status, stored.ScheduledAt)
	}
	if stored.PublishedAt == nil || !stored.PublishedAt.Equal(scheduledAt) {
		t.Fatalf("published_at = %v, want the original scheduled time %v", stored.PublishedAt, scheduledAt)
	}
}

func TestPostRepositoryMySQLPublishIfDueSkipsPostsThatAreNotDue(t *testing.T) {
	db, userID, categoryID := newPostSlugMySQLTestDB(t)
	repository := &PostRepository{db: db}

	future := time.Now().Add(time.Hour).Truncate(time.Second)
	posts := []PostModel{
		{Title: "Future", Slug: "future-post", Status: string(domain.PostStatusScheduled), AuthorID: userID, CategoryID: categoryID, ScheduledAt: &future},
		{Title: "Draft", Slug: "draft-post", Status: string(domain.PostStatusDraft), AuthorID: userID, CategoryID: categoryID},
	}
	for i := range posts {
		if err := db.Create(&posts[i]).Error; err != nil {
			t.Fatalf("create post %q: %v", posts[i].Title, err)
		}
		won, err := repository.PublishIfDue(context.Background(), posts[i].ID, time.Now())
		if err != nil || won {
			t.Fatalf("PublishIfDue(%q) = %v, %v; want false, nil", posts[i].Title, won, err)
		}
	}
}
