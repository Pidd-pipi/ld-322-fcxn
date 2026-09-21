package service

import (
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"log/slog"
	"math/rand"
	"time"
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
func (s *MonitoringService) Ingest(sensorID uint, value float64) (*model.SensorReading, *model.Alert, error) {
	sensor, err := s.sensorRepo.Get(sensorID)
	if err != nil {
		return nil, nil, err
	}
	reading := &model.SensorReading{SensorID: sensorID, Value: value, RecordedAt: time.Now()}
	if err = s.sensorRepo.AddReading(reading); err != nil {
		return nil, nil, err
	}
	var alert *model.Alert
	if value < sensor.Threshold.MinValue || value > sensor.Threshold.MaxValue {
		level := "warning"
		if value < sensor.Threshold.MinValue*.8 || value > sensor.Threshold.MaxValue*1.2 {
			level = "critical"
		}
		alert = &model.Alert{GreenhouseID: sensor.GreenhouseID, SensorID: sensor.ID, Level: level, Message: fmt.Sprintf("%s 当前值 %.2f%s 超出阈值 [%.2f, %.2f]", constants.SensorLabels[sensor.Type], value, sensor.Unit, sensor.Threshold.MinValue, sensor.Threshold.MaxValue), Value: value, Status: constants.AlertPending}
		if err = s.alertRepo.Create(alert); err != nil {
			return nil, nil, err
		}
		s.hub.Broadcast(constants.EventAlert, alert)
	}
	s.hub.Broadcast("reading.created", reading)
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
