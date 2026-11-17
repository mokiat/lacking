package sound_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/sound"
)

var _ = Describe("Music", func() {
	var (
		api   *fakeAPI
		scene *sound.Scene
		music sound.MusicView

		firstMedia  *fakeMedia
		secondMedia *fakeMedia

		started  []sound.MusicTrackID
		endings  []sound.MusicTrackID
		finished []sound.MusicTrackID
	)

	BeforeEach(func() {
		api = newFakeAPI()
		scene = sound.NewScene(api)
		music = scene.Music()

		firstMedia = newFakeMedia(10.0)
		secondMedia = newFakeMedia(20.0)

		started = nil
		endings = nil
		finished = nil
		music.SubscribeTrackStarted(func(id sound.MusicTrackID) {
			started = append(started, id)
		})
		music.SubscribeTrackEnding(func(id sound.MusicTrackID) {
			endings = append(endings, id)
		})
		music.SubscribeTrackFinished(func(id sound.MusicTrackID) {
			finished = append(finished, id)
		})
	})

	It("is silent initially", func() {
		scene.Update(frame)
		Expect(music.Current()).To(Equal(sound.NilMusicTrackID))
		Expect(music.QueueLength()).To(BeZero())
		Expect(started).To(BeEmpty())
	})

	Describe("queue", func() {
		var first, second sound.MusicTrackID

		BeforeEach(func() {
			first = music.Enqueue(sound.MusicTrackInfo{Media: firstMedia})
			second = music.Enqueue(sound.MusicTrackInfo{Media: secondMedia})
		})

		It("holds tracks until the scene is updated", func() {
			Expect(music.Current()).To(Equal(sound.NilMusicTrackID))
			Expect(music.QueueLength()).To(Equal(2))
			Expect(api.playbackOf(firstMedia).playing).To(BeFalse())
		})

		It("starts the first track on update", func() {
			scene.Update(frame)
			Expect(music.Current()).To(Equal(first))
			Expect(music.QueueLength()).To(Equal(1))
			Expect(started).To(Equal([]sound.MusicTrackID{first}))

			playback := api.playbackOf(firstMedia)
			Expect(playback.playing).To(BeTrue())
			Expect(playback.spatial).To(BeFalse())
			Expect(playback.gain).To(BeNumerically("~", 1.0))
			Expect(api.playbackOf(secondMedia).playing).To(BeFalse())
		})

		It("continues with the next track once the current one ends", func() {
			scene.Update(frame)
			api.playbackOf(firstMedia).finish()
			scene.Update(frame)

			Expect(finished).To(Equal([]sound.MusicTrackID{first}))
			Expect(started).To(Equal([]sound.MusicTrackID{first, second}))
			Expect(music.Current()).To(Equal(second))
			Expect(music.IsValid(first)).To(BeFalse())
			Expect(api.playbackOf(firstMedia).released).To(BeTrue())
			Expect(api.playbackOf(secondMedia).playing).To(BeTrue())
		})

		It("goes silent once the queue runs out", func() {
			scene.Update(frame)
			api.playbackOf(firstMedia).finish()
			scene.Update(frame)
			api.playbackOf(secondMedia).finish()
			scene.Update(frame)

			Expect(music.Current()).To(Equal(sound.NilMusicTrackID))
			Expect(finished).To(Equal([]sound.MusicTrackID{first, second}))
		})

		It("can skip to the next track", func() {
			scene.Update(frame)
			music.Skip()
			scene.Update(frame)

			Expect(music.Current()).To(Equal(second))
			Expect(finished).To(Equal([]sound.MusicTrackID{first}))
			Expect(api.playbackOf(firstMedia).playing).To(BeFalse())
			Expect(api.playbackOf(firstMedia).released).To(BeTrue())
			Expect(api.playbackOf(secondMedia).playing).To(BeTrue())
		})

		It("can remove a queued track", func() {
			scene.Update(frame)
			music.Remove(second)

			Expect(music.IsValid(second)).To(BeFalse())
			Expect(music.QueueLength()).To(BeZero())
			Expect(api.playbackOf(secondMedia).released).To(BeTrue())
			Expect(finished).To(BeEmpty())
		})

		It("moves on when the current track is removed", func() {
			scene.Update(frame)
			music.Remove(first)
			scene.Update(frame)
			Expect(music.Current()).To(Equal(second))
		})

		It("can clear the queue without affecting the current track", func() {
			scene.Update(frame)
			music.Clear()
			scene.Update(frame)

			Expect(music.Current()).To(Equal(first))
			Expect(music.QueueLength()).To(BeZero())
			Expect(music.IsValid(second)).To(BeFalse())
		})

		It("can be stopped", func() {
			scene.Update(frame)
			music.Stop(0)
			scene.Update(frame)

			Expect(music.Current()).To(Equal(sound.NilMusicTrackID))
			Expect(music.QueueLength()).To(BeZero())
			Expect(finished).To(Equal([]sound.MusicTrackID{first}))
			Expect(api.playbackOf(firstMedia).playing).To(BeFalse())
		})

		It("can be stopped with a fade out", func() {
			scene.Update(frame)
			music.Stop(2 * time.Second)

			scene.Update(time.Second)
			Expect(music.Current()).To(Equal(sound.NilMusicTrackID))
			Expect(music.IsValid(first)).To(BeTrue())
			Expect(api.playbackOf(firstMedia).playing).To(BeTrue())

			scene.Update(time.Second)
			Expect(api.playbackOf(firstMedia).gain).To(BeNumerically("~", 0.5, 1e-4))

			scene.Update(time.Second)
			Expect(music.IsValid(first)).To(BeFalse())
			Expect(api.playbackOf(firstMedia).playing).To(BeFalse())
			Expect(finished).To(Equal([]sound.MusicTrackID{first}))
		})

		It("can be replaced by a new track", func() {
			scene.Update(frame)
			thirdMedia := newFakeMedia(5.0)
			third := music.Play(sound.MusicTrackInfo{Media: thirdMedia})
			scene.Update(frame)

			Expect(music.Current()).To(Equal(third))
			Expect(music.QueueLength()).To(BeZero())
			Expect(music.IsValid(first)).To(BeFalse())
			Expect(music.IsValid(second)).To(BeFalse())
			Expect(api.playbackOf(thirdMedia).playing).To(BeTrue())
		})
	})

	Describe("track settings", func() {
		It("applies the gain and start offset", func() {
			music.Enqueue(sound.MusicTrackInfo{
				Media:   firstMedia,
				Gain:    opt.V(0.5),
				StartAt: 3 * time.Second,
			})
			scene.Update(frame)

			playback := api.playbackOf(firstMedia)
			Expect(playback.gain).To(BeNumerically("~", 0.5))
			Expect(playback.startedAt).To(BeNumerically("~", 3.0))
		})

		It("applies the loop region", func() {
			music.Enqueue(sound.MusicTrackInfo{
				Media:     firstMedia,
				Loop:      true,
				LoopStart: opt.V(2 * time.Second),
				LoopEnd:   opt.V(8 * time.Second),
			})

			playback := api.playbackOf(firstMedia)
			Expect(playback.looping).To(BeTrue())
			Expect(playback.loopStart).To(BeNumerically("~", 2.0))
			Expect(playback.loopEnd).To(BeNumerically("~", 8.0))
		})

		It("requires media", func() {
			Expect(func() {
				music.Enqueue(sound.MusicTrackInfo{})
			}).To(Panic())
		})
	})

	Describe("ending notification", func() {
		var first sound.MusicTrackID

		BeforeEach(func() {
			first = music.Enqueue(sound.MusicTrackInfo{Media: firstMedia})
			scene.Update(0)
		})

		It("is not sent while the end is far off", func() {
			scene.Update(4 * time.Second)
			Expect(endings).To(BeEmpty())
		})

		It("is sent once the lead time is reached", func() {
			scene.Update(4 * time.Second)
			scene.Update(2 * time.Second)
			Expect(endings).To(Equal([]sound.MusicTrackID{first}))
		})

		It("is sent only once per track", func() {
			scene.Update(6 * time.Second)
			scene.Update(time.Second)
			scene.Update(time.Second)
			Expect(endings).To(HaveLen(1))
		})

		It("honors the configured lead time", func() {
			music.SetEndingLeadTime(time.Second)
			Expect(music.EndingLeadTime()).To(Equal(time.Second))
			scene.Update(8 * time.Second)
			Expect(endings).To(BeEmpty())
			scene.Update(1500 * time.Millisecond)
			Expect(endings).To(HaveLen(1))
		})

		It("takes the start offset into account", func() {
			music.Play(sound.MusicTrackInfo{
				Media:   secondMedia,
				StartAt: 14 * time.Second,
			})
			scene.Update(0)
			endings = nil
			scene.Update(2 * time.Second)
			Expect(endings).To(HaveLen(1))
		})

		It("is not sent while the music is paused", func() {
			music.Bus().Pause()
			scene.Update(8 * time.Second)
			Expect(endings).To(BeEmpty())

			music.Bus().Resume()
			scene.Update(6 * time.Second)
			Expect(endings).To(HaveLen(1))
		})

		It("is not sent for looping tracks", func() {
			music.Play(sound.MusicTrackInfo{Media: secondMedia, Loop: true})
			scene.Update(0)
			endings = nil
			scene.Update(time.Minute)
			Expect(endings).To(BeEmpty())
		})

		It("allows the queue to be refilled from the callback", func() {
			var refill sound.MusicTrackID
			music.SubscribeTrackEnding(func(id sound.MusicTrackID) {
				refill = music.Enqueue(sound.MusicTrackInfo{Media: secondMedia})
			})

			scene.Update(6 * time.Second)
			Expect(music.QueueLength()).To(Equal(1))

			api.playbackOf(firstMedia).finish()
			scene.Update(frame)
			Expect(music.Current()).To(Equal(refill))
			Expect(api.playbackOf(secondMedia).playing).To(BeTrue())
		})
	})

	Describe("crossfade", func() {
		var first, second sound.MusicTrackID

		BeforeEach(func() {
			music.SetCrossfadeDuration(2 * time.Second)
			Expect(music.CrossfadeDuration()).To(Equal(2 * time.Second))

			first = music.Enqueue(sound.MusicTrackInfo{Media: firstMedia})
			second = music.Enqueue(sound.MusicTrackInfo{Media: secondMedia})
			scene.Update(0)
		})

		It("does not fade in the very first track", func() {
			Expect(api.playbackOf(firstMedia).gain).To(BeNumerically("~", 1.0))
		})

		It("starts the next track ahead of the end of the current one", func() {
			scene.Update(7 * time.Second)
			Expect(music.Current()).To(Equal(first))

			scene.Update(1500 * time.Millisecond)
			Expect(music.Current()).To(Equal(second))
			Expect(music.IsValid(first)).To(BeTrue())
			Expect(api.playbackOf(firstMedia).playing).To(BeTrue())
			Expect(api.playbackOf(secondMedia).playing).To(BeTrue())
			Expect(api.playbackOf(secondMedia).gain).To(BeNumerically("~", 0.0))
		})

		It("blends the two tracks over the remaining time", func() {
			scene.Update(8500 * time.Millisecond) // 1.5 seconds remain
			scene.Update(750 * time.Millisecond)
			Expect(api.playbackOf(firstMedia).gain).To(BeNumerically("~", 0.5, 1e-4))
			Expect(api.playbackOf(secondMedia).gain).To(BeNumerically("~", 0.5, 1e-4))

			scene.Update(750 * time.Millisecond)
			Expect(api.playbackOf(secondMedia).gain).To(BeNumerically("~", 1.0, 1e-4))
			Expect(music.IsValid(first)).To(BeFalse())
			Expect(api.playbackOf(firstMedia).released).To(BeTrue())
			Expect(finished).To(Equal([]sound.MusicTrackID{first}))
		})

		It("handles the outgoing track ending during the blend", func() {
			scene.Update(8500 * time.Millisecond)
			api.playbackOf(firstMedia).finish()
			scene.Update(frame)
			Expect(music.IsValid(first)).To(BeFalse())
			Expect(finished).To(Equal([]sound.MusicTrackID{first}))
			Expect(music.Current()).To(Equal(second))
		})

		It("lets the current track play out when nothing is queued", func() {
			music.Clear()
			scene.Update(9 * time.Second)
			Expect(music.Current()).To(Equal(first))
			Expect(api.playbackOf(firstMedia).gain).To(BeNumerically("~", 1.0))
		})

		It("is used when skipping", func() {
			music.Skip()
			scene.Update(frame)
			Expect(music.Current()).To(Equal(second))
			Expect(api.playbackOf(firstMedia).playing).To(BeTrue())

			scene.Update(time.Second)
			Expect(api.playbackOf(firstMedia).gain).To(BeNumerically("~", 0.5, 1e-4))
			Expect(api.playbackOf(secondMedia).gain).To(BeNumerically("~", 0.5, 1e-4))
		})

		It("does not progress while the music is paused", func() {
			music.Skip()
			scene.Update(frame)
			music.Bus().Pause()
			scene.Update(time.Minute)
			Expect(music.IsValid(first)).To(BeTrue())
			Expect(api.playbackOf(secondMedia).gain).To(BeNumerically("~", 0.0))
		})
	})
})
