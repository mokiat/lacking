package dto

// Sound represents a sound asset, such as a sound effect or a piece of
// music.
type Sound struct {
	SoundChunkHolder
}

// SoundChunkID is the chunk identifier under which a [SoundChunk] is stored.
const SoundChunkID = "lacking:sound"

// SoundChunkHolder holds the optional [SoundChunk] of an asset.
type SoundChunkHolder struct {

	// SoundChunk contains the audio data of the sound. It is nil when the
	// asset has no sound chunk.
	SoundChunk *SoundChunk `chunk:"lacking:sound"`
}

// SoundChunk contains the decoded audio data of a sound.
type SoundChunk struct {

	// SampleRate is the number of frames per second at which the sound is
	// meant to be played back.
	SampleRate uint32

	// Frames contains the decoded stereo PCM frames of the sound.
	Frames []SoundFrame
}

// SoundFrame is a single stereo PCM frame of a sound.
type SoundFrame struct {

	// Left is the sample value of the left channel.
	Left float32

	// Right is the sample value of the right channel.
	Right float32
}
