package audio

import "math"

// DBToGain converts a decibel value to a gain value.
func DBToGain(db float32) float32 {
	return float32(math.Pow(10.0, float64(db/20.0)))
}

// GainToDB converts a gain value to a decibel value.
func GainToDB(gain float32) float32 {
	return float32(20.0 * math.Log10(float64(max(0.0, gain))))
}

// SampleCount calculates the number of samples for a given duration in seconds
// given the used sample rate.
func SampleCount(seconds float32, sampleRate int) int {
	return int(float64(sampleRate) * float64(seconds))
}

// Seconds calculates the duration in seconds for a given number of samples
// and sample rate.
func Seconds(sampleCount, sampleRate int) float32 {
	return float32(sampleCount) / float32(sampleRate)
}
