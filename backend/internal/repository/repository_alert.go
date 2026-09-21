package repository

import (
	"errors"
	"fmt"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
	"time"
)

type AlertRepository struct{ db *gorm.DB }

func NewAlertRepository(db *gorm.DB) *AlertRepository { return &AlertRepository{db: db} }
func (r *AlertRepository) Create(row *model.Alert) error {
	if err := r.db.Create(row).Error; err != nil {
		return fmt.Errorf("create alert: %w", err)
	}
	return nil
}
func (r *AlertRepository) List(greenhouseID uint) ([]model.Alert, error) {
	var rows []model.Alert
	q := r.db.Preload("Sensor.Threshold").Order("created_at desc")
	if greenhouseID > 0 {
		q = q.Where("greenhouse_id = ?", greenhouseID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	return rows, nil
}
func (r *AlertRepository) Handle(id uint) (*model.Alert, error) {
	var row model.Alert
	err := r.db.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get alert: %w", err)
	}
	now := time.Now()
	row.Status = "handled"
	row.HandledAt = &now
	if err = r.db.Save(&row).Error; err != nil {
		return nil, fmt.Errorf("handle alert: %w", err)
	}
	return &row, nil
}
func (r *AlertRepository) CountBetween(greenhouseID uint, start, end time.Time) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Alert{}).Where("greenhouse_id=? AND created_at BETWEEN ? AND ?", greenhouseID, start, end).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count alerts: %w", err)
	}
	return count, nil
}
