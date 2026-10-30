package persistence

import (
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
)

func newPostSlugMySQLTestDB(t *testing.T, createLegacyIndex bool) (*gorm.DB, uint, uint) {
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

	if err := db.AutoMigrate(&UserModel{}, &CategoryModel{}, &TagModel{}, &PostModel{}); err != nil {
		t.Fatalf("auto-migrate test schema: %v", err)
	}
	if createLegacyIndex {
		if err := db.Exec("CREATE UNIQUE INDEX idx_posts_slug_deleted ON posts (slug, deleted_at)").Error; err != nil {
			t.Fatalf("create legacy slug index: %v", err)
		}
	} else if err := MigrateActivePostSlugIndex(db); err != nil {
		t.Fatalf("migrate active post slug index: %v", err)
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
	db, userID, categoryID := newPostSlugMySQLTestDB(t, false)
	repository := &PostRepository{db: db}

	first, err := repository.Create(newTestPost(t, userID, categoryID, "duplicate-slug"), nil)
	if err != nil {
		t.Fatalf("create first post: %v", err)
	}
	if _, err := repository.Create(newTestPost(t, userID, categoryID, "duplicate-slug"), nil); !errors.Is(err, domain.ErrPostSlugAlreadyExists) {
		t.Fatalf("duplicate active slug error = %v, want ErrPostSlugAlreadyExists", err)
	}

	second, err := repository.Create(newTestPost(t, userID, categoryID, "other-slug"), nil)
	if err != nil {
		t.Fatalf("create second post: %v", err)
	}
	secondData := second.Data()
	if err := second.Update("Updated title", "duplicate-slug", "updated content", domain.PostStatusDraft, categoryID); err != nil {
		t.Fatalf("update post domain data: %v", err)
	}
	if _, err := repository.Update(second, nil); !errors.Is(err, domain.ErrPostSlugAlreadyExists) {
		t.Fatalf("duplicate slug on update error = %v, want ErrPostSlugAlreadyExists", err)
	}

	second = domain.ReconstitutePost(secondData)
	if err := second.Update("Updated title", "other-slug", "updated content", domain.PostStatusDraft, categoryID); err != nil {
		t.Fatalf("update post without changing slug: %v", err)
	}
	if _, err := repository.Update(second, nil); err != nil {
		t.Fatalf("updating post with its existing slug should succeed: %v", err)
	}

	if err := repository.Delete(first.Data().ID); err != nil {
		t.Fatalf("soft-delete first post: %v", err)
	}
	reused, err := repository.Create(newTestPost(t, userID, categoryID, "duplicate-slug"), nil)
	if err != nil {
		t.Fatalf("reuse slug after deleting previous post: %v", err)
	}
	if err := repository.Delete(reused.Data().ID); err != nil {
		t.Fatalf("soft-delete reused post: %v", err)
	}
	if _, err := repository.Create(newTestPost(t, userID, categoryID, "duplicate-slug"), nil); err != nil {
		t.Fatalf("multiple deleted posts may retain the same slug: %v", err)
	}
}

func TestPostRepositoryMySQLConcurrentCreatesOnlyAllowOneActiveSlug(t *testing.T) {
	db, userID, categoryID := newPostSlugMySQLTestDB(t, false)
	repository := &PostRepository{db: db}

	const attempts = 2
	errorsFound := make(chan error, attempts)
	var waitGroup sync.WaitGroup
	for i := 0; i < attempts; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := repository.Create(newTestPost(t, userID, categoryID, "concurrent-slug"), nil)
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

func TestMigrateActivePostSlugIndexFailsWithLegacyDuplicatesAndCanRetry(t *testing.T) {
	db, userID, categoryID := newPostSlugMySQLTestDB(t, true)
	for _, title := range []string{"First", "Second"} {
		post := PostModel{
			Title: title, Slug: "legacy-duplicate", Status: "draft",
			AuthorID: userID, CategoryID: categoryID,
		}
		if err := db.Create(&post).Error; err != nil {
			t.Fatalf("insert legacy duplicate post: %v", err)
		}
	}

	if err := MigrateActivePostSlugIndex(db); err == nil {
		t.Fatal("migration should fail rather than accept duplicate active slugs")
	}
	var duplicateIDs []uint
	if err := db.Model(&PostModel{}).Where("slug = ?", "legacy-duplicate").Pluck("id", &duplicateIDs).Error; err != nil {
		t.Fatalf("query legacy duplicates: %v", err)
	}
	if len(duplicateIDs) != 2 {
		t.Fatalf("expected migration failure to preserve both legacy rows, got IDs %v", duplicateIDs)
	}
	if err := db.Delete(&PostModel{}, duplicateIDs[1]).Error; err != nil {
		t.Fatalf("resolve legacy duplicate: %v", err)
	}
	if err := MigrateActivePostSlugIndex(db); err != nil {
		t.Fatalf("migration should succeed after resolving duplicates: %v", err)
	}
	if !db.Migrator().HasIndex(&PostModel{}, "idx_posts_active_slug") {
		t.Fatal("migration did not create the active slug unique index")
	}
	if db.Migrator().HasIndex(&PostModel{}, "idx_posts_slug_deleted") {
		t.Fatal("migration did not remove the legacy composite index")
	}
}
