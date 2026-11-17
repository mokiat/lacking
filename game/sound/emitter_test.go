package sound_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("Emitter", func() {
	var (
		api      *fakeAPI
		scene    *sound.Scene
		listener sound.ListenerHandle
		emitter  sound.EmitterHandle
		effect   sound.EffectHandle
		playback *fakePlayback
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
		listener = scene.Listeners().CreateHandle(dprec.ZeroVec3(), dprec.IdentityQuat())

		// The emitter is placed 100 units along +X from the listener, so a
		// negative X velocity moves it towards the listener.
		emitter = scene.Emitters().CreateHandle(dprec.NewVec3(100.0, 0.0, 0.0))
		effect = scene.Effects().PlayHandle(sound.EffectInfo{
			Media:   newFakeMedia(1.0),
			Emitter: emitter.ID(),
			Loop:    true,
		})
		playback = api.lastPlayback()
	})

	It("has sensible defaults", func() {
		Expect(emitter.IsValid()).To(BeTrue())
		Expect(emitter.Position()).To(Equal(dprec.NewVec3(100.0, 0.0, 0.0)))
		Expect(emitter.Rotation()).To(Equal(dprec.IdentityQuat()))
		Expect(emitter.Velocity()).To(Equal(dprec.ZeroVec3()))
		Expect(emitter.Gain()).To(BeNumerically("~", 1.0))
		Expect(emitter.Doppler()).To(BeFalse())
		Expect(emitter.OuterConeAngle()).To(Equal(dprec.Degrees(360.0)))
	})

	It("can be enumerated", func() {
		other := scene.Emitters().Create(dprec.ZeroVec3())
		var ids []sound.EmitterID
		scene.Emitters().Each(func(id sound.EmitterID) {
			ids = append(ids, id)
		})
		Expect(ids).To(ConsistOf(emitter.ID(), other))
	})

	It("does not apply the Doppler effect by default", func() {
		emitter.SetVelocity(dprec.NewVec3(-34.3, 0.0, 0.0))
		listener.SetVelocity(dprec.NewVec3(34.3, 0.0, 0.0))
		scene.Update(frame)
		Expect(playback.rate).To(BeNumerically("~", 1.0))
	})

	Describe("Doppler effect", func() {
		BeforeEach(func() {
			emitter.SetDoppler(true)
		})

		It("does not change the rate when nothing moves", func() {
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 1.0))
		})

		It("raises the rate when the emitter approaches", func() {
			emitter.SetVelocity(dprec.NewVec3(-34.3, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 343.0/(343.0-34.3), 1e-4))
		})

		It("lowers the rate when the emitter recedes", func() {
			emitter.SetVelocity(dprec.NewVec3(34.3, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 343.0/(343.0+34.3), 1e-4))
		})

		It("raises the rate when the listener approaches", func() {
			listener.SetVelocity(dprec.NewVec3(34.3, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", (343.0+34.3)/343.0, 1e-4))
		})

		It("ignores motion that is perpendicular to the line of sight", func() {
			emitter.SetVelocity(dprec.NewVec3(0.0, 50.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 1.0))
		})

		It("combines with the pitch of the sound", func() {
			effect.SetPitch(2.0)
			emitter.SetVelocity(dprec.NewVec3(-34.3, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 2.0*343.0/(343.0-34.3), 1e-4))
		})

		It("limits the rate for supersonic emitters", func() {
			emitter.SetVelocity(dprec.NewVec3(-1000.0, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 2.0))

			emitter.SetVelocity(dprec.NewVec3(5000.0, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 0.5))
		})

		It("can be disabled per emitter", func() {
			emitter.SetDoppler(false)
			emitter.SetVelocity(dprec.NewVec3(-34.3, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 1.0))
		})

		It("can be disabled for the whole scene", func() {
			scene.SetDopplerFactor(0.0)
			emitter.SetVelocity(dprec.NewVec3(-34.3, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 1.0))
		})

		It("honors the speed of sound", func() {
			scene.SetSpeedOfSound(100.0)
			emitter.SetVelocity(dprec.NewVec3(-20.0, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 100.0/80.0, 1e-4))
		})

		It("is not applied when the emitter is at the listener", func() {
			emitter.SetPosition(dprec.ZeroVec3())
			emitter.SetVelocity(dprec.NewVec3(-34.3, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.rate).To(BeNumerically("~", 1.0))
		})
	})

	It("scales the gain of its sounds", func() {
		effect.SetGain(0.5)
		emitter.SetGain(0.5)
		scene.Update(frame)
		Expect(playback.gain).To(BeNumerically("~", 0.25))
	})

	It("applies its orientation and cone to its sounds", func() {
		rotation := dprec.RotationQuat(dprec.Degrees(90), dprec.BasisYVec3())
		emitter.SetRotation(rotation)
		emitter.SetInnerConeAngle(dprec.Degrees(30.0))
		emitter.SetOuterConeAngle(dprec.Degrees(90.0))
		emitter.SetOuterConeGain(0.25)
		scene.Update(frame)

		Expect(playback.rotation).To(Equal(dtos.Quat(rotation)))
		Expect(playback.innerConeAngle).To(Equal(sprec.Degrees(30.0)))
		Expect(playback.outerConeAngle).To(Equal(sprec.Degrees(90.0)))
		Expect(playback.outerConeGain).To(BeNumerically("~", 0.25))
	})

	Describe("Delete", func() {
		It("invalidates the emitter", func() {
			emitter.Delete()
			Expect(emitter.IsValid()).To(BeFalse())
			Expect(func() {
				emitter.Position()
			}).To(Panic())
		})

		It("stops the sounds that are playing on it", func() {
			ambience := scene.Ambience().PlayHandle(sound.AmbienceInfo{
				Media:   newFakeMedia(1.0),
				Emitter: emitter.ID(),
			})
			ambiencePlayback := api.lastPlayback()

			emitter.Delete()

			Expect(effect.IsValid()).To(BeFalse())
			Expect(playback.playing).To(BeFalse())
			Expect(playback.released).To(BeTrue())
			Expect(ambience.IsValid()).To(BeFalse())
			Expect(ambiencePlayback.released).To(BeTrue())
		})

		It("leaves the sounds of other emitters playing", func() {
			other := scene.Emitters().Create(dprec.ZeroVec3())
			scene.Emitters().Delete(other)
			scene.Update(frame)
			Expect(effect.IsValid()).To(BeTrue())
			Expect(playback.playing).To(BeTrue())
		})

		It("does not confuse a new emitter with the deleted one", func() {
			oldID := emitter.ID()
			emitter.Delete()
			newID := scene.Emitters().Create(dprec.ZeroVec3())
			Expect(newID).NotTo(Equal(oldID))
			Expect(scene.Emitters().IsValid(oldID)).To(BeFalse())
			Expect(scene.Emitters().IsValid(newID)).To(BeTrue())
		})
	})
})
