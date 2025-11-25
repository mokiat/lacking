package ui

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/audio"
)

var globalAudioGain = 1.0

func GlobalAudioGain() float64 {
	return globalAudioGain
}

func SetGlobalAudioGain(gain float64) {
	globalAudioGain = gain
}

func newSound(api audio.API, media audio.Media) *Sound {
	return &Sound{
		api:   api,
		media: media,
	}
}

type Sound struct {
	api   audio.API
	media audio.Media
}

func (s *Sound) Play(gain float64) {
	if s == nil {
		return
	}
	s.api.Play(s.media, audio.PlayInfo{
		Loop: false,
		Gain: opt.V(gain * globalAudioGain),
		Pan:  0.0,
	})
}
