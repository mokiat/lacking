package sound

import (
	"slices"
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
	"github.com/mokiat/lacking/util/observer"
)

const nilIndex = int32(-1)

const (
	// minDopplerRatio and maxDopplerRatio bound the playback rate change
	// that the Doppler effect is allowed to produce.
	minDopplerRatio = 0.5
	maxDopplerRatio = 2.0

	// minDopplerDistance is the distance between an emitter and the
	// listener below which the direction between the two is considered
	// undefined and the Doppler effect is not applied.
	minDopplerDistance = 0.001
)

// Scene represents a self-contained sound stage that holds listeners,
// emitters and everything that is playing: music, speech, effects and
// ambiences.
//
// A Scene and everything obtained from it must only be used from the
// thread that the [audio.API] is bound to. A Scene is not safe for
// concurrent use.
type Scene struct {
	api audio.API

	musicBus    busState
	speechBus   busState
	effectBus   busState
	ambienceBus busState

	freeListenerIndices   *ds.Stack[int32]
	freeEmitterIndices    *ds.Stack[int32]
	freeMusicTrackIndices *ds.Stack[int32]
	freeEffectIndices     *ds.Stack[int32]
	freeEffectPoolIndices *ds.Stack[int32]
	freeSpeechIndices     *ds.Stack[int32]
	freeAmbienceIndices   *ds.Stack[int32]

	listeners   []listenerState
	emitters    []emitterState
	musicTracks []musicTrackState
	effects     []effectState
	effectPools []effectPoolState
	speechLines []speechState
	ambiences   []ambienceState

	activeListenerID ListenerID

	speedOfSound  float64
	dopplerFactor float64

	musicQueue          []int32
	musicFadingIndices  []int32
	musicCurrentIndex   int32
	musicCrossfade      time.Duration
	musicEndingLeadTime time.Duration
	musicSkipRequested  bool
	musicSkipFade       time.Duration
	musicFadeIn         time.Duration

	musicTrackStartedSubscriptions  *observer.SubscriptionSet[MusicTrackStartedCallback]
	musicTrackEndingSubscriptions   *observer.SubscriptionSet[MusicTrackEndingCallback]
	musicTrackFinishedSubscriptions *observer.SubscriptionSet[MusicTrackFinishedCallback]
	effectFinishedSubscriptions     *observer.SubscriptionSet[EffectFinishedCallback]
	speechFinishedSubscriptions     *observer.SubscriptionSet[SpeechFinishedCallback]
}

// NewScene creates a new, empty [Scene] that plays through the specified
// audio API. See the individual setters for the default values.
//
// The scene takes over the spatial listener of the audio API, so only one
// Scene should be updated at a time per audio API.
func NewScene(api audio.API) *Scene {
	return &Scene{
		api: api,

		musicBus:    newBusState(api),
		speechBus:   newBusState(api),
		effectBus:   newBusState(api),
		ambienceBus: newBusState(api),

		freeListenerIndices:   ds.EmptyStack[int32](),
		freeEmitterIndices:    ds.EmptyStack[int32](),
		freeMusicTrackIndices: ds.EmptyStack[int32](),
		freeEffectIndices:     ds.EmptyStack[int32](),
		freeEffectPoolIndices: ds.EmptyStack[int32](),
		freeSpeechIndices:     ds.EmptyStack[int32](),
		freeAmbienceIndices:   ds.EmptyStack[int32](),

		listeners:   make([]listenerState, 0),
		emitters:    make([]emitterState, 0),
		musicTracks: make([]musicTrackState, 0),
		effects:     make([]effectState, 0),
		effectPools: make([]effectPoolState, 0),
		speechLines: make([]speechState, 0),
		ambiences:   make([]ambienceState, 0),

		activeListenerID: NilListenerID,

		speedOfSound:  343.0,
		dopplerFactor: 1.0,

		musicQueue:          make([]int32, 0),
		musicFadingIndices:  make([]int32, 0),
		musicCurrentIndex:   nilIndex,
		musicCrossfade:      0,
		musicEndingLeadTime: 5 * time.Second,

		musicTrackStartedSubscriptions:  observer.NewSubscriptionSet[MusicTrackStartedCallback](),
		musicTrackEndingSubscriptions:   observer.NewSubscriptionSet[MusicTrackEndingCallback](),
		musicTrackFinishedSubscriptions: observer.NewSubscriptionSet[MusicTrackFinishedCallback](),
		effectFinishedSubscriptions:     observer.NewSubscriptionSet[EffectFinishedCallback](),
		speechFinishedSubscriptions:     observer.NewSubscriptionSet[SpeechFinishedCallback](),
	}
}

