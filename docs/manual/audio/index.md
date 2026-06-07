---
title: Overview
---

# Audio

The `core/audio` package defines a mid-level, bus-based audio API. It is the audio equivalent of the `render` package: an abstract interface implemented separately for each platform (native desktop, web). Game code is not expected to use this API directly — a higher-level audio API is provided for that purpose — but there are cases where working directly with these primitives is necessary.

The API is obtained from the application window:

```go
api := window.AudioAPI()
```

## Core Concepts

| Concept | Description |
|---|---|
| **API** | Entry point. Creates media, buses, and playback instances. Exposes the master bus and spatial listener. |
| **MediaData** | Raw decoded audio frames together with the sample rate. |
| **Media** | A decoded audio clip loaded into the audio system. Acts as a data source for playback instances. |
| **Frame** | A single stereo audio frame consisting of left and right channel values. |
| **MasterBus** | The overall output sink. Controls master gain and compression. |
| **Bus** | A named group of sound sources with collective gain, reverb, compression, and pause/resume control. |
| **Playback** | A single playing instance of a `Media` on a `Bus`. Controls start/stop/pause and per-playback properties. |
| **SpatialPlayback** | A `Playback` that is also positioned and oriented in 3D space. |
| **SpatialListener** | The listener's position and orientation in 3D space for spatial audio. |

## Media

`Media` represents a decoded audio clip held in the audio system. It is created from a `MediaData` value, which holds raw PCM frames and a sample rate:

```go
// Decode from an encoded file (WAV, MP3, or any registered format).
data, _, err := audio.Decode(r)
if err != nil {
    // handle error
}

// Resample if necessary (e.g. if the audio system requires a specific rate).
// data.Frames = audio.Resample(data.Frames, data.SampleRate, targetRate)

media := api.CreateMedia(data)

// Release when no longer needed.
defer media.Release()
```

It is safe to release a `Media` after using it to create playback instances — existing playback is not affected.

`media.Length()` returns the duration of the clip in seconds.

### Frames and Sample Rate

Audio data is represented as a slice of `Frame` values, each holding a left and right channel sample:

```go
type Frame struct {
    Left  float32
    Right float32
}
```

The `Resample` utility converts frames between sample rates:

```go
resampled := audio.Resample(frames, originalRate, targetRate)
```

`SampleCount` and `Seconds` convert between frame counts and durations:

```go
count := audio.SampleCount(2.5, sampleRate) // frames for 2.5 seconds
dur   := audio.Seconds(count, sampleRate)   // back to seconds
```

### Decoding Audio Files

`audio.Decode` auto-detects the format by magic-byte prefix and decodes the data:

```go
data, format, err := audio.Decode(r) // format is e.g. "mp3" or "wav"
```

Format decoders self-register via package `init`. Import the sub-packages to enable them:

```go
import (
    _ "github.com/mokiat/lacking/core/audio/mp3"
    _ "github.com/mokiat/lacking/core/audio/wav"
)
```

Custom decoders can be added with `audio.RegisterDecoder`.

## Master Bus

`MasterBus` is the global output sink. It controls the overall gain and provides access to global compression:

```go
master := api.MasterBus()
master.SetGain(0.8)

comp := master.Compression()
comp.SetThreshold(-18.0)
comp.SetRatio(4.0)
```

## Buses

A `Bus` groups a set of sound sources for collective control. Create one with `CreateBus`, optionally enabling reverb and/or compression:

```go
musicBus := api.CreateBus(audio.BusSettings{})
sfxBus   := api.CreateBus(audio.BusSettings{UseCompression: true})
envBus   := api.CreateBus(audio.BusSettings{UseReverb: true, UseCompression: true})
defer musicBus.Release()
defer sfxBus.Release()
defer envBus.Release()
```

Releasing a bus stops all playback attached to it.

### Bus Controls

```go
bus.SetGain(0.5)  // half volume for everything on this bus

// Pause/resume all sources on the bus at once.
bus.Pause()
bus.Resume()
```

`bus.Compression()` and `bus.Reverb()` return `nil` if the bus was not created with those effects enabled.

### Reverb

Configure room characteristics on buses created with `UseReverb: true`:

| Parameter | Default | Range | Description |
|---|---|---|---|
| `RoomSize` | 0.3 | [0.0, 1.0] | Size of the virtual room. |
| `Damping` | 0.5 | [0.0, 1.0] | High-frequency absorption. Higher values simulate softer surfaces. |
| `Dry` | 1.0 | [0.0, 1.0] | Level of the unprocessed signal. |
| `Wet` | 0.5 | [0.0, 1.0] | Level of the reverberated signal. |

