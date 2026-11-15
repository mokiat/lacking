package sound_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("Effect", func() {
	var (
		api      *fakeAPI
		scene    *sound.Scene
		effects  sound.EffectView
		finished []sound.EffectID
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
		effects = scene.Effects()

		finished = nil
		effects.SubscribeFinished(func(id sound.EffectID) {
			finished = append(finished, id)
		})
	})

	It("starts playing right away", func() {
		id := effects.Play(sound.EffectInfo{Media: newFakeMedia(1.0)})
		Expect(id).NotTo(Equal(sound.NilEffectID))
		Expect(effects.IsValid(id)).To(BeTrue())

		playback := api.lastPlayback()
		Expect(playback.playing).To(BeTrue())
		Expect(playback.looping).To(BeFalse())
		Expect(playback.gain).To(BeNumerically("~", 1.0))
		Expect(playback.rate).To(BeNumerically("~", 1.0))
	})

	It("is not spatial without an emitter", func() {
		effects.Play(sound.EffectInfo{Media: newFakeMedia(1.0)})
		Expect(api.lastPlayback().spatial).To(BeFalse())
	})

	It("is spatial on an emitter", func() {
		emitter := scene.Emitters().Create(dprec.ZeroVec3())
		effects.Play(sound.EffectInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		Expect(api.lastPlayback().spatial).To(BeTrue())
	})

	It("applies the specified gain and pitch", func() {
		effects.Play(sound.EffectInfo{
			Media: newFakeMedia(1.0),
			Gain:  opt.V(0.5),
			Pitch: opt.V(1.5),
		})
		playback := api.lastPlayback()
		Expect(playback.gain).To(BeNumerically("~", 0.5))
		Expect(playback.rate).To(BeNumerically("~", 1.5))
	})

	It("requires media", func() {
		Expect(func() {
			effects.Play(sound.EffectInfo{})
		}).To(Panic())
	})

	It("rejects a deleted emitter", func() {
		emitter := scene.Emitters().Create(dprec.ZeroVec3())
		scene.Emitters().Delete(emitter)
		Expect(func() {
			effects.Play(sound.EffectInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		}).To(Panic())
	})

	It("can have its gain and pitch changed while playing", func() {
		effect := effects.PlayHandle(sound.EffectInfo{Media: newFakeMedia(1.0), Loop: true})
		playback := api.lastPlayback()

		effect.SetGain(0.25)
		effect.SetPitch(0.75)

		Expect(effect.Gain()).To(BeNumerically("~", 0.25))
		Expect(effect.Pitch()).To(BeNumerically("~", 0.75))
		Expect(playback.gain).To(BeNumerically("~", 0.25))
		Expect(playback.rate).To(BeNumerically("~", 0.75))
	})

	Describe("finishing", func() {
		var (
			id       sound.EffectID
			playback *fakePlayback
		)

		BeforeEach(func() {
			id = effects.Play(sound.EffectInfo{Media: newFakeMedia(1.0)})
			playback = api.lastPlayback()
		})

		It("stays valid while playing", func() {
			scene.Update(frame)
			Expect(effects.IsValid(id)).To(BeTrue())
			Expect(finished).To(BeEmpty())
		})

		It("is removed on the update after it has ended", func() {
			playback.finish()
			Expect(finished).To(BeEmpty())

			scene.Update(frame)
			Expect(effects.IsValid(id)).To(BeFalse())
			Expect(playback.released).To(BeTrue())
			Expect(finished).To(Equal([]sound.EffectID{id}))
		})

		It("is reported only once", func() {
			playback.finish()
			scene.Update(frame)
			scene.Update(frame)
			Expect(finished).To(HaveLen(1))
		})

		It("does not confuse a new effect with the finished one", func() {
			playback.finish()
			scene.Update(frame)

			newID := effects.Play(sound.EffectInfo{Media: newFakeMedia(1.0)})
			Expect(newID).NotTo(Equal(id))
			Expect(effects.IsValid(newID)).To(BeTrue())
			Expect(effects.IsValid(id)).To(BeFalse())

			// A late notification for the old playback must not affect the
			// new effect.
			playback.finish()
			scene.Update(frame)
			Expect(effects.IsValid(newID)).To(BeTrue())
		})

		It("allows effects to be played from the finished callback", func() {
			var chained sound.EffectID
			effects.SubscribeFinished(func(sound.EffectID) {
				chained = effects.Play(sound.EffectInfo{Media: newFakeMedia(1.0)})
			})
			playback.finish()
			scene.Update(frame)
			Expect(effects.IsValid(chained)).To(BeTrue())
		})
	})

	Describe("Stop", func() {
		It("stops a looping effect without a notification", func() {
			effect := effects.PlayHandle(sound.EffectInfo{Media: newFakeMedia(1.0), Loop: true})
			playback := api.lastPlayback()
			Expect(playback.looping).To(BeTrue())

			effect.Stop()
			scene.Update(frame)

			Expect(effect.IsValid()).To(BeFalse())
			Expect(playback.playing).To(BeFalse())
			Expect(playback.released).To(BeTrue())
			Expect(finished).To(BeEmpty())
		})

		It("rejects an effect that is no longer playing", func() {
			effect := effects.PlayHandle(sound.EffectInfo{Media: newFakeMedia(1.0)})
			effect.Stop()
			Expect(func() {
				effect.Stop()
			}).To(Panic())
		})
	})
})
