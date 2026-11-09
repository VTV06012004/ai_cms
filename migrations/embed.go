// Package migrations chứa các file migration SQL, nhúng sẵn vào binary migrate.
//
// Quy tắc: đã chạy ở bất kỳ môi trường nào thì KHÔNG sửa file; muốn đổi schema
// thì thêm file mới với số thứ tự tiếp theo (0002_..., 0003_...).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
