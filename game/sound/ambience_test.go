package sound_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("Ambience", func() {
	var (
		api   *fakeAPI
		scene *sound.Scene
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
	})

	It("loops its media", func() {
		ambience := scene.Ambience().PlayHandle(sound.AmbienceInfo{
			Media: newFakeMedia(1.0),
			Gain:  opt.V(0.5),
		})
		playback := api.lastPlayback()

		Expect(ambience.IsValid()).To(BeTrue())
		Expect(ambience.Gain()).To(BeNumerically("~", 0.5))
		Expect(playback.playing).To(BeTrue())
		Expect(playback.looping).To(BeTrue())
		Expect(playback.spatial).To(BeFalse())
		Expect(playback.gain).To(BeNumerically("~", 0.5))
	})

	It("is spatial on an emitter", func() {
		emitter := scene.Emitters().Create(dprec.ZeroVec3())
		scene.Ambience().Play(sound.AmbienceInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		Expect(api.lastPlayback().spatial).To(BeTrue())
	})

	It("requires media", func() {
		Expect(func() {
			scene.Ambience().Play(sound.AmbienceInfo{})
		}).To(Panic())
	})

	It("can have its gain changed while playing", func() {
		ambience := scene.Ambience().PlayHandle(sound.AmbienceInfo{Media: newFakeMedia(1.0)})
		ambience.SetGain(0.25)
		Expect(ambience.Gain()).To(BeNumerically("~", 0.25))
		Expect(api.lastPlayback().gain).To(BeNumerically("~", 0.25))
	})

	It("can be faded in", func() {
		scene.Ambience().Play(sound.AmbienceInfo{
			Media:  newFakeMedia(1.0),
			Gain:   opt.V(0.8),
			FadeIn: 2 * time.Second,
		})
		playback := api.lastPlayback()
		Expect(playback.playing).To(BeTrue())
		Expect(playback.gain).To(BeNumerically("~", 0.0))

		scene.Update(time.Second)
		Expect(playback.gain).To(BeNumerically("~", 0.4, 1e-4))

		scene.Update(time.Second)
		Expect(playback.gain).To(BeNumerically("~", 0.8, 1e-4))

		scene.Update(time.Second)
		Expect(playback.gain).To(BeNumerically("~", 0.8, 1e-4))
	})

	It("can be stopped right away", func() {
		ambience := scene.Ambience().PlayHandle(sound.AmbienceInfo{Media: newFakeMedia(1.0)})
		playback := api.lastPlayback()

		ambience.Stop(0)

		Expect(ambience.IsValid()).To(BeFalse())
		Expect(playback.playing).To(BeFalse())
		Expect(playback.released).To(BeTrue())
	})

	It("can be faded out", func() {
		ambience := scene.Ambience().PlayHandle(sound.AmbienceInfo{Media: newFakeMedia(1.0)})
		playback := api.lastPlayback()

		ambience.Stop(2 * time.Second)
		Expect(ambience.IsValid()).To(BeTrue())
		Expect(playback.playing).To(BeTrue())

		scene.Update(time.Second)
		Expect(ambience.IsValid()).To(BeTrue())
		Expect(playback.gain).To(BeNumerically("~", 0.5, 1e-4))

		scene.Update(time.Second)
		Expect(ambience.IsValid()).To(BeFalse())
		Expect(playback.playing).To(BeFalse())
		Expect(playback.released).To(BeTrue())
	})

	It("fades out from the level it has reached while fading in", func() {
		ambience := scene.Ambience().PlayHandle(sound.AmbienceInfo{
			Media:  newFakeMedia(1.0),
			FadeIn: 2 * time.Second,
		})
		playback := api.lastPlayback()
		scene.Update(time.Second) // half way in

		ambience.Stop(2 * time.Second)
		scene.Update(time.Second)
		Expect(playback.gain).To(BeNumerically("~", 0.25, 1e-4))
	})

	It("does not progress fades while paused", func() {
		ambience := scene.Ambience().PlayHandle(sound.AmbienceInfo{Media: newFakeMedia(1.0)})
		ambience.Stop(time.Second)

		scene.Ambience().Bus().Pause()
		scene.Update(time.Minute)
		Expect(ambience.IsValid()).To(BeTrue())

		scene.Ambience().Bus().Resume()
		scene.Update(time.Second)
		Expect(ambience.IsValid()).To(BeFalse())
	})

	It("can stop all ambiences", func() {
		first := scene.Ambience().PlayHandle(sound.AmbienceInfo{Media: newFakeMedia(1.0)})
		second := scene.Ambience().PlayHandle(sound.AmbienceInfo{Media: newFakeMedia(1.0)})

		scene.Ambience().StopAll(time.Second)
		Expect(first.IsValid()).To(BeTrue())
		Expect(second.IsValid()).To(BeTrue())

		scene.Update(time.Second)
		Expect(first.IsValid()).To(BeFalse())
		Expect(second.IsValid()).To(BeFalse())
	})
})
