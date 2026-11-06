// Package migrate chạy các file migration SQL có đánh số phiên bản.
//
// Lịch sử đã chạy được lưu trong bảng schema_migrations. Mỗi file được ghi kèm
// checksum; nếu file đã chạy bị sửa thì Run từ chối chạy tiếp.
package migrate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const lockName = "ai_cms_schema_migrations"

// ErrLegacySchema báo DB đã có bảng nhưng chưa có lịch sử migration.
var ErrLegacySchema = errors.New("database đã có bảng nhưng chưa có lịch sử migration (schema_migrations); " +
	"nếu DB này do bản dùng AutoMigrate tạo, hãy chạy: migrate -baseline 1")

var fileNamePattern = regexp.MustCompile(`^(\d{4,})_([A-Za-z0-9_-]+)\.sql$`)

// Migration là một file .sql đã đọc vào bộ nhớ.
type Migration struct {
	Version  int
	Name     string
	SQL      string
	Checksum string
}

// Load đọc các file NNNN_ten.sql ở thư mục gốc của fsys. Số phiên bản phải liên tục
// từ 1, không trùng và không bỏ số.
func Load(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("đọc thư mục migration: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		match := fileNamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			return nil, fmt.Errorf("tên file migration %q sai định dạng, cần NNNN_ten.sql", entry.Name())
		}
		version, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("số phiên bản của %q: %w", entry.Name(), err)
		}
		content, err := fs.ReadFile(fsys, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("đọc %q: %w", entry.Name(), err)
		}
		sum := sha256.Sum256(content)
		migrations = append(migrations, Migration{
			Version:  version,
			Name:     strings.TrimSuffix(entry.Name(), ".sql"),
			SQL:      string(content),
			Checksum: hex.EncodeToString(sum[:]),
		})
	}

	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	for i, m := range migrations {
		if m.Version != i+1 {
			return nil, fmt.Errorf("số phiên bản migration phải liên tục từ 1: đang thiếu hoặc trùng quanh phiên bản %d (%s)", m.Version, m.Name)
		}
	}
	return migrations, nil
}

// WithMultiStatements bật multiStatements trong DSN, vì mỗi file migration có nhiều câu lệnh.
func WithMultiStatements(dsn string) (string, error) {
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", fmt.Errorf("DB_DSN không hợp lệ: %w", err)
	}
	config.MultiStatements = true
	return config.FormatDSN(), nil
}

type appliedRow struct {
	Name     string
	Checksum string
}

// Run chạy các migration chưa áp dụng theo thứ tự. Chạy lại khi không có gì mới thì không làm gì.
// MySQL tự commit khi chạy DDL nên một file lỗi giữa chừng không tự hoàn tác được; khi đó
// phải xử lý tay rồi chạy lại (file lỗi chưa được ghi vào schema_migrations).
func Run(ctx context.Context, db *sql.DB, migrations []Migration, logf func(format string, args ...any)) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("mở kết nối: %w", err)
	}
	defer conn.Close()

	release, err := acquireLock(ctx, conn)
	if err != nil {
		return err
	}
	defer release()

	if err := ensureTable(ctx, conn); err != nil {
		return err
	}
	applied, err := loadApplied(ctx, conn)
	if err != nil {
		return err
	}
	if err := verifyApplied(applied, migrations); err != nil {
		return err
	}
	if len(applied) == 0 {
		legacy, err := tableExists(ctx, conn, "users")
		if err != nil {
			return err
		}
		if legacy {
			return ErrLegacySchema
		}
	}

	ran := 0
	for _, m := range migrations {
		if _, done := applied[m.Version]; done {
			continue
		}
		logf("đang chạy %s", m.Name)
		if _, err := conn.ExecContext(ctx, m.SQL); err != nil {
			return fmt.Errorf("migration %s lỗi (chưa được ghi nhận, cần kiểm tra DB trước khi chạy lại): %w", m.Name, err)
		}
		if err := record(ctx, conn, m); err != nil {
			return err
		}
		ran++
	}
	if ran == 0 {
		logf("không có migration mới")
	} else {
		logf("đã chạy %d migration", ran)
	}
	return nil
}

