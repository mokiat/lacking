package sound

import "github.com/mokiat/lacking/util/observer"

// MusicTrackStartedCallback is a mechanism to receive notifications about
// a music track that has started playing.
type MusicTrackStartedCallback func(id MusicTrackID)

// MusicTrackStartedSubscription represents a notification subscription for
// music tracks that start playing.
type MusicTrackStartedSubscription = observer.Subscription[MusicTrackStartedCallback]

// MusicTrackEndingCallback is a mechanism to receive notifications about
// a music track that is about to reach its end.
//
// The callback may enqueue further tracks through [MusicView.Enqueue].
type MusicTrackEndingCallback func(id MusicTrackID)

// MusicTrackEndingSubscription represents a notification subscription for
// music tracks that are about to reach their end.
type MusicTrackEndingSubscription = observer.Subscription[MusicTrackEndingCallback]

// MusicTrackFinishedCallback is a mechanism to receive notifications about
// a music track that is no longer playing. By the time the callback is
// called the ID no longer references a valid track.
type MusicTrackFinishedCallback func(id MusicTrackID)

// MusicTrackFinishedSubscription represents a notification subscription
// for music tracks that are no longer playing.
type MusicTrackFinishedSubscription = observer.Subscription[MusicTrackFinishedCallback]

// EffectFinishedCallback is a mechanism to receive notifications about a
// sound effect that has played to its end. By the time the callback is
// called the ID no longer references a valid effect.
type EffectFinishedCallback func(id EffectID)

// EffectFinishedSubscription represents a notification subscription for
// sound effects that have played to their end.
type EffectFinishedSubscription = observer.Subscription[EffectFinishedCallback]

// SpeechFinishedCallback is a mechanism to receive notifications about a
// speech line that is no longer playing. The completed argument is true
// if the line was played to its end and false if it was cut short. By the
// time the callback is called the ID no longer references a valid line.
type SpeechFinishedCallback func(id SpeechID, completed bool)

// SpeechFinishedSubscription represents a notification subscription for
// speech lines that are no longer playing.
type SpeechFinishedSubscription = observer.Subscription[SpeechFinishedCallback]
