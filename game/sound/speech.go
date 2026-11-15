package sound

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/audio"
)

// SpeechID uniquely identifies a speech line that was started through
// [SpeechView.Play].
//
// The zero value, also available as [NilSpeechID], does not reference a
// valid speech line.
type SpeechID struct {
	index    int32
	revision int32
}

// NilSpeechID is a [SpeechID] that never references a valid speech line.
var NilSpeechID = SpeechID{}

// SpeechInfo describes a speech line to be played through
// [SpeechView.Play].
type SpeechInfo struct {

	// Media is the recording of the line. It must be specified.
	Media audio.Media

	// Emitter is the emitter that speaks the line. If it is
	// [NilEmitterID], the line is spoken by a narrator, which is not
	// spatial and is heard the same regardless of where the listener is.
	Emitter EmitterID

	// Gain is the gain of the line.
	//
	// Defaults to 1.
	Gain opt.T[float64]
}

// SpeechView provides access to the speech lines that are playing in a
// [Scene], and to the mixing controls that they share (see
// [SpeechView.Bus]).
//
// A speaker, be it an emitter or the narrator, says one line at a time.
// Playing a line on a speaker that is already speaking cuts the previous
// line short.
//
// A SpeechView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Speech].
type SpeechView struct {
	scene *Scene
}

// Bus returns a [BusView] with the mixing controls that are shared by all
// speech.
func (v SpeechView) Bus() BusView {
	return BusView{
		bus: &v.scene.speechBus,
	}
}

// Play starts a new speech line and returns its ID. If the speaker is
// already speaking, its previous line is cut short and reported to the
// subscribers of [SpeechView.SubscribeFinished] before the new one starts.
//
// The ID stays valid for as long as the line is playing. Once the line
// reaches its end, it is removed during the next [Scene.Update] and its
// ID becomes invalid.
//
// It panics if no media is specified, or if the emitter is neither
// [NilEmitterID] nor a valid emitter.
func (v SpeechView) Play(info SpeechInfo) SpeechID {
	if info.Media == nil {
		panic("media must be specified")
	}
	if info.Emitter != NilEmitterID {
		v.scene.Emitters().resolve(info.Emitter, true)
	}
	v.scene.stopSpeechOnEmitter(info.Emitter)

	playback := v.scene.createPlayback(&v.scene.speechBus, info.Media, info.Emitter)
	playback.gain = max(0.0, info.Gain.ValueOrDefault(1.0))

	index, line := allocateSlot(&v.scene.speechLines, v.scene.freeSpeechIndices)
	*line = speechState{
		revision: line.revision + 1, // progress revision to valid (odd) value
		playback: playback,
	}
	id := SpeechID{
		index:    index,
		revision: line.revision,
	}
	scene := v.scene
	playback.setOnFinished(func() {
		if line := scene.Speech().resolve(id, false); line != nil {
			line.finished = true
		}
	})
	scene.syncPlayback(&line.playback)
	playback.start(0.0)

	return id
}

// Stop cuts short the speech line with the specified ID, which makes the
// ID invalid, and reports it to the subscribers of
// [SpeechView.SubscribeFinished].
//
// It panics if the ID does not reference a valid speech line. Use
// [SpeechView.IsValid] first, since a line becomes invalid on its own once
// it reaches its end.
func (v SpeechView) Stop(id SpeechID) {
	v.resolve(id, true)
	v.scene.finishSpeech(id.index, false)
}

// StopAll cuts short all speech lines that are playing and reports them
// to the subscribers of [SpeechView.SubscribeFinished].
func (v SpeechView) StopAll() {
	for i := 0; i < len(v.scene.speechLines); i++ {
		if v.scene.speechLines[i].isValid() {
			v.scene.finishSpeech(int32(i), false)
		}
	}
}

// IsValid returns whether the specified ID references a speech line that
// is still playing.
func (v SpeechView) IsValid(id SpeechID) bool {
	line := v.resolve(id, false)
	return line != nil
}

// IsSpeaking returns whether a speech line is playing on the emitter with
// the specified ID. Passing [NilEmitterID] returns whether the narrator is
// speaking.
func (v SpeechView) IsSpeaking(emitterID EmitterID) bool {
	for i := range v.scene.speechLines {
		if line := &v.scene.speechLines[i]; line.isValid() && line.playback.emitter == emitterID {
			return true
		}
	}
	return false
}

// SubscribeFinished registers a callback that is invoked whenever a
// speech line stops playing. Lines that play to their end are reported
// during [Scene.Update]. Lines that are cut short are reported by the
// call that causes it.
//
// Call [SpeechFinishedSubscription.Delete] on the returned subscription to
// stop receiving notifications.
func (v SpeechView) SubscribeFinished(callback SpeechFinishedCallback) *SpeechFinishedSubscription {
	return v.scene.speechFinishedSubscriptions.Subscribe(callback)
}

// resolve looks up the speechState referenced by id. If id is stale or
// otherwise invalid, resolve panics when required is true, or returns nil
// otherwise.
func (v SpeechView) resolve(id SpeechID, required bool) *speechState {
	if id.revision == 0 {
		if required {
			panic("invalid speech ID")
		}
		return nil
	}
	line := &v.scene.speechLines[id.index]
	if line.revision != id.revision {
		if required {
			panic("invalid speech ID")
		}
		return nil
	}
	return line
}

// speechState holds the internal state of a single playing speech line,
// as tracked by a [Scene].
type speechState struct {
	revision int32
	playback playbackState
	finished bool // set once the audio API reports that playback has ended
}

// isValid returns whether this state is currently backing a playing speech
// line, as opposed to a freed slot awaiting reuse.
func (s *speechState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
