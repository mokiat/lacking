package ecs

// Option is a configuration function that can be used to customize the
// behavior of the ECS engine.
type Option func(*config)

type config struct{}
