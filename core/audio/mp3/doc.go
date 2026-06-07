// Package mp3 provides an MP3 audio decoder for the audio package.
// Importing this package is sufficient to register the decoder — the init
// function calls [audio.RegisterDecoder] so that [audio.Decode] can handle
// MP3 data identified by the "ID3" magic prefix.
package mp3
