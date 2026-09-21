package model

import "time"

type Greenhouse struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);uniqueIndex" json:"name"`
	Location  string    `gorm:"type:varchar(255)" json:"location"`
	Area      float64   `json:"area"`
	CreatedAt time.Time `json:"createdAt"`
	Sensors   []Sensor  `json:"sensors,omitempty"`
	Devices   []Device  `json:"devices,omitempty"`
}
