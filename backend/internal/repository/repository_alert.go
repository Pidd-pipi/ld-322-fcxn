package repository

import (
	"errors"
	"fmt"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"gorm.io/gorm"
	"sync"
	"time"
)

// AlertAction 描述一次读数评估对报警产生的影响。
type AlertAction string

const (
	AlertActionNone      AlertAction = "none"
	AlertActionCreated   AlertAction = "created"
	AlertActionUpdated   AlertAction = "updated"
	AlertActionRecovered AlertAction = "recovered"
)

type AlertRepository struct {
	db *gorm.DB
	// sensorLocks 串行化同一传感器的读数评估，避免并发读数重复创建未恢复报警。
	sensorMuLock sync.Mutex
	sensorLocks  map[uint]*sync.Mutex
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db, sensorLocks: make(map[uint]*sync.Mutex)}
}

func (r *AlertRepository) sensorLock(sensorID uint) *sync.Mutex {
	r.sensorMuLock.Lock()
	defer r.sensorMuLock.Unlock()
	mu, ok := r.sensorLocks[sensorID]
	if !ok {
		mu = &sync.Mutex{}
		r.sensorLocks[sensorID] = mu
	}
	return mu
}

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

// Reconcile 根据最新读数评估报警闭环，整个过程在单事务内完成：
//   - 超限且存在未恢复报警：更新最新值、级别并累计次数；
//   - 超限但没有未恢复报警：创建一条新报警；
//   - 回到阈值内：将未恢复报警置为 recovered，保存恢复值与恢复时间。
//
// 同一传感器的评估按进程内互斥串行；更新只写相关列，不会覆盖并发的人工确认状态。
func (r *AlertRepository) Reconcile(sensor *model.Sensor, value float64, message string) (*model.Alert, AlertAction, error) {
	outOfRange := value < sensor.Threshold.MinValue || value > sensor.Threshold.MaxValue
	level := alertLevel(sensor, value)
	lock := r.sensorLock(sensor.ID)
	lock.Lock()
	defer lock.Unlock()

	var alert *model.Alert
	var action AlertAction
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var open model.Alert
		result := tx.Where("sensor_id = ? AND status <> ?", sensor.ID, constants.AlertRecovered).Order("created_at desc").Limit(1).Find(&open)
		if result.Error != nil {
			return fmt.Errorf("load open alert: %w", result.Error)
		}
		hasOpen := result.RowsAffected > 0
		switch {
		case outOfRange && hasOpen:
			updated := tx.Model(&model.Alert{}).Where("id = ? AND status <> ?", open.ID, constants.AlertRecovered).
				Updates(map[string]any{"level": level, "message": message, "value": value, "count": gorm.Expr("count + 1")})
			if updated.Error != nil {
				return fmt.Errorf("update alert: %w", updated.Error)
			}
			if updated.RowsAffected == 0 {
				return apperrors.ErrAlertConflict
			}
			if err := tx.First(&open, open.ID).Error; err != nil {
				return fmt.Errorf("reload alert: %w", err)
			}
			alert, action = &open, AlertActionUpdated
		case outOfRange:
			row := &model.Alert{
				GreenhouseID: sensor.GreenhouseID,
				SensorID:     sensor.ID,
				Level:        level,
				Message:      message,
				Value:        value,
				Count:        1,
				Status:       constants.AlertPending,
			}
			if err := tx.Create(row).Error; err != nil {
				return fmt.Errorf("create alert: %w", err)
			}
			alert, action = row, AlertActionCreated
		case hasOpen:
			recoveredAt := time.Now()
			recovered := tx.Model(&model.Alert{}).Where("id = ? AND status <> ?", open.ID, constants.AlertRecovered).
				Updates(map[string]any{"status": constants.AlertRecovered, "recovered_at": recoveredAt, "recovered_value": value})
			if recovered.Error != nil {
				return fmt.Errorf("recover alert: %w", recovered.Error)
			}
			if recovered.RowsAffected == 0 {
				return apperrors.ErrAlertConflict
			}
			if err := tx.First(&open, open.ID).Error; err != nil {
				return fmt.Errorf("reload recovered alert: %w", err)
			}
			alert, action = &open, AlertActionRecovered
		default:
			action = AlertActionNone
		}
		return nil
	})
	if err != nil {
		return nil, AlertActionNone, err
	}
	return alert, action, nil
}

// Handle 人工确认一条仍未恢复的报警；重复确认或确认已恢复报警返回业务冲突，状态不变。
func (r *AlertRepository) Handle(id uint) (*model.Alert, error) {
	var row model.Alert
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&row, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrRecordNotFound
			}
			return fmt.Errorf("get alert: %w", err)
		}
		if row.Status != constants.AlertPending {
			return apperrors.ErrAlertConflict
		}
		now := time.Now()
		result := tx.Model(&model.Alert{}).Where("id = ? AND status = ?", id, constants.AlertPending).
			Updates(map[string]any{"status": constants.AlertHandled, "handled_at": &now})
		if result.Error != nil {
			return fmt.Errorf("handle alert: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return apperrors.ErrAlertConflict
		}
		row.Status = constants.AlertHandled
		row.HandledAt = &now
		return nil
	})
	if err != nil {
		return nil, err
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

// alertLevel 按读数相对阈值的偏离程度计算报警级别。
func alertLevel(sensor *model.Sensor, value float64) string {
	if value < sensor.Threshold.MinValue*0.8 || value > sensor.Threshold.MaxValue*1.2 {
		return "critical"
	}
	return "warning"
}
