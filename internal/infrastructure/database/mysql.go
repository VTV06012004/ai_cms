package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMySQLConnection khởi tạo kết nối GORM tới cơ sở dữ liệu MySQL
func NewMySQLConnection() (*gorm.DB, error) {
	// Lấy DSN từ biến môi trường; thiếu cấu hình thì không khởi động với mật khẩu mặc định.
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DB_DSN is required")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới database: %v", err)
	}

	log.Println(" kết nối cơ sở dữ liệu MySQL thành công!")
	return db, nil
}