// Listeners returns a [ListenerView] through which the listeners of this
// scene can be created and managed.
func (s *Scene) Listeners() ListenerView {
	return ListenerView{
		scene: s,
	}
}

// Emitters returns an [EmitterView] through which the emitters of this
// scene can be created and managed.
func (s *Scene) Emitters() EmitterView {
	return EmitterView{
		scene: s,
	}
}

// Music returns a [MusicView] through which the music of this scene can be
// queued and controlled.
func (s *Scene) Music() MusicView {
	return MusicView{
		scene: s,
	}
}

// Speech returns a [SpeechView] through which speech lines can be played
// in this scene.
func (s *Scene) Speech() SpeechView {
	return SpeechView{
		scene: s,
	}
}

// Effects returns an [EffectView] through which individual sound effects
// can be played in this scene.
func (s *Scene) Effects() EffectView {
	return EffectView{
		scene: s,
	}
}

// EffectPools returns an [EffectPoolView] through which pools of
// frequently played sound effects can be created and managed.
func (s *Scene) EffectPools() EffectPoolView {
	return EffectPoolView{
		scene: s,
	}
}

// Ambience returns an [AmbienceView] through which looping background
// sounds can be played in this scene.
func (s *Scene) Ambience() AmbienceView {
	return AmbienceView{
		scene: s,
	}
}

// ActiveListener returns the ID of the listener from which the scene is
// heard, or [NilListenerID] if there is none.
func (s *Scene) ActiveListener() ListenerID {
	return s.activeListenerID
}

// SetActiveListener changes the listener from which the scene is heard.
// Passing [NilListenerID] leaves the scene without an active listener, in
// which case spatial sounds are heard as if from the world origin, facing
// along the default orientation.
//
// It panics if the ID is neither NilListenerID nor a valid listener.
func (s *Scene) SetActiveListener(id ListenerID) {
	if id != NilListenerID {
		s.Listeners().resolve(id, true)
	}
	s.activeListenerID = id
}

// SpeedOfSound returns the speed at which sound propagates, in units per
// second, which is used to calculate the Doppler effect.
//
// Defaults to 343, which is the speed of sound in air in meters per
// second.
func (s *Scene) SpeedOfSound() float64 {
	return s.speedOfSound
}

// SetSpeedOfSound changes the speed at which sound propagates, in units
// per second. The value must be positive.
func (s *Scene) SetSpeedOfSound(speed float64) {
	s.speedOfSound = speed
}

// DopplerFactor returns the amount by which the Doppler effect is
// exaggerated.
//
// Defaults to 1, which is physically accurate.
func (s *Scene) DopplerFactor() float64 {
	return s.dopplerFactor
}

// SetDopplerFactor changes the amount by which the Doppler effect is
// exaggerated. Values above 1 make the effect more pronounced, values
// below 1 make it more subtle and a value of 0 disables it for the whole
// scene. Negative values are clamped to 0.
func (s *Scene) SetDopplerFactor(factor float64) {
	s.dopplerFactor = max(0.0, factor)
}

// Update advances the scene by elapsedTime. It should be called once per
// frame.
//
// It applies the current placement of the active listener and of the
// emitters to the sounds that are playing, progresses fades and the music
// queue, and notifies subscribers of sounds that have finished playing
// since the previous call.
func (s *Scene) Update(elapsedTime time.Duration) {
	elapsedSeconds := elapsedTime.Seconds()
	s.updateListener()
	s.updateEffects()
	s.updatePooledEffects()
	s.updateSpeech()
	s.updateAmbiences(elapsedSeconds)
	s.updateMusic(elapsedSeconds)
}

