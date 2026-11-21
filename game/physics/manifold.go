package physics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

const (
	// maxManifoldPoints is the largest number of contact points that are
	// tracked for a single pair of shapes. Four points are sufficient to
	// describe a stable face-to-face contact.
	maxManifoldPoints = 4

	// maxManifoldIdleSteps is the number of consecutive steps without a
	// detected contact after which the contact points that are tracked for
	// a pair of shapes are forgotten.
	maxManifoldIdleSteps = 8

	// manifoldBreakingDistance is the distance by which the two sides of a
	// tracked contact point can move relative to each other - away from
	// each other, into each other or sideways along the contact surface -
	// before the point is considered to no longer describe the contact and
	// is discarded. It is also the distance within which two contact points
	// are considered to be the same point.
	manifoldBreakingDistance = 0.02

	// contactSlop is the penetration depth that is tolerated at a contact
	// point, instead of being corrected. Were penetration corrected in full,
	// two shapes at rest would end up exactly touching, at which point
	// collision detection no longer reports them as being in contact and
	// their tracked contact points are lost, only for the shapes to sink
	// into each other again.
	contactSlop = 0.001
)

// correctableContactDepth returns the part of the specified penetration
// depth of a contact that should be corrected.
func correctableContactDepth(depth float64) float64 {
	return max(0.0, depth-contactSlop)
}

// manifoldFrame is the reference frame of one of the two sides of a contact.
// Contact points are stored relative to it, so that they follow the side
// they belong to as it moves.
type manifoldFrame struct {
	position dprec.Vec3
	rotation dprec.Quat
}

// staticManifoldFrame returns the [manifoldFrame] of a side that never
// moves, for which local space and world space are one and the same.
func staticManifoldFrame() manifoldFrame {
	return manifoldFrame{
		position: dprec.ZeroVec3(),
		rotation: dprec.IdentityQuat(),
	}
}

func (f manifoldFrame) localPoint(point dprec.Vec3) dprec.Vec3 {
	return f.localDirection(dprec.Vec3Diff(point, f.position))
}

func (f manifoldFrame) worldPoint(point dprec.Vec3) dprec.Vec3 {
	return dprec.Vec3Sum(f.position, f.worldDirection(point))
}

func (f manifoldFrame) localDirection(direction dprec.Vec3) dprec.Vec3 {
	return dprec.QuatVec3Rotation(dprec.ConjugateQuat(f.rotation), direction)
}

func (f manifoldFrame) worldDirection(direction dprec.Vec3) dprec.Vec3 {
	return dprec.QuatVec3Rotation(f.rotation, direction)
}

// manifoldPoint is a single contact point of a [manifold].
type manifoldPoint struct {

	// sourcePoint is the contact point on the source side, in the local
	// space of the source frame.
	sourcePoint dprec.Vec3

	// targetPoint is the contact point on the target side, in the local
	// space of the target frame.
	targetPoint dprec.Vec3

	// targetNormal is the contact normal of the target side, in the local
	// space of the target frame.
	targetNormal dprec.Vec3

	// contact is the world-space description of the point, as of the last
	// time the point was detected or refreshed.
	contact shape3d.Contact
}

// manifold holds the contact points that have been gathered, over a number
// of consecutive simulation steps, for the two sides identified by its key.
type manifold[K comparable] struct {
	key        K
	revision   uint64
	pointCount int
	points     [maxManifoldPoints]manifoldPoint
}

// refresh brings the points up to date with the current frames of the two
// sides, discarding the ones whose two sides have moved apart.
func (m *manifold[K]) refresh(source, target manifoldFrame) {
	const sqrBreakingDistance = manifoldBreakingDistance * manifoldBreakingDistance

	keptCount := 0
	for i := range m.pointCount {
		point := m.points[i]

		sourcePoint := source.worldPoint(point.sourcePoint)
		targetPoint := target.worldPoint(point.targetPoint)
		targetNormal := target.worldDirection(point.targetNormal)

		delta := dprec.Vec3Diff(targetPoint, sourcePoint)
		depth := dprec.Vec3Dot(delta, targetNormal)
		if dprec.Abs(depth) > manifoldBreakingDistance {
			// The two sides have either separated or have moved into each
			// other by more than what a point that is extrapolated from a
			// past step can be trusted to describe. In the latter case the
			// penetration is left to newly detected contacts.
			continue
		}
		slide := dprec.Vec3Diff(delta, dprec.Vec3Prod(targetNormal, depth))
		if slide.SqrLength() > sqrBreakingDistance {
			continue // the two sides have slid past each other
		}

		point.contact = shape3d.Contact{
			TargetPoint:  targetPoint,
			TargetNormal: targetNormal,
			Depth:        depth,
		}
		m.points[keptCount] = point
		keptCount++
	}
	m.pointCount = keptCount
}

// add records a newly detected contact, which is expected to be expressed in
// world space and to be consistent with the specified frames.
func (m *manifold[K]) add(source, target manifoldFrame, contact shape3d.Contact) {
	point := manifoldPoint{
		sourcePoint:  source.localPoint(contact.EvalSourcePoint()),
		targetPoint:  target.localPoint(contact.TargetPoint),
		targetNormal: target.localDirection(contact.TargetNormal),
		contact:      contact,
	}
	m.points[m.slotFor(point)] = point
}

