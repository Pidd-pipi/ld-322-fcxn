package repository

import (
	"errors"
	"fmt"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
)

type DeviceRepository struct{ db *gorm.DB }

func NewDeviceRepository(db *gorm.DB) *DeviceRepository { return &DeviceRepository{db: db} }
func (r *DeviceRepository) List(greenhouseID uint) ([]model.Device, error) {
	var rows []model.Device
	if err := r.db.Where("greenhouse_id=?", greenhouseID).Order("id asc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	return rows, nil
}
func (r *DeviceRepository) Toggle(id uint, status string) (*model.Device, error) {
	var row model.Device
	err := r.db.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	row.Status = status
	if err = r.db.Save(&row).Error; err != nil {
		return nil, fmt.Errorf("save device: %w", err)
	}
	if err = r.db.Create(&model.DeviceAction{DeviceID: id, Action: status, Operator: "admin"}).Error; err != nil {
		return nil, fmt.Errorf("create device action: %w", err)
	}
	return &row, nil
}
func (r *DeviceRepository) CreateSchedule(row *model.Schedule) error {
	if err := r.db.Create(row).Error; err != nil {
		return fmt.Errorf("create schedule: %w", err)
	}
	return nil
}
func (r *DeviceRepository) ListSchedules(deviceID uint) ([]model.Schedule, error) {
	var rows []model.Schedule
	if err := r.db.Where("device_id=?", deviceID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	return rows, nil
}
