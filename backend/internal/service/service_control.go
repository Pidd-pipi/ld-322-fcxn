package service

import (
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"log/slog"
)

type ControlService struct {
	repo   *repository.DeviceRepository
	logger *slog.Logger
	hub    *ws.Hub
}

func NewControlService(r *repository.DeviceRepository, l *slog.Logger, h *ws.Hub) *ControlService {
	return &ControlService{r, l, h}
}
func (s *ControlService) List(gid uint) ([]model.Device, error) { return s.repo.List(gid) }
func (s *ControlService) Toggle(id uint, status string) (*model.Device, error) {
	d, e := s.repo.Toggle(id, status)
	if e == nil {
		s.hub.Broadcast(constants.EventDevice, d)
	}
	return d, e
}
func (s *ControlService) Schedule(row *model.Schedule) error { return s.repo.CreateSchedule(row) }
func (s *ControlService) Schedules(id uint) ([]model.Schedule, error) {
	return s.repo.ListSchedules(id)
}
