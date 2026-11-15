package sound_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("Speech", func() {
	type finishedEvent struct {
		id        sound.SpeechID
		completed bool
	}

	var (
		api      *fakeAPI
		scene    *sound.Scene
		speech   sound.SpeechView
		emitter  sound.EmitterID
		finished []finishedEvent
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
		speech = scene.Speech()
		emitter = scene.Emitters().Create(dprec.ZeroVec3())

		finished = nil
		speech.SubscribeFinished(func(id sound.SpeechID, completed bool) {
			finished = append(finished, finishedEvent{id, completed})
		})
	})

	It("plays a line on an emitter", func() {
		id := speech.Play(sound.SpeechInfo{
			Media:   newFakeMedia(1.0),
			Emitter: emitter,
			Gain:    opt.V(0.5),
		})
		Expect(speech.IsValid(id)).To(BeTrue())
		Expect(speech.IsSpeaking(emitter)).To(BeTrue())
		Expect(speech.IsSpeaking(sound.NilEmitterID)).To(BeFalse())

		playback := api.lastPlayback()
		Expect(playback.spatial).To(BeTrue())
		Expect(playback.playing).To(BeTrue())
		Expect(playback.gain).To(BeNumerically("~", 0.5))
	})

	It("plays a narrator line without an emitter", func() {
		speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})
		Expect(api.lastPlayback().spatial).To(BeFalse())
		Expect(speech.IsSpeaking(sound.NilEmitterID)).To(BeTrue())
		Expect(speech.IsSpeaking(emitter)).To(BeFalse())
	})

	It("requires media", func() {
		Expect(func() {
			speech.Play(sound.SpeechInfo{})
		}).To(Panic())
	})

	It("reports a line that plays to its end as completed", func() {
		id := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		playback := api.lastPlayback()

		playback.finish()
		Expect(finished).To(BeEmpty())

		scene.Update(frame)
		Expect(finished).To(Equal([]finishedEvent{{id, true}}))
		Expect(speech.IsValid(id)).To(BeFalse())
		Expect(speech.IsSpeaking(emitter)).To(BeFalse())
		Expect(playback.released).To(BeTrue())
	})

	It("cuts short the previous line of the same emitter", func() {
		first := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		firstPlayback := api.lastPlayback()
		second := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})

		Expect(finished).To(Equal([]finishedEvent{{first, false}}))
		Expect(speech.IsValid(first)).To(BeFalse())
		Expect(speech.IsValid(second)).To(BeTrue())
		Expect(firstPlayback.playing).To(BeFalse())
		Expect(firstPlayback.released).To(BeTrue())
		Expect(api.lastPlayback().playing).To(BeTrue())
	})

	It("cuts short the previous line of the narrator", func() {
		first := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})
		speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})
		Expect(finished).To(Equal([]finishedEvent{{first, false}}))
	})

	It("lets different speakers talk at the same time", func() {
		other := scene.Emitters().Create(dprec.ZeroVec3())
		first := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		second := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: other})
		third := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})

		Expect(finished).To(BeEmpty())
		Expect(speech.IsValid(first)).To(BeTrue())
		Expect(speech.IsValid(second)).To(BeTrue())
		Expect(speech.IsValid(third)).To(BeTrue())
	})

	It("can stop a line", func() {
		id := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		speech.Stop(id)

		Expect(finished).To(Equal([]finishedEvent{{id, false}}))
		Expect(speech.IsValid(id)).To(BeFalse())
		Expect(api.lastPlayback().playing).To(BeFalse())
		Expect(func() {
			speech.Stop(id)
		}).To(Panic())
	})

	It("can stop all lines", func() {
		first := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		second := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})
		speech.StopAll()

		Expect(finished).To(ConsistOf(finishedEvent{first, false}, finishedEvent{second, false}))
		Expect(speech.IsSpeaking(emitter)).To(BeFalse())
		Expect(speech.IsSpeaking(sound.NilEmitterID)).To(BeFalse())
	})

	It("cuts short the line of an emitter that is deleted", func() {
		id := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		narration := speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0)})

		scene.Emitters().Delete(emitter)

		Expect(finished).To(Equal([]finishedEvent{{id, false}}))
		Expect(speech.IsValid(narration)).To(BeTrue())
	})

	It("allows the next line to be played from the finished callback", func() {
		var next sound.SpeechID
		speech.SubscribeFinished(func(id sound.SpeechID, completed bool) {
			if completed && next == sound.NilSpeechID {
				next = speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
			}
		})

		speech.Play(sound.SpeechInfo{Media: newFakeMedia(1.0), Emitter: emitter})
		api.lastPlayback().finish()
		scene.Update(frame)

		Expect(speech.IsValid(next)).To(BeTrue())
		Expect(api.lastPlayback().playing).To(BeTrue())
		Expect(finished).To(HaveLen(1))
	})
})
