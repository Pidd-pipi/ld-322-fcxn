package service

import (
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	"log/slog"
)

type AlertService struct {
	repo   *repository.AlertRepository
	logger *slog.Logger
}

func NewAlertService(r *repository.AlertRepository, l *slog.Logger) *AlertService {
	return &AlertService{r, l}
}
func (s *AlertService) List(id uint) ([]model.Alert, error)  { return s.repo.List(id) }
func (s *AlertService) Handle(id uint) (*model.Alert, error) { return s.repo.Handle(id) }
