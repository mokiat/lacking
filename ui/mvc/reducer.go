package mvc

import (
	co "github.com/mokiat/lacking/ui/component"
)

type reducerRegistrationKey struct{}

type reducerRegistration struct {
	parent  *reducerRegistration
	reducer Reducer
}

// Reducer represents an algorithm that can process actions.
type Reducer interface {

	// Reduce attempts to process the specified action. If the Reducer does not
	// recognize the action, then it should return false, in which case a
	// parent reducer will be invoked to try and handle it.
	Reduce(action Action) bool
}

// ReducerFunc is a helper func wrapper that implements the Reducer interface.
type ReducerFunc func(action Action) bool

func (f ReducerFunc) Reduce(action Action) bool {
	return f(action)
}

// Dispatch propagates the specified Action through the chain of Reducers
// registered on the specified Scope.
func Dispatch(scope co.Scope, action Action) {
	registration := co.GetScopeValue[*reducerRegistration](scope, reducerRegistrationKey{})
	for registration != nil {
		if registration.reducer.Reduce(action) {
			return
		}
		registration = registration.parent
	}
	logger.Warn("No reducer found that can process action (type: %T)!", action)
}

// UseReducer extends the specified Scope with the specified Reducer. Dispatch
// calls to the returned scope will first be processed by the specified
// Reducer, followed by any parent Reducer registrations in case the current
// one does not recognize the action.
//
// NOTE: Make sure to attach the returned Scope to any child components in the
// hierarchy, otherwise they would not have access to the Reducer.
func UseReducer(scope co.Scope, reducer Reducer) co.Scope {
	parentRegistration := co.GetScopeValue[*reducerRegistration](scope, reducerRegistrationKey{})
	registration := &reducerRegistration{
		parent:  parentRegistration,
		reducer: reducer,
	}
	return co.ValueScope(scope, reducerRegistrationKey{}, registration)
}
