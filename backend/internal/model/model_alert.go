package model

import "time"

type Alert struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	GreenhouseID uint       `gorm:"index" json:"greenhouseId"`
	SensorID     uint       `gorm:"index" json:"sensorId"`
	Level        string     `gorm:"type:varchar(32)" json:"level"`
	Message      string     `gorm:"type:varchar(500)" json:"message"`
	Value        float64    `json:"value"`
	Status       string     `gorm:"type:varchar(32);index" json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	HandledAt    *time.Time `json:"handledAt,omitempty"`
	Sensor       Sensor     `json:"sensor,omitempty"`
}
