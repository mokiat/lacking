package audio

import (
	"time"

	"github.com/mokiat/gomath/sprec"
)

// NewNopAPI returns an API that does nothing.
func NewNopAPI() API {
	return &nopAPI{
		listener: &nopListener{
			rotation: sprec.IdentityQuat(),
		},
	}
}

type nopAPI struct {
	listener *nopListener
}

func (a *nopAPI) SampleRate() int {
	return 44100
}

func (a *nopAPI) CreateMedia(samples []Sample) Media {
	return NewNopMedia()
}

func (a *nopAPI) ParseMedia(info MediaInfo) Media {
	return NewNopMedia()
}

func (a *nopAPI) Output() Node {
	return nil
}

func (a *nopAPI) SpatialListener() SpatialListener {
	return a.listener
}

func (a *nopAPI) CreatePlaybackNode(media Media, loop bool) PlaybackNode {
	return &nopPlaybackNode{}
}

func (a *nopAPI) CreateOscillatorNode() OscillatorNode {
	return &nopOscillatorNode{}
}

func (a *nopAPI) CreateGainNode() GainNode {
	return &nopGainNode{}
}

func (a *nopAPI) CreatePanNode() PanNode {
	return &nopPanNode{}
}

func (a *nopAPI) CreateSpatialNode() SpatialNode {
	return &nopSpatialNode{}
}

func (a *nopAPI) CreateHighPassNode() HighPassNode {
	return &nopHighPassNode{}
}

func (a *nopAPI) CreateLowPassNode() LowPassNode {
	return &nopLowPassNode{}
}

func (a *nopAPI) CreateDelayNode() DelayNode {
	return &nopDelayNode{}
}

func (a *nopAPI) CreateReverbNode() ReverbNode {
	return &nopReverbNode{}
}

func (a *nopAPI) CreateCompressorNode() CompressorNode {
	return &nopCompressorNode{}
}

func (a *nopAPI) CreateConnectorNode() ConnectorNode {
	return &nopUserNode{}
}

func (a *nopAPI) Chain(nodes ...Node) {}

func (a *nopAPI) Connect(source, target Node) {}

func (a *nopAPI) Disconnect(source, target Node) {}

func (a *nopAPI) Play(media Media, info PlayInfo) Playback {
	return &nopPlayback{}
}

func NewNopMedia() Media {
	return &nopMedia{}
}

type nopMedia struct{}

func (m *nopMedia) Length() time.Duration {
	return time.Millisecond
}

func (m *nopMedia) Delete() {}

type nopListener struct {
	position sprec.Vec3
	rotation sprec.Quat
}

func (l *nopListener) Position() sprec.Vec3 {
	return l.position
}

func (l *nopListener) SetPosition(position sprec.Vec3) {
	l.position = position
}

func (l *nopListener) Rotation() sprec.Quat {
	return l.rotation
}

func (l *nopListener) SetRotation(rotation sprec.Quat) {
	l.rotation = rotation
}

type nopUserNode struct{}

func (n *nopUserNode) Delete() {}

type nopPlaybackNode struct {
	nopUserNode
	loop      bool
	loopStart float32
	loopEnd   float32
}

func (n *nopPlaybackNode) Start(offset float32) {}

func (n *nopPlaybackNode) Stop() {}

func (n *nopPlaybackNode) Resume() {}

func (n *nopPlaybackNode) Pause() {}

func (n *nopPlaybackNode) IsPlaying() bool {
	return false
}

func (n *nopPlaybackNode) IsLoop() bool {
	return n.loop
}

func (n *nopPlaybackNode) SetLoop(loop bool) {
	n.loop = loop
}

func (n *nopPlaybackNode) LoopStart() float32 {
	return n.loopStart
}

func (n *nopPlaybackNode) SetLoopStart(loopStart float32) {
	n.loopStart = loopStart
}

func (n *nopPlaybackNode) LoopEnd() float32 {
	return n.loopEnd
}

func (n *nopPlaybackNode) SetLoopEnd(loopEnd float32) {
	n.loopEnd = loopEnd
}

type nopOscillatorNode struct {
	nopUserNode
	frequency float32
}

func (n *nopOscillatorNode) Frequency() float32 {
	return n.frequency
}

func (n *nopOscillatorNode) SetFrequency(frequency float32) {
	n.frequency = frequency
}

type nopGainNode struct {
	nopUserNode
	gain float32
}

func (n *nopGainNode) Gain() float32 {
	return n.gain
}

func (n *nopGainNode) SetGain(gain float32) {
	n.gain = gain
}

type nopPanNode struct {
	nopUserNode
	pan float32
}

func (n *nopPanNode) Pan() float32 {
	return n.pan
}

func (n *nopPanNode) SetPan(pan float32) {
	n.pan = pan
}

type nopSpatialNode struct {
	nopUserNode
	position sprec.Vec3
}

func (n *nopSpatialNode) Position() sprec.Vec3 {
	return n.position
}

func (n *nopSpatialNode) SetPosition(position sprec.Vec3) {
	n.position = position
}

type nopHighPassNode struct {
	nopUserNode
	cutoffFrequency float32
}

func (n *nopHighPassNode) CutoffFrequency() float32 {
	return n.cutoffFrequency
}

func (n *nopHighPassNode) SetCutoffFrequency(cutoffFrequency float32) {
	n.cutoffFrequency = cutoffFrequency
}

type nopLowPassNode struct {
	nopUserNode
	cutoffFrequency float32
}

func (n *nopLowPassNode) CutoffFrequency() float32 {
	return n.cutoffFrequency
}

func (n *nopLowPassNode) SetCutoffFrequency(cutoffFrequency float32) {
	n.cutoffFrequency = cutoffFrequency
}

type nopDelayNode struct {
	nopUserNode
	delayTime float32
}

func (n *nopDelayNode) DelayTime() float32 {
	return n.delayTime
}

func (n *nopDelayNode) SetDelayTime(delayTime float32) {
	n.delayTime = delayTime
}

type nopReverbNode struct {
	nopUserNode
	roomSize float32
}

func (n *nopReverbNode) RoomSize() float32 {
	return n.roomSize
}

func (n *nopReverbNode) SetRoomSize(roomSize float32) {
	n.roomSize = roomSize
}

type nopCompressorNode struct {
	nopUserNode
	threshold float32
}

func (n *nopCompressorNode) Threshold() float32 {
	return n.threshold
}

func (n *nopCompressorNode) SetThreshold(threshold float32) {
	n.threshold = threshold
}

type nopPlayback struct{}

func (p *nopPlayback) Stop() {}
