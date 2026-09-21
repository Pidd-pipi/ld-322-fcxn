package service

import (
	"io"
	"log/slog"
	"testing"

	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	apperrors "github.com/cygreenenv/greenhouse-panel/internal/errors"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newIngestFixture(t *testing.T, dsn string, min, max float64) (*gorm.DB, *MonitoringService, model.Sensor, model.Greenhouse) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	g := model.Greenhouse{Name: "测试温室"}
	if err = db.Create(&g).Error; err != nil {
		t.Fatal(err)
	}
	sensor := model.Sensor{GreenhouseID: g.ID, Name: "温度", Type: "temperature", Unit: "°C", Status: "online"}
	if err = db.Create(&sensor).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&model.Threshold{SensorID: sensor.ID, MinValue: min, MaxValue: max}).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewMonitoringService(repository.NewGreenhouseRepository(db), repository.NewSensorRepository(db), repository.NewAlertRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)), ws.NewHub())
	return db, svc, sensor, g
}

func TestIngestKeepsSingleOpenAlertForContinuousViolation(t *testing.T) {
	db, svc, sensor, g := newIngestFixture(t, "file:ingest_single?mode=memory&cache=shared", 10, 30)
	repo := repository.NewAlertRepository(db)

	if _, alert, err := svc.Ingest(sensor.ID, 31); err != nil {
		t.Fatal(err)
	} else if alert == nil || alert.Status != constants.AlertPending || alert.OccurrenceCount != 1 || alert.Level != constants.AlertLevelWarning {
		t.Fatalf("unexpected first alert: %#v", alert)
	}
	if _, alert, err := svc.Ingest(sensor.ID, 33); err != nil {
		t.Fatal(err)
	} else if alert == nil || alert.OccurrenceCount != 2 || alert.Value != 33 {
		t.Fatalf("second reading should update the open alert: %#v", alert)
	}
	if _, alert, err := svc.Ingest(sensor.ID, 40); err != nil {
		t.Fatal(err)
	} else if alert == nil || alert.OccurrenceCount != 3 || alert.Level != constants.AlertLevelCritical {
		t.Fatalf("third reading should bump count and escalate level: %#v", alert)
	}
	alerts, err := repo.List(g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 {
		t.Fatalf("continuous violation must keep one open alert, got %d", len(alerts))
	}
	if alerts[0].Value != 40 || alerts[0].Level != constants.AlertLevelCritical {
		t.Fatalf("open alert should hold the latest value and level: %#v", alerts[0])
	}
	var readings int64
	if err = db.Model(&model.SensorReading{}).Count(&readings).Error; err != nil {
		t.Fatal(err)
	}
	if readings != 3 {
		t.Fatalf("every reading must be persisted, want 3, got %d", readings)
	}
}

func TestIngestAutoRecoversOpenAlert(t *testing.T) {
	db, svc, sensor, g := newIngestFixture(t, "file:ingest_recover?mode=memory&cache=shared", 10, 30)
	repo := repository.NewAlertRepository(db)

	if _, alert, err := svc.Ingest(sensor.ID, 35); err != nil {
		t.Fatal(err)
	} else if alert == nil {
		t.Fatal("expected alert for out-of-range reading")
	}
	if _, alert, err := svc.Ingest(sensor.ID, 25); err != nil {
		t.Fatal(err)
	} else if alert == nil || alert.Status != constants.AlertRecovered {
		t.Fatalf("reading back in range should recover the alert: %#v", alert)
	} else {
		if alert.RecoveredAt == nil {
			t.Fatal("recoveredAt must be saved")
		}
		if alert.RecoveredValue == nil || *alert.RecoveredValue != 25 {
			t.Fatalf("recoveredValue must record the restoring reading, got %#v", alert.RecoveredValue)
		}
	}
	alerts, err := repo.List(g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Status != constants.AlertRecovered {
		t.Fatalf("the same alert row must be marked recovered: %#v", alerts)
	}
	// A new violation after recovery opens a brand new alert instead of
	// resurrecting the recovered one.
	if _, alert, err := svc.Ingest(sensor.ID, 36); err != nil {
		t.Fatal(err)
	} else if alert == nil || alert.ID == alerts[0].ID || alert.OccurrenceCount != 1 {
		t.Fatalf("re-violation should create a new open alert: %#v", alert)
	}
	if alerts, err = repo.List(g.ID); err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 2 {
		t.Fatalf("want one recovered and one open alert, got %d", len(alerts))
	}
}

func TestAcknowledgeConflictsDoNotChangeState(t *testing.T) {
	db, svc, sensor, _ := newIngestFixture(t, "file:ingest_ack?mode=memory&cache=shared", 10, 30)
	alertSvc := NewAlertService(repository.NewAlertRepository(db), slog.New(slog.NewTextHandler(io.Discard, nil)))

	if _, alert, err := svc.Ingest(sensor.ID, 35); err != nil {
		t.Fatal(err)
	} else {
		handled, err := alertSvc.Acknowledge(alert.ID)
		if err != nil {
			t.Fatalf("first acknowledgement should succeed: %v", err)
		}
		if handled.Status != constants.AlertHandled || handled.AcknowledgedAt == nil || handled.HandledAt == nil {
			t.Fatalf("acknowledged alert must record handle metadata: %#v", handled)
		}
		if _, err = alertSvc.Acknowledge(alert.ID); err != apperrors.ErrAlertHandled {
			t.Fatalf("duplicate acknowledgement should return conflict, got %v", err)
		}
	}

	// A recovered alert cannot be acknowledged either.
	if _, fresh, err := svc.Ingest(sensor.ID, 40); err != nil {
		// sensor now has a handled alert, so 40 opens a fresh pending alert
		t.Fatal(err)
	} else if fresh == nil {
		t.Fatal("re-violation after handled alert should open a new alert")
	} else {
		if _, alert, err := svc.Ingest(sensor.ID, 20); err != nil {
			t.Fatal(err)
		} else if alert == nil || alert.Status != constants.AlertRecovered {
			t.Fatalf("expected recovery, got %#v", alert)
		}
		if _, err = alertSvc.Acknowledge(fresh.ID); err != apperrors.ErrAlertRecovered {
			t.Fatalf("acknowledging a recovered alert should return conflict, got %v", err)
		}
	}
}

func TestIngestRollsBackOnFailure(t *testing.T) {
	db, svc, sensor, _ := newIngestFixture(t, "file:ingest_rollback?mode=memory&cache=shared", 10, 30)

	var before int64
	if err := db.Model(&model.SensorReading{}).Count(&before).Error; err != nil {
		t.Fatal(err)
	}
	// Drop the alerts table so the alert write fails mid-transaction; the
	// reading written earlier in the same transaction must be rolled back.
	if err := db.Migrator().DropTable("alerts"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Ingest(sensor.ID, 35); err == nil {
		t.Fatal("expected ingestion to fail when alert persistence fails")
	}
	var after int64
	if err := db.Model(&model.SensorReading{}).Count(&after).Error; err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("failed ingestion must roll back the reading, before=%d after=%d", before, after)
	}
}
