package audio

import "time"

// MediaDataType indicates the type of media data contained in a data block.
type MediaDataType int8

const (
	MediaDataTypeAuto MediaDataType = iota
	MediaDataTypeWAV
	MediaDataTypeMP3
)

// MediaInfo contains the necessary information to create a Media.
type MediaInfo struct {
	Data     []byte
	DataType MediaDataType
}

// Media represents a playable audio sequence.
type Media interface {

	// Length returns the duration of the media.
	Length() time.Duration

	// Delete frees any resources used by this media.
	Delete()
}
