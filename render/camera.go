package render

import "github.com/mokiat/gomath/sprec"

type Camera struct {
	projectionMatrix sprec.Mat4
	matrix           sprec.Mat4
	viewMatrix       sprec.Mat4
}

func NewCamera() *Camera {
	return &Camera{
		projectionMatrix: sprec.IdentityMat4(),
		matrix:           sprec.IdentityMat4(),
		viewMatrix:       sprec.IdentityMat4(),
	}
}

func (c *Camera) SetProjectionMatrix(matrix sprec.Mat4) {
	c.projectionMatrix = matrix
}

func (c *Camera) ProjectionMatrix() sprec.Mat4 {
	return c.projectionMatrix
}

func (c *Camera) SetMatrix(matrix sprec.Mat4) {
	c.matrix = matrix
	c.viewMatrix = sprec.InverseMat4(matrix)
}

func (c *Camera) Matrix() sprec.Mat4 {
	return c.matrix
}

func (c *Camera) ViewMatrix() sprec.Mat4 {
	return c.viewMatrix
}
