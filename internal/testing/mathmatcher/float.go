package mathmatcher

import (
	"fmt"

	"github.com/mokiat/lacking/internal/testing"
	"github.com/mokiat/lacking/math"
	"github.com/onsi/gomega/types"
)

const float32Margin = 0.00001

func areEqualFloat32(a, b float32) bool {
	return math.Abs32(a-b) < float32Margin
}

func EqualFloat32(expectedValue float32) types.GomegaMatcher {
	return testing.SimpleMatcher(func(actualValue interface{}) (testing.MatchStatus, error) {
		floatValue, ok := actualValue.(float32)
		if !ok {
			return testing.MatchStatus{}, fmt.Errorf("EqualFloat32 matcher expects a float32")
		}

		if !areEqualFloat32(floatValue, expectedValue) {
			return testing.FailureMatchStatus(
				fmt.Sprintf("Expected\n\t%f\nto equal\n\t%f", floatValue, expectedValue),
				fmt.Sprintf("Expected\n\t%f\nnot to equal\n\t%f", floatValue, expectedValue),
			), nil
		}
		return testing.SuccessMatchStatus(), nil
	})
}
