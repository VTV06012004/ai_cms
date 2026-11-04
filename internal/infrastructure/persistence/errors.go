package persistence

import (
	"errors"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// notFoundOr đổi "record not found" thành lỗi nghiệp vụ,
// các lỗi khác (ví dụ mất kết nối DB) giữ nguyên.
func notFoundOr(err, notFound error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFound
	}
	return err
}

// isDuplicateKey kiểm tra lỗi trùng khóa của MySQL (mã 1062).
func isDuplicateKey(err error, indexHint string) bool {
	var mysqlError *mysqlDriver.MySQLError
	return errors.As(err, &mysqlError) &&
		mysqlError.Number == 1062 &&
		strings.Contains(mysqlError.Message, indexHint)
}

// isForeignKeyViolation báo lỗi MySQL 1451: xóa bản ghi còn được bảng khác tham chiếu.
func isForeignKeyViolation(err error) bool {
	var mysqlError *mysqlDriver.MySQLError
	return errors.As(err, &mysqlError) && mysqlError.Number == 1451
}
