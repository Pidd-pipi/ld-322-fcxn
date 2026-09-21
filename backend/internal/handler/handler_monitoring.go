package handler

import (
	"github.com/cygreenenv/greenhouse-panel/internal/dto"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MonitoringHandler struct {
	service   *service.MonitoringService
	validator *validator.Validate
}

func NewMonitoringHandler(s *service.MonitoringService, v *validator.Validate) *MonitoringHandler {
	return &MonitoringHandler{s, v}
}
func (h *MonitoringHandler) Ingest(c *gin.Context) {
	var req dto.ReadingRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil {
		Fail(c, apperrors.ErrValidation)
		return
	}
	reading, alert, err := h.service.Ingest(req.SensorID, req.Value)
	if err != nil {
		Fail(c, err)
		return
	}
	Created(c, gin.H{"reading": reading, "alert": alert})
}
func (h *MonitoringHandler) Latest(c *gin.Context) {
	id, ok := queryID(c, "greenhouse_id")
	if !ok {
		return
	}
	rows, err := h.service.Latest(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
func (h *MonitoringHandler) History(c *gin.Context) {
	id, ok := queryID(c, "greenhouse_id")
	if !ok {
		return
	}
	end := time.Now()
	start := end.Add(-24 * time.Hour)
	if value := c.Query("start"); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			start = parsed
		}
	}
	if value := c.Query("end"); value != "" {
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			end = parsed
		}
	}
	types := []string{}
	if raw := c.Query("types"); raw != "" {
		types = strings.Split(raw, ",")
	}
	rows, err := h.service.History(id, types, start, end)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, rows)
}
func (h *MonitoringHandler) Threshold(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil || h.validator.Struct(req) != nil || req.MinValue >= req.MaxValue {
		Fail(c, apperrors.ErrValidation)
		return
	}
	row, err := h.service.UpdateThreshold(id, req.MinValue, req.MaxValue)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, row)
}
func (h *MonitoringHandler) Simulate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	count, err := h.service.Simulate(id)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, gin.H{"created": count})
}
func parseID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40001, "message": "请求参数不合法", "data": nil})
		return 0, false
	}
	return uint(value), true
}
func queryID(c *gin.Context, key string) (uint, bool) {
	value, err := strconv.ParseUint(c.Query(key), 10, 64)
	if err != nil || value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40001, "message": "请求参数不合法", "data": nil})
		return 0, false
	}
	return uint(value), true
}
