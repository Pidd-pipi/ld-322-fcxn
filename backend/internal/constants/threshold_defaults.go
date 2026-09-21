package constants

type ThresholdRange struct {
	Min float64
	Max float64
}

var DefaultThresholds = map[string]ThresholdRange{
	SensorTemperature: {Min: 15, Max: 32}, SensorHumidity: {Min: 45, Max: 80}, SensorLight: {Min: 6000, Max: 30000}, SensorCO2: {Min: 400, Max: 1500}, SensorSoil: {Min: 35, Max: 75},
}
