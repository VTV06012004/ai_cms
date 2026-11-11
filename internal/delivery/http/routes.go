package http

import (
	"api/internal/delivery/http/middleware"
	"api/internal/infrastructure/security"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	credentials *security.Credentials,
	authH *AuthHandler,
	postH *PostHandler,
	catH *CategoryHandler,
	tagH *TagHandler,
	menuH *MenuHandler,
	revH *RevisionHandler,
	auditH *AuditHandler,
	mediaH *MediaHandler,
	userH *UserHandler,
) {
	// Public Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Gin + MySQL 9.7 Backend is running!"})
	})
	r.GET("/docs", ServeDocs)
	r.GET("/docs/openapi.yaml", ServeOpenAPISpec)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authH.RateLimitIP(), authH.Register)
		auth.POST("/login", authH.RateLimitIP(), authH.Login)
		auth.POST("/refresh", authH.RateLimitSession(), authH.Refresh)
		auth.POST("/logout", authH.RateLimitSession(), authH.Logout)
	}

	r.GET("/api/posts", middleware.OptionalAuthMiddleware(credentials), postH.GetAll)
	r.GET("/api/posts/:id", middleware.OptionalAuthMiddleware(credentials), postH.GetByID)
	r.GET("/api/categories", catH.GetAll)
	r.GET("/api/tags", tagH.GetAll)
	r.GET("/api/menus", menuH.GetAll)
	r.GET("/uploads/*filepath", mediaH.ServeUploadedFile)

	// Protected Routes (JWT)
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(credentials))
	{
		protected.POST("/posts", postH.Create)
		protected.PUT("/posts/:id", postH.Update)
		protected.DELETE("/posts/:id", postH.Delete)
		protected.GET("/posts-revisions/:postId", revH.GetByPostID)
		protected.POST("/revisions/:id/restore", revH.Restore)
		protected.POST("/upload", mediaH.UploadImage)
		protected.GET("/media", mediaH.List)
		protected.GET("/revisions/compare", revH.Compare)
	}

	// Admin / Moderator Routes
	moderationGroup := r.Group("/api")
	moderationGroup.Use(middleware.AuthMiddleware(credentials), middleware.RequireRoles("admin", "editor"))
	{
		moderationGroup.PUT("/posts/:id/moderate", postH.Moderate)
		moderationGroup.POST("/categories", catH.Create)
		moderationGroup.PUT("/categories/:id", catH.Update)
		moderationGroup.DELETE("/categories/:id", catH.Delete)
		moderationGroup.POST("/tags", tagH.Create)
		moderationGroup.PUT("/tags/:id", tagH.Update)
		moderationGroup.DELETE("/tags/:id", tagH.Delete)
		moderationGroup.DELETE("/media/:name", mediaH.Delete)
	}

	adminProtected := r.Group("/api")
	adminProtected.Use(middleware.AuthMiddleware(credentials), middleware.RequireRoles("admin"))
	{
		adminProtected.GET("/audit-logs", auditH.GetAuditLogs)
		adminProtected.POST("/menus", menuH.Create)
		adminProtected.DELETE("/menus/:id", menuH.Delete)
		adminProtected.PUT("/menus/:id", menuH.Update)
	}

	adminUsers := r.Group("/api/admin")
	adminUsers.Use(middleware.AuthMiddleware(credentials), middleware.AdminOnlyMiddleware())
	{
		adminUsers.GET("/users", userH.GetAllUsers)
		adminUsers.PUT("/users/:id/role", userH.UpdateRole)
	}
}
