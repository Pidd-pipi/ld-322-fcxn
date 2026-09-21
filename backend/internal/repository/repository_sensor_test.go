package repository

import (
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestSensorRepositoryHistory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:repo_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	g := model.Greenhouse{Name: "测试温室"}
	if err = db.Create(&g).Error; err != nil {
		t.Fatal(err)
	}
	s := model.Sensor{GreenhouseID: g.ID, Name: "温度", Type: "temperature", Unit: "°C", Status: "online"}
	if err = db.Create(&s).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for _, value := range []float64{21, 22} {
		if err = db.Create(&model.SensorReading{SensorID: s.ID, Value: value, RecordedAt: now}).Error; err != nil {
			t.Fatal(err)
		}
	}
	repo := NewSensorRepository(db)
	rows, err := repo.History(g.ID, []string{"temperature"}, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 readings, got %d", len(rows))
	}
}

func TestSensorRepositoryLatestLoadsThreshold(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:latest_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	g := model.Greenhouse{Name: "最新数据温室"}
	if err = db.Create(&g).Error; err != nil {
		t.Fatal(err)
	}
	s := model.Sensor{GreenhouseID: g.ID, Name: "温度", Type: "temperature", Unit: "°C", Status: "online"}
	if err = db.Create(&s).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&model.Threshold{SensorID: s.ID, MinValue: 15, MaxValue: 32}).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&model.SensorReading{SensorID: s.ID, Value: 24, RecordedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	rows, err := NewSensorRepository(db).LatestForGreenhouse(g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Sensor.Threshold.MinValue != 15 || rows[0].Sensor.Threshold.MaxValue != 32 {
		t.Fatalf("latest reading did not include the expected threshold: %#v", rows)
	}
}
