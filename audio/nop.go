package audio

import (
	"time"

	"github.com/mokiat/gomath/sprec"
)

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

func (a *nopAPI) CreatePlaybackNode(media Media, loop bool) PlaybackNode {
	return nil
}

func (a *nopAPI) CreateOscillatorNode() OscillatorNode {
	return nil
}

func (a *nopAPI) CreateGainNode() GainNode {
	return nil
}

func (a *nopAPI) CreatePanNode() PanNode {
	return nil
}

func (a *nopAPI) CreateSpatialNode() SpatialNode {
	return nil
}

func (a *nopAPI) Chain(nodes ...Node) {}

func (a *nopAPI) Connect(source, target Node) {}

func (a *nopAPI) Disconnect(source, target Node) {}

func (a *nopAPI) SpatialListener() SpatialListener {
	return &nopListener{}
}

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

type nopListener struct{}

func (l *nopListener) Position() sprec.Vec3 {
	return sprec.ZeroVec3()
}

func (l *nopListener) SetPosition(position sprec.Vec3) {}

func (l *nopListener) Rotation() sprec.Quat {
	return sprec.IdentityQuat()
}

func (l *nopListener) SetRotation(rotation sprec.Quat) {}
