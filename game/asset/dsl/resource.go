package dsl

import "fmt"

var resourceProviders = make(map[string]Provider[any])

// Save saves the specified resource to an asset at the specified path.
func Save[T any](path string, provider Provider[T]) any {
	if _, ok := resourceProviders[path]; ok {
		panic(fmt.Sprintf("provider for asset at path %q already exists", path))
	}
	resourceProviders[path] = OnceProvider(FuncProvider(
		// get function
		func() (any, error) {
			resource, err := provider.Get()
			if err != nil {
				return nil, fmt.Errorf("error getting resource: %w", err)
			}
			return resource, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("save-asset", path, provider)
		},
	))
	return nil
}
