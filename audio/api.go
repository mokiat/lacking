package audio

// API provides access to a low-level audio manipulation and playback.
//
// All methods must be called from the UI thread.
type API interface {

	// SampleRate returns the audio sample rate used by the API (i.e. how
	// many samples there are in a single second).
	SampleRate() int

	// CreateMedia creates a new Media object from the specified samples. This
	// function assumes that the samples match the API's sample rate.
	//
	// Keep in mind that the implementation may keep a reference to the provided
	// samples slice, so it should not be modified after being passed to this
	// method.
	CreateMedia(samples []Sample) Media

	// ParseMedia creates a new Media object based on the specified raw data info
	// by parsing it according to its format.
	ParseMedia(info MediaInfo) Media

	// Output returns the output audio node.
	Output() Node

	// SpatialListener returns the spatial listener used for 3D audio.
	SpatialListener() SpatialListener

	// CreatePlaybackNode creates a new playback node for the specified media.
	//
	// It is safe to delete the media after creating the playback node.
	CreatePlaybackNode(media Media, loop bool) PlaybackNode

	// CreateOscillatorNode creates a new oscillator node.
	CreateOscillatorNode() OscillatorNode

	// CreateGainNode creates a new gain node.
	CreateGainNode() GainNode

	// CreatePan creates a new pan node.
	CreatePanNode() PanNode

	// CreateSpatialNode creates a new spatial audio node.
	CreateSpatialNode() SpatialNode

	// CreateHighPassNode creates a new high-pass filter node.
	CreateHighPassNode() HighPassNode

	// CreateLowPassNode creates a new low-pass filter node.
	CreateLowPassNode() LowPassNode

	// CreateDelayNode creates a new delay node.
	CreateDelayNode() DelayNode

	// CreateReverbNode creates a new reverb node.
	CreateReverbNode() ReverbNode

	// CreateCompressorNode creates a new compressor node.
	CreateCompressorNode() CompressorNode

	// CreateConnectorNode creates a new connector node. It is a pass-through
	// node that forwards its input signal unchanged, useful as a named
	// connection point in a larger node graph.
	CreateConnectorNode() ConnectorNode

	// Chain connects the specified nodes in sequence. This is a convenience
	// function that uses [API.Connect]. Beware that it may incur allocations
	// due to variadic parameters.
	Chain(nodes ...Node)

	// Connect connects the source node to the target node.
	//
	// The audio signal from the source node will be added to the audio
	// signal from any other nodes that are already connected to the target node.
	Connect(source, target Node)

	// Disconnect disconnects the source node from the target node.
	Disconnect(source, target Node)

	// Play plays the specified media as soon as possible.
	//
	// TODO: REMOVE THIS!!!!
	Play(media Media, info PlayInfo) Playback
}
