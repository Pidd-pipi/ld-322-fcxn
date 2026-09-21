package constants

type ThresholdRange struct {
	Min float64
	Max float64
}

var DefaultThresholds = map[string]ThresholdRange{
	SensorTemperature: {Min: 15, Max: 32}, SensorHumidity: {Min: 45, Max: 80}, SensorLight: {Min: 6000, Max: 30000}, SensorCO2: {Min: 400, Max: 1500}, SensorSoil: {Min: 35, Max: 75},
}

// CriticalLevelFactor turns a threshold boundary into the critical-level
// boundary: readings beyond 80%/120% of the allowed range are critical.
const CriticalLevelFactor = 0.8

const (
	AlertLevelWarning  = "warning"
	AlertLevelCritical = "critical"
)
