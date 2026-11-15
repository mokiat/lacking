// Package sound provides a game-oriented sound stage on top of the
// low-level [audio] package.
//
// # Scene
//
// A [Scene] is a self-contained sound stage, created through [NewScene]
// for a given [audio.API]. It holds listeners, emitters and everything
// that is playing, and it is advanced through [Scene.Update], which should
// be called once per frame. A Scene must only be used from the thread
// that its audio API is bound to.
//
// Sounds are represented by [audio.Media], which is created through the
// audio API and is owned by the caller. The Scene never releases media.
//
// # Categories
//
// Everything that plays belongs to one of four categories, each with its
// own view and its own mixing controls (gain, volume, mute and pause),
// which are available through the Bus method of the view (see [BusView]),
// so that a game can offer separate volume settings and can, for example,
// pause effects while the music keeps playing. The categories have no
// reverb or compression of their own:
//
//   - Music (see [MusicView]) plays tracks from a queue, one after the
//     other, with optional crossfading. Subscribers are told when the
//     current track is about to end, so that they can keep the queue
//     filled.
//   - Speech (see [SpeechView]) plays spoken lines, one at a time per
//     speaker, and reports when each line stops.
//   - Effects (see [EffectView]) plays individual one-shot or looping
//     sounds. Sounds that are played often, such as footsteps, are better
//     served by an effect pool (see [EffectPoolView]), which reuses a
//     fixed set of playbacks and so has no per-play allocation cost.
//   - Ambience (see [AmbienceView]) plays looping background sounds that
//     can be faded in and out.
//
// # Emitters and listeners
//
// An emitter (see [EmitterView]) is a point in space on which speech,
// effects and ambiences can be played. Such sounds are spatial: they are
// heard relative to the active listener of the scene (see [ListenerView]
// and [Scene.SetActiveListener]) and they follow the emitter as it moves.
// A sound that is played without an emitter, by passing [NilEmitterID],
// is not spatial. Music never is.
//
// Both emitters and listeners have a velocity. It does not move them, but
// it is used to apply the Doppler effect to the sounds of emitters that
// have it enabled (see [EmitterView.SetDoppler],
// [Scene.SetSpeedOfSound] and [Scene.SetDopplerFactor]).
//
// # IDs, views and handles
//
// Every entity in a Scene is referred to by a small, comparable ID value,
// such as [EmitterID] or [EffectID]. An ID holds no pointers, so it is
// cheap to store, and it becomes invalid once its entity is deleted or
// has finished playing, even if the underlying storage is later reused.
// The zero value of every ID type (e.g. [NilEmitterID]) never refers to a
// valid entity.
//
// An ID is resolved through a view of the owning Scene, such as
// [Scene.Emitters] or [Scene.Effects]. Most view methods panic when given
// an invalid ID; use the view's IsValid method if an ID might be stale,
// which is common for sounds that end on their own.
//
// A handle, such as [EmitterHandle], binds an ID together with its view,
// for callers that prefer an object-oriented API. Handles are convenient
// but keep the Scene reachable, so prefer storing IDs in long-lived data.
package sound
