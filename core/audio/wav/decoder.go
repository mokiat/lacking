package wav

import (
	"bytes"
	"io"

	"github.com/go-audio/wav"
	"github.com/mokiat/lacking/core/audio"
)

// Decode decodes WAV data from the provided reader and returns the decoded
// audio frames.
func Decode(in io.Reader) (audio.MediaData, error) {
	raw, err := io.ReadAll(in)
	if err != nil {
		return audio.MediaData{}, err
	}

	decoder := wav.NewDecoder(bytes.NewReader(raw))
	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return audio.MediaData{}, err
	}
	flBuffer := buffer.AsFloat32Buffer()

	length := flBuffer.NumFrames()
	frames := make([]audio.Frame, length)
	if buffer.Format.NumChannels > 0 {
		if buffer.Format.NumChannels == 1 {
			for i := range length {
				value := flBuffer.Data[i]
				frames[i] = audio.Frame{
					Left:  value,
					Right: value,
				}
			}
		} else {
			offset := 0
			for i := range length {
				frames[i] = audio.Frame{
					Left:  flBuffer.Data[offset+0],
					Right: flBuffer.Data[offset+1],
				}
				offset += buffer.Format.NumChannels
			}
		}
	}

	return audio.MediaData{
		Frames:     frames,
		SampleRate: buffer.Format.SampleRate,
	}, nil
}
