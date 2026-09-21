package service

import (
	"log/slog"

	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
)

type AlertService struct {
	repo   *repository.AlertRepository
	logger *slog.Logger
}

func NewAlertService(r *repository.AlertRepository, l *slog.Logger) *AlertService {
	return &AlertService{r, l}
}
func (s *AlertService) List(id uint) ([]model.Alert, error) { return s.repo.List(id) }

// Acknowledge marks an open alert as handled by an operator. Duplicate
// acknowledgement and acknowledging a recovered alert are business conflicts.
func (s *AlertService) Acknowledge(id uint) (*model.Alert, error) {
	return s.repo.Acknowledge(id)
}
