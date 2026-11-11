package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/VTV06012004/ai_cms/internal/usecase"
)

const refreshCookieName = "refresh_token"

type loginService interface {
	Register(ctx context.Context, input usecase.RegisterDTO, ipAddress string) (uint, error)
	Login(ctx context.Context, input usecase.LoginDTO, ipAddress string) (*usecase.LoginResponse, error)
}

type sessionService interface {
	Issue(ctx context.Context, userID uint) (string, error)
	Refresh(ctx context.Context, rawToken string) (*usecase.RefreshedSession, error)
	Revoke(ctx context.Context, rawToken string) error
}

type AuthHandler struct {
	userUseCase    loginService
	sessionUseCase sessionService
	rateLimiter    *authRateLimiter
	sessionLimiter *authRateLimiter
	allowedOrigins map[string]struct{}
	cookieSecure   bool
}

func NewAuthHandler(uc loginService, sessions sessionService, allowedOrigins []string, cookieSecure bool) *AuthHandler {
	origins := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origins[origin] = struct{}{}
	}
	return &AuthHandler{
		userUseCase: uc, sessionUseCase: sessions, rateLimiter: newAuthRateLimiter(),
		sessionLimiter: newRateLimiter(sessionRateLimit),
		allowedOrigins: origins, cookieSecure: cookieSecure,
	}
}

func (h *AuthHandler) RateLimitIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.rateLimiter.allowIP(c.ClientIP(), time.Now()) {
			writeError(c, errAuthRateLimitExceeded)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitSession giới hạn riêng cho làm mới phiên và đăng xuất, nới hơn đăng nhập.
func (h *AuthHandler) RateLimitSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.sessionLimiter.allowIP(c.ClientIP(), time.Now()) {
			writeError(c, errAuthRateLimitExceeded)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	if !h.rateLimiter.allowEmail(request.Email, time.Now()) {
		writeError(c, errAuthRateLimitExceeded)
		return
	}

	userID, err := h.userUseCase.Register(c.Request.Context(), usecase.RegisterDTO{
		Name: request.Name, Email: request.Email, Password: request.Password,
	}, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Đăng ký tài khoản thành công!", "user_id": userID})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		writeInvalidRequest(c)
		return
	}
	if !h.rateLimiter.allowEmail(request.Email, time.Now()) {
		writeError(c, errAuthRateLimitExceeded)
		return
	}

	res, err := h.userUseCase.Login(c.Request.Context(), usecase.LoginDTO{Email: request.Email, Password: request.Password}, c.ClientIP())
	if err != nil {
		writeError(c, err)
		return
	}
	refreshToken, err := h.sessionUseCase.Issue(c.Request.Context(), res.User.Data().ID)
	if err != nil {
		writeError(c, err)
		return
	}
	h.setRefreshCookie(c, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng nhập thành công",
		"token":   res.Token,
		"user": gin.H{
			"id":    res.User.Data().ID,
			"name":  res.User.Data().Name,
			"email": string(res.User.Data().Email),
			"role":  string(res.User.Data().Role),
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	if !h.hasAllowedOrigin(c) {
		writeError(c, errForbidden)
		return
	}
	cookie, err := c.Cookie(refreshCookieName)
	if err != nil {
		h.clearRefreshCookie(c)
		writeError(c, usecase.ErrInvalidRefreshToken)
		return
	}
	session, err := h.sessionUseCase.Refresh(c.Request.Context(), cookie)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidRefreshToken) {
			h.clearRefreshCookie(c)
		}
		writeError(c, err)
		return
	}
	h.setRefreshCookie(c, session.RefreshToken)
	data := session.User.Data()
	c.JSON(http.StatusOK, gin.H{
		"token": session.AccessToken,
		"user": gin.H{
			"id":    data.ID,
			"name":  data.Name,
			"email": string(data.Email),
			"role":  string(data.Role),
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if !h.hasAllowedOrigin(c) {
		writeError(c, errForbidden)
		return
	}
	cookie, err := c.Cookie(refreshCookieName)
	if err == nil {
		if err := h.sessionUseCase.Revoke(c.Request.Context(), cookie); err != nil {
			h.clearRefreshCookie(c)
			writeError(c, err)
			return
		}
	} else if !errors.Is(err, http.ErrNoCookie) {
		writeError(c, err)
		return
	}
	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) hasAllowedOrigin(c *gin.Context) bool {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	_, allowed := h.allowedOrigins[origin]
	return origin != "" && allowed
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(h.refreshCookieSameSite())
	c.SetCookie(refreshCookieName, token, int(usecase.RefreshTokenLifetime.Seconds()),
		"/api/auth", "", h.cookieSecure, true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(h.refreshCookieSameSite())
	c.SetCookie(refreshCookieName, "", -1, "/api/auth", "", h.cookieSecure, true)
}

func (h *AuthHandler) refreshCookieSameSite() http.SameSite {
	if h.cookieSecure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}
