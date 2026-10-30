package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

func MigrateActivePostSlugIndex(db *gorm.DB) error {
	var duplicateSlug string
	if err := db.Raw(
		"SELECT slug FROM posts WHERE deleted_at IS NULL GROUP BY slug HAVING COUNT(*) > 1 LIMIT 1",
	).Scan(&duplicateSlug).Error; err != nil {
		return fmt.Errorf("kiểm tra slug bài viết hiện có: %w", err)
	}
	if duplicateSlug != "" {
		return fmt.Errorf("không thể tạo unique index slug: đang có nhiều bài chưa xóa cùng slug %q; hãy xử lý dữ liệu trùng trước khi khởi động", duplicateSlug)
	}

	if !db.Migrator().HasColumn(&PostModel{}, "active_slug") {
		if err := db.Exec(
			"ALTER TABLE posts ADD COLUMN active_slug VARCHAR(255) GENERATED ALWAYS AS (CASE WHEN deleted_at IS NULL THEN slug ELSE NULL END) STORED",
		).Error; err != nil {
			return fmt.Errorf("tạo cột slug cho bài viết chưa xóa: %w", err)
		}
	}
	if !db.Migrator().HasIndex(&PostModel{}, "idx_posts_active_slug") {
		if err := db.Exec("CREATE UNIQUE INDEX idx_posts_active_slug ON posts (active_slug)").Error; err != nil {
			return fmt.Errorf("tạo unique index slug bài viết: %w", err)
		}
	}
	if db.Migrator().HasIndex(&PostModel{}, "idx_posts_slug_deleted") {
		if err := db.Migrator().DropIndex(&PostModel{}, "idx_posts_slug_deleted"); err != nil {
			return fmt.Errorf("xóa unique index slug cũ: %w", err)
		}
	}
	return nil
}
