package model

import "time"

type Device struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	GreenhouseID uint      `gorm:"index" json:"greenhouseId"`
	Name         string    `gorm:"type:varchar(100)" json:"name"`
	Type         string    `gorm:"type:varchar(50)" json:"type"`
	Status       string    `gorm:"type:varchar(32)" json:"status"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
type DeviceAction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  uint      `gorm:"index" json:"deviceId"`
	Action    string    `gorm:"type:varchar(32)" json:"action"`
	Operator  string    `gorm:"type:varchar(100)" json:"operator"`
	CreatedAt time.Time `json:"createdAt"`
}
type Schedule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  uint      `gorm:"index" json:"deviceId"`
	Cron      string    `gorm:"type:varchar(100)" json:"cron"`
	Action    string    `gorm:"type:varchar(32)" json:"action"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
}
type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"type:varchar(100);uniqueIndex" json:"username"`
	PasswordHash string `gorm:"type:varchar(255)" json:"-"`
	Role         string `gorm:"type:varchar(32)" json:"role"`
}
