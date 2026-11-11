package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var id [16]byte
		if _, err := rand.Read(id[:]); err != nil {
			slog.Error("generate request id", "error", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		requestID := hex.EncodeToString(id[:])
		c.Set("request_id", requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		slog.Info("http request",
			"request_id", requestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(started),
			"client_ip", c.ClientIP(),
		)
	}
}

func StructuredRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("http panic recovered",
			"request_id", requestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", recovered,
			"stack", string(debug.Stack()),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func requestID(c *gin.Context) string {
	value, _ := c.Get("request_id")
	id, _ := value.(string)
	return id
}
