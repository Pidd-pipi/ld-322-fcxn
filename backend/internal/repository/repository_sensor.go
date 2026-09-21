package repository

import (
	"errors"
	"fmt"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
	"time"
)

type SensorRepository struct{ db *gorm.DB }

func NewSensorRepository(db *gorm.DB) *SensorRepository { return &SensorRepository{db: db} }
func (r *SensorRepository) Create(sensor *model.Sensor, threshold *model.Threshold) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sensor).Error; err != nil {
			return fmt.Errorf("create sensor: %w", err)
		}
		threshold.SensorID = sensor.ID
		if err := tx.Create(threshold).Error; err != nil {
			return fmt.Errorf("create threshold: %w", err)
		}
		return nil
	})
}
func (r *SensorRepository) Get(id uint) (*model.Sensor, error) {
	var sensor model.Sensor
	err := r.db.Preload("Threshold").First(&sensor, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get sensor: %w", err)
	}
	return &sensor, nil
}
func (r *SensorRepository) AddReading(reading *model.SensorReading) error {
	if err := r.db.Create(reading).Error; err != nil {
		return fmt.Errorf("create reading: %w", err)
	}
	return nil
}
func (r *SensorRepository) UpdateThreshold(id uint, min, max float64) (*model.Threshold, error) {
	var t model.Threshold
	err := r.db.Where("sensor_id = ?", id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get threshold: %w", err)
	}
	t.MinValue = min
	t.MaxValue = max
	if err = r.db.Save(&t).Error; err != nil {
		return nil, fmt.Errorf("update threshold: %w", err)
	}
	return &t, nil
}
func (r *SensorRepository) History(greenhouseID uint, types []string, start, end time.Time) ([]model.SensorReading, error) {
	q := r.db.Joins("JOIN sensors ON sensors.id = sensor_readings.sensor_id").Where("sensors.greenhouse_id = ? AND sensor_readings.recorded_at BETWEEN ? AND ?", greenhouseID, start, end).Preload("Sensor").Order("sensor_readings.recorded_at asc")
	if len(types) > 0 {
		q = q.Where("sensors.type IN ?", types)
	}
	var rows []model.SensorReading
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	return rows, nil
}
func (r *SensorRepository) LatestForGreenhouse(greenhouseID uint) ([]model.SensorReading, error) {
	var rows []model.SensorReading
	sql := `SELECT sr.* FROM sensor_readings sr JOIN sensors s ON s.id=sr.sensor_id JOIN (SELECT s2.type, MAX(sr2.recorded_at) latest FROM sensor_readings sr2 JOIN sensors s2 ON s2.id=sr2.sensor_id WHERE s2.greenhouse_id=? GROUP BY s2.type) l ON l.type=s.type AND l.latest=sr.recorded_at WHERE s.greenhouse_id=?`
	if err := r.db.Raw(sql, greenhouseID, greenhouseID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("latest readings: %w", err)
	}
	for i := range rows {
		if err := r.db.Preload("Threshold").First(&rows[i].Sensor, rows[i].SensorID).Error; err != nil {
			return nil, fmt.Errorf("load latest sensor: %w", err)
		}
	}
	return rows, nil
}