// Baseline đánh dấu các migration từ 1 đến upTo là đã áp dụng mà không chạy chúng.
// Chỉ dùng cho DB cũ đã có đúng schema đó và chưa có lịch sử migration.
func Baseline(ctx context.Context, db *sql.DB, migrations []Migration, upTo int, logf func(format string, args ...any)) error {
	if upTo < 1 || upTo > len(migrations) {
		return fmt.Errorf("baseline %d không hợp lệ: chỉ có %d migration", upTo, len(migrations))
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("mở kết nối: %w", err)
	}
	defer conn.Close()

	release, err := acquireLock(ctx, conn)
	if err != nil {
		return err
	}
	defer release()

	if err := ensureTable(ctx, conn); err != nil {
		return err
	}
	applied, err := loadApplied(ctx, conn)
	if err != nil {
		return err
	}
	if len(applied) > 0 {
		return errors.New("schema_migrations đã có dữ liệu, không thể baseline")
	}
	for _, m := range migrations[:upTo] {
		if err := record(ctx, conn, m); err != nil {
			return err
		}
		logf("đã đánh dấu %s là đã áp dụng (không chạy)", m.Name)
	}
	return nil
}

func acquireLock(ctx context.Context, conn *sql.Conn) (func(), error) {
	var got sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 30)", lockName).Scan(&got); err != nil {
		return nil, fmt.Errorf("lấy khóa migration: %w", err)
	}
	if !got.Valid || got.Int64 != 1 {
		return nil, errors.New("đang có tiến trình migration khác chạy, thử lại sau")
	}
	return func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", lockName)
	}, nil
}

func ensureTable(ctx context.Context, conn *sql.Conn) error {
	_, err := conn.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations ("+
		"version INT NOT NULL PRIMARY KEY, "+
		"name VARCHAR(255) NOT NULL, "+
		"checksum CHAR(64) NOT NULL, "+
		"applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"+
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	if err != nil {
		return fmt.Errorf("tạo bảng schema_migrations: %w", err)
	}
	return nil
}

func loadApplied(ctx context.Context, conn *sql.Conn) (map[int]appliedRow, error) {
	rows, err := conn.QueryContext(ctx, "SELECT version, name, checksum FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("đọc schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[int]appliedRow{}
	for rows.Next() {
		var version int
		var row appliedRow
		if err := rows.Scan(&version, &row.Name, &row.Checksum); err != nil {
			return nil, fmt.Errorf("đọc schema_migrations: %w", err)
		}
		applied[version] = row
	}
	return applied, rows.Err()
}

func verifyApplied(applied map[int]appliedRow, migrations []Migration) error {
	byVersion := make(map[int]Migration, len(migrations))
	for _, m := range migrations {
		byVersion[m.Version] = m
	}
	for version, row := range applied {
		m, ok := byVersion[version]
		if !ok {
			return fmt.Errorf("DB đã chạy migration %d (%s) nhưng bản này không có file đó; hãy dùng bản code mới hơn", version, row.Name)
		}
		if m.Checksum != row.Checksum {
			return fmt.Errorf("file %s đã bị sửa sau khi chạy (checksum khác); không được sửa file đã chạy, hãy thêm file migration mới", m.Name)
		}
	}
	return nil
}

func tableExists(ctx context.Context, conn *sql.Conn, name string) (bool, error) {
	var count int
	err := conn.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", name,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("kiểm tra bảng %s: %w", name, err)
	}
	return count > 0, nil
}

func record(ctx context.Context, conn *sql.Conn, m Migration) error {
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, name, checksum) VALUES (?, ?, ?)",
		m.Version, m.Name, m.Checksum,
	); err != nil {
		return fmt.Errorf("ghi nhận migration %s: %w", m.Name, err)
	}
	return nil
}
