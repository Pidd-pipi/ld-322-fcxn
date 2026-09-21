package service

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"gorm.io/gorm"
)

type MonitoringService struct {
	greenhouseRepo *repository.GreenhouseRepository
	sensorRepo     *repository.SensorRepository
	alertRepo      *repository.AlertRepository
	logger         *slog.Logger
	hub            *ws.Hub
}

func NewMonitoringService(g *repository.GreenhouseRepository, s *repository.SensorRepository, a *repository.AlertRepository, l *slog.Logger, h *ws.Hub) *MonitoringService {
	return &MonitoringService{g, s, a, l, h}
}
func (s *MonitoringService) ListGreenhouses() ([]model.Greenhouse, error) {
	return s.greenhouseRepo.List()
}
func (s *MonitoringService) Detail(id uint) (*model.Greenhouse, error) {
	return s.greenhouseRepo.Get(id)
}

// alertLevel classifies an out-of-range reading into warning or critical.
func alertLevel(value float64, threshold model.Threshold) string {
	if value < threshold.MinValue*constants.CriticalLevelFactor || value > threshold.MaxValue*(2-constants.CriticalLevelFactor) {
		return constants.AlertLevelCritical
	}
	return constants.AlertLevelWarning
}

func alertMessage(sensor model.Sensor, value float64, threshold model.Threshold) string {
	return fmt.Sprintf("%s 当前值 %.2f%s 超出阈值 [%.2f, %.2f]", constants.SensorLabels[sensor.Type], value, sensor.Unit, threshold.MinValue, threshold.MaxValue)
}

// Ingest persists a reading and keeps exactly one un-recovered alert per
// sensor: repeated violations update the open alert's value/level/count, and a
// reading back inside the threshold auto-recovers it. Reading and alert are
// written in one transaction, so a failure rolls the whole ingestion back.
func (s *MonitoringService) Ingest(sensorID uint, value float64) (*model.SensorReading, *model.Alert, error) {
	sensor, err := s.sensorRepo.Get(sensorID)
	if err != nil {
		return nil, nil, err
	}
	reading := &model.SensorReading{SensorID: sensorID, Value: value, RecordedAt: time.Now()}
	var alert *model.Alert
	violation := value < sensor.Threshold.MinValue || value > sensor.Threshold.MaxValue
	txErr := s.alertRepo.InTransaction(func(tx *gorm.DB) error {
		if err := s.sensorRepo.AddReading(tx, reading); err != nil {
			return err
		}
		open, err := s.alertRepo.OpenBySensor(tx, sensorID)
		if err != nil {
			return err
		}
		switch {
		case violation && open != nil:
			open.Value = value
			open.Level = alertLevel(value, sensor.Threshold)
			open.Message = alertMessage(*sensor, value, sensor.Threshold)
			open.OccurrenceCount++
			alert = open
			if err := s.alertRepo.Save(tx, open); err != nil {
				return err
			}
		case violation:
			alert = &model.Alert{
				GreenhouseID:    sensor.GreenhouseID,
				SensorID:        sensor.ID,
				Level:           alertLevel(value, sensor.Threshold),
				Message:         alertMessage(*sensor, value, sensor.Threshold),
				Value:           value,
				Status:          constants.AlertPending,
				OccurrenceCount: 1,
			}
			if err := s.alertRepo.Create(tx, alert); err != nil {
				return err
			}
		case !violation && open != nil:
			now := time.Now()
			recoveredValue := value
			open.Status = constants.AlertRecovered
			open.RecoveredAt = &now
			open.RecoveredValue = &recoveredValue
			alert = open
			if err := s.alertRepo.Save(tx, open); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, nil, txErr
	}
	if alert != nil {
		if alert.Status == constants.AlertRecovered {
			s.hub.Broadcast(constants.EventAlertRecovered, alert)
		} else if alert.OccurrenceCount > 1 {
			s.hub.Broadcast(constants.EventAlertUpdate, alert)
		} else {
			s.hub.Broadcast(constants.EventAlert, alert)
		}
	}
	s.hub.Broadcast(constants.EventReading, reading)
	return reading, alert, nil
}
func (s *MonitoringService) History(greenhouseID uint, types []string, start, end time.Time) ([]model.SensorReading, error) {
	return s.sensorRepo.History(greenhouseID, types, start, end)
}
func (s *MonitoringService) Latest(greenhouseID uint) ([]model.SensorReading, error) {
	return s.sensorRepo.LatestForGreenhouse(greenhouseID)
}
func (s *MonitoringService) UpdateThreshold(sensorID uint, min, max float64) (*model.Threshold, error) {
	return s.sensorRepo.UpdateThreshold(sensorID, min, max)
}
func (s *MonitoringService) Simulate(greenhouseID uint) (int, error) {
	g, err := s.greenhouseRepo.Get(greenhouseID)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, sensor := range g.Sensors {
		base := (sensor.Threshold.MinValue + sensor.Threshold.MaxValue) / 2
		span := (sensor.Threshold.MaxValue - sensor.Threshold.MinValue) / 3
		value := base + (rand.Float64()-.5)*span
		if _, _, err = s.Ingest(sensor.ID, value); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
func (s *MonitoringService) CreateGreenhouse(row *model.Greenhouse) error {
	return s.greenhouseRepo.Create(row)
}
func (s *MonitoringService) CreateSensor(greenhouseID uint, name, sensorType string, min, max float64) (*model.Sensor, error) {
	sensor := &model.Sensor{GreenhouseID: greenhouseID, Name: name, Type: sensorType, Unit: constants.SensorUnits[sensorType], Status: constants.StatusOnline}
	threshold := &model.Threshold{MinValue: min, MaxValue: max}
	if err := s.sensorRepo.Create(sensor, threshold); err != nil {
		return nil, err
	}
	sensor.Threshold = *threshold
	return sensor, nil
}
