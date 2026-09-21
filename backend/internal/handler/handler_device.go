package handler

import (
	"github.com/cygreenenv/greenhouse-panel/internal/dto"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type DeviceHandler struct {
	service   *service.ControlService
	validator *validator.Validate
}

func NewDeviceHandler(s *service.ControlService, v *validator.Validate) *DeviceHandler {
	return &DeviceHandler{s, v}
}
func (h *DeviceHandler) List(c *gin.Context) {
	id, ok := queryID(c, "greenhouse_id")
	if !ok {
		return
	}
	rows, err := h.service.List(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
func (h *DeviceHandler) Toggle(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.DeviceToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil {
		Fail(c, apperrors.ErrValidation)
		return
	}
	row, err := h.service.Toggle(id, req.Status)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, row)
}
func (h *DeviceHandler) Schedule(c *gin.Context) {
	var req dto.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil {
		Fail(c, apperrors.ErrValidation)
		return
	}
	row := &model.Schedule{DeviceID: req.DeviceID, Cron: req.Cron, Action: req.Action, Enabled: true}
	if err := h.service.Schedule(row); err != nil {
		Fail(c, err)
		return
	}
	Created(c, row)
}
func (h *DeviceHandler) Schedules(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	rows, err := h.service.Schedules(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
