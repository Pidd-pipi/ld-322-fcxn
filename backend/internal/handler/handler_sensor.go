package handler

import (
	"github.com/cygreenenv/greenhouse-panel/internal/dto"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SensorHandler struct {
	service   *service.MonitoringService
	validator *validator.Validate
}

func NewSensorHandler(s *service.MonitoringService, v *validator.Validate) *SensorHandler {
	return &SensorHandler{s, v}
}
func (h *SensorHandler) Create(c *gin.Context) {
	var req dto.SensorRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil || req.MinValue >= req.MaxValue {
		Fail(c, apperrors.ErrValidation)
		return
	}
	sensor, err := h.service.CreateSensor(req.GreenhouseID, req.Name, req.Type, req.MinValue, req.MaxValue)
	if err != nil {
		Fail(c, err)
		return
	}
	Created(c, sensor)
}
