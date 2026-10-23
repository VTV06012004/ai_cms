package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"api/handlers"
	"api/middleware"
	"api/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "cms_user:cms_password@tcp(db:3306)/ai_cms_db?parseTime=true&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Không thể kết nối Database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Lỗi lấy instance DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Không thể ping tới MySQL trong Docker: %v", err)
	}
	fmt.Println("👉 Kết nối CSDL MySQL 9.7 LTS thành công!")

	// Tự động tạo/cập nhật bảng trong DB
	db.AutoMigrate(&models.User{}, &models.Category{}, &models.Post{}, &models.Tag{}, &models.Revision{}, &models.AuditLog{}, &models.Menu{}, &models.MenuItem{})

	// Tự động tạo thư mục chứa ảnh upload nếu chưa tồn tại
	os.MkdirAll("uploads", 0755)

	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count == 0 {
		defaultCategories := []models.Category{
			{Name: "News", Slug: "news"},
			{Name: "Technology", Slug: "technology"},
			{Name: "Lifestyle", Slug: "lifestyle"},
		}
		db.Create(&defaultCategories)
		println("Đã tự động tạo dữ liệu danh mục mẫu thành công!")
	}
	var tagCount int64
	db.Model(&models.Tag{}).Count(&tagCount)
	if tagCount == 0 {
		defaultTags := []models.Tag{
			{Name: "Golang", Slug: "golang"},
			{Name: "Next.js", Slug: "nextjs"},
			{Name: "Web Development", Slug: "web-development"},
		}
		db.Create(&defaultTags)
		println("Đã tự động tạo dữ liệu thẻ mẫu thành công!")
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Handlers
	authH := handlers.NewAuthHandler(db)
	catH := handlers.NewCategoryHandler(db)
	postH := handlers.NewPostHandler(db)
	revH := handlers.NewRevisionHandler(db)
	mediaH := handlers.NewMediaHandler()
	tagH := handlers.NewTagHandler(db)
	auditH := handlers.NewAuditHandler(db) // Khởi tạo audit handler
	menuH := handlers.NewMenuHandler(db)

	// Public Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Gin + MySQL 9.7 Backend is running!"})
	})
	r.POST("/api/auth/register", authH.Register)
	r.POST("/api/auth/login", authH.Login)

	// Public APIs cho Frontend đọc dữ liệu
	r.GET("/api/posts", postH.GetAll)
	r.GET("/api/posts/:id", postH.GetByID)
	r.GET("/api/categories", catH.GetAll)
	r.GET("/api/tags", tagH.GetAll)
	r.GET("/api/menus", menuH.GetAll)
	// Phục vụ thư mục static cho ảnh đã upload
	r.Static("/uploads", "./uploads")

	// Protected Routes (Bắt buộc phải có Token JWT)
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// CRUD Categories
		protected.POST("/categories", middleware.RequireRoles("admin", "editor"), catH.Create)

		// CRUD Posts
		protected.POST("/posts", postH.Create)
		protected.PUT("/posts/:id", postH.Update)
		protected.DELETE("/posts/:id", postH.Delete)
		protected.GET("/posts-revisions/:postId", revH.GetByPostID)
		protected.POST("/revisions/:id/restore", revH.Restore)
		// Upload Media
		protected.POST("/upload", mediaH.UploadImage)

		// CRUD Menus (Dành cho Admin)
		protected.POST("/menus", middleware.RequireRoles("admin"), menuH.Create)
		protected.DELETE("/menus/:id", middleware.RequireRoles("admin"), menuH.Delete)
	}

	// Route dành riêng cho Admin (Xem Nhật ký hệ thống / Audit Logs)
	adminProtected := r.Group("/api")
	adminProtected.Use(middleware.AuthMiddleware(), middleware.RequireRoles("admin"))
	{
		adminProtected.GET("/audit-logs", auditH.GetAuditLogs)
	}

	// Nhóm router cho Admin quản lý users
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminOnlyMiddleware())
	{
		adminGroup.GET("/users", handlers.GetAllUsers(db))
		adminGroup.PUT("/users/:id/role", handlers.UpdateUserRole(db))
	}

	// Nhóm route dành cho kiểm duyệt (Admin hoặc User được phân quyền duyệt)
	moderationGroup := r.Group("/api")
	moderationGroup.Use(middleware.AuthMiddleware(), middleware.RequireRoles("admin", "editor")) // Hoặc kiểm tra thêm cờ can_approve
	{
		moderationGroup.PUT("/posts/:id/moderate", postH.ModeratePost)
	}

	log.Println("Backend đang chạy tại cổng 8080...")
	r.Run("0.0.0.0:8080")
}
