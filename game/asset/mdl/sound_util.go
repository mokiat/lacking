package mdl

import (
	"fmt"
	"io"

	"github.com/mokiat/lacking/core/audio"
	_ "github.com/mokiat/lacking/core/audio/mp3"
	_ "github.com/mokiat/lacking/core/audio/wav"
)

// ParseSound reads and decodes audio data from the specified reader into
// a new [Sound].
func ParseSound(in io.Reader) (*Sound, error) {
	data, _, err := audio.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio: %w", err)
	}

	sound := NewSound()
	sound.SetSampleRate(data.SampleRate)
	sound.SetFrames(data.Frames)
	return sound, nil
}
