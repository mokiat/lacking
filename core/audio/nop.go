package audio

import "github.com/mokiat/gomath/sprec"

type nopAPI struct {
	masterBus *nopMasterBus
	listener  *nopSpatialListener
}

var _ API = (*nopAPI)(nil)

// NewNopAPI returns a no-op implementation of the API interface.
func NewNopAPI() API {
	return &nopAPI{
		masterBus: &nopMasterBus{
			compression: &nopCompression{},
		},
		listener: &nopSpatialListener{},
	}
}

func (a *nopAPI) CreateMedia(data MediaData) Media {
	return &nopMedia{}
}

func (a *nopAPI) CreateBus(settings BusSettings) Bus {
	b := &nopBus{}
	if settings.UseCompression {
		b.compression = &nopCompression{}
	}
	if settings.UseReverb {
		b.reverb = &nopReverb{}
	}
	return b
}

func (a *nopAPI) CreatePlayback(bus Bus, media Media, settings PlaybackSettings) Playback {
	return newNopPlayback(settings)
}

func (a *nopAPI) CreateSpatialPlayback(bus Bus, media Media, settings PlaybackSettings) SpatialPlayback {
	return newNopSpatialPlayback(settings)
}

func (a *nopAPI) MasterBus() MasterBus {
	return a.masterBus
}

func (a *nopAPI) SpatialListener() SpatialListener {
	return a.listener
}

var _ Media = (*nopMedia)(nil)

type nopMedia struct{}

func (m *nopMedia) Length() float64 {
	return 0.0
}

func (m *nopMedia) Release() {}

var _ MasterBus = (*nopMasterBus)(nil)

type nopMasterBus struct {
	gain        float32
	compression *nopCompression
}

func (b *nopMasterBus) Gain() float32 {
	return b.gain
}

func (b *nopMasterBus) SetGain(gain float32) {
	b.gain = gain
}

func (b *nopMasterBus) Compression() Compression {
	return b.compression
}

var _ Bus = (*nopBus)(nil)

type nopBus struct {
	gain        float32
	compression *nopCompression
	reverb      *nopReverb
}

func (b *nopBus) Gain() float32 {
	return b.gain
}

func (b *nopBus) SetGain(gain float32) {
	b.gain = gain
}

func (b *nopBus) Compression() Compression {
	if b.compression == nil {
		return nil
	}
	return b.compression
}

func (b *nopBus) Reverb() Reverb {
	if b.reverb == nil {
		return nil
	}
	return b.reverb
}

func (b *nopBus) Pause() {}

func (b *nopBus) Resume() {}

func (b *nopBus) Release() {}

var _ Reverb = (*nopReverb)(nil)

type nopReverb struct {
	roomSize float32
	damping  float32
	dry      float32
	wet      float32
}

func (r *nopReverb) RoomSize() float32 {
	return r.roomSize
}

func (r *nopReverb) SetRoomSize(size float32) {
	r.roomSize = size
}

func (r *nopReverb) Damping() float32 {
	return r.damping
}

func (r *nopReverb) SetDamping(damping float32) {
	r.damping = damping
}

func (r *nopReverb) Dry() float32 {
	return r.dry
}

func (r *nopReverb) SetDry(dry float32) {
	r.dry = dry
}

func (r *nopReverb) Wet() float32 {
	return r.wet
}

func (r *nopReverb) SetWet(wet float32) {
	r.wet = wet
}

var _ Compression = (*nopCompression)(nil)

type nopCompression struct {
	attack    float32
	release   float32
	ratio     float32
	knee      float32
	threshold float32
}

func (c *nopCompression) Attack() float32 {
	return c.attack
}

func (c *nopCompression) SetAttack(attack float32) {
	c.attack = attack
}

func (c *nopCompression) Release() float32 {
	return c.release
}

func (c *nopCompression) SetRelease(release float32) {
	c.release = release
}

func (c *nopCompression) Ratio() float32 {
	return c.ratio
}

func (c *nopCompression) SetRatio(ratio float32) {
	c.ratio = ratio
}

func (c *nopCompression) Knee() float32 {
	return c.knee
}

func (c *nopCompression) SetKnee(knee float32) {
	c.knee = knee
}

func (c *nopCompression) Threshold() float32 {
	return c.threshold
}

func (c *nopCompression) SetThreshold(threshold float32) {
	c.threshold = threshold
}

var _ FrequencyFilter = (*nopFrequencyFilter)(nil)

type nopFrequencyFilter struct {
	frequency float32
}

func (f *nopFrequencyFilter) Frequency() float32 {
	return f.frequency
}

