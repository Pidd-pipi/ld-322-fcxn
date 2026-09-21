package service

import (
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
	"math"
	"time"
)

func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Greenhouse{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count seed greenhouses: %w", err)
	}
	if count > 0 {
		return nil
	}
	items := []struct {
		Name, Location string
		Area           float64
	}{{"一号番茄温室", "东区 A-01", 680}, {"二号叶菜温室", "东区 A-02", 520}}
	for index, item := range items {
		g := model.Greenhouse{Name: item.Name, Location: item.Location, Area: item.Area}
		if err := db.Create(&g).Error; err != nil {
			return fmt.Errorf("seed greenhouse: %w", err)
		}
		if err := seedGreenhouse(db, g, index); err != nil {
			return err
		}
	}
	return db.Create(&model.User{Username: "admin", PasswordHash: "demo", Role: constants.RoleAdmin}).Error
}
func seedGreenhouse(db *gorm.DB, g model.Greenhouse, offset int) error {
	types := []string{constants.SensorTemperature, constants.SensorHumidity, constants.SensorLight, constants.SensorCO2, constants.SensorSoil}
	for idx, typ := range types {
		rangeDef := constants.DefaultThresholds[typ]
		sensor := model.Sensor{GreenhouseID: g.ID, Name: constants.SensorLabels[typ] + "传感器", Type: typ, Unit: constants.SensorUnits[typ], Status: constants.StatusOnline}
		if err := db.Create(&sensor).Error; err != nil {
			return fmt.Errorf("seed sensor: %w", err)
		}
		if err := db.Create(&model.Threshold{SensorID: sensor.ID, MinValue: rangeDef.Min, MaxValue: rangeDef.Max}).Error; err != nil {
			return fmt.Errorf("seed threshold: %w", err)
		}
		for hour := 24; hour >= 0; hour-- {
			center := (rangeDef.Min + rangeDef.Max) / 2
			variation := (rangeDef.Max - rangeDef.Min) * 0.12 * math.Sin(float64(hour+idx+offset))
			reading := model.SensorReading{SensorID: sensor.ID, Value: center + variation, RecordedAt: time.Now().Add(-time.Duration(hour) * time.Hour)}
			if err := db.Create(&reading).Error; err != nil {
				return fmt.Errorf("seed reading: %w", err)
			}
		}
	}
	devices := []model.Device{{GreenhouseID: g.ID, Name: "循环风机", Type: "fan", Status: constants.StatusOn}, {GreenhouseID: g.ID, Name: "遮阳帘", Type: "shade", Status: constants.StatusOff}, {GreenhouseID: g.ID, Name: "灌溉泵", Type: "pump", Status: constants.StatusOff}, {GreenhouseID: g.ID, Name: "补光灯", Type: "light", Status: constants.StatusOn}}
	if err := db.Create(&devices).Error; err != nil {
		return fmt.Errorf("seed devices: %w", err)
	}
	return nil
}
