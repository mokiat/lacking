package sound_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("EffectPool", func() {
	var (
		api     *fakeAPI
		scene   *sound.Scene
		emitter sound.EmitterID
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
		scene.Listeners().Create(dprec.ZeroVec3(), dprec.IdentityQuat())
		emitter = scene.Emitters().Create(dprec.NewVec3(1.0, 2.0, 3.0))
	})

	It("requires media and a positive capacity", func() {
		Expect(func() {
			scene.EffectPools().Create(sound.EffectPoolInfo{Capacity: 1})
		}).To(Panic())
		Expect(func() {
			scene.EffectPools().Create(sound.EffectPoolInfo{Media: newFakeMedia(1.0)})
		}).To(Panic())
	})

	Describe("spatial pool", func() {
		var pool sound.EffectPoolHandle

		BeforeEach(func() {
			pool = scene.EffectPools().CreateHandle(sound.EffectPoolInfo{
				Media:    newFakeMedia(1.0),
				Capacity: 2,
				Gain:     opt.V(0.5),
				Pitch:    opt.V(2.0),
			})
		})

		It("allocates all of its playbacks up front", func() {
			Expect(api.playbacks).To(HaveLen(2))
			for _, playback := range api.playbacks {
				Expect(playback.spatial).To(BeTrue())
				Expect(playback.playing).To(BeFalse())
			}
			Expect(pool.PlayingCount()).To(BeZero())
		})

		It("plays on the specified emitter", func() {
			Expect(pool.Play(emitter)).To(BeTrue())
			Expect(pool.PlayingCount()).To(Equal(1))

			playback := api.playbacks[0]
			Expect(playback.playing).To(BeTrue())
			Expect(playback.position).To(Equal(sprec.NewVec3(1.0, 2.0, 3.0)))
			Expect(playback.gain).To(BeNumerically("~", 0.5))
			Expect(playback.rate).To(BeNumerically("~", 2.0))
		})

		It("follows the emitter on update", func() {
			pool.Play(emitter)
			scene.Emitters().SetPosition(emitter, dprec.NewVec3(4.0, 5.0, 6.0))
			scene.Update(frame)
			Expect(api.playbacks[0].position).To(Equal(sprec.NewVec3(4.0, 5.0, 6.0)))
		})

		It("does not create audio resources when playing", func() {
			for range 10 {
				pool.Play(emitter)
			}
			Expect(api.playbacks).To(HaveLen(2))
		})

		It("does not allocate memory when playing", func() {
			allocs := testing.AllocsPerRun(100, func() {
				pool.Play(emitter)
			})
			Expect(allocs).To(BeZero())
		})

		It("uses a separate playback for overlapping plays", func() {
			pool.Play(emitter)
			pool.Play(emitter)
			Expect(api.playbacks[0].startCount).To(Equal(1))
			Expect(api.playbacks[1].startCount).To(Equal(1))
			Expect(pool.PlayingCount()).To(Equal(2))
		})

		It("steals the oldest playback when all are in use", func() {
			pool.Play(emitter)
			pool.Play(emitter)

			Expect(pool.Play(emitter)).To(BeTrue())
			Expect(api.playbacks[0].startCount).To(Equal(2))
			Expect(api.playbacks[1].startCount).To(Equal(1))

			Expect(pool.Play(emitter)).To(BeTrue())
			Expect(api.playbacks[0].startCount).To(Equal(2))
			Expect(api.playbacks[1].startCount).To(Equal(2))
		})

		It("reuses a playback that has finished", func() {
			pool.Play(emitter)
			pool.Play(emitter)
			api.playbacks[1].finish()
			Expect(pool.PlayingCount()).To(Equal(1))

			pool.Play(emitter)
			Expect(api.playbacks[0].startCount).To(Equal(1))
			Expect(api.playbacks[1].startCount).To(Equal(2))
		})

		It("requires a valid emitter", func() {
			Expect(func() {
				pool.Play(sound.NilEmitterID)
			}).To(Panic())
		})

		It("stops the playbacks of an emitter that is deleted", func() {
			other := scene.Emitters().Create(dprec.ZeroVec3())
			pool.Play(emitter)
			pool.Play(other)

			scene.Emitters().Delete(emitter)
			scene.Update(frame)

			Expect(api.playbacks[0].playing).To(BeFalse())
			Expect(api.playbacks[1].playing).To(BeTrue())
			Expect(pool.PlayingCount()).To(Equal(1))
		})

		It("can stop all of its playbacks", func() {
			pool.Play(emitter)
			pool.Play(emitter)
			pool.StopAll()
			Expect(pool.PlayingCount()).To(BeZero())
			Expect(api.playbacks[0].playing).To(BeFalse())
			Expect(api.playbacks[1].playing).To(BeFalse())
		})

		It("releases its playbacks when deleted", func() {
			pool.Play(emitter)
			pool.Delete()

			Expect(pool.IsValid()).To(BeFalse())
			for _, playback := range api.playbacks {
				Expect(playback.playing).To(BeFalse())
				Expect(playback.released).To(BeTrue())
			}
			Expect(func() {
				pool.Play(emitter)
			}).To(Panic())
		})

		It("can be enumerated", func() {
			var ids []sound.EffectPoolID
			scene.EffectPools().Each(func(id sound.EffectPoolID) {
				ids = append(ids, id)
			})
			Expect(ids).To(ConsistOf(pool.ID()))
		})
	})

	It("can skip plays when all playbacks are in use", func() {
		pool := scene.EffectPools().CreateHandle(sound.EffectPoolInfo{
			Media:    newFakeMedia(1.0),
			Capacity: 1,
			Overflow: sound.EffectPoolOverflowSkip,
		})
		Expect(pool.Play(emitter)).To(BeTrue())
		Expect(pool.Play(emitter)).To(BeFalse())
		Expect(api.playbacks[0].startCount).To(Equal(1))

		api.playbacks[0].finish()
		Expect(pool.Play(emitter)).To(BeTrue())
		Expect(api.playbacks[0].startCount).To(Equal(2))
	})

	It("can play without an emitter when not spatial", func() {
		pool := scene.EffectPools().CreateHandle(sound.EffectPoolInfo{
			Media:    newFakeMedia(1.0),
			Capacity: 1,
			Spatial:  opt.V(false),
		})
		Expect(api.playbacks[0].spatial).To(BeFalse())
		Expect(pool.Play(sound.NilEmitterID)).To(BeTrue())
		Expect(api.playbacks[0].playing).To(BeTrue())

		scene.Update(frame)
		Expect(api.playbacks[0].playing).To(BeTrue())
	})

	It("varies the gain and pitch within the specified bounds", func() {
		pool := scene.EffectPools().CreateHandle(sound.EffectPoolInfo{
			Media:         newFakeMedia(1.0),
			Capacity:      1,
			Gain:          opt.V(0.5),
			Pitch:         opt.V(2.0),
			GainVariance:  0.2,
			PitchVariance: 0.1,
		})
		playback := api.playbacks[0]

		gains := make(map[float32]struct{})
		for range 100 {
			pool.Play(emitter)
			Expect(playback.gain).To(BeNumerically(">=", 0.4-1e-4))
			Expect(playback.gain).To(BeNumerically("<=", 0.6+1e-4))
			Expect(playback.rate).To(BeNumerically(">=", 1.8-1e-4))
			Expect(playback.rate).To(BeNumerically("<=", 2.2+1e-4))
			gains[playback.gain] = struct{}{}
		}
		Expect(len(gains)).To(BeNumerically(">", 1))
	})
})
