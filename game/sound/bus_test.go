package sound_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("Mixing controls", func() {
	var (
		api   *fakeAPI
		scene *sound.Scene
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
	})

	// Each entry provides the controls of a category and starts a sound in
	// that category, so that the bus backing it can be looked up.
	DescribeTable("per category",
		func(setup func() (sound.BusView, *fakeBus)) {
			controls, bus := setup()

			By("starting off at unit gain")
			Expect(controls.Gain()).To(BeNumerically("~", 1.0))
			Expect(controls.Volume()).To(BeNumerically("~", 1.0))
			Expect(bus.gain).To(BeNumerically("~", 1.0))

			By("changing the gain")
			controls.SetGain(0.25)
			Expect(controls.Gain()).To(BeNumerically("~", 0.25))
			Expect(controls.Volume()).To(BeNumerically("~", 0.5))
			Expect(bus.gain).To(BeNumerically("~", 0.25))

			By("changing the volume")
			controls.SetVolume(0.8)
			Expect(controls.Volume()).To(BeNumerically("~", 0.8, 1e-6))
			Expect(controls.Gain()).To(BeNumerically("~", 0.64, 1e-6))
			Expect(bus.gain).To(BeNumerically("~", 0.64, 1e-6))

			By("clamping negative values")
			controls.SetGain(-1.0)
			Expect(controls.Gain()).To(BeZero())
			controls.SetGain(0.5)

			By("muting")
			controls.SetMuted(true)
			Expect(controls.Muted()).To(BeTrue())
			Expect(controls.Gain()).To(BeNumerically("~", 0.5))
			Expect(bus.gain).To(BeZero())
			controls.SetMuted(false)
			Expect(bus.gain).To(BeNumerically("~", 0.5))

			By("pausing")
			controls.Pause()
			Expect(controls.Paused()).To(BeTrue())
			Expect(bus.paused).To(BeTrue())
			controls.Resume()
			Expect(controls.Paused()).To(BeFalse())
			Expect(bus.paused).To(BeFalse())
		},
		Entry("music", func() (sound.BusView, *fakeBus) {
			scene.Music().Enqueue(sound.MusicTrackInfo{Media: newFakeMedia(1.0)})
			return scene.Music().Bus(), api.lastPlayback().bus
		}),
		Entry("speech", func() (sound.BusView, *fakeBus) {
			scene.Speech().Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})
			return scene.Speech().Bus(), api.lastPlayback().bus
		}),
		Entry("effects", func() (sound.BusView, *fakeBus) {
			scene.Effects().Play(sound.EffectInfo{Media: newFakeMedia(1.0)})
			return scene.Effects().Bus(), api.lastPlayback().bus
		}),
		Entry("pooled effects", func() (sound.BusView, *fakeBus) {
			scene.EffectPools().Create(sound.EffectPoolInfo{Media: newFakeMedia(1.0), Capacity: 1})
			return scene.Effects().Bus(), api.lastPlayback().bus
		}),
		Entry("ambience", func() (sound.BusView, *fakeBus) {
			scene.Ambience().Play(sound.AmbienceInfo{Media: newFakeMedia(1.0)})
			return scene.Ambience().Bus(), api.lastPlayback().bus
		}),
	)

	It("uses a separate bus for each category", func() {
		scene.Music().Enqueue(sound.MusicTrackInfo{Media: newFakeMedia(1.0)})
		scene.Speech().Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})
		scene.Effects().Play(sound.EffectInfo{Media: newFakeMedia(1.0)})
		scene.Ambience().Play(sound.AmbienceInfo{Media: newFakeMedia(1.0)})

		buses := make(map[*fakeBus]struct{})
		for _, playback := range api.playbacks {
			buses[playback.bus] = struct{}{}
		}
		Expect(buses).To(HaveLen(4))
	})
})
