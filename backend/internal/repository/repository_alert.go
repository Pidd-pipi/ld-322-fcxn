package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AlertRepository struct{ db *gorm.DB }

func NewAlertRepository(db *gorm.DB) *AlertRepository { return &AlertRepository{db: db} }

// InTransaction runs fn inside a single database transaction and rolls back on
// any error so a failed alert evaluation never leaves a half-written state.
func (r *AlertRepository) InTransaction(fn func(tx *gorm.DB) error) error {
	if err := r.db.Transaction(fn); err != nil {
		return fmt.Errorf("alert transaction: %w", err)
	}
	return nil
}

func (r *AlertRepository) Create(tx *gorm.DB, row *model.Alert) error {
	if err := tx.Create(row).Error; err != nil {
		return fmt.Errorf("create alert: %w", err)
	}
	return nil
}

func (r *AlertRepository) Save(tx *gorm.DB, row *model.Alert) error {
	if err := tx.Save(row).Error; err != nil {
		return fmt.Errorf("save alert: %w", err)
	}
	return nil
}

// OpenBySensor returns the single un-recovered (pending) alert of a sensor, or
// nil when none exists. Only one open alert is kept per sensor.
func (r *AlertRepository) OpenBySensor(tx *gorm.DB, sensorID uint) (*model.Alert, error) {
	var row model.Alert
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sensor_id = ? AND status = ?", sensorID, constants.AlertPending).Order("created_at desc").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find open alert: %w", err)
	}
	return &row, nil
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

// Acknowledge marks an open alert as handled by an operator. Repeated
// acknowledgement or acknowledging an already recovered alert is a business
// conflict; in those cases no state is changed.
func (r *AlertRepository) Acknowledge(id uint) (*model.Alert, error) {
	var row model.Alert
	err := r.db.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get alert: %w", err)
	}
	switch row.Status {
	case constants.AlertHandled:
		return nil, apperrors.ErrAlertHandled
	case constants.AlertRecovered:
		return nil, apperrors.ErrAlertRecovered
	}
	now := time.Now()
	row.Status = constants.AlertHandled
	row.HandledAt = &now
	row.AcknowledgedAt = &now
	if err = r.db.Save(&row).Error; err != nil {
		return nil, fmt.Errorf("acknowledge alert: %w", err)
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
