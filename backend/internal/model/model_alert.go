package model

import "time"

type Alert struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	GreenhouseID    uint       `gorm:"index" json:"greenhouseId"`
	SensorID        uint       `gorm:"index" json:"sensorId"`
	Level           string     `gorm:"type:varchar(32)" json:"level"`
	Message         string     `gorm:"type:varchar(500)" json:"message"`
	Value           float64    `json:"value"`
	Status          string     `gorm:"type:varchar(32);index" json:"status"`
	OccurrenceCount int        `gorm:"default:1" json:"occurrenceCount"`
	CreatedAt       time.Time  `json:"createdAt"`
	AcknowledgedAt  *time.Time `json:"acknowledgedAt,omitempty"`
	RecoveredAt     *time.Time `json:"recoveredAt,omitempty"`
	RecoveredValue  *float64   `json:"recoveredValue,omitempty"`
	HandledAt       *time.Time `json:"handledAt,omitempty"`
	Sensor          Sensor     `json:"sensor,omitempty"`
}
