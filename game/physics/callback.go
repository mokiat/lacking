package physics

import "github.com/mokiat/lacking/util/observer"

// SoloCollisionCallback is a mechanism to receive notifications
// about collisions between a body and a terrain in the scene.
type SoloCollisionCallback func(bodyID BodyID, terrainID TerrainID, active bool)

// SoloCollisionSubscription represents a notification subscription
// for single body collisions.
type SoloCollisionSubscription = observer.Subscription[SoloCollisionCallback]

// PairCollisionCallback is a mechanism to receive notifications
// about collisions between two bodies.
type PairCollisionCallback func(firstBodyID, secondBodyID BodyID, active bool)

// PairCollisionSubscription represents a notification subscription
// for double body collisions.
type PairCollisionSubscription = observer.Subscription[PairCollisionCallback]