// Delete stops everything that is playing in this scene and releases the
// audio resources that the scene has allocated. No subscribers are
// notified. The scene must not be used afterwards.
//
// The media that was played through the scene is not released, since it is
// owned by the caller.
func (s *Scene) Delete() {
	for i := range s.effects {
		if effect := &s.effects[i]; effect.isValid() {
			effect.playback.stop()
			effect.playback.release()
		}
	}
	for i := range s.effectPools {
		if pool := &s.effectPools[i]; pool.isValid() {
			for j := range pool.playbacks {
				pool.playbacks[j].playback.stop()
				pool.playbacks[j].playback.release()
			}
		}
	}
	for i := range s.speechLines {
		if line := &s.speechLines[i]; line.isValid() {
			line.playback.stop()
			line.playback.release()
		}
	}
	for i := range s.ambiences {
		if ambience := &s.ambiences[i]; ambience.isValid() {
			ambience.playback.stop()
			ambience.playback.release()
		}
	}
	for i := range s.musicTracks {
		if track := &s.musicTracks[i]; track.isValid() {
			track.playback.Stop()
			track.playback.Release()
		}
	}
	s.listeners = nil
	s.emitters = nil
	s.musicTracks = nil
	s.effects = nil
	s.effectPools = nil
	s.speechLines = nil
	s.ambiences = nil
	s.musicQueue = nil
	s.musicFadingIndices = nil
	s.musicCurrentIndex = nilIndex
	s.activeListenerID = NilListenerID

	s.musicTrackStartedSubscriptions.Clear()
	s.musicTrackEndingSubscriptions.Clear()
	s.musicTrackFinishedSubscriptions.Clear()
	s.effectFinishedSubscriptions.Clear()
	s.speechFinishedSubscriptions.Clear()

	s.musicBus.release()
	s.speechBus.release()
	s.effectBus.release()
	s.ambienceBus.release()
}

// updateListener sends the orientation of the active listener to the audio
// API. The audio listener is kept at the origin, since the positions of
// spatial playbacks are expressed relative to the active listener.
func (s *Scene) updateListener() {
	rotation := sprec.IdentityQuat()
	if s.activeListenerID != NilListenerID {
		rotation = dtos.Quat(s.listeners[s.activeListenerID.index].rotation)
	}
	listener := s.api.SpatialListener()
	listener.SetPosition(sprec.ZeroVec3())
	listener.SetRotation(rotation)
}

// releaseMusicTrack releases the audio resources of the track at the
// specified index and frees its slot.
func (s *Scene) releaseMusicTrack(index int32) {
	track := &s.musicTracks[index]
	track.playback.Release()
	*track = musicTrackState{
		revision: track.revision + 1, // progress revision to invalid (even) value
	}
	s.freeMusicTrackIndices.Push(index)
}

// finishMusicTrack stops and removes the track at the specified index,
// which must have been playing, and notifies subscribers.
func (s *Scene) finishMusicTrack(index int32) {
	track := &s.musicTracks[index]
	id := MusicTrackID{
		index:    index,
		revision: track.revision,
	}
	track.playback.Stop()
	s.releaseMusicTrack(index)

	for callback := range s.musicTrackFinishedSubscriptions.CallbacksIter() {
		callback(id)
	}
}

// updateMusic progresses the music: fades, the current track and the
// queue.
func (s *Scene) updateMusic(elapsedSeconds float64) {
	if s.musicBus.paused {
		elapsedSeconds = 0.0 // the music does not progress while paused
	}

	for i := 0; i < len(s.musicFadingIndices); {
		index := s.musicFadingIndices[i]
		track := &s.musicTracks[index]
		track.fade.advance(elapsedSeconds)
		if track.finished || track.fade.isDone() {
			s.musicFadingIndices = slices.Delete(s.musicFadingIndices, i, i+1)
			s.finishMusicTrack(index)
			continue
		}
		track.playback.SetGain(float32(track.gain) * track.fade.value)
		i++
	}

	if s.musicCurrentIndex != nilIndex {
		switch {
		case s.musicTracks[s.musicCurrentIndex].finished:
			index := s.musicCurrentIndex
			s.musicCurrentIndex = nilIndex
			s.finishMusicTrack(index)
		case s.musicSkipRequested:
			s.retireCurrentMusicTrack(s.musicSkipFade)
		default:
			s.advanceCurrentMusicTrack(elapsedSeconds)
		}
	}
	s.musicSkipRequested = false

	if s.musicCurrentIndex == nilIndex && len(s.musicQueue) > 0 {
		s.startNextMusicTrack()
	}
	s.musicFadeIn = 0
}

