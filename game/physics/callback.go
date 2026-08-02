package physics

import "github.com/mokiat/lacking/util/observer"

// SoloBodyCollisionCallback is a mechanism to receive notifications
// about collisions between a body and a prop in the scene.
type SoloBodyCollisionCallback func(bodyID BodyID, terrainID TerrainID, active bool)

// SoloBodyCollisionSubscription represents a notification subscription
// for single body collisions.
type SoloBodyCollisionSubscription = observer.Subscription[SoloBodyCollisionCallback]

// PairBodyCollisionCallback is a mechanism to receive notifications
// about collisions between two bodies.
type PairBodyCollisionCallback func(firstBodyID, secondBodyID BodyID, active bool)

// PairBodyCollisionSubscription represents a notification subscription
// for double body collisions.
type PairBodyCollisionSubscription = observer.Subscription[PairBodyCollisionCallback]
