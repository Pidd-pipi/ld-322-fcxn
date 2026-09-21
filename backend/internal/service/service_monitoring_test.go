package service

import (
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"testing"
)

func TestIngestCreatesAlertBeyondThreshold(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:service_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	g := model.Greenhouse{Name: "测试温室"}
	db.Create(&g)
	sensor := model.Sensor{GreenhouseID: g.ID, Name: "温度", Type: "temperature", Unit: "°C", Status: "online"}
	db.Create(&sensor)
	db.Create(&model.Threshold{SensorID: sensor.ID, MinValue: 10, MaxValue: 30})
	svc := NewMonitoringService(repository.NewGreenhouseRepository(db), repository.NewSensorRepository(db), repository.NewAlertRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)), ws.NewHub())
	_, alert, err := svc.Ingest(sensor.ID, 35)
	if err != nil {
		t.Fatal(err)
	}
	if alert == nil {
		t.Fatal("expected alert for out-of-range reading")
	}
	alerts, err := repository.NewAlertRepository(db).List(g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 {
		t.Fatalf("want 1 alert, got %d", len(alerts))
	}
}
