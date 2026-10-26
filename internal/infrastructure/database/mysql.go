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
	// Lấy DSN từ biến môi trường hoặc dùng giá trị mặc định từ docker-compose
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "cms_user:cms_password@tcp(db:3306)/ai_cms_db?parseTime=true&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới database: %v", err)
	}

	log.Println(" kết nối cơ sở dữ liệu MySQL thành công!")
	return db, nil
}
