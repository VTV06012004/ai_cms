package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMySQLConnection khởi tạo kết nối GORM tới MySQL bằng DSN đã được kiểm tra ở config.
func NewMySQLConnection(dsn string) (*gorm.DB, error) {
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