// advanceCurrentMusicTrack progresses the current track, notifies
// subscribers once it is about to end, and starts the crossfade into the
// next track once it is due.
func (s *Scene) advanceCurrentMusicTrack(elapsedSeconds float64) {
	track := &s.musicTracks[s.musicCurrentIndex]
	track.position += elapsedSeconds
	if !track.fade.isDone() {
		track.fade.advance(elapsedSeconds)
		track.playback.SetGain(float32(track.gain) * track.fade.value)
	}
	if track.looping {
		return
	}
	remainingSeconds := max(0.0, track.length-track.position)

	if !track.endingNotified && remainingSeconds <= s.musicEndingLeadTime.Seconds() {
		track.endingNotified = true
		id := MusicTrackID{
			index:    s.musicCurrentIndex,
			revision: track.revision,
		}
		for callback := range s.musicTrackEndingSubscriptions.CallbacksIter() {
			callback(id)
		}
		// NOTE: The track pointer must not be used past this point, since
		// callbacks are allowed to enqueue tracks.
	}

	if s.musicCrossfade > 0 && len(s.musicQueue) > 0 && remainingSeconds <= s.musicCrossfade.Seconds() {
		s.retireCurrentMusicTrack(time.Duration(remainingSeconds * float64(time.Second)))
	}
}

// retireCurrentMusicTrack makes the current track fade out over the
// specified time, leaving the scene without a current track. The track
// that takes over fades in over the same time.
func (s *Scene) retireCurrentMusicTrack(fade time.Duration) {
	index := s.musicCurrentIndex
	s.musicCurrentIndex = nilIndex
	if fade <= 0 {
		s.finishMusicTrack(index)
		return
	}
	s.musicTracks[index].fade.fadeTo(0.0, fade)
	s.musicFadingIndices = append(s.musicFadingIndices, index)
	s.musicFadeIn = fade
}

// startNextMusicTrack makes the first track in the queue the current one
// and starts playing it.
func (s *Scene) startNextMusicTrack() {
	index := s.musicQueue[0]
	s.musicQueue = slices.Delete(s.musicQueue, 0, 1)
	s.musicCurrentIndex = index

	track := &s.musicTracks[index]
	track.fade = newFader(0.0)
	track.fade.fadeTo(1.0, s.musicFadeIn)
	track.playback.SetGain(float32(track.gain) * track.fade.value)
	track.playback.Start(track.position)

	id := MusicTrackID{
		index:    index,
		revision: track.revision,
	}
	for callback := range s.musicTrackStartedSubscriptions.CallbacksIter() {
		callback(id)
	}
}

// finishSpeech removes the speech line at the specified index, stopping it
// first if it has not completed on its own, and notifies subscribers.
func (s *Scene) finishSpeech(index int32, completed bool) {
	line := &s.speechLines[index]
	id := SpeechID{
		index:    index,
		revision: line.revision,
	}
	if !completed {
		line.playback.stop()
	}
	line.playback.release()
	*line = speechState{
		revision: line.revision + 1, // progress revision to invalid (even) value
	}
	s.freeSpeechIndices.Push(index)

	for callback := range s.speechFinishedSubscriptions.CallbacksIter() {
		callback(id, completed)
	}
}

// updateSpeech removes the speech lines that have finished playing,
// notifying subscribers, and applies emitter and listener changes to the
// rest.
func (s *Scene) updateSpeech() {
	for i := 0; i < len(s.speechLines); i++ {
		line := &s.speechLines[i]
		if !line.isValid() {
			continue
		}
		if line.finished {
			s.finishSpeech(int32(i), true)
		} else {
			s.syncPlayback(&line.playback)
		}
	}
}

// stopSpeechOnEmitter cuts short the speech lines that are playing on the
// emitter with the specified ID, where [NilEmitterID] stands for the
// narrator.
func (s *Scene) stopSpeechOnEmitter(emitterID EmitterID) {
	for i := 0; i < len(s.speechLines); i++ {
		if line := &s.speechLines[i]; line.isValid() && line.playback.emitter == emitterID {
			s.finishSpeech(int32(i), false)
		}
	}
}

// releaseEffect releases the audio resources of the effect at the
// specified index and frees its slot.
func (s *Scene) releaseEffect(index int32) {
	effect := &s.effects[index]
	effect.playback.release()
	*effect = effectState{
		revision: effect.revision + 1, // progress revision to invalid (even) value
	}
	s.freeEffectIndices.Push(index)
}

