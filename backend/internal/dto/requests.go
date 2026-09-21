package dto

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=3,max=100"`
}
type GreenhouseRequest struct {
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Location string  `json:"location" validate:"required,max=100"`
	Area     float64 `json:"area" validate:"gt=0"`
}
type SensorRequest struct {
	GreenhouseID uint    `json:"greenhouseId" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	Type         string  `json:"type" validate:"required,oneof=temperature humidity light co2 soil_moisture"`
	MinValue     float64 `json:"minValue"`
	MaxValue     float64 `json:"maxValue" validate:"gt=0"`
}
type ReadingRequest struct {
	SensorID uint    `json:"sensorId" validate:"required"`
	Value    float64 `json:"value" validate:"required"`
}
type ThresholdRequest struct {
	MinValue float64 `json:"minValue"`
	MaxValue float64 `json:"maxValue" validate:"gt=0"`
}
type DeviceToggleRequest struct {
	Status string `json:"status" validate:"required,oneof=on off"`
}
type ScheduleRequest struct {
	DeviceID uint   `json:"deviceId" validate:"required"`
	Cron     string `json:"cron" validate:"required,max=50"`
	Action   string `json:"action" validate:"required,oneof=on off"`
}
