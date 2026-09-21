package router

import (
	"github.com/cygreenenv/greenhouse-panel/internal/config"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"github.com/cygreenenv/greenhouse-panel/internal/handler"
	"github.com/cygreenenv/greenhouse-panel/internal/middleware"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log/slog"
)

type Dependencies struct {
	Config     config.Config
	Logger     *slog.Logger
	Auth       *service.AuthService
	Monitoring *service.MonitoringService
	Alerts     *service.AlertService
	Control    *service.ControlService
	Reports    *service.ReportService
	Hub        *ws.Hub
}

func New(d Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestLogger(d.Logger))
	v := validator.New()
	auth := handler.NewAuthHandler(d.Auth, v)
	greenhouse := handler.NewGreenhouseHandler(d.Monitoring, v)
	sensors := handler.NewSensorHandler(d.Monitoring, v)
	monitoring := handler.NewMonitoringHandler(d.Monitoring, v)
	alerts := handler.NewAlertHandler(d.Alerts)
	devices := handler.NewDeviceHandler(d.Control, v)
	reports := handler.NewReportHandler(d.Reports)
	r.GET(constants.HealthPath, func(c *gin.Context) { handler.Success(c, gin.H{"status": "healthy"}) })
	r.GET(constants.WebSocketPath, func(c *gin.Context) { d.Hub.Handle(c.Writer, c.Request) })
	api := r.Group(constants.APIPrefix)
	api.POST("/auth/login", auth.Login)
	api.GET("/greenhouses", greenhouse.List)
	api.GET("/greenhouses/:id", greenhouse.Detail)
	api.GET("/readings/latest", monitoring.Latest)
	api.GET("/readings/history", monitoring.History)
	api.GET("/alerts", alerts.List)
	api.GET("/devices", devices.List)
	api.GET("/reports/environment", reports.Get)
	secured := api.Group("")
	secured.Use(middleware.Auth(d.Auth))
	secured.POST("/greenhouses", greenhouse.Create)
	secured.POST("/sensors", sensors.Create)
	secured.POST("/readings", monitoring.Ingest)
	secured.POST("/greenhouses/:id/simulate", monitoring.Simulate)
	secured.PUT("/sensors/:id/threshold", monitoring.Threshold)
	secured.PATCH("/alerts/:id/handle", alerts.Handle)
	secured.PATCH("/devices/:id/toggle", devices.Toggle)
	secured.POST("/schedules", devices.Schedule)
	secured.GET("/devices/:id/schedules", devices.Schedules)
	return r
}
