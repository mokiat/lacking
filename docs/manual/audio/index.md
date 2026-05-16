---
title: Overview
---

# Audio

The `audio` package defines a low-level, node-graph-based audio API. It is the audio equivalent of the `render` package: an abstract interface that is implemented separately for each platform (native desktop, web). Game code is not expected to use this API directly — a higher-level audio API will be provided for that purpose.

The API is obtained from the application window:

```go
api := window.AudioAPI()
```

## Core Concepts

| Concept | Description |
|---|---|
| **API** | Entry point. Creates media and nodes, exposes the output node and spatial listener. |
| **Media** | A decoded audio clip loaded into memory. Acts as a data source for `PlaybackNode`. |
| **Sample** | A single stereo audio sample consisting of left and right channel values. |
| **Node** | An element in the audio graph that produces or processes a signal. |
| **UserNode** | A node that owns resources and must be explicitly deleted when no longer needed. |
| **Output** | The terminal sink node of the graph. Signals connected to it are sent to the speakers. |
| **SpatialListener** | Represents the listener's position and orientation in 3D space for spatial audio. |

## Media

`Media` represents a decoded audio clip held in memory. It is created from raw PCM samples or from an encoded file (WAV, MP3):

```go
// From raw PCM samples (must match the API's sample rate).
media := api.CreateMedia(samples)

// From encoded file bytes.
media := api.ParseMedia(audio.MediaInfo{
    Data:     fileBytes,
    DataType: audio.MediaDataTypeWAV, // or MediaDataTypeMP3, MediaDataTypeAuto
})

// Release when no longer needed.
defer media.Delete()
```

It is safe to delete a `Media` after using it to create a `PlaybackNode` — the node retains its own reference to the underlying data.

### Sample Rate

The API operates at a fixed sample rate, available via `api.SampleRate()`. Raw samples passed to `CreateMedia` must already be at this rate. The `Resample` utility can be used to convert:

```go
resampled := audio.Resample(samples, originalRate, api.SampleRate())
```

`SampleCount` calculates how many samples correspond to a given duration:

```go
count := audio.SampleCount(2.5, api.SampleRate()) // samples for 2.5 seconds
```

## Node Graph

Audio flows through a directed graph of nodes. Sources generate signals, processing nodes transform them, and everything ultimately connects to the output node. When multiple sources are connected to the same target their signals are mixed additively.

```go
output := api.Output()

// Connect source directly to output.
api.Connect(source, output)

// Or use Chain for a linear sequence.
api.Chain(source, gainNode, reverbNode, output)

// Disconnect when done.
api.Disconnect(source, output)
```

All nodes that are no longer needed must be deleted to avoid resource leaks:

```go
gainNode.Delete()
```

## Node Types

The following node types are available:

| Node | Factory | Purpose |
|---|---|---|
| `PlaybackNode` | `CreatePlaybackNode` | Plays back a `Media` clip. |
| `OscillatorNode` | `CreateOscillatorNode` | Generates a periodic waveform. |
| `GainNode` | `CreateGainNode` | Scales the signal amplitude. |
| `PanNode` | `CreatePanNode` | Pans the signal between left and right channels. |
| `SpatialNode` | `CreateSpatialNode` | Applies 3D positional audio effects. |
| `HighPassNode` | `CreateHighPassNode` | Removes frequencies below a cutoff. |
| `LowPassNode` | `CreateLowPassNode` | Removes frequencies above a cutoff. |
| `DelayNode` | `CreateDelayNode` | Adds a time delay to the signal. |
| `ReverbNode` | `CreateReverbNode` | Applies a room reverb effect. |
| `CompressorNode` | `CreateCompressorNode` | Applies dynamic range compression. |
| `ConnectorNode` | `CreateConnectorNode` | Pass-through node useful as a named connection point. |

### PlaybackNode

Plays back a `Media` clip. Created with an initial loop setting:

```go
node := api.CreatePlaybackNode(media, false)
defer node.Delete()

api.Connect(node, api.Output())

node.Start(0)   // start from the beginning
node.Pause()    // pause; resumes from the same position
node.Resume()   // resume from where it was paused
node.Stop()     // stop and reset position
```

Loop playback can be configured after creation, including a sub-range of the clip:

```go
node.SetLoop(true)
node.SetLoopStart(1.0) // loop from 1.0 s
node.SetLoopEnd(4.5)   // to 4.5 s
```

### OscillatorNode

