package service

import (
	"errors"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"path/filepath"
	"sync"
	"testing"
)

func newLifecycleService(t *testing.T) (*gorm.DB, *MonitoringService, *AlertService, model.Sensor) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+filepath.Join(t.TempDir(), "lifecycle.db")+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	g := model.Greenhouse{Name: "闭环测试温室"}
	if err = db.Create(&g).Error; err != nil {
		t.Fatal(err)
	}
	sensor := model.Sensor{GreenhouseID: g.ID, Name: "温度", Type: "temperature", Unit: "°C", Status: "online"}
	if err = db.Create(&sensor).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&model.Threshold{SensorID: sensor.ID, MinValue: 10, MaxValue: 30}).Error; err != nil {
		t.Fatal(err)
	}
	alertRepo := repository.NewAlertRepository(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	monitoring := NewMonitoringService(repository.NewGreenhouseRepository(db), repository.NewSensorRepository(db), alertRepo, logger, ws.NewHub())
	alerts := NewAlertService(alertRepo, logger)
	return db, monitoring, alerts, sensor
}

func listAlerts(t *testing.T, db *gorm.DB, sensorID uint) []model.Alert {
	t.Helper()
	var rows []model.Alert
	if err := db.Where("sensor_id = ?", sensorID).Order("id asc").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}

// 持续超限只保留一条未恢复报警，新读数更新数值、级别和次数。
func TestPersistentExceedanceKeepsSingleAlert(t *testing.T) {
	db, monitoring, _, sensor := newLifecycleService(t)

	if _, alert, err := monitoring.Ingest(sensor.ID, 33); err != nil || alert == nil {
		t.Fatalf("first exceedance should create alert: %v %v", alert, err)
	}
	if _, alert, err := monitoring.Ingest(sensor.ID, 35); err != nil || alert == nil {
		t.Fatalf("second exceedance should return the open alert: %v %v", alert, err)
	}
	if _, alert, err := monitoring.Ingest(sensor.ID, 40); err != nil || alert == nil {
		t.Fatalf("third exceedance should return the open alert: %v %v", alert, err)
	}

	rows := listAlerts(t, db, sensor.ID)
	if len(rows) != 1 {
		t.Fatalf("persistent exceedance must keep a single alert, got %d", len(rows))
	}
	row := rows[0]
	if row.Count != 3 {
		t.Fatalf("want count 3, got %d", row.Count)
	}
	if row.Value != 40 {
		t.Fatalf("value should track the latest reading, got %.2f", row.Value)
	}
	if row.Level != "critical" {
		t.Fatalf("40 exceeds 30*1.2=36, want critical level, got %s", row.Level)
	}
	if row.Status != constants.AlertPending {
		t.Fatalf("alert must remain pending, got %s", row.Status)
	}
	if row.RecoveredAt != nil {
		t.Fatalf("open alert must not carry a recovery time")
	}
}

// 回到阈值内自动恢复，并保存恢复值与恢复时间。
func TestReadingWithinRangeRecoversAlert(t *testing.T) {
	db, monitoring, _, sensor := newLifecycleService(t)

	if _, _, err := monitoring.Ingest(sensor.ID, 35); err != nil {
		t.Fatal(err)
	}
	if _, alert, err := monitoring.Ingest(sensor.ID, 25); err != nil {
		t.Fatalf("in-range reading failed: %v", err)
	} else if alert == nil || alert.Status != constants.AlertRecovered {
		t.Fatalf("expected recovered alert, got %#v", alert)
	}

	rows := listAlerts(t, db, sensor.ID)
	if len(rows) != 1 {
		t.Fatalf("recovery must not create a new record, got %d", len(rows))
	}
	row := rows[0]
	if row.RecoveredAt == nil {
		t.Fatal("recoveredAt must be persisted")
	}
	if row.RecoveredValue == nil || *row.RecoveredValue != 25 {
		t.Fatalf("recoveredValue must be 25, got %v", row.RecoveredValue)
	}
	if row.Count != 1 {
		t.Fatalf("recovery reading must not increment count, got %d", row.Count)
	}

	// 恢复后读数继续正常，不产生新报警。
	if _, alert, err := monitoring.Ingest(sensor.ID, 22); err != nil || alert != nil {
		t.Fatalf("in-range readings after recovery must yield no alert, got %#v", alert)
	}
	if rows = listAlerts(t, db, sensor.ID); len(rows) != 1 {
		t.Fatalf("still one alert expected, got %d", len(rows))
	}

	// 恢复后再次超限应创建一条全新的报警。
	if _, alert, err := monitoring.Ingest(sensor.ID, 31); err != nil || alert == nil || alert.Count != 1 {
		t.Fatalf("a new episode must create a fresh alert with count 1, got %#v", alert)
	}
	if rows = listAlerts(t, db, sensor.ID); len(rows) != 2 {
		t.Fatalf("new episode should add a second alert, got %d", len(rows))
	}
}

