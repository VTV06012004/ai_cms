package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	sqlfiles "api/migrations"
)

// newIsolatedMySQL tạo một database riêng cho từng test và xóa đi khi test xong.
func newIsolatedMySQL(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("set MYSQL_TEST_DSN to a disposable MySQL 8.4 server to run integration tests")
	}
	config, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("parse MYSQL_TEST_DSN: %v", err)
	}
	admin, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("connect to MySQL test server: %v", err)
	}
	t.Cleanup(func() { _ = admin.Close() })

	name := fmt.Sprintf("ai_cms_migrate_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4"); err != nil {
		t.Fatalf("create isolated database: %v", err)
	}
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE `" + name + "`") })

	config.DBName = name
	config.MultiStatements = true
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		t.Fatalf("connect to isolated database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func countRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return count
}

func TestRunAppliesAllMigrationsAndIsRepeatable(t *testing.T) {
	db := newIsolatedMySQL(t)
	migrations, err := Load(sqlfiles.FS)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if err := Run(context.Background(), db, migrations, t.Logf); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if got := countRows(t, db, "SELECT COUNT(*) FROM schema_migrations"); got != len(migrations) {
		t.Fatalf("schema_migrations has %d rows, want %d", got, len(migrations))
	}
	for _, table := range []string{"users", "posts", "post_tags", "revisions", "audit_logs"} {
		if got := countRows(t, db,
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table); got != 1 {
			t.Fatalf("table %s was not created", table)
		}
	}

	if err := Run(context.Background(), db, migrations, t.Logf); err != nil {
		t.Fatalf("second Run should be a no-op: %v", err)
	}
	if got := countRows(t, db, "SELECT COUNT(*) FROM schema_migrations"); got != len(migrations) {
		t.Fatalf("second Run changed schema_migrations: %d rows, want %d", got, len(migrations))
	}
}

func TestRunRejectsEditedMigrationFile(t *testing.T) {
	db := newIsolatedMySQL(t)
	migrations, err := Load(sqlfiles.FS)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := Run(context.Background(), db, migrations, t.Logf); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := db.Exec("UPDATE schema_migrations SET checksum = 'tampered' WHERE version = 1"); err != nil {
		t.Fatalf("tamper checksum: %v", err)
	}
	if err := Run(context.Background(), db, migrations, t.Logf); err == nil {
		t.Fatal("Run accepted a migration whose recorded checksum no longer matches the file")
	}
}

func TestRunRefusesLegacyDatabaseUntilBaselined(t *testing.T) {
	db := newIsolatedMySQL(t)
	migrations, err := Load(sqlfiles.FS)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Mô phỏng DB cũ do AutoMigrate tạo: đã có bảng nhưng chưa có schema_migrations.
	if _, err := db.Exec("CREATE TABLE users (id INT PRIMARY KEY)"); err != nil {
		t.Fatalf("create legacy table: %v", err)
	}

	if err := Run(context.Background(), db, migrations, t.Logf); !errors.Is(err, ErrLegacySchema) {
		t.Fatalf("Run on legacy database = %v, want ErrLegacySchema", err)
	}
	if err := Baseline(context.Background(), db, migrations, len(migrations), t.Logf); err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if err := Run(context.Background(), db, migrations, t.Logf); err != nil {
		t.Fatalf("Run after Baseline should do nothing: %v", err)
	}
	if got := countRows(t, db, "SELECT COUNT(*) FROM schema_migrations"); got != len(migrations) {
		t.Fatalf("schema_migrations has %d rows, want %d", got, len(migrations))
	}
	if err := Baseline(context.Background(), db, migrations, 1, t.Logf); err == nil {
		t.Fatal("Baseline should refuse to run twice")
	}
}
