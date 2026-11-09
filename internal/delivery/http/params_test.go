package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseIDParamRejectsZeroAndMalformedValuesThroughWriteError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/:id", func(c *gin.Context) {
		id, ok := parseIDParam(c, "id")
		if ok {
			c.JSON(http.StatusOK, gin.H{"id": id})
		}
	})

	for _, path := range []string{"/0", "/not-a-number", "/4294967296"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s status = %d, want %d", path, response.Code, http.StatusBadRequest)
		}
	}
}
