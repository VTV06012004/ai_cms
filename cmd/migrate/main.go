// Lệnh migrate chạy các file SQL trong thư mục migrations/ vào database.
// Chạy riêng, TRƯỚC khi chạy/deploy server; server không tự migrate.
//
//	migrate              chạy các migration chưa áp dụng
//	migrate -baseline N  đánh dấu migration 1..N là đã áp dụng (cho DB cũ, không chạy SQL)
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"api/internal/infrastructure/migrate"
	sqlfiles "api/migrations"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	baseline := flag.Int("baseline", 0, "đánh dấu migration 1..N là đã áp dụng mà không chạy (chỉ cho DB cũ đã có đúng schema)")
	flag.Parse()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return errors.New("DB_DSN is required")
	}
	dsn, err := migrate.WithMultiStatements(dsn)
	if err != nil {
		return err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("mở kết nối database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := waitForDB(ctx, db, 60*time.Second); err != nil {
		return err
	}

	migrations, err := migrate.Load(sqlfiles.FS)
	if err != nil {
		return err
	}
	if *baseline > 0 {
		return migrate.Baseline(ctx, db, migrations, *baseline, log.Printf)
	}
	return migrate.Run(ctx, db, migrations, log.Printf)
}

// waitForDB chờ MySQL nhận kết nối, vì lần đầu khởi tạo volume MySQL mở cổng muộn.
func waitForDB(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		err := db.PingContext(ctx)
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("không kết nối được database sau %s: %w", timeout, err)
		}
		time.Sleep(2 * time.Second)
	}
}
