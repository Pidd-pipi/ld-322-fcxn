package handler

import (
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type AlertHandler struct{ service *service.AlertService }

func NewAlertHandler(s *service.AlertService) *AlertHandler { return &AlertHandler{s} }
func (h *AlertHandler) List(c *gin.Context) {
	var id uint
	if raw := c.Query("greenhouse_id"); raw != "" {
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || value == 0 {
			Fail(c, apperrors.ErrValidation)
			return
		}
		id = uint(value)
	}
	rows, err := h.service.List(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
func (h *AlertHandler) Handle(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	row, err := h.service.Handle(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, row)
}
