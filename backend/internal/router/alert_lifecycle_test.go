package router_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	"github.com/cygreenenv/greenhouse-panel/internal/router"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type apiEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type alertDTO struct {
	ID              uint     `json:"id"`
	GreenhouseID    uint     `json:"greenhouseId"`
	SensorID        uint     `json:"sensorId"`
	Level           string   `json:"level"`
	Message         string   `json:"message"`
	Value           float64  `json:"value"`
	Status          string   `json:"status"`
	OccurrenceCount int      `json:"occurrenceCount"`
	CreatedAt       string   `json:"createdAt"`
	AcknowledgedAt  *string  `json:"acknowledgedAt"`
	RecoveredAt     *string  `json:"recoveredAt"`
	RecoveredValue  *float64 `json:"recoveredValue"`
	HandledAt       *string  `json:"handledAt"`
}

type ingestData struct {
	Reading json.RawMessage `json:"reading"`
	Alert   *alertDTO       `json:"alert"`
}

func setupLifecycleRouter(t *testing.T) (http.Handler, *gorm.DB, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:router_lifecycle?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	g := model.Greenhouse{Name: "联调温室", Location: "测试区", Area: 100}
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
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	greenhouseRepo := repository.NewGreenhouseRepository(db)
	sensorRepo := repository.NewSensorRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	hub := ws.NewHub()
	auth := service.NewAuthService("integration-secret")
	monitoring := service.NewMonitoringService(greenhouseRepo, sensorRepo, alertRepo, logger, hub)
	control := service.NewControlService(deviceRepo, logger, hub)
	alerts := service.NewAlertService(alertRepo, logger)
	reports := service.NewReportService(sensorRepo, alertRepo)
	engine := router.New(router.Dependencies{Logger: logger, Auth: auth, Monitoring: monitoring, Alerts: alerts, Control: control, Reports: reports, Hub: hub})
	return engine, db, sensor.ID
}

func doJSON(t *testing.T, engine http.Handler, method, path, token string, body any) (int, apiEnvelope) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	var envelope apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("response was not the standard envelope: %s", rec.Body.String())
	}
	return rec.Code, envelope
}

func loginToken(t *testing.T, engine http.Handler) string {
	t.Helper()
	status, envelope := doJSON(t, engine, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"username": "admin", "password": "admin123"})
	if status != http.StatusOK {
		t.Fatalf("login failed: %d %s", status, envelope.Message)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data.Token
}

func ingest(t *testing.T, engine http.Handler, token string, sensorID uint, value float64) *alertDTO {
	t.Helper()
	status, envelope := doJSON(t, engine, http.MethodPost, "/api/v1/readings", token, map[string]any{"sensorId": sensorID, "value": value})
	if status != http.StatusCreated {
		t.Fatalf("ingest %.1f failed: %d %s", value, status, envelope.Message)
	}
	var data ingestData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data.Alert
}

