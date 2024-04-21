package graphics

type SkyInfo struct {
	Definition *SkyDefinition
}

func newSky(scene *Scene, info SkyInfo) *Sky {
	result := &Sky{
		scene:      scene,
		definition: info.Definition,
		active:     true,
	}
	scene.skies.Add(result)
	return result
}

type Sky struct {
	scene      *Scene
	definition *SkyDefinition
	active     bool
}

func (s *Sky) Active() bool {
	return s.active
}

func (s *Sky) SetActive(active bool) {
	s.active = active
}

func (s *Sky) Delete() {
	s.scene.skies.Remove(s)
	s.definition = nil
	s.scene = nil
}

func (s *Sky) Definition() *SkyDefinition {
	return s.definition
}