// slotFor returns the index at which the specified new point should be
// stored, increasing the number of points if a free slot is used.
func (m *manifold[K]) slotFor(point manifoldPoint) int {
	const sqrBreakingDistance = manifoldBreakingDistance * manifoldBreakingDistance

	// A point that is already being tracked is replaced by its newer state.
	for i := range m.pointCount {
		offset := dprec.Vec3Diff(m.points[i].sourcePoint, point.sourcePoint)
		if offset.SqrLength() <= sqrBreakingDistance {
			return i
		}
	}

	if m.pointCount < maxManifoldPoints {
		m.pointCount++
		return m.pointCount - 1
	}

	// The manifold is full, so an existing point has to make room. The
	// deepest point is always preserved, since it matters the most for
	// resolving the penetration. Among the rest, the one that is given up is
	// the one without which the manifold covers the largest area, since a
	// wide support is what keeps the two sides stable.
	deepestIndex := 0
	for i := 1; i < maxManifoldPoints; i++ {
		if m.points[i].contact.Depth > m.points[deepestIndex].contact.Depth {
			deepestIndex = i
		}
	}
	bestIndex := -1
	bestArea := -1.0
	for i := range maxManifoldPoints {
		if i == deepestIndex {
			continue
		}
		candidates := m.points
		candidates[i] = point
		area := manifoldArea(
			candidates[0].sourcePoint,
			candidates[1].sourcePoint,
			candidates[2].sourcePoint,
			candidates[3].sourcePoint,
		)
		if area > bestArea {
			bestArea = area
			bestIndex = i
		}
	}
	return bestIndex
}

// manifoldArea returns a measure of the area that is covered by the
// specified four points, regardless of their order. The measure is only
// meaningful for comparing sets of points with one another.
func manifoldArea(a, b, c, d dprec.Vec3) float64 {
	return max(
		dprec.Vec3Cross(dprec.Vec3Diff(a, b), dprec.Vec3Diff(c, d)).SqrLength(),
		dprec.Vec3Cross(dprec.Vec3Diff(a, c), dprec.Vec3Diff(b, d)).SqrLength(),
		dprec.Vec3Cross(dprec.Vec3Diff(a, d), dprec.Vec3Diff(b, c)).SqrLength(),
	)
}

// manifoldCache tracks the contact points between pairs of sides, each pair
// being identified by a key of type K, across simulation steps.
//
// Collision detection reports a single contact point for a pair of shapes,
// which is not enough to keep one shape resting on a face of another: the
// supported shape pivots around that single point. The cache accumulates the
// points reported over consecutive steps instead, for as long as they remain
// valid, so that such shapes end up being supported at several points at
// once.
//
// On each simulation step, [manifoldCache.Track] has to be called for every
// detected contact, followed by a single call to [manifoldCache.Commit].
type manifoldCache[K comparable] struct {
	revision  uint64
	indices   map[K]int
	manifolds []manifold[K]
}

func newManifoldCache[K comparable]() *manifoldCache[K] {
	return &manifoldCache[K]{
		indices:   make(map[K]int),
		manifolds: make([]manifold[K], 0),
	}
}

// Track records a contact that was detected during the current step between
// the two sides that are identified by key, whose current frames are source
// and target. The contact has to be expressed in world space.
func (c *manifoldCache[K]) Track(key K, source, target manifoldFrame, contact shape3d.Contact) {
	index, ok := c.indices[key]
	if !ok {
		index = len(c.manifolds)
		c.indices[key] = index
		c.manifolds = append(c.manifolds, manifold[K]{
			key:      key,
			revision: c.revision,
		})
	}
	manifold := &c.manifolds[index]
	if manifold.revision != c.revision {
		manifold.revision = c.revision
		manifold.refresh(source, target)
	}
	manifold.add(source, target, contact)
}

// Commit concludes the current step. It calls yield for every contact point
// that is in touch, expressed in world space, of the pairs for which a
// contact was tracked during the step.
//
// The remaining pairs are not in contact, though their points are held on
// to for a few more steps, in case the contact gets reestablished. This is
// what keeps two sides that barely touch from having to gather their
// contact points from scratch whenever they momentarily come apart.
func (c *manifoldCache[K]) Commit(yield func(key K, contact shape3d.Contact)) {
	for i := len(c.manifolds) - 1; i >= 0; i-- {
		if c.revision-c.manifolds[i].revision > maxManifoldIdleSteps {
			c.remove(i)
		}
	}
	for i := range c.manifolds {
		manifold := &c.manifolds[i]
		if manifold.revision != c.revision {
			continue
		}
		for _, point := range manifold.points[:manifold.pointCount] {
			if point.contact.Depth >= 0.0 {
				yield(manifold.key, point.contact)
			}
		}
	}
	c.revision++
}

// remove deletes the manifold at the specified index by having the last
// manifold take its place.
func (c *manifoldCache[K]) remove(index int) {
	lastIndex := len(c.manifolds) - 1
	delete(c.indices, c.manifolds[index].key)
	if index != lastIndex {
		c.manifolds[index] = c.manifolds[lastIndex]
		c.indices[c.manifolds[index].key] = index
	}
	c.manifolds = c.manifolds[:lastIndex]
}
