package ecs

type entityDescriptor struct {
	revision int32

	archetype       *componentArchetype
	archetypeOffset uint32
}
