package main

import (
	"context"
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/config"
	"github.com/cygreenenv/greenhouse-panel/internal/logger"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	"github.com/cygreenenv/greenhouse-panel/internal/router"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	l := logger.New()
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN()), &gorm.Config{})
	if err != nil {
		log.Fatal(fmt.Errorf("open mysql: %w", err))
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		log.Fatal(fmt.Errorf("migrate database: %w", err))
	}
	if err = service.Seed(db); err != nil {
		log.Fatal(fmt.Errorf("seed database: %w", err))
	}
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress()})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if err = redisClient.Ping(ctx).Err(); err != nil {
		l.Warn("redis unavailable", "error", err)
	}
	greenhouses := repository.NewGreenhouseRepository(db)
	sensors := repository.NewSensorRepository(db)
	alerts := repository.NewAlertRepository(db)
	devices := repository.NewDeviceRepository(db)
	hub := ws.NewHub()
	auth := service.NewAuthService(cfg.JWTSecret)
	monitoring := service.NewMonitoringService(greenhouses, sensors, alerts, l, hub)
	control := service.NewControlService(devices, l, hub)
	alertService := service.NewAlertService(alerts, l)
	reports := service.NewReportService(sensors, alerts)
	engine := router.New(router.Dependencies{Config: cfg, Logger: l, Auth: auth, Monitoring: monitoring, Alerts: alertService, Control: control, Reports: reports, Hub: hub})
	l.Info("server started", "port", cfg.ServerPort)
	if err := engine.Run(fmt.Sprintf(":%d", cfg.ServerPort)); err != nil {
		log.Fatal(err)
	}
}
