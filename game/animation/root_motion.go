package animation

import (
	"math"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

func NewRootMotion(source Source, bone string) *RootMotion {
	return &RootMotion{
		source: source,
		bone:   bone,
	}
}

type RootMotion struct {
	source Source
	bone   string
}

func (m *RootMotion) GetDeltaTransform(from, to float64) NodeTransform {
	// FIXME: This is an ugly approach to this.
	oldPosition := m.source.Position()
	defer m.source.SetPosition(oldPosition)

	resultMatrix := dprec.IdentityMat4()

	length := m.source.Length()
	modFrom := math.Mod(from, length)
	modTo := math.Mod(to, length)

	fromMatrix := m.getMatrixAt(modFrom)

	for (modTo < modFrom) && (to > from) {
		toMatrix := m.getMatrixAt(length - 0.00001) // prevent mod down to zero

		deltaMatrix := dprec.Mat4Prod(
			toMatrix,
			dprec.InverseMat4(fromMatrix),
		)

		resultMatrix = dprec.Mat4Prod(
			deltaMatrix,
			resultMatrix,
		)

		fromMatrix = m.getMatrixAt(0.0)

		to -= length
	}

	toMatrix := m.getMatrixAt(modTo)

	deltaMatrix := dprec.Mat4Prod(
		toMatrix,
		dprec.InverseMat4(fromMatrix),
	)

	resultMatrix = dprec.Mat4Prod(
		deltaMatrix,
		resultMatrix,
	)

	t, r, s := resultMatrix.TRS()
	return NodeTransform{
		Translation: opt.V(t),
		Rotation:    opt.V(r),
		Scale:       opt.V(s),
	}
}

func (m *RootMotion) getMatrixAt(t float64) dprec.Mat4 {
	// FIXME: This is an ugly approach to this.
	oldPosition := m.source.Position()
	defer m.source.SetPosition(oldPosition)

	m.source.SetPosition(t)
	transform := m.source.NodeTransform(m.bone)

	return dprec.TRSMat4(
		transform.Translation.ValueOrDefault(dprec.ZeroVec3()),
		transform.Rotation.ValueOrDefault(dprec.IdentityQuat()),
		transform.Scale.ValueOrDefault(dprec.NewVec3(1.0, 1.0, 1.0)),
	)
}