// updateEffects removes the effects that have finished playing, notifying
// subscribers, and applies emitter and listener changes to the rest.
func (s *Scene) updateEffects() {
	for i := 0; i < len(s.effects); i++ {
		effect := &s.effects[i]
		if !effect.isValid() {
			continue
		}
		if !effect.finished {
			s.syncPlayback(&effect.playback)
			continue
		}
		id := EffectID{
			index:    int32(i),
			revision: effect.revision,
		}
		s.releaseEffect(id.index)
		for callback := range s.effectFinishedSubscriptions.CallbacksIter() {
			callback(id)
		}
	}
}

// stopEffectsOnEmitter stops all effects that are playing on the emitter
// with the specified ID.
func (s *Scene) stopEffectsOnEmitter(emitterID EmitterID) {
	for i := range s.effects {
		if effect := &s.effects[i]; effect.isValid() && effect.playback.emitter == emitterID {
			effect.playback.stop()
			s.releaseEffect(int32(i))
		}
	}
}

// updatePooledEffects applies emitter and listener changes to the pooled
// playbacks that are playing.
func (s *Scene) updatePooledEffects() {
	for i := range s.effectPools {
		pool := &s.effectPools[i]
		for j := range pool.playbacks {
			if pooled := &pool.playbacks[j]; pooled.active {
				s.syncPlayback(&pooled.playback)
			}
		}
	}
}

// stopPooledEffectsOnEmitter stops all pooled playbacks that are playing on
// the emitter with the specified ID.
func (s *Scene) stopPooledEffectsOnEmitter(emitterID EmitterID) {
	for i := range s.effectPools {
		pool := &s.effectPools[i]
		for j := range pool.playbacks {
			if pooled := &pool.playbacks[j]; pooled.active && pooled.playback.emitter == emitterID {
				pooled.playback.stop()
				pooled.active = false
			}
		}
	}
}

// releaseAmbience releases the audio resources of the ambience at the
// specified index and frees its slot.
func (s *Scene) releaseAmbience(index int32) {
	ambience := &s.ambiences[index]
	ambience.playback.release()
	*ambience = ambienceState{
		revision: ambience.revision + 1, // progress revision to invalid (even) value
	}
	s.freeAmbienceIndices.Push(index)
}

// updateAmbiences progresses the fades of the ambiences, removes the ones
// that have faded out, and applies emitter and listener changes to the
// rest.
func (s *Scene) updateAmbiences(elapsedSeconds float64) {
	if s.ambienceBus.paused {
		elapsedSeconds = 0.0 // fades do not progress while paused
	}
	for i := range s.ambiences {
		ambience := &s.ambiences[i]
		if !ambience.isValid() {
			continue
		}
		ambience.fade.advance(elapsedSeconds)
		if ambience.stopping && ambience.fade.isDone() {
			ambience.playback.stop()
			s.releaseAmbience(int32(i))
			continue
		}
		ambience.playback.fade = ambience.fade.value
		s.syncPlayback(&ambience.playback)
	}
}

// stopAmbiencesOnEmitter stops all ambiences that are playing on the
// emitter with the specified ID.
func (s *Scene) stopAmbiencesOnEmitter(emitterID EmitterID) {
	for i := range s.ambiences {
		if ambience := &s.ambiences[i]; ambience.isValid() && ambience.playback.emitter == emitterID {
			ambience.playback.stop()
			s.releaseAmbience(int32(i))
		}
	}
}

// createPlayback creates a playback for the specified media on the
// specified bus. The playback is spatial if emitterID is not
// [NilEmitterID], in which case the ID must be valid.
func (s *Scene) createPlayback(bus *busState, media audio.Media, emitterID EmitterID) playbackState {
	if media == nil {
		panic("media must be specified")
	}
	result := playbackState{
		emitter: emitterID,
		gain:    1.0,
		fade:    1.0,
		pitch:   1.0,
	}
	if emitterID != NilEmitterID {
		s.Emitters().resolve(emitterID, true)
		playback := s.api.CreateSpatialPlayback(bus.bus, media, audio.PlaybackSettings{})
		result.playback = playback
		result.spatial = playback
	} else {
		result.playback = s.api.CreatePlayback(bus.bus, media, audio.PlaybackSettings{})
	}
	return result
}

