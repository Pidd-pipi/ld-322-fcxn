package constants

const (
	SensorTemperature = "temperature"
	SensorHumidity    = "humidity"
	SensorLight       = "light"
	SensorCO2         = "co2"
	SensorSoil        = "soil_moisture"
)

var SensorUnits = map[string]string{SensorTemperature: "°C", SensorHumidity: "%", SensorLight: "lux", SensorCO2: "ppm", SensorSoil: "%"}
var SensorLabels = map[string]string{SensorTemperature: "温度", SensorHumidity: "湿度", SensorLight: "光照", SensorCO2: "CO₂", SensorSoil: "土壤湿度"}
