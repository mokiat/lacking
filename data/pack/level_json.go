package pack

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/json"
	"github.com/mokiat/lacking/util/resource"
)

type OpenLevelResourceAction struct {
	locator resource.ReadLocator
	uri     string
	level   *Level
}

func (a *OpenLevelResourceAction) Describe() string {
	return fmt.Sprintf("open_level_resource(%q)", a.uri)
}

func (a *OpenLevelResourceAction) Level() *Level {
	if a.level == nil {
		panic("reading data from unprocessed action")
	}
	return a.level
}

func (a *OpenLevelResourceAction) Run() error {
	in, err := a.locator.ReadResource(a.uri)
	if err != nil {
		return fmt.Errorf("failed to open level resource %q: %w", a.uri, err)
	}
	defer in.Close()

	jsonLevel, err := json.NewLevelDecoder().Decode(in)
	if err != nil {
		return fmt.Errorf("failed to decode level %q: %w", a.uri, err)
	}

	a.level = &Level{
		StaticEntities: make([]*LevelEntity, len(jsonLevel.StaticEntities)),
	}
	for i, jsonStaticEntity := range jsonLevel.StaticEntities {
		a.level.StaticEntities[i] = &LevelEntity{
			Model: jsonStaticEntity.Model,
			Matrix: sprec.NewMat4(
				jsonStaticEntity.Matrix[0], jsonStaticEntity.Matrix[4], jsonStaticEntity.Matrix[8], jsonStaticEntity.Matrix[12],
				jsonStaticEntity.Matrix[1], jsonStaticEntity.Matrix[5], jsonStaticEntity.Matrix[9], jsonStaticEntity.Matrix[13],
				jsonStaticEntity.Matrix[2], jsonStaticEntity.Matrix[6], jsonStaticEntity.Matrix[10], jsonStaticEntity.Matrix[14],
				jsonStaticEntity.Matrix[3], jsonStaticEntity.Matrix[7], jsonStaticEntity.Matrix[11], jsonStaticEntity.Matrix[15],
			),
		}
	}

	return nil
}
