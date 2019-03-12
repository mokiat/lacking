package testing

import "github.com/onsi/gomega/types"

type MatchStatus struct {
	success                bool
	failureMessage         string
	negativeFailureMessage string
}

func SuccessMatchStatus() MatchStatus {
	return MatchStatus{
		success: true,
	}
}

func FailureMatchStatus(failureMessage, negativeFailureMessage string) MatchStatus {
	return MatchStatus{
		success:                false,
		failureMessage:         failureMessage,
		negativeFailureMessage: negativeFailureMessage,
	}
}

type MatchCheckFunc func(actualValue interface{}) (MatchStatus, error)

func SimpleMatcher(checkFunc MatchCheckFunc) types.GomegaMatcher {
	return &simpleMatcher{
		checkFunc: checkFunc,
	}
}

type simpleMatcher struct {
	checkFunc MatchCheckFunc
	status    MatchStatus
}

func (m *simpleMatcher) Match(actualValue interface{}) (bool, error) {
	var err error
	if m.status, err = m.checkFunc(actualValue); err != nil {
		return false, err
	}
	return m.status.success, nil
}

func (m *simpleMatcher) FailureMessage(actualValue interface{}) string {
	return m.status.failureMessage
}

func (m *simpleMatcher) NegatedFailureMessage(actualValue interface{}) string {
	return m.status.negativeFailureMessage
}
