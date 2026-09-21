package handler

import (
	"errors"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": constants.SuccessMessage, "data": data})
}
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": constants.SuccessMessage, "data": data})
}
func Fail(c *gin.Context, err error) {
	var business *apperrors.BusinessError
	if errors.As(err, &business) {
		c.JSON(business.Status, gin.H{"code": business.Code, "message": business.Message, "data": nil})
		return
	}
	if errors.Is(err, apperrors.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"code": 40401, "message": "资源不存在", "data": nil})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": 50001, "message": "服务器内部错误", "data": nil})
}
