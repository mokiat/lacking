package graphics

type ShadowStageInput struct {
}

func newShadowStage(input ShadowStageInput) *ShadowStage {
	return &ShadowStage{}
}

var _ Stage = (*ShadowStage)(nil)

// ShadowStage is a stage that renders shadows.
type ShadowStage struct {
}

func (s *ShadowStage) Allocate() {
}

func (s *ShadowStage) Release() {
}

func (s *ShadowStage) PreRender(width, height uint32) {
}

func (s *ShadowStage) Render(ctx StageContext) {
}

func (s *ShadowStage) PostRender() {
	// Nothing to do here.
}
