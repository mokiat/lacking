package audio

// API provides access to a low-level audio manipulation and playback.
type API interface {

	// CreateMedia creates a new Media object based on the specified info.
	CreateMedia(info MediaInfo) Media

	// Play plays the specified media as soon as possible.
	Play(media Media, info PlayInfo) Playback
}
