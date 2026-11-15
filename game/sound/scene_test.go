package sound_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/sound"
)

const frame = 16 * time.Millisecond

var _ = Describe("Scene", func() {
	var (
		api   *fakeAPI
		scene *sound.Scene
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
	})

	Describe("active listener", func() {
		It("is not available initially", func() {
			Expect(scene.ActiveListener()).To(Equal(sound.NilListenerID))
		})

		It("defaults to the first listener that is created", func() {
			first := scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
			scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
			Expect(scene.ActiveListener()).To(Equal(first))
		})

		It("can be changed", func() {
			scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
			second := scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
			scene.SetActiveListener(second)
			Expect(scene.ActiveListener()).To(Equal(second))
		})

		It("is cleared when the listener is deleted", func() {
			listener := scene.Listeners().CreateHandle(dprec.ZeroVec3(), dprec.IdentityQuat())
			listener.Delete()
			Expect(listener.IsValid()).To(BeFalse())
			Expect(scene.ActiveListener()).To(Equal(sound.NilListenerID))
		})

		It("rejects a deleted listener", func() {
			listener := scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
			scene.Listeners().Delete(listener)
			Expect(func() {
				scene.SetActiveListener(listener)
			}).To(Panic())
		})

		It("applies its rotation to the audio listener on update", func() {
			rotation := dprec.RotationQuat(dprec.Degrees(90), dprec.BasisYVec3())
			scene.Listeners().Create(dprec.NewVec3(5.0, 0.0, 0.0), rotation)
			scene.Update(frame)
			Expect(api.listener.rotation).To(Equal(dtos.Quat(rotation)))
			Expect(api.listener.position).To(Equal(sprec.ZeroVec3()))
		})
	})

	Describe("spatial placement", func() {
		var (
			listener sound.ListenerHandle
			emitter  sound.EmitterHandle
			playback *fakePlayback
		)

		BeforeEach(func() {
			listener = scene.Listeners().CreateHandle(dprec.NewVec3(1000.0, 0.0, 0.0), dprec.IdentityQuat())
			emitter = scene.Emitters().CreateHandle(dprec.NewVec3(1003.0, 4.0, 0.0))
			scene.Effects().Play(sound.EffectInfo{
				Media:   newFakeMedia(1.0),
				Emitter: emitter.ID(),
			})
			playback = api.lastPlayback()
		})

		It("positions sounds relative to the active listener", func() {
			Expect(playback.spatial).To(BeTrue())
			Expect(playback.position).To(Equal(sprec.NewVec3(3.0, 4.0, 0.0)))
		})

		It("follows the emitter on update", func() {
			emitter.SetPosition(dprec.NewVec3(1000.0, 0.0, -2.0))
			scene.Update(frame)
			Expect(playback.position).To(Equal(sprec.NewVec3(0.0, 0.0, -2.0)))
		})

		It("follows the listener on update", func() {
			listener.SetPosition(dprec.NewVec3(1003.0, 0.0, 0.0))
			scene.Update(frame)
			Expect(playback.position).To(Equal(sprec.NewVec3(0.0, 4.0, 0.0)))
		})

		It("uses the world origin when there is no active listener", func() {
			scene.SetActiveListener(sound.NilListenerID)
			scene.Update(frame)
			Expect(playback.position).To(Equal(sprec.NewVec3(1003.0, 4.0, 0.0)))
		})
	})

	Describe("Delete", func() {
		It("stops and releases everything", func() {
			emitter := scene.Emitters().Create(dprec.ZeroVec3())
			scene.Effects().Play(sound.EffectInfo{Media: newFakeMedia(1.0), Emitter: emitter})
			scene.Speech().Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})
			scene.Ambience().Play(sound.AmbienceInfo{Media: newFakeMedia(1.0)})
			scene.Music().Enqueue(sound.MusicTrackInfo{Media: newFakeMedia(1.0)})
			scene.EffectPools().Create(sound.EffectPoolInfo{Media: newFakeMedia(1.0), Capacity: 2})
			scene.Update(frame)

			scene.Delete()

			Expect(api.playbacks).To(HaveLen(6))
			for _, playback := range api.playbacks {
				Expect(playback.playing).To(BeFalse())
				Expect(playback.released).To(BeTrue())
			}
			Expect(api.buses).To(HaveLen(4))
			for _, bus := range api.buses {
				Expect(bus.released).To(BeTrue())
			}
		})
	})
})