// syncPlayback sends the gain, playback rate and, for spatial playbacks,
// the placement of the playback to the audio API, skipping values that
// have not changed since the last call.
func (s *Scene) syncPlayback(v *playbackState) {
	gain := float32(v.gain) * v.fade
	rate := float32(v.pitch)

	if v.spatial != nil {
		emitter := &s.emitters[v.emitter.index]
		listenerPosition, listenerVelocity := s.listenerMotion()

		gain *= float32(emitter.gain)
		rate *= s.dopplerRatio(emitter, listenerPosition, listenerVelocity)

		// The position is sent relative to the listener, in order to
		// preserve precision when far away from the world origin.
		position := dtos.Vec3(dprec.Vec3Diff(emitter.position, listenerPosition))
		if !v.synced || position != v.appliedPosition {
			v.spatial.SetPosition(position)
			v.appliedPosition = position
		}
		rotation := dtos.Quat(emitter.rotation)
		if !v.synced || rotation != v.appliedRotation {
			v.spatial.SetRotation(rotation)
			v.appliedRotation = rotation
		}
		innerConeAngle := dtos.Angle(emitter.innerConeAngle)
		if !v.synced || innerConeAngle != v.appliedInnerConeAngle {
			v.spatial.SetInnerConeAngle(innerConeAngle)
			v.appliedInnerConeAngle = innerConeAngle
		}
		outerConeAngle := dtos.Angle(emitter.outerConeAngle)
		if !v.synced || outerConeAngle != v.appliedOuterConeAngle {
			v.spatial.SetOuterConeAngle(outerConeAngle)
			v.appliedOuterConeAngle = outerConeAngle
		}
		outerConeGain := float32(emitter.outerConeGain)
		if !v.synced || outerConeGain != v.appliedOuterConeGain {
			v.spatial.SetOuterConeGain(outerConeGain)
			v.appliedOuterConeGain = outerConeGain
		}
	}

	if !v.synced || gain != v.appliedGain {
		v.playback.SetGain(gain)
		v.appliedGain = gain
	}
	if !v.synced || rate != v.appliedRate {
		v.playback.SetPlaybackRate(rate)
		v.appliedRate = rate
	}
	v.synced = true
}

// listenerMotion returns the position and velocity of the active listener,
// or zero vectors if there is no active listener.
func (s *Scene) listenerMotion() (dprec.Vec3, dprec.Vec3) {
	if s.activeListenerID == NilListenerID {
		return dprec.ZeroVec3(), dprec.ZeroVec3()
	}
	listener := &s.listeners[s.activeListenerID.index]
	return listener.position, listener.velocity
}

// dopplerRatio returns the factor by which the playback rate of a sound on
// the specified emitter needs to be scaled so that it is perceived with the
// correct Doppler shift by a listener with the specified motion.
func (s *Scene) dopplerRatio(emitter *emitterState, listenerPosition, listenerVelocity dprec.Vec3) float32 {
	if !emitter.doppler || s.dopplerFactor <= 0.0 {
		return 1.0
	}
	delta := dprec.Vec3Diff(listenerPosition, emitter.position)
	distance := delta.Length()
	if distance < minDopplerDistance {
		return 1.0
	}
	direction := dprec.Vec3Quot(delta, distance) // from emitter to listener

	// Both speeds are measured along the direction of sound propagation.
	listenerSpeed := s.dopplerFactor * dprec.Vec3Dot(listenerVelocity, direction)
	emitterSpeed := s.dopplerFactor * dprec.Vec3Dot(emitter.velocity, direction)

	numerator := s.speedOfSound - listenerSpeed
	denominator := s.speedOfSound - emitterSpeed
	if numerator <= 0.0 {
		return minDopplerRatio // listener outruns the sound
	}
	if denominator <= 0.0 {
		return maxDopplerRatio // emitter is supersonic towards the listener
	}
	return float32(dprec.Clamp(numerator/denominator, minDopplerRatio, maxDopplerRatio))
}

// allocateSlot returns the index of an unused slot, and a pointer to it,
// reusing a previously freed slot if there is one.
//
// The returned pointer is only valid until the next allocation, since
// growing the slice may move the slots.
func allocateSlot[T any](slots *[]T, freeIndices *ds.Stack[int32]) (int32, *T) {
	var index int32
	if freeIndices.IsEmpty() {
		index = int32(len(*slots))
		var zero T
		*slots = append(*slots, zero)
	} else {
		index = freeIndices.Pop()
	}
	return index, &(*slots)[index]
}
