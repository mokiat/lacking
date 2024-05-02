package dsl

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
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

// TODO: Rework based on latest changes
// func AddBox(opts ...Operation) Operation {
// 	return FuncOperation(
// 		// apply function
// 		func(target any) error {
// 			cfg := addBoxConfig{
// 				width:  1.0,
// 				height: 1.0,
// 				length: 1.0,
// 			}
// 			for _, opt := range opts {
// 				if err := opt.Apply(&cfg); err != nil {
// 					return err
// 				}
// 			}

// 			fragment, ok := target.(*mdl.Fragment)
// 			if !ok {
// 				return fmt.Errorf("target %T is not a fragment", target)
// 			}

// 			// front
// 			frontVertexOffset := fragment.VertexOffset()
// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(-1.0, 1.0, 1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, 1.0),
// 				TexCoord: dprec.NewVec2(0.0, 1.0),
// 			}.Scale(cfg.halfSize()))

// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(-1.0, -1.0, 1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, 1.0),
// 				TexCoord: dprec.NewVec2(0.0, 0.0),
// 			}.Scale(cfg.halfSize()))

// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(1.0, -1.0, 1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, 1.0),
// 				TexCoord: dprec.NewVec2(1.0, 0.0),
// 			}.Scale(cfg.halfSize()))

// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(1.0, 1.0, 1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, 1.0),
// 				TexCoord: dprec.NewVec2(1.0, 1.0),
// 			}.Scale(cfg.halfSize()))

// 			fragment.AddIndex(frontVertexOffset + 0)
// 			fragment.AddIndex(frontVertexOffset + 1)
// 			fragment.AddIndex(frontVertexOffset + 2)
// 			fragment.AddIndex(frontVertexOffset + 0)
// 			fragment.AddIndex(frontVertexOffset + 2)
// 			fragment.AddIndex(frontVertexOffset + 3)

// 			// back
// 			backVertexOffset := fragment.VertexOffset()
// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(-1.0, 1.0, -1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, -1.0),
// 				TexCoord: dprec.NewVec2(0.0, 1.0),
// 			}.Scale(cfg.halfSize()))

// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(-1.0, -1.0, -1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, -1.0),
// 				TexCoord: dprec.NewVec2(0.0, 0.0),
// 			}.Scale(cfg.halfSize()))

// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(1.0, -1.0, -1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, -1.0),
// 				TexCoord: dprec.NewVec2(1.0, 0.0),
// 			}.Scale(cfg.halfSize()))

// 			fragment.AddVertex(mdl.Vertex{
// 				Coord:    dprec.NewVec3(1.0, 1.0, -1.0),
// 				Normal:   dprec.NewVec3(0.0, 0.0, -1.0),
// 				TexCoord: dprec.NewVec2(1.0, 1.0),
// 			}.Scale(cfg.halfSize()))

// 			// TODO: Add other sides...

// 			fragment.AddIndex(backVertexOffset + 0)
// 			fragment.AddIndex(backVertexOffset + 1)
// 			fragment.AddIndex(backVertexOffset + 2)
// 			fragment.AddIndex(backVertexOffset + 0)
// 			fragment.AddIndex(backVertexOffset + 2)
// 			fragment.AddIndex(backVertexOffset + 3)

// 			return nil
// 		},

// 		// digest function
// 		func() ([]byte, error) {
// 			return CreateDigest("add-box", opts)
// 		},
// 	)
// }

type addBoxConfig struct {
	width  float64
	height float64
	length float64
}

func (c *addBoxConfig) Width() float64 {
	return c.width
}

func (c *addBoxConfig) SetWidth(width float64) {
	c.width = width
}

func (c *addBoxConfig) Height() float64 {
	return c.height
}

func (c *addBoxConfig) SetHeight(height float64) {
	c.height = height
}

func (c *addBoxConfig) Length() float64 {
	return c.length
}

func (c *addBoxConfig) SetLength(length float64) {
	c.length = length
}

func (c *addBoxConfig) halfSize() dprec.Vec3 {
	return dprec.NewVec3(c.width/2.0, c.height/2.0, c.length/2.0)
}