```go
reverb := bus.Reverb()
reverb.SetRoomSize(0.8)
reverb.SetDamping(0.3)
reverb.SetWet(0.4)
```

### Compression

Available on both `Bus` (when `UseCompression: true`) and `MasterBus`:

| Parameter | Default | Range | Description |
|---|---|---|---|
| `Threshold` | -24.0 dB | [-100.0, 0.0] | Level above which compression is applied. |
| `Ratio` | 12.0 | [1.0, 20.0] | Compression ratio (input dB : output dB above threshold). |
| `Knee` | 30.0 dB | [0.0, 40.0] | Width of the soft-knee transition around the threshold. |
| `Attack` | 0.003 s | [0.0, 1.0] | Time for compression to engage. |
| `Release` | 0.25 s | [0.0, 1.0] | Time for compression to disengage. |

```go
comp := bus.Compression()
comp.SetThreshold(-18.0)
comp.SetRatio(4.0)
```

## Playback

A `Playback` is a single instance of a `Media` playing on a `Bus`. Create one with `CreatePlayback`:

```go
playback := api.CreatePlayback(bus, media, audio.PlaybackSettings{})
defer playback.Release()

playback.Start(0)    // start from the beginning
playback.Pause()     // pause; position is preserved
playback.Resume()    // resume from paused position
playback.Stop()      // stop and reset position
```

`playback.Playing()` reports whether the playback is currently active.

### Loop Control

```go
playback.SetLooping(true)
playback.SetLoopStart(1.0) // loop from 1.0 s
playback.SetLoopEnd(4.5)   // to 4.5 s
```

### Per-Playback Controls

```go
playback.SetGain(0.7)          // individual volume
playback.SetPlaybackRate(1.5)  // 1.5× speed; pitch shifts accordingly
```

`DBToGain` and `GainToDB` convert between decibels and linear gain:

```go
playback.SetGain(audio.DBToGain(-6.0)) // -6 dB
```

### Completion Callback

```go
playback.SetOnFinished(func() {
    // called when the media plays through naturally (not on Stop or Pause,
    // and not on each loop iteration)
})
```

### Per-Playback Filters

Low-pass and high-pass filters can be enabled at creation time:

```go
playback := api.CreatePlayback(bus, media, audio.PlaybackSettings{
    UseLowPassFilter:  true,
    UseHighPassFilter: true,
})

playback.LowPassFilter().SetFrequency(8000.0)  // remove hiss above 8 kHz
playback.HighPassFilter().SetFrequency(80.0)   // remove rumble below 80 Hz
```

`LowPassFilter()` and `HighPassFilter()` return `nil` if the respective filter was not enabled at creation.

## Spatial Audio

`CreateSpatialPlayback` returns a `SpatialPlayback`, which combines `Playback` with `SpatialEmitter`. The sound source is positioned in 3D space and attenuated relative to the `SpatialListener`.

```go
// Update the listener each frame to match the camera.
listener := api.SpatialListener()
listener.SetPosition(cameraPosition)
listener.SetRotation(cameraRotation)

// Create a positioned sound source.
spatial := api.CreateSpatialPlayback(bus, media, audio.PlaybackSettings{})
defer spatial.Release()

spatial.SetPosition(sprec.Vec3{X: 10, Y: 0, Z: -5})
spatial.Start(0)
```

### Directional Emission (Cone)

By default a spatial source emits equally in all directions (`OuterConeAngle` = 360°). Narrowing the cone makes the source directional:

```go
spatial.SetInnerConeAngle(sprec.Degrees(30)) // full gain within 30°
spatial.SetOuterConeAngle(sprec.Degrees(90)) // fade to outer gain by 90°
spatial.SetOuterConeGain(0.0)                // silent outside the outer cone
```

| Property | Default | Description |
|---|---|---|
| `InnerConeAngle` | — | Within this angle the emitter plays at full gain. |
| `OuterConeAngle` | 360° | Beyond this angle the gain is `OuterConeGain`. Between inner and outer the gain is linearly interpolated. |
| `OuterConeGain` | 0.0 | Gain applied when the listener is outside the outer cone. |

## No-op Implementation

`NewNopAPI` returns a fully functional but silent implementation. All methods work correctly and return valid objects; no audio is produced. Useful for headless environments and tests:

```go
api := audio.NewNopAPI()
```
