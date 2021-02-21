package resource

import (
	"sync"

	"github.com/mokiat/lacking/async"
)

type Set struct {
	registry *Registry

	resourcesMU sync.Mutex
	resources   map[string]int
}

func NewSet(registry *Registry) *Set {
	return &Set{
		registry:  registry,
		resources: make(map[string]int),
	}
}

func (s *Set) Open(uri string, allocator Allocator, releaser Releaser, inject func(value interface{})) async.Eventual {
	s.resourcesMU.Lock()
	defer s.resourcesMU.Unlock()
	if count, ok := s.resources[uri]; ok {
		s.resources[uri] = count + 1
	} else {
		s.resources[uri] = 1
	}
	return s.registry.allocate(s, uri, allocator, releaser, inject)
}

func (s *Set) OpenProgram(uri string, program **Program) async.Eventual {
	allocator := s.registry.programOperator.Allocator(uri)
	releaser := s.registry.programOperator.Releaser()
	return s.Open(uri, allocator, releaser, func(value interface{}) {
		*program = value.(*Program)
	})
}

func (s *Set) OpenTwoDTexture(uri string, texture **TwoDTexture) async.Eventual {
	allocator := s.registry.twodTextureOperator.Allocator(uri)
	releaser := s.registry.twodTextureOperator.Releaser()
	return s.Open(uri, allocator, releaser, func(value interface{}) {
		*texture = value.(*TwoDTexture)
	})
}

func (s *Set) OpenCubeTexture(uri string, texture **CubeTexture) async.Eventual {
	allocator := s.registry.cubeTextureOperator.Allocator(uri)
	releaser := s.registry.cubeTextureOperator.Releaser()
	return s.Open(uri, allocator, releaser, func(value interface{}) {
		*texture = value.(*CubeTexture)
	})
}

func (s *Set) OpenModel(uri string, model **Model) async.Eventual {
	allocator := s.registry.modelOperator.Allocator(uri)
	releaser := s.registry.modelOperator.Releaser()
	return s.Open(uri, allocator, releaser, func(value interface{}) {
		*model = value.(*Model)
	})
}

func (s *Set) OpenLevel(uri string, level **Level) async.Eventual {
	allocator := s.registry.levelOperator.Allocator(uri)
	releaser := s.registry.levelOperator.Releaser()
	return s.Open(uri, allocator, releaser, func(value interface{}) {
		*level = value.(*Level)
	})
}

func (s *Set) CreateShader(info ShaderInfo, shader **Shader) async.Eventual {
	allocator := s.registry.shaderOperator.Allocator(info)
	releaser := s.registry.shaderOperator.Releaser()
	return s.Open(info.ID(), allocator, releaser, func(value interface{}) {
		*shader = value.(*Shader)
	})
}

func (s *Set) Release() async.Eventual {
	s.resourcesMU.Lock()
	defer s.resourcesMU.Unlock()

	eventuals := make([]async.Eventual, 0, len(s.resources))
	for uri, count := range s.resources {
		eventuals = append(eventuals, s.registry.release(uri, count))
	}
	s.resources = nil
	return async.NewCompositeEventual(eventuals...)
}
