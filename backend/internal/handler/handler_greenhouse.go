package handler

import (
	"github.com/cygreenenv/greenhouse-panel/internal/dto"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type GreenhouseHandler struct {
	service   *service.MonitoringService
	validator *validator.Validate
}

func NewGreenhouseHandler(s *service.MonitoringService, v *validator.Validate) *GreenhouseHandler {
	return &GreenhouseHandler{s, v}
}
func (h *GreenhouseHandler) List(c *gin.Context) {
	rows, err := h.service.ListGreenhouses()
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
func (h *GreenhouseHandler) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	row, err := h.service.Detail(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, row)
}
func (h *GreenhouseHandler) Create(c *gin.Context) {
	var req dto.GreenhouseRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil {
		Fail(c, apperrors.ErrValidation)
		return
	}
	row := &model.Greenhouse{Name: req.Name, Location: req.Location, Area: req.Area}
	if err := h.service.CreateGreenhouse(row); err != nil {
		Fail(c, err)
		return
	}
	Created(c, row)
}
