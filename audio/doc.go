// Package audio defines the audio API used by the engine.
//
// The API is built around a node graph model. Audio sources (e.g.
// [PlaybackNode], [OscillatorNode]) produce signals that flow through
// processing nodes (e.g. [GainNode], [ReverbNode]) and ultimately reach the
// output node returned by [API.Output]. Nodes are connected with [API.Connect]
// and disconnected with [API.Disconnect]. When multiple sources are connected
// to the same target their signals are mixed additively.
//
// All created nodes implement [UserNode] and must be explicitly deleted via
// [UserNode.Delete] when no longer needed, otherwise resources will leak.
//
// The nop implementation ([NewNopAPI]) provides a fully functional but silent
// API suitable for headless operation and testing.
package audio
