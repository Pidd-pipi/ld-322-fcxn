package constants

const (
	APIPrefix           = "/api/v1"
	HealthPath          = "/healthz"
	WebSocketPath       = "/ws"
	DefaultPage         = 1
	DefaultPageSize     = 100
	MaxPageSize         = 500
	StatusOnline        = "online"
	StatusOff           = "off"
	StatusOn            = "on"
	AlertPending        = "pending"
	AlertHandled        = "handled"
	AlertRecovered      = "recovered"
	RoleAdmin           = "admin"
	SuccessMessage      = "ok"
	EventAlert          = "alert.created"
	EventAlertUpdate    = "alert.updated"
	EventAlertRecovered = "alert.recovered"
	EventReading        = "reading.created"
	EventDevice         = "device.updated"
)
