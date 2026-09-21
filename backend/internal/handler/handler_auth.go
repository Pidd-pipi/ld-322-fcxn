package handler

import (
	"github.com/cygreenenv/greenhouse-panel/internal/dto"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type AuthHandler struct {
	service   *service.AuthService
	validator *validator.Validate
}

func NewAuthHandler(s *service.AuthService, v *validator.Validate) *AuthHandler {
	return &AuthHandler{s, v}
}
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil {
		Fail(c, apperrors.ErrValidation)
		return
	}
	token, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"token": token, "user": gin.H{"username": req.Username, "role": "admin"}}})
}
