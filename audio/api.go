package audio

// API provides access to a low-level audio manipulation and playback.
type API interface {

	// CreateMedia creates a new Media object based on the specified info.
	CreateMedia(info MediaInfo) Media

	// Play plays the specified media as soon as possible.
	//
	// TODO: REMOVE THIS!!!!
	Play(media Media, info PlayInfo) Playback

	// CreatePlayback creates a new playback node for the specified media.
	CreatePlayback(media Media, loop bool) PlaybackNode

	// CreateOscillator creates a new oscillator node.
	CreateOscillator() OscillatorNode

	// CreateGain creates a new gain node.
	CreateGain() GainNode

	// CreatePan creates a new pan node.
	CreatePan() PanNode

	// Chain connects the specified nodes in sequence. This is a convenience
	// function that uses the Connect method of the API. Beware that it may
	// incur allocations due to variadic parameters.
	Chain(nodes ...Node)

	// Connect connects the source node to the target node.
	Connect(source, target Node)

	// Disconnect disconnects the source node from the target node.
	Disconnect(source, target Node)

	// Output returns the output audio node.
	Output() Node
}
