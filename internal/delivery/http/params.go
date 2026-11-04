package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseIDParam đọc ID dương trong đường dẫn; sai thì tự trả 400 và báo false.
func parseIDParam(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID không hợp lệ"})
		return 0, false
	}
	return uint(id), true
}

func currentUserID(c *gin.Context) uint {
	value, _ := c.Get("user_id")
	id, _ := value.(uint)
	return id
}
