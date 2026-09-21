package service

import (
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	"time"
)

type ReportService struct {
	sensorRepo *repository.SensorRepository
	alertRepo  *repository.AlertRepository
}
type Report struct {
	GreenhouseID uint              `json:"greenhouseId"`
	Range        string            `json:"range"`
	GeneratedAt  time.Time         `json:"generatedAt"`
	Alerts       int64             `json:"alerts"`
	Metrics      map[string]Metric `json:"metrics"`
}
type Metric struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Unit    string  `json:"unit"`
}

func NewReportService(s *repository.SensorRepository, a *repository.AlertRepository) *ReportService {
	return &ReportService{s, a}
}
func (s *ReportService) Generate(id uint, rangeName string, start, end time.Time) (*Report, error) {
	rows, err := s.sensorRepo.History(id, nil, start, end)
	if err != nil {
		return nil, fmt.Errorf("report history: %w", err)
	}
	report := &Report{GreenhouseID: id, Range: rangeName, GeneratedAt: time.Now(), Metrics: map[string]Metric{}}
	sums := map[string]float64{}
	counts := map[string]int{}
	for _, r := range rows {
		m, ok := report.Metrics[r.Sensor.Type]
		if !ok {
			m = Metric{Min: r.Value, Max: r.Value, Unit: r.Sensor.Unit}
		}
		if r.Value < m.Min {
			m.Min = r.Value
		}
		if r.Value > m.Max {
			m.Max = r.Value
		}
		sums[r.Sensor.Type] += r.Value
		counts[r.Sensor.Type]++
		m.Average = sums[r.Sensor.Type] / float64(counts[r.Sensor.Type])
		report.Metrics[r.Sensor.Type] = m
	}
	report.Alerts, err = s.alertRepo.CountBetween(id, start, end)
	if err != nil {
		return nil, err
	}
	return report, nil
}

var _ = model.Alert{}
