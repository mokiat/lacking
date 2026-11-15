package sound_test

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

// fakeAPI is an in-memory implementation of audio.API that records what
// is requested of it, so that specs can verify how a sound scene drives
// the audio layer.
type fakeAPI struct {
	buses     []*fakeBus
	playbacks []*fakePlayback
	listener  *fakeListener
}

var _ audio.API = (*fakeAPI)(nil)

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		listener: &fakeListener{
			rotation: sprec.IdentityQuat(),
		},
	}
}

func (a *fakeAPI) CreateMedia(data audio.MediaData) audio.Media {
	return &fakeMedia{}
}

func (a *fakeAPI) CreateBus(settings audio.BusSettings) audio.Bus {
	bus := &fakeBus{
		gain: 1.0,
	}
	a.buses = append(a.buses, bus)
	return bus
}

func (a *fakeAPI) CreatePlayback(bus audio.Bus, media audio.Media, settings audio.PlaybackSettings) audio.Playback {
	return a.createPlayback(bus, media, false)
}

func (a *fakeAPI) CreateSpatialPlayback(bus audio.Bus, media audio.Media, settings audio.PlaybackSettings) audio.SpatialPlayback {
	return a.createPlayback(bus, media, true)
}

func (a *fakeAPI) MasterBus() audio.MasterBus {
	return nil
}

func (a *fakeAPI) SpatialListener() audio.SpatialListener {
	return a.listener
}

func (a *fakeAPI) createPlayback(bus audio.Bus, media audio.Media, spatial bool) *fakePlayback {
	playback := &fakePlayback{
		bus:     bus.(*fakeBus),
		media:   media.(*fakeMedia),
		spatial: spatial,
		rate:    1.0,
		gain:    1.0,
	}
	a.playbacks = append(a.playbacks, playback)
	return playback
}

// lastPlayback returns the playback that was created most recently.
func (a *fakeAPI) lastPlayback() *fakePlayback {
	return a.playbacks[len(a.playbacks)-1]
}

// playbackOf returns the only playback that was created for the specified
// media.
func (a *fakeAPI) playbackOf(media audio.Media) *fakePlayback {
	var result *fakePlayback
	for _, playback := range a.playbacks {
		if audio.Media(playback.media) == media {
			if result != nil {
				panic("more than one playback for media")
			}
			result = playback
		}
	}
	if result == nil {
		panic("no playback for media")
	}
	return result
}

type fakeMedia struct {
	length float64
}

var _ audio.Media = (*fakeMedia)(nil)

// newFakeMedia creates a media with the specified length in seconds.
func newFakeMedia(length float64) *fakeMedia {
	return &fakeMedia{
		length: length,
	}
}

func (m *fakeMedia) Length() float64 {
	return m.length
}

func (m *fakeMedia) Release() {}

type fakeBus struct {
	gain     float32
	paused   bool
	released bool
}

var _ audio.Bus = (*fakeBus)(nil)

func (b *fakeBus) Gain() float32 {
	return b.gain
}

func (b *fakeBus) SetGain(gain float32) {
	b.gain = gain
}

func (b *fakeBus) Compression() audio.Compression {
	return nil
}

func (b *fakeBus) Reverb() audio.Reverb {
	return nil
}

func (b *fakeBus) Pause() {
	b.paused = true
}

func (b *fakeBus) Resume() {
	b.paused = false
}

func (b *fakeBus) Release() {
	b.released = true
}

type fakePlayback struct {
	bus     *fakeBus
	media   *fakeMedia
	spatial bool

	playing    bool
	startCount int
	startedAt  float64
	released   bool
	looping    bool
	loopStart  float64
	loopEnd    float64
	rate       float32
	gain       float32
	onFinished func()

	position       sprec.Vec3
	rotation       sprec.Quat
	innerConeAngle sprec.Angle
	outerConeAngle sprec.Angle
	outerConeGain  float32
}

var _ audio.SpatialPlayback = (*fakePlayback)(nil)

// finish simulates the playback reaching the end of its media.
func (p *fakePlayback) finish() {
	p.playing = false
	if p.onFinished != nil {
		p.onFinished()
	}
}

func (p *fakePlayback) Start(at float64) {
	p.playing = true
	p.startCount++
	p.startedAt = at
}

func (p *fakePlayback) Stop() {
	p.playing = false
}

func (p *fakePlayback) Pause() {}

func (p *fakePlayback) Resume() {}

func (p *fakePlayback) Looping() bool {
	return p.looping
}

func (p *fakePlayback) SetLooping(loop bool) {
	p.looping = loop
}

func (p *fakePlayback) LoopStart() float64 {
	return p.loopStart
}

func (p *fakePlayback) SetLoopStart(loopStart float64) {
	p.loopStart = loopStart
}

func (p *fakePlayback) LoopEnd() float64 {
	return p.loopEnd
}

func (p *fakePlayback) SetLoopEnd(loopEnd float64) {
	p.loopEnd = loopEnd
}

func (p *fakePlayback) Playing() bool {
	return p.playing
}

func (p *fakePlayback) PlaybackRate() float32 {
	return p.rate
}

func (p *fakePlayback) SetPlaybackRate(rate float32) {
	p.rate = rate
}

func (p *fakePlayback) Gain() float32 {
	return p.gain
}

func (p *fakePlayback) SetGain(gain float32) {
	p.gain = gain
}

func (p *fakePlayback) LowPassFilter() audio.FrequencyFilter {
	return nil
}

func (p *fakePlayback) HighPassFilter() audio.FrequencyFilter {
	return nil
}

func (p *fakePlayback) SetOnFinished(onFinished func()) {
	p.onFinished = onFinished
}

func (p *fakePlayback) Release() {
	p.released = true
}

func (p *fakePlayback) Position() sprec.Vec3 {
	return p.position
}

func (p *fakePlayback) SetPosition(position sprec.Vec3) {
	p.position = position
}

func (p *fakePlayback) Rotation() sprec.Quat {
	return p.rotation
}

func (p *fakePlayback) SetRotation(rotation sprec.Quat) {
	p.rotation = rotation
}

func (p *fakePlayback) InnerConeAngle() sprec.Angle {
	return p.innerConeAngle
}

func (p *fakePlayback) SetInnerConeAngle(angle sprec.Angle) {
	p.innerConeAngle = angle
}

func (p *fakePlayback) OuterConeAngle() sprec.Angle {
	return p.outerConeAngle
}

func (p *fakePlayback) SetOuterConeAngle(angle sprec.Angle) {
	p.outerConeAngle = angle
}

func (p *fakePlayback) OuterConeGain() float32 {
	return p.outerConeGain
}

func (p *fakePlayback) SetOuterConeGain(gain float32) {
	p.outerConeGain = gain
}

type fakeListener struct {
	position sprec.Vec3
	rotation sprec.Quat
}

var _ audio.SpatialListener = (*fakeListener)(nil)

func (l *fakeListener) Position() sprec.Vec3 {
	return l.position
}

func (l *fakeListener) SetPosition(position sprec.Vec3) {
	l.position = position
}

func (l *fakeListener) Rotation() sprec.Quat {
	return l.rotation
}

func (l *fakeListener) SetRotation(rotation sprec.Quat) {
	l.rotation = rotation
}
