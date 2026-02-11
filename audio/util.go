package audio

import "math"

// DBToGain converts a decibel value to a gain value.
func DBToGain(db float32) float32 {
	return float32(math.Pow(10.0, float64(db/20.0)))
}

// GainToDB converts a gain value to a decibel value.
func GainToDB(gain float32) float32 {
	return float32(20.0 * math.Log10(float64(gain)))
}
