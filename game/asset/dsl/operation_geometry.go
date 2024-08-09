package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// SetVertexFormat sets the vertex format of the target geometry.
func SetVertexFormat(format mdl.VertexFormat) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			geometry, ok := target.(*mdl.Geometry)
			if !ok {
				return fmt.Errorf("target %T is not a geometry", target)
			}
			geometry.SetFormat(format)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("set-vertex-format", format)
		},
	)
}

// AddFragment adds the specified fragment to the target geometry.
func AddFragment(fragmentProvider Provider[*mdl.Fragment]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			geometry, ok := target.(*mdl.Geometry)
			if !ok {
				return fmt.Errorf("target %T is not a geometry", target)
			}

			fragment, err := fragmentProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting fragment: %w", err)
			}
			geometry.AddFragment(fragment)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("add-fragment", fragmentProvider)
		},
	)
}

// TODO: Make it possible to add pre-defined shapes to a geometry:
// func AddBox(opts ...Operation) Operation {
// }

// type addBoxConfig struct {
// 	width  float64
// 	height float64
// 	length float64
// }

// func (c *addBoxConfig) SetWidth(width float64) {
// 	c.width = width
// }

// func (c *addBoxConfig) SetHeight(height float64) {
// 	c.height = height
// }

// func (c *addBoxConfig) SetLength(length float64) {
// 	c.length = length
// }

// func (c *addBoxConfig) halfSize() dprec.Vec3 {
// 	return dprec.NewVec3(c.width/2.0, c.height/2.0, c.length/2.0)
// }
