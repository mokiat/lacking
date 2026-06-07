package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/audio"
)

type Mixer struct {
	api audio.API
	bus audio.Bus
}

func newMixer(api audio.API) *Mixer {
	return &Mixer{
		api: api,
		bus: api.CreateBus(audio.BusSettings{}),
	}
}

func (m *Mixer) Gain() float32 {
	return m.bus.Gain()
}

func (m *Mixer) SetGain(gain float32) {
	m.bus.SetGain(gain)
}

func (m *Mixer) Volume() float32 {
	gain := max(0.0, m.Gain())
	return sprec.Pow(gain, 1.0/2.0)
}

func (m *Mixer) SetVolume(volume float32) {
	volume = max(0.0, volume)
	gain := sprec.Pow(volume, 2.0)
	m.SetGain(gain)
}

func (m *Mixer) PlaySound(sound *Sound) {
	playback := m.api.CreatePlayback(m.bus, sound.media, audio.PlaybackSettings{})
	playback.SetOnFinished(func() {
		playback.Release()
	})
	playback.Start(0.0)
}

func newSound(media audio.Media) *Sound {
	return &Sound{
		media: media,
	}
}

type Sound struct {
	media audio.Media
}
