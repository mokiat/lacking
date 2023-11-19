package audio

import "time"

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

type nopMedia struct{}

func (m *nopMedia) Length() time.Duration {
	return time.Millisecond
}

func (m *nopMedia) Delete() {}

type nopPlayback struct{}

func (p *nopPlayback) Stop() {}
