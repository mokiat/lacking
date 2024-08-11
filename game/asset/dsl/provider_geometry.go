package dsl

import "github.com/mokiat/lacking/game/asset/mdl"

// CreateGeometry creates a new geometry.
func CreateGeometry(name string, opts ...Operation) Provider[*mdl.Geometry] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Geometry, error) {
			var geometry mdl.Geometry
			geometry.SetName(name)
			geometry.SetFormat(mdl.VertexFormatCoord)
			for _, opt := range opts {
				if err := opt.Apply(&geometry); err != nil {
					return nil, err
				}
			}
			return &geometry, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-geometry", name, opts)
		},
	))
}

// CreateFragment creates a new geometry fragment.
func CreateFragment(name string, topology mdl.Topology, opts ...Operation) Provider[*mdl.Fragment] {
	return OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Fragment, error) {
			var fragment mdl.Fragment
			fragment.SetName(name)
			fragment.SetTopology(topology)
			for _, opt := range opts {
				if err := opt.Apply(&fragment); err != nil {
					return nil, err
				}
			}
			return &fragment, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-fragment", name, topology, opts)
		},
	))
}
