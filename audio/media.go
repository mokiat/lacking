package audio

import "time"

// MediaDataType indicates the type of media data contained in a data block.
type MediaDataType int8

const (
	// MediaDataTypeAuto indicates that the media data type should be
	// automatically detected based on the data.
	MediaDataTypeAuto MediaDataType = iota

	// MediaDataTypeWAV indicates that the media data is in WAV format.
	MediaDataTypeWAV

	// MediaDataTypeMP3 indicates that the media data is in MP3 format.
	MediaDataTypeMP3
)

// MediaInfo contains the necessary information to create a Media.
type MediaInfo struct {

	// Data is the raw media data.
	Data []byte

	// DataType indicates the type of media data.
	DataType MediaDataType
}

// Media represents a playable audio sequence.
type Media interface {

	// Length returns the duration of the media.
	Length() time.Duration

	// Delete frees any resources used by this media.
	Delete()
}