func (f *nopFrequencyFilter) SetFrequency(frequency float32) {
	f.frequency = frequency
}

var _ Playback = (*nopPlayback)(nil)

type nopPlayback struct {
	playing        bool
	looping        bool
	loopStart      float64
	loopEnd        float64
	playbackRate   float32
	gain           float32
	lowPassFilter  *nopFrequencyFilter
	highPassFilter *nopFrequencyFilter
	onFinished     func()
}

func newNopPlayback(settings PlaybackSettings) *nopPlayback {
	p := &nopPlayback{}
	if settings.UseLowPassFilter {
		p.lowPassFilter = &nopFrequencyFilter{}
	}
	if settings.UseHighPassFilter {
		p.highPassFilter = &nopFrequencyFilter{}
	}
	return p
}

func (p *nopPlayback) Start(at float64) {
	p.playing = true
}

func (p *nopPlayback) Stop() {
	p.playing = false
}

func (p *nopPlayback) Pause() {
	p.playing = false
}

func (p *nopPlayback) Resume() {}

func (p *nopPlayback) Looping() bool {
	return p.looping
}

func (p *nopPlayback) SetLooping(loop bool) {
	p.looping = loop
}

func (p *nopPlayback) LoopStart() float64 {
	return p.loopStart
}

func (p *nopPlayback) SetLoopStart(s float64) {
	p.loopStart = s
}

func (p *nopPlayback) LoopEnd() float64 {
	return p.loopEnd
}

func (p *nopPlayback) SetLoopEnd(e float64) {
	p.loopEnd = e
}

func (p *nopPlayback) Playing() bool {
	return p.playing
}

func (p *nopPlayback) PlaybackRate() float32 {
	return p.playbackRate
}

func (p *nopPlayback) SetPlaybackRate(r float32) {
	p.playbackRate = r
}

func (p *nopPlayback) Gain() float32 {
	return p.gain
}

func (p *nopPlayback) SetGain(gain float32) {
	p.gain = gain
}

func (p *nopPlayback) LowPassFilter() FrequencyFilter {
	if p.lowPassFilter == nil {
		return nil
	}
	return p.lowPassFilter
}
func (p *nopPlayback) HighPassFilter() FrequencyFilter {
	if p.highPassFilter == nil {
		return nil
	}
	return p.highPassFilter
}

func (p *nopPlayback) SetOnFinished(fn func()) {
	p.onFinished = fn
}

func (p *nopPlayback) Release() {}

var _ SpatialPlayback = (*nopSpatialPlayback)(nil)

type nopSpatialPlayback struct {
	*nopPlayback
	position       sprec.Vec3
	rotation       sprec.Quat
	innerConeAngle sprec.Angle
	outerConeAngle sprec.Angle
	outerConeGain  float32
}

func newNopSpatialPlayback(settings PlaybackSettings) *nopSpatialPlayback {
	return &nopSpatialPlayback{
		nopPlayback: newNopPlayback(settings),
	}
}

func (p *nopSpatialPlayback) Position() sprec.Vec3 {
	return p.position
}

func (p *nopSpatialPlayback) SetPosition(pos sprec.Vec3) {
	p.position = pos
}

func (p *nopSpatialPlayback) Rotation() sprec.Quat {
	return p.rotation
}

func (p *nopSpatialPlayback) SetRotation(rot sprec.Quat) {
	p.rotation = rot
}

func (p *nopSpatialPlayback) InnerConeAngle() sprec.Angle {
	return p.innerConeAngle
}

func (p *nopSpatialPlayback) SetInnerConeAngle(a sprec.Angle) {
	p.innerConeAngle = a
}

func (p *nopSpatialPlayback) OuterConeAngle() sprec.Angle {
	return p.outerConeAngle
}

func (p *nopSpatialPlayback) SetOuterConeAngle(a sprec.Angle) {
	p.outerConeAngle = a
}

func (p *nopSpatialPlayback) OuterConeGain() float32 {
	return p.outerConeGain
}

func (p *nopSpatialPlayback) SetOuterConeGain(gain float32) {
	p.outerConeGain = gain
}

var _ SpatialListener = (*nopSpatialListener)(nil)

type nopSpatialListener struct {
	position sprec.Vec3
	rotation sprec.Quat
}

func (l *nopSpatialListener) Position() sprec.Vec3 {
	return l.position
}

func (l *nopSpatialListener) SetPosition(pos sprec.Vec3) {
	l.position = pos
}

func (l *nopSpatialListener) Rotation() sprec.Quat {
	return l.rotation
}

func (l *nopSpatialListener) SetRotation(rot sprec.Quat) {
	l.rotation = rot
}
