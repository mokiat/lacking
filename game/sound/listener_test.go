package sound_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("Listener", func() {
	var (
		api      *fakeAPI
		scene    *sound.Scene
		rotation dprec.Quat
		listener sound.ListenerHandle
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
		rotation = dprec.RotationQuat(dprec.Degrees(90), dprec.BasisYVec3())
		listener = scene.Listeners().CreateHandle(dprec.NewVec3(1.0, 2.0, 3.0), rotation)
	})

	It("starts with what it was created with", func() {
		Expect(listener.IsValid()).To(BeTrue())
		Expect(listener.ID()).ToNot(Equal(sound.NilListenerID))
		Expect(listener.Position()).To(Equal(dprec.NewVec3(1.0, 2.0, 3.0)))
		Expect(listener.Rotation()).To(Equal(rotation))
		Expect(listener.Velocity()).To(Equal(dprec.ZeroVec3()))
	})

	It("can be moved", func() {
		listener.SetPosition(dprec.NewVec3(-4.0, 5.0, 6.0))
		Expect(listener.Position()).To(Equal(dprec.NewVec3(-4.0, 5.0, 6.0)))
	})

	It("can be rotated", func() {
		newRotation := dprec.RotationQuat(dprec.Degrees(-45), dprec.BasisXVec3())
		listener.SetRotation(newRotation)
		Expect(listener.Rotation()).To(Equal(newRotation))
	})

	It("can be given a velocity", func() {
		listener.SetVelocity(dprec.NewVec3(0.0, 0.0, -20.0))
		Expect(listener.Velocity()).To(Equal(dprec.NewVec3(0.0, 0.0, -20.0)))

		By("not being moved by it")
		scene.Update(frame)
		Expect(listener.Position()).To(Equal(dprec.NewVec3(1.0, 2.0, 3.0)))
	})

	It("can be accessed through the view", func() {
		listeners := scene.Listeners()
		id := listener.ID()

		listeners.SetPosition(id, dprec.NewVec3(7.0, 8.0, 9.0))
		listeners.SetRotation(id, dprec.IdentityQuat())
		listeners.SetVelocity(id, dprec.NewVec3(1.0, 0.0, 0.0))

		Expect(listeners.IsValid(id)).To(BeTrue())
		Expect(listeners.Position(id)).To(Equal(dprec.NewVec3(7.0, 8.0, 9.0)))
		Expect(listeners.Rotation(id)).To(Equal(dprec.IdentityQuat()))
		Expect(listeners.Velocity(id)).To(Equal(dprec.NewVec3(1.0, 0.0, 0.0)))

		By("seeing the same listener through a handle")
		handle := listeners.Handle(id)
		Expect(handle.ID()).To(Equal(id))
		Expect(handle.Position()).To(Equal(dprec.NewVec3(7.0, 8.0, 9.0)))
	})

	It("can be enumerated", func() {
		other := scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
		deleted := scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
		scene.Listeners().Delete(deleted)

		var ids []sound.ListenerID
		scene.Listeners().Each(func(id sound.ListenerID) {
			ids = append(ids, id)
		})
		Expect(ids).To(ConsistOf(listener.ID(), other))
	})

	Describe("validity", func() {
		It("does not consider the nil ID valid", func() {
			Expect(scene.Listeners().IsValid(sound.NilListenerID)).To(BeFalse())
		})

		It("is lost on deletion", func() {
			id := listener.ID()
			listener.Delete()
			Expect(listener.IsValid()).To(BeFalse())
			Expect(scene.Listeners().IsValid(id)).To(BeFalse())
		})

		It("is not regained when another listener takes its place", func() {
			id := listener.ID()
			listener.Delete()
			replacement := scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
			Expect(replacement).ToNot(Equal(id))
			Expect(scene.Listeners().IsValid(replacement)).To(BeTrue())
			Expect(scene.Listeners().IsValid(id)).To(BeFalse())
		})

		It("is required in order to access a listener", func() {
			listener.Delete()
			Expect(func() { listener.Position() }).To(Panic())
			Expect(func() { listener.SetPosition(dprec.ZeroVec3()) }).To(Panic())
			Expect(func() { listener.Rotation() }).To(Panic())
			Expect(func() { listener.SetRotation(dprec.IdentityQuat()) }).To(Panic())
			Expect(func() { listener.Velocity() }).To(Panic())
			Expect(func() { listener.SetVelocity(dprec.ZeroVec3()) }).To(Panic())
			Expect(func() { listener.Delete() }).To(Panic())
			Expect(func() { scene.Listeners().Position(sound.NilListenerID) }).To(Panic())
		})
	})

	Describe("hearing", func() {
		var (
			other    sound.ListenerHandle
			playback *fakePlayback
		)

		BeforeEach(func() {
			other = scene.Listeners().CreateHandle(dprec.NewVec3(100.0, 0.0, 0.0), dprec.IdentityQuat())
			emitter := scene.Emitters().Create(dprec.NewVec3(11.0, 2.0, 3.0))
			scene.Effects().Play(sound.EffectInfo{
				Media:   newFakeMedia(1.0),
				Emitter: emitter,
				Loop:    true,
			})
			playback = api.lastPlayback()
			scene.Update(frame)
		})

		It("is done from the listener that was created first", func() {
			Expect(scene.ActiveListener()).To(Equal(listener.ID()))
			Expect(api.listener.rotation).To(Equal(dtos.Quat(rotation)))
			Expect(playback.position).To(Equal(sprec.NewVec3(10.0, 0.0, 0.0)))
		})

		It("follows the rotation of the active listener", func() {
			newRotation := dprec.RotationQuat(dprec.Degrees(180), dprec.BasisYVec3())
			listener.SetRotation(newRotation)
			scene.Update(frame)
			Expect(api.listener.rotation).To(Equal(dtos.Quat(newRotation)))
		})

		It("follows the position of the active listener", func() {
			listener.SetPosition(dprec.NewVec3(11.0, 2.0, 8.0))
			scene.Update(frame)
			Expect(playback.position).To(Equal(sprec.NewVec3(0.0, 0.0, -5.0)))
		})

		It("is not affected by a listener that is not active", func() {
			other.SetPosition(dprec.NewVec3(11.0, 2.0, 3.0))
			other.SetRotation(dprec.RotationQuat(dprec.Degrees(180), dprec.BasisYVec3()))
			other.SetVelocity(dprec.NewVec3(50.0, 0.0, 0.0))
			scene.Update(frame)
			Expect(api.listener.rotation).To(Equal(dtos.Quat(rotation)))
			Expect(playback.position).To(Equal(sprec.NewVec3(10.0, 0.0, 0.0)))
		})

		It("moves over to a listener that is made active", func() {
			scene.SetActiveListener(other.ID())
			scene.Update(frame)
			Expect(api.listener.rotation).To(Equal(sprec.IdentityQuat()))
			Expect(playback.position).To(Equal(sprec.NewVec3(-89.0, 2.0, 3.0)))
		})

		It("does not move over to another listener when the active one is deleted", func() {
			listener.Delete()
			scene.Update(frame)
			Expect(scene.ActiveListener()).To(Equal(sound.NilListenerID))
			Expect(playback.position).To(Equal(sprec.NewVec3(11.0, 2.0, 3.0)))
		})
	})
})
