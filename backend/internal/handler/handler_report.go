package handler

import (
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	"github.com/gin-gonic/gin"
	"time"
)

type ReportHandler struct{ service *service.ReportService }

func NewReportHandler(s *service.ReportService) *ReportHandler { return &ReportHandler{s} }
func (h *ReportHandler) Get(c *gin.Context) {
	id, ok := queryID(c, "greenhouse_id")
	if !ok {
		return
	}
	rangeName := c.DefaultQuery("range", "day")
	end := time.Now()
	start := end.AddDate(0, 0, -1)
	if rangeName == "week" {
		start = end.AddDate(0, 0, -7)
	}
	if rangeName == "month" {
		start = end.AddDate(0, -1, 0)
	}
	report, err := h.service.Generate(id, rangeName, start, end)
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, report)
}
