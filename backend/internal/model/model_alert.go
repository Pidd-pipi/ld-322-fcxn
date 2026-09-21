package model

import "time"

type Alert struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	GreenhouseID   uint       `gorm:"index" json:"greenhouseId"`
	SensorID       uint       `gorm:"index" json:"sensorId"`
	Level          string     `gorm:"type:varchar(32)" json:"level"`
	Message        string     `gorm:"type:varchar(500)" json:"message"`
	Value          float64    `json:"value"`
	Count          int        `gorm:"default:1;not null" json:"count"`
	Status         string     `gorm:"type:varchar(32);index" json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	HandledAt      *time.Time `json:"handledAt,omitempty"`
	RecoveredAt    *time.Time `json:"recoveredAt,omitempty"`
	RecoveredValue *float64   `json:"recoveredValue,omitempty"`
	Sensor         Sensor     `json:"sensor,omitempty"`
}

// Active 表示报警尚未随读数恢复，仍处于打开状态（待处理或已确认）。
func (a *Alert) Active() bool { return a.Status != "recovered" }
