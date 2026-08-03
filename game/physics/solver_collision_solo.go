package physics

import "github.com/mokiat/gomath/dprec"

type SoloCollisionSolverConfig struct {
	TerrainFrictionCoefficient    float64
	TerrainRestitutionCoefficient float64
	TerrainContactNormal          dprec.Vec3

	BodyFrictionCoefficient    float64
	BodyRestitutionCoefficient float64
	BodyContactPoint           dprec.Vec3

	Depth float64
}

type SoloCollisionSolver struct {
}

var _ SoloConstraintSolver = (*SoloCollisionSolver)(nil)

func (s *SoloCollisionSolver) Init(config SoloCollisionSolverConfig) {

}

func (s *SoloCollisionSolver) Reset(ctx SoloConstraintContext) {

}
func (s *SoloCollisionSolver) ApplyImpulses(ctx SoloConstraintContext) {

}
func (s *SoloCollisionSolver) ApplyNudges(ctx SoloConstraintContext) {

}
