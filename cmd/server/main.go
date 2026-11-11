package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VTV06012004/ai_cms/internal/config"
	delivery "github.com/VTV06012004/ai_cms/internal/delivery/http"
	"github.com/VTV06012004/ai_cms/internal/delivery/http/middleware"
	"github.com/VTV06012004/ai_cms/internal/infrastructure/database"
	"github.com/VTV06012004/ai_cms/internal/infrastructure/persistence"
	"github.com/VTV06012004/ai_cms/internal/infrastructure/security"
	"github.com/VTV06012004/ai_cms/internal/infrastructure/storage"
	"github.com/VTV06012004/ai_cms/internal/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	credentials, err := security.NewCredentials(cfg.JWTSecret)
	if err != nil {
		return err
	}

	db, err := database.NewMySQLConnection(cfg.DatabaseDSN)
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

	if err := protectExistingAdmin(db); err != nil {
		return err
	}
	if err := seedDefaults(db); err != nil {
		return err
	}
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return err
	}

	userRepo := persistence.NewUserRepository(db)
	refreshTokenRepo := persistence.NewRefreshTokenRepository(db)
	postRepo := persistence.NewPostRepository(db)
	categoryRepo := persistence.NewCategoryRepository(db)
	tagRepo := persistence.NewTagRepository(db)
	menuRepo := persistence.NewMenuRepository(db)
	revisionRepo := persistence.NewRevisionRepository(db)
	auditRepo := persistence.NewAuditRepository(db)
	unitOfWork := persistence.NewUnitOfWork(db)
	userUC := usecase.NewUserUseCase(userRepo, credentials, unitOfWork)
	sessionUC := usecase.NewSessionUseCase(userRepo, refreshTokenRepo, credentials)
	if err := userUC.EnsureInitialAdmin(context.Background(), usecase.InitialAdminConfig{
		Name:     cfg.InitialAdmin.Name,
		Email:    cfg.InitialAdmin.Email,
		Password: cfg.InitialAdmin.Password,
	}); err != nil {
		return err
	}
	postUC := usecase.NewPostUseCase(postRepo, unitOfWork)
	categoryUC := usecase.NewCategoryUseCase(categoryRepo, unitOfWork)
	tagUC := usecase.NewTagUseCase(tagRepo, unitOfWork)
	menuUC := usecase.NewMenuUseCase(menuRepo, categoryRepo, unitOfWork)
	revisionUC := usecase.NewRevisionUseCase(revisionRepo, postRepo, unitOfWork)
	auditUC := usecase.NewAuditUseCase(auditRepo)
	mediaUC := usecase.NewMediaUseCase(storage.NewLocalFileStorage("uploads"))

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.RequestID(), middleware.StructuredLogger(), middleware.StructuredRecovery())
	router.Use(corsMiddleware(cfg.CORSAllowedOrigins))
	delivery.RegisterRoutes(
		router,
		credentials,
		delivery.NewAuthHandler(userUC, sessionUC, cfg.CORSAllowedOrigins, cfg.AuthCookieSecure),
		delivery.NewPostHandler(postUC),
		delivery.NewCategoryHandler(categoryUC),
		delivery.NewTagHandler(tagUC),
		delivery.NewMenuHandler(menuUC),
		delivery.NewRevisionHandler(revisionUC),
		delivery.NewAuditHandler(auditUC),
		delivery.NewMediaHandler(mediaUC, cfg.PublicBaseURL),
		delivery.NewUserHandler(userUC),
	)

	server := &http.Server{
		Addr:              "0.0.0.0:" + cfg.Port,
		Handler:           router,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Nhận Ctrl+C hoặc tín hiệu dừng của Docker (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	schedulerDone := make(chan struct{})
	go func() {
		defer close(schedulerDone)
		runScheduler(ctx, postUC)
	}()
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", "port", cfg.Port)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}
	stop()
	slog.Info("shutdown signal received")

	// Cho các request đang xử lý tối đa 15 giây để hoàn tất.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("tắt server không thành công: %w", err)
	}
	<-schedulerDone
	slog.Info("server shut down cleanly")
	return nil
}

// runScheduler quét bài hẹn giờ mỗi phút cho đến khi ctx bị hủy.
func runScheduler(ctx context.Context, posts *usecase.PostUseCase) {
	tick := func() {
		count, err := posts.PublishDuePosts(ctx, time.Now())
		if err != nil {
			slog.Error("scheduled post scan failed", "error", err)
			return
		}
		if count > 0 {
			slog.Info("scheduled posts published", "count", count)
		}
	}
	tick()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick()
		}
	}
}

func protectExistingAdmin(db *gorm.DB) error {
	var rootAdminCount int64
	if err := db.Model(&persistence.UserModel{}).Where("root_admin_key = ?", 1).Count(&rootAdminCount).Error; err != nil {
		return err
	}
	if rootAdminCount > 0 {
		return nil
	}

	var admin persistence.UserModel
	err := db.Where("role = ?", "admin").Order("created_at ASC, id ASC").First(&admin).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	key := uint(1)
	return db.Model(&admin).Update("root_admin_key", key).Error
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

func corsMiddleware(origins []string) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowedOrigins[origin] = struct{}{}
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
