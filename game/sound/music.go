package sound

import (
	"slices"
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/core/audio"
)

// MusicTrackID uniquely identifies a music track that was added through
// [MusicView.Enqueue] or [MusicView.Play].
//
// The zero value, also available as [NilMusicTrackID], does not reference
// a valid track.
type MusicTrackID struct {
	index    int32
	revision int32
}

// NilMusicTrackID is a [MusicTrackID] that never references a valid track.
var NilMusicTrackID = MusicTrackID{}

// MusicTrackInfo describes a music track to be added through
// [MusicView.Enqueue] or [MusicView.Play].
type MusicTrackInfo struct {

	// Media is the music to be played. It must be specified.
	Media audio.Media

	// Gain is the gain of the track, which allows tracks that were
	// mastered at different levels to be evened out.
	//
	// Defaults to 1.
	Gain opt.T[float64]

	// StartAt is the offset into the media at which the track starts
	// playing.
	StartAt time.Duration

	// Loop indicates whether the track should repeat until the music is
	// moved on through [MusicView.Skip], [MusicView.Play] or
	// [MusicView.Stop]. Tracks queued behind a looping track do not start
	// on their own.
	Loop bool

	// LoopStart is the offset into the media to which a looping track
	// returns, which allows a track to have an intro that is played once.
	//
	// Defaults to the start of the media.
	LoopStart opt.T[time.Duration]

	// LoopEnd is the offset into the media at which a looping track
	// returns to LoopStart.
	//
	// Defaults to the end of the media.
	LoopEnd opt.T[time.Duration]
}

// MusicView provides access to the music of a [Scene], and to its mixing
// controls (see [MusicView.Bus]).
//
// Music is not spatial and plays from a queue, one track at a time. Once
// the current track ends, the next one in the queue takes over. All
// changes to what is playing take place during [Scene.Update], which is
// also when subscribers are notified.
//
// A MusicView is a lightweight accessor around a [Scene] and can be
// obtained through [Scene.Music].
type MusicView struct {
	scene *Scene
}

// Bus returns a [BusView] with the mixing controls of the music.
//
// Pausing the bus also suspends the fades of the music and the progress of
// its queue. Muting it does neither.
func (v MusicView) Bus() BusView {
	return BusView{
		bus: &v.scene.musicBus,
	}
}

// Enqueue adds a track to the end of the queue and returns its ID. If no
// music is playing, the track starts during the next [Scene.Update].
//
// The ID stays valid until the track stops playing or is removed from the
// queue.
//
// It panics if no media is specified.
func (v MusicView) Enqueue(info MusicTrackInfo) MusicTrackID {
	if info.Media == nil {
		panic("media must be specified")
	}
	playback := v.scene.api.CreatePlayback(v.scene.musicBus.bus, info.Media, audio.PlaybackSettings{})
	playback.SetLooping(info.Loop)
	if loopStart, ok := info.LoopStart.Unwrap(); ok {
		playback.SetLoopStart(loopStart.Seconds())
	}
	if loopEnd, ok := info.LoopEnd.Unwrap(); ok {
		playback.SetLoopEnd(loopEnd.Seconds())
	}

	index, track := allocateSlot(&v.scene.musicTracks, v.scene.freeMusicTrackIndices)
	*track = musicTrackState{
		revision: track.revision + 1, // progress revision to valid (odd) value
		playback: playback,
		gain:     max(0.0, info.Gain.ValueOrDefault(1.0)),
		looping:  info.Loop,
		length:   info.Media.Length(),
		position: info.StartAt.Seconds(),
	}
	id := MusicTrackID{
		index:    index,
		revision: track.revision,
	}

	scene := v.scene
	playback.SetOnFinished(func() {
		if track := scene.Music().resolve(id, false); track != nil {
			track.finished = true
		}
	})
	scene.musicQueue = append(scene.musicQueue, index)

	return id
}

// Play replaces the queue with the specified track and moves the music on
// to it during the next [Scene.Update], crossfading from the current track
// according to [MusicView.CrossfadeDuration]. It returns the ID of the new
// track.
//
// It panics if no media is specified.
func (v MusicView) Play(info MusicTrackInfo) MusicTrackID {
	v.Clear()
	id := v.Enqueue(info)
	v.Skip()
	return id
}

// Skip moves the music on to the next track in the queue during the next
// [Scene.Update], crossfading from the current track according to
// [MusicView.CrossfadeDuration]. If the queue is empty, the music stops.
func (v MusicView) Skip() {
	v.scene.musicSkipRequested = true
	v.scene.musicSkipFade = v.scene.musicCrossfade
}

// Stop empties the queue and stops the current track during the next
// [Scene.Update], fading it out over the specified time. If fadeOut is
// not positive, the track is stopped without a fade.
func (v MusicView) Stop(fadeOut time.Duration) {
	v.Clear()
	v.scene.musicSkipRequested = true
	v.scene.musicSkipFade = fadeOut
}

// Remove removes the specified track from the queue, which makes its ID
// invalid. If the track is the one that is currently playing, the music is
// moved on instead, as if through [MusicView.Skip].
//
// It panics if the ID does not reference a valid track.
func (v MusicView) Remove(id MusicTrackID) {
	v.resolve(id, true)
	if id.index == v.scene.musicCurrentIndex {
		v.Skip()
		return
	}
	if position := slices.Index(v.scene.musicQueue, id.index); position >= 0 {
		v.scene.musicQueue = slices.Delete(v.scene.musicQueue, position, position+1)
		v.scene.releaseMusicTrack(id.index)
	}
}

