package conv

import (
	"fmt"
	"math"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/core/audio"
	"github.com/mokiat/lacking/game/asset/dsl"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
)

var _ SoundSource = (*mdl.Sound)(nil)

// SoundSource is an asset that provides decoded audio data, such as
// [mdl.Sound].
type SoundSource interface {
	SampleRate() int
	Frames() []audio.Frame
}

// SoundConverter converts a [SoundSource] into a [dto.SoundChunk].
type SoundConverter struct{}

var _ dsl.Converter = (*SoundConverter)(nil)

// NewSoundConverter creates a new [SoundConverter].
func NewSoundConverter() *SoundConverter {
	return &SoundConverter{}
}

// Convert appends a [dto.SoundChunk] to the target if the asset is a
// [SoundSource]. Other assets are ignored.
func (c *SoundConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(SoundSource)
	if !ok {
		return nil
	}
	chunk, err := c.CreateSoundChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(dto.SoundChunkID, chunk))
	return nil
}

// CreateSoundChunk creates a [dto.SoundChunk] from the specified source.
// It returns an error if the sample rate is not positive or does not fit
// into a uint32.
func (c *SoundConverter) CreateSoundChunk(src SoundSource) (*dto.SoundChunk, error) {
	sampleRate := src.SampleRate()
	if sampleRate <= 0 || sampleRate > math.MaxUint32 {
		return nil, fmt.Errorf("invalid sample rate %d", sampleRate)
	}
	frames := src.Frames()
	dtoFrames := make([]dto.SoundFrame, len(frames))
	for i, frame := range frames {
		dtoFrames[i] = dto.SoundFrame{
			Left:  frame.Left,
			Right: frame.Right,
		}
	}
	return &dto.SoundChunk{
		SampleRate: uint32(sampleRate),
		Frames:     dtoFrames,
	}, nil
}
