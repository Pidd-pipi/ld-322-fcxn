package model

func All() []any {
	return []any{&Greenhouse{}, &Sensor{}, &SensorReading{}, &Threshold{}, &Alert{}, &Device{}, &DeviceAction{}, &Schedule{}, &User{}}
}