// Clear removes all tracks from the queue, which makes their IDs invalid.
// The track that is currently playing is not affected.
func (v MusicView) Clear() {
	for _, index := range v.scene.musicQueue {
		v.scene.releaseMusicTrack(index)
	}
	v.scene.musicQueue = v.scene.musicQueue[:0]
}

// Current returns the ID of the track that is currently playing, or
// [NilMusicTrackID] if there is none.
func (v MusicView) Current() MusicTrackID {
	if v.scene.musicCurrentIndex == nilIndex {
		return NilMusicTrackID
	}
	return MusicTrackID{
		index:    v.scene.musicCurrentIndex,
		revision: v.scene.musicTracks[v.scene.musicCurrentIndex].revision,
	}
}

// QueueLength returns the number of tracks that are waiting to be played,
// which does not include the track that is currently playing.
func (v MusicView) QueueLength() int {
	return len(v.scene.musicQueue)
}

// IsValid returns whether the specified ID references a track that is
// either queued or still playing.
func (v MusicView) IsValid(id MusicTrackID) bool {
	track := v.resolve(id, false)
	return track != nil
}

// CrossfadeDuration returns the time over which the music transitions
// from one track to the next.
//
// Defaults to 0.
func (v MusicView) CrossfadeDuration() time.Duration {
	return v.scene.musicCrossfade
}

// SetCrossfadeDuration changes the time over which the music transitions
// from one track to the next. Negative values are clamped to 0.
//
// With a positive duration, the next track in the queue starts that much
// ahead of the end of the current one and the two are blended.
//
// With a duration of 0, the next track starts during the first
// [Scene.Update] after the current one has ended, which can leave a gap
// of up to one frame between the two.
func (v MusicView) SetCrossfadeDuration(duration time.Duration) {
	v.scene.musicCrossfade = max(0, duration)
}

// EndingLeadTime returns how far ahead of the end of a track the
// subscribers of [MusicView.SubscribeTrackEnding] are notified.
//
// Defaults to 5 seconds.
func (v MusicView) EndingLeadTime() time.Duration {
	return v.scene.musicEndingLeadTime
}

// SetEndingLeadTime changes how far ahead of the end of a track the
// subscribers of [MusicView.SubscribeTrackEnding] are notified. Negative
// values are clamped to 0.
//
// It should be longer than [MusicView.CrossfadeDuration] if tracks that
// are enqueued in response to the notification are to be crossfaded.
func (v MusicView) SetEndingLeadTime(duration time.Duration) {
	v.scene.musicEndingLeadTime = max(0, duration)
}

// SubscribeTrackStarted registers a callback that is invoked, during
// [Scene.Update], whenever a track starts playing.
//
// Call [MusicTrackStartedSubscription.Delete] on the returned subscription
// to stop receiving notifications.
func (v MusicView) SubscribeTrackStarted(callback MusicTrackStartedCallback) *MusicTrackStartedSubscription {
	return v.scene.musicTrackStartedSubscriptions.Subscribe(callback)
}

// SubscribeTrackEnding registers a callback that is invoked, during
// [Scene.Update], once the time left until the end of the current track
// drops below [MusicView.EndingLeadTime]. It is invoked at most once per
// track and never for looping tracks.
//
// The callback is the opportunity to keep the music going by adding a
// track through [MusicView.Enqueue]. The number of tracks that are already
// queued is available through [MusicView.QueueLength].
//
// Call [MusicTrackEndingSubscription.Delete] on the returned subscription
// to stop receiving notifications.
func (v MusicView) SubscribeTrackEnding(callback MusicTrackEndingCallback) *MusicTrackEndingSubscription {
	return v.scene.musicTrackEndingSubscriptions.Subscribe(callback)
}

// SubscribeTrackFinished registers a callback that is invoked, during
// [Scene.Update], whenever a track that was playing no longer is, be it
// because it has ended or because the music was moved on or stopped.
// Tracks that are removed from the queue without ever playing are not
// reported.
//
// Call [MusicTrackFinishedSubscription.Delete] on the returned
// subscription to stop receiving notifications.
func (v MusicView) SubscribeTrackFinished(callback MusicTrackFinishedCallback) *MusicTrackFinishedSubscription {
	return v.scene.musicTrackFinishedSubscriptions.Subscribe(callback)
}

// resolve looks up the musicTrackState referenced by id. If id is stale or
// otherwise invalid, resolve panics when required is true, or returns nil
// otherwise.
func (v MusicView) resolve(id MusicTrackID, required bool) *musicTrackState {
	if id.revision == 0 {
		if required {
			panic("invalid music track ID")
		}
		return nil
	}
	track := &v.scene.musicTracks[id.index]
	if track.revision != id.revision {
		if required {
			panic("invalid music track ID")
		}
		return nil
	}
	return track
}

// musicTrackState holds the internal state of a single music track, as
// tracked by a [Scene]. A track is either waiting in the queue, is the
// current track, or is fading out after having been the current track.
type musicTrackState struct {
	revision       int32
	playback       audio.Playback
	gain           float64
	looping        bool
	length         float64 // in seconds
	position       float64 // in seconds; estimated while playing
	fade           fader
	endingNotified bool
	finished       bool // set once the audio API reports that playback has ended
}

// isValid returns whether this state is currently backing a live track,
// as opposed to a freed slot awaiting reuse.
func (s *musicTrackState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