// 人工确认只作用于未恢复报警；重复确认或确认已恢复报警返回业务冲突且不改变状态。
func TestHandleConflictsDoNotChangeState(t *testing.T) {
	db, monitoring, alerts, sensor := newLifecycleService(t)

	if _, _, err := monitoring.Ingest(sensor.ID, 33); err != nil {
		t.Fatal(err)
	}
	row := listAlerts(t, db, sensor.ID)[0]

	handled, err := alerts.Handle(row.ID)
	if err != nil {
		t.Fatalf("first handle should succeed: %v", err)
	}
	if handled.Status != constants.AlertHandled || handled.HandledAt == nil {
		t.Fatalf("handle should mark alert handled with timestamp, got %#v", handled)
	}

	// 重复确认 -> 业务冲突，状态保持 handled。
	if _, err = alerts.Handle(row.ID); !errors.Is(err, apperrors.ErrAlertConflict) {
		t.Fatalf("duplicate handle must return ErrAlertConflict, got %v", err)
	}
	again := listAlerts(t, db, sensor.ID)[0]
	if again.Status != constants.AlertHandled {
		t.Fatalf("failed handle must not change status, got %s", again.Status)
	}

	// 确认已恢复报警 -> 业务冲突。
	if _, _, err = monitoring.Ingest(sensor.ID, 25); err != nil {
		t.Fatal(err)
	}
	recovered := listAlerts(t, db, sensor.ID)[0]
	if recovered.Status != constants.AlertRecovered {
		t.Fatalf("handled alert should auto-recover in range, got %s", recovered.Status)
	}
	if _, err = alerts.Handle(recovered.ID); !errors.Is(err, apperrors.ErrAlertConflict) {
		t.Fatalf("handling a recovered alert must return ErrAlertConflict, got %v", err)
	}
	after := listAlerts(t, db, sensor.ID)[0]
	if after.Status != constants.AlertRecovered || after.HandledAt == nil {
		t.Fatalf("conflict must leave recovered alert intact, got %#v", after)
	}

	// 不存在的报警 -> not found。
	if _, err = alerts.Handle(99999); !errors.Is(err, apperrors.ErrRecordNotFound) {
		t.Fatalf("missing alert must return ErrRecordNotFound, got %v", err)
	}
}

// 已确认但未恢复的报警继续超限时仍更新同一条记录；恢复后次数与恢复值均保留。
func TestHandledAlertKeepsAccumulatingUntilRecovery(t *testing.T) {
	db, monitoring, alerts, sensor := newLifecycleService(t)

	if _, _, err := monitoring.Ingest(sensor.ID, 32); err != nil {
		t.Fatal(err)
	}
	row := listAlerts(t, db, sensor.ID)[0]
	if _, err := alerts.Handle(row.ID); err != nil {
		t.Fatal(err)
	}
	if _, alert, err := monitoring.Ingest(sensor.ID, 34); err != nil {
		t.Fatal(err)
	} else if alert == nil || alert.Count != 2 {
		t.Fatalf("handled open alert must keep accumulating, got %#v", alert)
	}
	rows := listAlerts(t, db, sensor.ID)
	if len(rows) != 1 || rows[0].Status != constants.AlertHandled || rows[0].Value != 34 {
		t.Fatalf("alert should remain handled with the newest value, got %#v", rows)
	}
}

// 并发读数：同一传感器并发超限时仍然只有一条未恢复报警，次数恰好累计。
func TestConcurrentIngestKeepsSingleOpenAlert(t *testing.T) {
	db, monitoring, _, sensor := newLifecycleService(t)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := monitoring.Ingest(sensor.ID, 34); err != nil {
				t.Errorf("concurrent ingest failed: %v", err)
			}
		}()
	}
	wg.Wait()

	var openCount int64
	if err := db.Model(&model.Alert{}).Where("sensor_id = ? AND status <> ?", sensor.ID, constants.AlertRecovered).Count(&openCount).Error; err != nil {
		t.Fatal(err)
	}
	if openCount != 1 {
		t.Fatalf("concurrent readings must keep exactly 1 open alert, got %d", openCount)
	}
	row := listAlerts(t, db, sensor.ID)[0]
	if row.Count != 20 {
		t.Fatalf("count must accumulate every concurrent reading, got %d", row.Count)
	}
}
