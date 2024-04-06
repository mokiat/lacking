package graphics

type SkyInfo struct {
	Definition *SkyDefinition
}

func newSky(scene *Scene, info SkyInfo) *Sky {
	result := &Sky{
		scene:      scene,
		definition: info.Definition,
	}
	scene.skies.Add(result)
	return result
}

type Sky struct {
	scene      *Scene
	definition *SkyDefinition
}

func (s *Sky) Delete() {
	s.scene.skies.Remove(s)
	s.definition = nil
	s.scene = nil
}

func (s *Sky) Definition() *SkyDefinition {
	return s.definition
}
