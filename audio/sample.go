package audio

import (
	"math"

	"github.com/mokiat/gomath/sprec"
)

// Sample represents a single audio sample with left and right channel data.
type Sample struct {

	// Left is the left channel sample value.
	Left float32

	// Right is the right channel sample value.
	Right float32
}

// SampleCount calculates the number of samples for a given duration in seconds
// given the used sample rate.
func SampleCount(seconds float32, sampleRate int) int {
	return int(float32(sampleRate) * seconds)
}

// Seconds calculates the duration in seconds for a given number of samples
// and sample rate.
func Seconds(sampleCount, sampleRate int) float32 {
	return float32(sampleCount) / float32(sampleRate)
}

// Resample resamples the given audio samples from one sample rate to another.
func Resample(samples []Sample, fromRate int, toRate int) []Sample {
	if (fromRate == toRate) || (len(samples) == 0) {
		return samples
	}
	if fromRate <= 0 || toRate <= 0 {
		panic("invalid sample rate")
	}
	oldLength := len(samples)

	scale := float64(toRate) / float64(fromRate)
	newLength := int(float64(len(samples))*scale + 0.5)
	if newLength <= 0 {
		return nil
	}
	if newLength == 1 {
		return []Sample{samples[0]}
	}

	result := make([]Sample, newLength)
	step := float64(oldLength-1) / float64(newLength-1)
	for i := range newLength {
		srcPosition, srcFraction := math.Modf(float64(i) * step)
		srcIndexPrev := min(int(srcPosition), oldLength-1)
		srcIndexNext := min(srcIndexPrev+1, oldLength-1)
		if srcIndexPrev == srcIndexNext {
			result[i] = samples[srcIndexPrev]
		} else {
			prev := samples[srcIndexPrev]
			next := samples[srcIndexNext]
			result[i] = Sample{
				Left:  sprec.Mix(prev.Left, next.Left, float32(srcFraction)),
				Right: sprec.Mix(prev.Right, next.Right, float32(srcFraction)),
			}
		}
	}
	return result
}
