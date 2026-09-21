package repository

import (
	"errors"
	"fmt"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
)

type GreenhouseRepository struct{ db *gorm.DB }

func NewGreenhouseRepository(db *gorm.DB) *GreenhouseRepository { return &GreenhouseRepository{db: db} }
func (r *GreenhouseRepository) List() ([]model.Greenhouse, error) {
	var rows []model.Greenhouse
	if err := r.db.Preload("Sensors.Threshold").Preload("Devices").Order("id asc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list greenhouses: %w", err)
	}
	return rows, nil
}
func (r *GreenhouseRepository) Create(row *model.Greenhouse) error {
	if err := r.db.Create(row).Error; err != nil {
		return fmt.Errorf("create greenhouse: %w", err)
	}
	return nil
}
func (r *GreenhouseRepository) Get(id uint) (*model.Greenhouse, error) {
	var row model.Greenhouse
	err := r.db.Preload("Sensors.Threshold").Preload("Devices").First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get greenhouse: %w", err)
	}
	return &row, nil
}
