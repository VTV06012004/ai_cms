package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	delivery "api/internal/delivery/http"
	"api/internal/infrastructure/database"
	"api/internal/infrastructure/persistence"
	"api/internal/infrastructure/security"
	"api/internal/infrastructure/storage"
	"api/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	db, err := database.NewMySQLConnection()
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := migrate(db); err != nil {
		return err
	}
	if err := seedDefaults(db); err != nil {
		return err
	}
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return err
	}

	userRepo := persistence.NewUserRepository(db)
	postRepo := persistence.NewPostRepository(db)
	categoryRepo := persistence.NewCategoryRepository(db)
	tagRepo := persistence.NewTagRepository(db)
	menuRepo := persistence.NewMenuRepository(db)
	revisionRepo := persistence.NewRevisionRepository(db)
	auditRepo := persistence.NewAuditRepository(db)
	unitOfWork := persistence.NewUnitOfWork(db)
	credentials := security.NewCredentials()

	userUC := usecase.NewUserUseCase(userRepo, credentials, unitOfWork)
	postUC := usecase.NewPostUseCase(postRepo, unitOfWork)
	categoryUC := usecase.NewCategoryUseCase(categoryRepo)
	tagUC := usecase.NewTagUseCase(tagRepo)
	menuUC := usecase.NewMenuUseCase(menuRepo, categoryRepo, unitOfWork)
	revisionUC := usecase.NewRevisionUseCase(revisionRepo, postRepo, unitOfWork)
	auditUC := usecase.NewAuditUseCase(auditRepo)
	mediaUC := usecase.NewMediaUseCase(storage.NewLocalFileStorage("uploads"))

	router := gin.Default()
	router.Use(corsMiddleware())
	delivery.RegisterRoutes(
		router,
		delivery.NewAuthHandler(userUC),
		delivery.NewPostHandler(postUC),
		delivery.NewCategoryHandler(categoryUC),
		delivery.NewTagHandler(tagUC),
		delivery.NewMenuHandler(menuUC),
		delivery.NewRevisionHandler(revisionUC),
		delivery.NewAuditHandler(auditUC),
		delivery.NewMediaHandler(mediaUC),
		delivery.NewUserHandler(userUC),
	)

	log.Println("Backend đang chạy tại cổng 8080...")
	return router.Run("0.0.0.0:8080")
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&persistence.UserModel{},
		&persistence.CategoryModel{},
		&persistence.PostModel{},
		&persistence.TagModel{},
		&persistence.RevisionModel{},
		&persistence.AuditLogModel{},
		&persistence.MenuModel{},
		&persistence.MenuItemModel{},
	)
}

func seedDefaults(db *gorm.DB) error {
	var categoryCount int64
	if err := db.Model(&persistence.CategoryModel{}).Count(&categoryCount).Error; err != nil {
		return err
	}
	if categoryCount == 0 {
		categories := []persistence.CategoryModel{
			{Name: "News", Slug: "news"},
			{Name: "Technology", Slug: "technology"},
			{Name: "Lifestyle", Slug: "lifestyle"},
		}
		if err := db.Create(&categories).Error; err != nil {
			return err
		}
	}

	var tagCount int64
	if err := db.Model(&persistence.TagModel{}).Count(&tagCount).Error; err != nil {
		return err
	}
	if tagCount == 0 {
		tags := []persistence.TagModel{
			{Name: "Golang", Slug: "golang"},
			{Name: "Next.js", Slug: "nextjs"},
			{Name: "Web Development", Slug: "web-development"},
		}
		if err := db.Create(&tags).Error; err != nil {
			return err
		}
	}
	return nil
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{
		"http://localhost:3000": {},
		"http://localhost:3001": {},
	}
	if configuredOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); configuredOrigins != "" {
		allowedOrigins = make(map[string]struct{})
		for _, origin := range strings.Split(configuredOrigins, ",") {
			if origin = strings.TrimSpace(origin); origin != "" {
				allowedOrigins[origin] = struct{}{}
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, allowed := allowedOrigins[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Expose-Headers", "Content-Length")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "43200")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
