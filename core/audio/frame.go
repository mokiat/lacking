package audio

import (
	"math"

	"github.com/mokiat/gomath/sprec"
)

// Frame represents a PCM frame with left and right channel data.
type Frame struct {

	// Left is the left channel sample value.
	Left float32

	// Right is the right channel sample value.
	Right float32
}

// Resample resamples the given audio frames from one sample rate to another.
func Resample(frames []Frame, fromRate int, toRate int) []Frame {
	if (fromRate == toRate) || (len(frames) == 0) {
		return frames
	}
	if fromRate <= 0 || toRate <= 0 {
		panic("invalid sample rate")
	}
	oldLength := len(frames)

	scale := float64(toRate) / float64(fromRate)
	newLength := int(float64(len(frames))*scale + 0.5)
	if newLength <= 0 {
		return nil
	}
	if newLength == 1 {
		return []Frame{frames[0]}
	}

	result := make([]Frame, newLength)
	step := float64(oldLength-1) / float64(newLength-1)
	for i := range newLength {
		srcPosition, srcFraction := math.Modf(float64(i) * step)
		srcIndexPrev := min(int(srcPosition), oldLength-1)
		srcIndexNext := min(srcIndexPrev+1, oldLength-1)
		if srcIndexPrev == srcIndexNext {
			result[i] = frames[srcIndexPrev]
		} else {
			prev := frames[srcIndexPrev]
			next := frames[srcIndexNext]
			result[i] = Frame{
				Left:  sprec.Mix(prev.Left, next.Left, float32(srcFraction)),
				Right: sprec.Mix(prev.Right, next.Right, float32(srcFraction)),
			}
		}
	}
	return result
}
