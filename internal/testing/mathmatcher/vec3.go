package mathmatcher

import (
	"fmt"

	"github.com/mokiat/lacking/internal/testing"
	"github.com/mokiat/lacking/math"
	"github.com/onsi/gomega/types"
)

func HaveVec3Coords(expectedX, expectedY, expectedZ float32) types.GomegaMatcher {
	return testing.SimpleMatcher(func(actualValue interface{}) (testing.MatchStatus, error) {
		vector, ok := actualValue.(math.Vec3)
		if !ok {
			return testing.MatchStatus{}, fmt.Errorf("HaveVec3Coords matcher expects a math.Vec3")
		}

		matches := areEqualFloat32(vector.X, expectedX) &&
			areEqualFloat32(vector.Y, expectedY) &&
			areEqualFloat32(vector.Z, expectedZ)

		if !matches {
			return testing.FailureMatchStatus(
				fmt.Sprintf("Expected\n\t%#v\nto have coords\n\t(%f, %f, %f)", vector, expectedX, expectedY, expectedZ),
				fmt.Sprintf("Expected\n\t%#v\nnot to have coords\n\t(%f, %f, %f)", vector, expectedX, expectedY, expectedZ),
			), nil
		}
		return testing.SuccessMatchStatus(), nil
	})
}
