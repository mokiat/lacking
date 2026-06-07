// Package wav provides a WAV audio decoder for the audio package.
// Importing this package is sufficient to register the decoder — the init
// function calls [audio.RegisterDecoder] so that [audio.Decode] can handle
// WAV data identified by the "RIFF" magic prefix.
package wav
