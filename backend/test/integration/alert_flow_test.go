package integration

import (
	"bytes"
	"encoding/json"
	"github.com/cygreenenv/greenhouse-panel/internal/constants"
	"github.com/cygreenenv/greenhouse-panel/internal/model"
	"github.com/cygreenenv/greenhouse-panel/internal/repository"
	"github.com/cygreenenv/greenhouse-panel/internal/router"
	"github.com/cygreenenv/greenhouse-panel/internal/service"
	ws "github.com/cygreenenv/greenhouse-panel/internal/websocket"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type apiEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func setupServer(t *testing.T) (*httptest.Server, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:http_"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(model.All()...); err != nil {
		t.Fatal(err)
	}
	g := model.Greenhouse{Name: "HTTP 闭环温室", Location: "测试区", Area: 100}
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
	auth := service.NewAuthService("test-secret")
	monitoring := service.NewMonitoringService(greenhouseRepo, sensorRepo, alertRepo, logger, hub)
	alertService := service.NewAlertService(alertRepo, logger)
	control := service.NewControlService(deviceRepo, logger, hub)
	reports := service.NewReportService(sensorRepo, alertRepo)
	engine := router.New(router.Dependencies{Logger: logger, Auth: auth, Monitoring: monitoring, Alerts: alertService, Control: control, Reports: reports, Hub: hub})
	return httptest.NewServer(engine), db
}

func doJSON(t *testing.T, srv *httptest.Server, method, path, token string, body any) (int, apiEnvelope) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, srv.URL+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var env apiEnvelope
	if err = json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, env
}

// 实测完整闭环：持续超限合并 → 人工确认 → 重复确认冲突 → 自动恢复 → 确认已恢复冲突，刷新后状态一致。
func TestAlertLifecycleOverHTTP(t *testing.T) {
	srv, db := setupServer(t)
	defer srv.Close()

	_, loginEnv := doJSON(t, srv, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"username": "admin", "password": "admin123"})
	var auth struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginEnv.Data, &auth); err != nil || auth.Token == "" {
		t.Fatalf("login failed: %s", loginEnv.Message)
	}
	token := auth.Token

	ingest := func(value float64) {
		t.Helper()
		status, env := doJSON(t, srv, http.MethodPost, "/api/v1/readings", token, map[string]any{"sensorId": 1, "value": value})
		if status != http.StatusCreated {
			t.Fatalf("ingest %.1f -> status %d, message %s", value, status, env.Message)
		}
	}
	listAlerts := func() []model.Alert {
		t.Helper()
		status, env := doJSON(t, srv, http.MethodGet, "/api/v1/alerts?greenhouse_id=1", token, nil)
		if status != http.StatusOK {
			t.Fatalf("list alerts -> status %d", status)
		}
		var rows []model.Alert
		if err := json.Unmarshal(env.Data, &rows); err != nil {
			t.Fatal(err)
		}
		return rows
	}

	// 未鉴权写入应被拒绝，且不产生报警。
	if status, _ := doJSON(t, srv, http.MethodPost, "/api/v1/readings", "", map[string]any{"sensorId": 1, "value": 99}); status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated ingest should be 401, got %d", status)
	}

	// 持续超限：三次读数只保留一条未恢复报警，数值、级别、次数持续更新。
	ingest(33)
	ingest(35)
	ingest(40)
	rows := listAlerts()
	if len(rows) != 1 {
		t.Fatalf("want exactly 1 open alert, got %d: %#v", len(rows), rows)
	}
	alert := rows[0]
	if alert.Count != 3 || alert.Value != 40 || alert.Level != "critical" || alert.Status != constants.AlertPending {
		t.Fatalf("alert not merged correctly: %#v", alert)
	}

	// 人工确认成功。
	status, env := doJSON(t, srv, http.MethodPatch, "/api/v1/alerts/1/handle", token, nil)
	if status != http.StatusOK {
		t.Fatalf("first handle should be 200, got %d %s", status, env.Message)
	}
	// 重复确认：业务冲突 409，且刷新后状态仍是 handled、HandledAt 未被改动。
	status, env = doJSON(t, srv, http.MethodPatch, "/api/v1/alerts/1/handle", token, nil)
	if status != http.StatusConflict || env.Code != 40901 {
		t.Fatalf("duplicate handle should be 409/40901, got %d/%d", status, env.Code)
	}
	rows = listAlerts()
	if len(rows) != 1 || rows[0].Status != constants.AlertHandled || rows[0].HandledAt == nil || rows[0].Count != 3 {
		t.Fatalf("state after failed duplicate handle changed unexpectedly: %#v", rows)
	}

	// 读数回到阈值内：自动恢复，保存恢复值与时间。
	ingest(25)
	rows = listAlerts()
	if len(rows) != 1 {
		t.Fatalf("recovery must keep one record, got %d", len(rows))
	}
	alert = rows[0]
	if alert.Status != constants.AlertRecovered || alert.RecoveredAt == nil || alert.RecoveredValue == nil || *alert.RecoveredValue != 25 {
		t.Fatalf("recovery fields incorrect: %#v", alert)
	}

	// 确认已恢复报警：业务冲突，状态保持 recovered（失败回滚）。
	status, env = doJSON(t, srv, http.MethodPatch, "/api/v1/alerts/1/handle", token, nil)
	if status != http.StatusConflict {
		t.Fatalf("handling recovered alert should be 409, got %d/%d", status, env.Code)
	}
	var stored model.Alert
	if err := db.First(&stored, 1).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != constants.AlertRecovered {
		t.Fatalf("conflict must roll back with status recovered, got %s", stored.Status)
	}

	// 新的超限时段开启第二条报警，历史恢复记录仍保留。
	ingest(31)
	rows = listAlerts()
	if len(rows) != 2 {
		t.Fatalf("new episode should create a second alert, got %d", len(rows))
	}
	if rows[0].Status != constants.AlertPending || rows[0].Count != 1 {
		t.Fatalf("newest alert should be a fresh pending one, got %#v", rows[0])
	}
}
