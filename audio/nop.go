package audio

import "time"

// NewNopAPI returns an API that does nothing.
func NewNopAPI() API {
	return &nopAPI{}
}

type nopAPI struct{}

func (a *nopAPI) CreateMedia(info MediaInfo) Media {
	return &nopMedia{}
}

func (a *nopAPI) Play(media Media, info PlayInfo) Playback {
	return &nopPlayback{}
}

func (a *nopAPI) CreatePlayback(media Media, loop bool) PlaybackNode {
	return nil
}

func (a *nopAPI) CreateOscillator() OscillatorNode {
	return nil
}

func (a *nopAPI) CreateGain() GainNode {
	return nil
}

func (a *nopAPI) CreatePan() PanNode {
	return nil
}

func (a *nopAPI) Chain(nodes ...Node) {}

func (a *nopAPI) Connect(source, target Node) {}

func (a *nopAPI) Disconnect(source, target Node) {}

func (a *nopAPI) Output() Node {
	return nil
}

type nopMedia struct{}

func (m *nopMedia) Length() time.Duration {
	return time.Millisecond
}

func (m *nopMedia) Delete() {}

type nopPlayback struct{}

func (p *nopPlayback) Stop() {}
