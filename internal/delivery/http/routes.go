package http

import (
	"api/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterRoutes(
	r *gin.Engine,
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
	r.POST("/api/auth/register", authH.Register)
	r.POST("/api/auth/login", authH.Login)

	r.GET("/api/posts", postH.GetAll)
	r.GET("/api/posts/:id", postH.GetByID)
	r.GET("/api/categories", catH.GetAll)
	r.GET("/api/tags", tagH.GetAll)
	r.GET("/api/menus", menuH.GetAll)
	r.GET("/uploads/*filepath", mediaH.ServeUploadedFile)

	// Protected Routes (JWT)
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/posts", postH.Create)
		protected.PUT("/posts/:id", postH.Update)
		protected.DELETE("/posts/:id", postH.Delete)
		protected.GET("/posts-revisions/:postId", revH.GetByPostID)
		protected.POST("/revisions/:id/restore", revH.Restore)
		protected.POST("/upload", mediaH.UploadImage)
	}

	// Admin / Moderator Routes
	moderationGroup := r.Group("/api")
	moderationGroup.Use(middleware.AuthMiddleware(), middleware.RequireRoles("admin", "editor"))
	{
		moderationGroup.PUT("/posts/:id/moderate", postH.Moderate)
		moderationGroup.POST("/categories", catH.Create)
	}

	adminProtected := r.Group("/api")
	adminProtected.Use(middleware.AuthMiddleware(), middleware.RequireRoles("admin"))
	{
		adminProtected.GET("/audit-logs", auditH.GetAuditLogs)
		adminProtected.POST("/menus", menuH.Create)
		adminProtected.DELETE("/menus/:id", menuH.Delete)
	}

	adminUsers := r.Group("/api/admin")
	adminUsers.Use(middleware.AuthMiddleware(), middleware.AdminOnlyMiddleware())
	{
		adminUsers.GET("/users", userH.GetAllUsers)
		adminUsers.PUT("/users/:id/role", userH.UpdateRole)
	}
}
