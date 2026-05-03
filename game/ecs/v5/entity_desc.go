package ecs

import "github.com/mokiat/lacking/game/ecs/v5/internal"

type entityDescriptor struct {
	revision int32

	archetype    *internal.Archetype
	archetypeRow internal.ArchetypeRow
}
