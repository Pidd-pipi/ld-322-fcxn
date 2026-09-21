package middleware

import (
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/handler"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"strings"
)

func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			handler.Fail(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}
		if _, err := auth.Parse(parts[1]); err != nil {
			handler.Fail(c, err)
			c.Abort()
			return
		}
		c.Next()
	}
}
