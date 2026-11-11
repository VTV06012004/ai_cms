package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseIDParam đọc ID dương trong đường dẫn; sai thì tự trả 400 và báo false.
func parseIDParam(c *gin.Context, name string) (uint, bool) {
	id, err := parseIDValue(c.Param(name))
	if err != nil {
		writeError(c, err)
		return 0, false
	}
	return id, true
}

func parseIDValue(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil || id == 0 {
		return 0, errInvalidID
	}
	return uint(id), nil
}

func currentUserID(c *gin.Context) uint {
	value, _ := c.Get("user_id")
	id, _ := value.(uint)
	return id
}
