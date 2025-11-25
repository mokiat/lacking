package component

// ScopeModifier represents a function that takes a scope and returns a modified
// version of it.
//
// Modifiers are not usually applied directly but rather through helper render
// functions.
type ScopeModifier func(Scope) Scope

// ScopeValueModifier creates a ScopeModifier that adds a value to the scope
// it is applied to using the specified key.
//
// This function would normally not be used by user code, instead a rendering
// function such as WithScopeValue would be used.
func ScopeValueModifier(key, value any) ScopeModifier {
	return func(scope Scope) Scope {
		return ValueScope(scope, key, value)
	}
}

// TypedScopeValueModifier creates a ScopeModifier that adds a value to the
// scope it is applied to using the value's type as the key.
//
// This function would normally not be used by user code, instead a rendering
// function such as WithTypedScopeValue would be used.
func TypedScopeValueModifier[T any](value T) ScopeModifier {
	return func(scope Scope) Scope {
		return TypedValueScope(scope, value)
	}
}

// ChainScopeModifier creates a new ScopeModifier that applies the parent
// modifier first and then the current modifier.
// If the parent modifier is nil, only the current modifier is used.
//
// This function would normally not be used by user code, instead a rendering
// function would be used to chain scope modifications.
func ChainScopeModifier(parent, current ScopeModifier) ScopeModifier {
	if parent == nil {
		return current
	}
	return func(scope Scope) Scope {
		scope = parent(scope)
		scope = current(scope)
		return scope
	}
}
