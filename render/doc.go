// Package render exposes an abstraction API over the GPU so that various
// implementations (e.g. WebGPU, OpenGL, WebGL) can be substituted.
//
// The API is heavily inspired by the WebGPU API but takes a step back
// towards the OpenGL API in some areas in order to make it easier to
// implement on top of existing APIs.
package render