func listAlerts(t *testing.T, engine http.Handler, token string, greenhouseID uint) []alertDTO {
	t.Helper()
	status, envelope := doJSON(t, engine, http.MethodGet, "/api/v1/alerts?greenhouse_id=1", token, nil)
	if status != http.StatusOK {
		t.Fatalf("list alerts failed: %d", status)
	}
	var rows []alertDTO
	if err := json.Unmarshal(envelope.Data, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

// TestAlertLifecycleOverHTTP drives the closed loop through the real router:
// continuous violation, auto recovery, duplicate/recovered acknowledgement
// conflicts and rollback-safe state.
func TestAlertLifecycleOverHTTP(t *testing.T) {
	engine, _, sensorID := setupLifecycleRouter(t)
	token := loginToken(t, engine)

	// 1) First out-of-range reading opens the only pending alert.
	first := ingest(t, engine, token, sensorID, 35)
	if first == nil || first.Status != "pending" || first.OccurrenceCount != 1 || first.Level != "warning" {
		t.Fatalf("unexpected first alert: %#v", first)
	}
	if first.GreenhouseID != 1 || first.SensorID != sensorID || first.CreatedAt == "" || first.Message == "" {
		t.Fatalf("alert must keep legacy wire fields: %#v", first)
	}

	// 2) Continuous violation updates value/level/count on the same row.
	second := ingest(t, engine, token, sensorID, 40)
	if second == nil || second.ID != first.ID {
		t.Fatalf("repeated violation must reuse the open alert, got %#v", second)
	}
	if second.OccurrenceCount != 2 || second.Value != 40 || second.Level != "critical" {
		t.Fatalf("open alert should track latest reading, got %#v", second)
	}
	if rows := listAlerts(t, engine, token, 1); len(rows) != 1 {
		t.Fatalf("continuous violation must keep a single alert, got %d", len(rows))
	}

	// 3) Acknowledge once, then a duplicate must conflict without changing state.
	if status, envelope := doJSON(t, engine, http.MethodPatch, "/api/v1/alerts/1/handle", token, nil); status != http.StatusOK || envelope.Code != 0 {
		t.Fatalf("first acknowledge should succeed, got %d %s", status, envelope.Message)
	}
	rows := listAlerts(t, engine, token, 1)
	handledAt := rows[0].HandledAt
	if rows[0].Status != "handled" || handledAt == nil || rows[0].AcknowledgedAt == nil {
		t.Fatalf("acknowledge metadata missing: %#v", rows[0])
	}
	status, envelope := doJSON(t, engine, http.MethodPatch, "/api/v1/alerts/1/handle", token, nil)
	if status != http.StatusConflict || envelope.Code != 40901 {
		t.Fatalf("duplicate acknowledge must be a 409/40901 conflict, got %d %d", status, envelope.Code)
	}
	rows = listAlerts(t, engine, token, 1)
	if rows[0].Status != "handled" || rows[0].HandledAt == nil || *rows[0].HandledAt != *handledAt {
		t.Fatalf("failed acknowledge must not change state: %#v", rows[0])
	}

	// 4) Re-violation opens a new alert; returning in range auto-recovers it.
	next := ingest(t, engine, token, sensorID, 36)
	if next == nil || next.ID == first.ID || next.Status != "pending" || next.OccurrenceCount != 1 {
		t.Fatalf("re-violation after acknowledgement should open a new alert, got %#v", next)
	}
	recovered := ingest(t, engine, token, sensorID, 25)
	if recovered == nil || recovered.ID != next.ID || recovered.Status != "recovered" {
		t.Fatalf("in-range reading should recover the open alert, got %#v", recovered)
	}
	if recovered.RecoveredAt == nil || recovered.RecoveredValue == nil || *recovered.RecoveredValue != 25 {
		t.Fatalf("recovery must save value and time: %#v", recovered)
	}
	status, envelope = doJSON(t, engine, http.MethodPatch, "/api/v1/alerts/2/handle", token, nil)
	if status != http.StatusConflict || envelope.Code != 40902 {
		t.Fatalf("acknowledging a recovered alert must be 409/40902, got %d %d", status, envelope.Code)
	}

	// 5) The list reflects confirmation, recovery and counts and stays stable
	// across refreshes.
	live := listAlerts(t, engine, token, 1)
	if len(live) != 2 {
		t.Fatalf("want 2 alert rows, got %d", len(live))
	}
	byID := map[uint]alertDTO{}
	for _, row := range live {
		byID[row.ID] = row
	}
	if byID[1].Status != "handled" || byID[1].OccurrenceCount != 2 {
		t.Fatalf("first row mismatch: %#v", byID[1])
	}
	if byID[2].Status != "recovered" || byID[2].OccurrenceCount != 1 || byID[2].RecoveredValue == nil {
		t.Fatalf("second row mismatch: %#v", byID[2])
	}
	if refreshed := listAlerts(t, engine, token, 1); len(refreshed) != 2 ||
		refreshed[0].Status != live[0].Status || refreshed[0].OccurrenceCount != live[0].OccurrenceCount {
		t.Fatalf("alert list must be consistent after refresh: %#v vs %#v", refreshed, live)
	}

	// 6) Unauthenticated writes are rejected and leave state untouched.
	if status, _ := doJSON(t, engine, http.MethodPatch, "/api/v1/alerts/2/handle", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("acknowledging without a token must be 401, got %d", status)
	}
	finalRows := listAlerts(t, engine, token, 1)
	finalByID := map[uint]alertDTO{}
	for _, row := range finalRows {
		finalByID[row.ID] = row
	}
	if finalByID[1].Status != "handled" || finalByID[2].Status != "recovered" {
		t.Fatalf("rejected request must not mutate state: %#v", finalRows)
	}
}
