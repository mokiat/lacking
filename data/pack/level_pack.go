package pack

import "github.com/mokiat/gomath/sprec"

type LevelProvider interface {
	Level() *Level
}

type Level struct {
	StaticEntities []*LevelEntity
}

type LevelEntity struct {
	Model  string
	Matrix sprec.Mat4
}
