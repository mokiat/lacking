package game

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mokiat/lacking/core/audio"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/storage/chunked"
)

func newSoundResourceLoader() ResourceLoader[any] {
	return GenericResourceLoader(&soundResourceLoader{
		resourceType: reflect.TypeFor[audio.Media](),
	})
}

type soundResourceLoader struct {
	resourceType reflect.Type
}

func (l *soundResourceLoader) ApplicableType() reflect.Type {
	return l.resourceType
}

func (l *soundResourceLoader) LoadResource(loader *AssetLoader, asset *chunked.Asset) (audio.Media, error) {
	audioAPI := loader.Engine().Audio()
	if audioAPI == nil {
		return nil, errors.New("engine has no audio API")
	}

	var dtoSound dto.Sound
	if err := asset.Read(&dtoSound); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}
	chunk := dtoSound.SoundChunk
	if chunk == nil {
		return nil, errors.New("asset has no sound chunk")
	}

	frames := make([]audio.Frame, len(chunk.Frames))
	for i, frame := range chunk.Frames {
		frames[i] = audio.Frame{
			Left:  frame.Left,
			Right: frame.Right,
		}
	}
	data := audio.MediaData{
		Frames:     frames,
		SampleRate: int(chunk.SampleRate),
	}

	var media audio.Media
	if err := loader.ScheduleMain(func() error {
		media = audioAPI.CreateMedia(data)
		return nil
	}).Wait(); err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}
	return media, nil
}

func (l *soundResourceLoader) UnloadResource(loader *AssetLoader, resource audio.Media) error {
	return loader.ScheduleMain(func() error {
		resource.Release()
		return nil
	}).Wait()
}
