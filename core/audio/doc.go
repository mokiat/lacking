// Package audio provides a platform-agnostic audio API covering media loading,
// playback control, spatial (3D) audio, and a format decoder registry.
// Platform-specific implementations satisfy the [API] interface; a no-op
// implementation ([NewNopAPI]) is available for headless or test operation.
// Format decoder plugins (e.g. the mp3 and wav sub-packages) self-register
// via their package init functions and are selected at decode time by
// magic-byte prefix matching.
package audio
