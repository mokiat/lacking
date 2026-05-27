package mp3

import (
	"io"
	"math"

	"github.com/hajimehoshi/go-mp3"
	"github.com/mokiat/gblob"
	"github.com/mokiat/lacking/audio"
)

func init() {
	magics := []string{
		"ID3",
	}
	audio.RegisterFormat("mp3", magics, Decode)
}

// Decode decodes MP3 data from the provided reader and returns the decoded
// audio frames.
func Decode(in io.Reader) (audio.MediaData, error) {
	decoder, err := mp3.NewDecoder(in)
	if err != nil {
		return audio.MediaData{}, err
	}

	// TODO: There must be a faster and cheaper way to do this.
	// 1. The decoder has overhead from being able to Seek. If an implementation
	// is used/written that doesn't support seeking, some overhead can be avoided.
	// 2. It might be possible to decode directly via a LittleEndian decoder
	// without having to read the entire data into memory.
	data, err := io.ReadAll(decoder)
	if err != nil {
		return audio.MediaData{}, err
	}
	buffer := gblob.LittleEndianBlock(data)

	length := len(data) / 4
	frames := make([]audio.Frame, length)
	for i := range length {
		leftInt16 := buffer.Int16(i*4 + 0)
		rightInt16 := buffer.Int16(i*4 + 2)
		frames[i] = audio.Frame{
			Left:  int16ToFloat32(leftInt16),
			Right: int16ToFloat32(rightInt16),
		}
	}

	return audio.MediaData{
		Frames:     frames,
		SampleRate: decoder.SampleRate(),
	}, nil
}

func int16ToFloat32(value int16) float32 {
	if value >= 0 {
		return float32(value) / float32(math.MaxInt16)
	} else {
		return -float32(value) / float32(math.MinInt16)
	}
}