Generates a continuous periodic waveform at a configurable frequency. Default frequency is 440 Hz (A4).

```go
node := api.CreateOscillatorNode()
defer node.Delete()

node.SetFrequency(220.0) // A3
api.Connect(node, api.Output())
```

### GainNode

Scales the signal amplitude. A gain of `1.0` is unity (no change); `0.0` is silence; values above `1.0` amplify. Default is `1.0`.

```go
gain := api.CreateGainNode()
defer gain.Delete()

gain.SetGain(0.5) // half volume
api.Chain(source, gain, api.Output())
```

The `DBToGain` and `GainToDB` utilities convert between decibels and linear gain:

```go
gain.SetGain(audio.DBToGain(-6.0)) // -6 dB
```

### PanNode

Distributes the signal between left and right channels. The range is `[-1.0, 1.0]` where `-1.0` is full left, `0.0` is center, and `1.0` is full right. Default is `0.0`.

```go
pan := api.CreatePanNode()
defer pan.Delete()

pan.SetPan(-0.5) // slightly left
api.Chain(source, pan, api.Output())
```

### Filter Nodes

`HighPassNode` and `LowPassNode` each expose a single `CutoffFrequency` parameter (in Hz). Default cutoff is 350 Hz for both.

```go
hp := api.CreateHighPassNode()
defer hp.Delete()
hp.SetCutoffFrequency(80.0) // remove rumble below 80 Hz

lp := api.CreateLowPassNode()
defer lp.Delete()
lp.SetCutoffFrequency(8000.0) // remove hiss above 8 kHz
```

### DelayNode

Adds a time delay to the signal. Default delay is `0.0` seconds. Implementations must support at least 1 second of delay.

```go
delay := api.CreateDelayNode()
defer delay.Delete()

delay.SetDelayTime(0.3) // 300 ms
api.Chain(source, delay, api.Output())
```

### ReverbNode

Applies a reverb effect with configurable room characteristics and dry/wet mix.

| Parameter | Default | Range | Description |
|---|---|---|---|
| `RoomSize` | 0.3 | [0.0, 1.0] | Size of the virtual room. |
| `Damping` | 0.5 | [0.0, 1.0] | High-frequency absorption. Higher values simulate softer surfaces. |
| `Dry` | 1.0 | [0.0, 1.0] | Level of the unprocessed signal. |
| `Wet` | 0.5 | [0.0, 1.0] | Level of the reverberated signal. |

```go
reverb := api.CreateReverbNode()
defer reverb.Delete()

reverb.SetRoomSize(0.8)
reverb.SetDamping(0.3)
reverb.SetWet(0.4)

api.Chain(source, reverb, api.Output())
```

### CompressorNode

Applies dynamic range compression, attenuating signals that exceed a threshold.

| Parameter | Default | Range | Description |
|---|---|---|---|
| `Threshold` | -24.0 dB | [-100.0, 0.0] | Level above which compression is applied. |
| `Ratio` | 12.0 | [1.0, 20.0] | Compression ratio (input dB : output dB above threshold). |
| `Knee` | 30.0 dB | [0.0, 40.0] | Width of the soft-knee transition around the threshold. |
| `Attack` | 0.003 s | [0.0, 1.0] | Time for compression to engage after the threshold is exceeded. |
| `Release` | 0.25 s | [0.0, 1.0] | Time for compression to disengage after the signal drops below the threshold. |

```go
comp := api.CreateCompressorNode()
defer comp.Delete()

comp.SetThreshold(-18.0)
comp.SetRatio(4.0)
api.Chain(source, comp, api.Output())
```

## Spatial Audio

`SpatialNode` wraps a signal source and applies 3D positional effects relative to the `SpatialListener`. The attenuation model is inverse distance: `gain = 1.0 / max(1.0, distance)`.

```go
// Configure the listener (typically updated each frame to match the camera).
listener := api.SpatialListener()
listener.SetPosition(cameraPosition)
listener.SetRotation(cameraRotation)

// Place a sound source in the world.
spatial := api.CreateSpatialNode()
defer spatial.Delete()

spatial.SetPosition(sprec.Vec3{X: 10, Y: 0, Z: -5})
api.Chain(playbackNode, spatial, api.Output())
```

## No-op Implementation

`NewNopAPI` returns a fully functional but silent implementation. All node factories return working objects that store and return their configured values; no audio is produced. This is useful for headless environments and tests:

```go
api := audio.NewNopAPI()
```
