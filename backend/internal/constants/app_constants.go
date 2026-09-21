package constants

const (
	APIPrefix       = "/api/v1"
	HealthPath      = "/healthz"
	WebSocketPath   = "/ws"
	DefaultPage     = 1
	DefaultPageSize = 100
	MaxPageSize     = 500
	StatusOnline    = "online"
	StatusOff       = "off"
	StatusOn        = "on"
	AlertPending    = "pending"
	AlertHandled    = "handled"
	AlertRecovered  = "recovered"
	RoleAdmin       = "admin"
	SuccessMessage  = "ok"
	EventAlert      = "alert.created"
	EventAlertUpd   = "alert.updated"
	EventAlertRecv  = "alert.recovered"
	EventDevice     = "device.updated"
)
