package mathmatcher

import (
	"fmt"

	"github.com/mokiat/lacking/internal/testing"
	"github.com/mokiat/lacking/math"
	"github.com/onsi/gomega/types"
)

func HaveVec2Coords(expectedX, expectedY float32) types.GomegaMatcher {
	return testing.SimpleMatcher(func(actualValue interface{}) (testing.MatchStatus, error) {
		vector, ok := actualValue.(math.Vec2)
		if !ok {
			return testing.MatchStatus{}, fmt.Errorf("HaveVec2Coords matcher expects a math.Vec2")
		}

		matches := areEqualFloat32(vector.X, expectedX) && areEqualFloat32(vector.Y, expectedY)
		if !matches {
			return testing.FailureMatchStatus(
				fmt.Sprintf("Expected\n\t%#v\nto have coords\n\t(%f, %f)", vector, expectedX, expectedY),
				fmt.Sprintf("Expected\n\t%#v\nnot to have coords\n\t(%f, %f)", vector, expectedX, expectedY),
			), nil
		}
		return testing.SuccessMatchStatus(), nil
	})
}
