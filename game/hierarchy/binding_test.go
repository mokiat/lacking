package hierarchy_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/hierarchy"
)

// fullSolver implements all four binding solver interfaces, delegating to
// optional callbacks so that individual specs can observe only what they need.
type fullSolver struct {
	onDelete func(id hierarchy.NodeID, value string)
	onSource func(id hierarchy.NodeID, value string)
	onTarget func(id hierarchy.NodeID, value string)
	onInterp func(id hierarchy.NodeID, value string, fraction float64)
}

func (s *fullSolver) OnDelete(_ *hierarchy.Scene, id hierarchy.NodeID, value string) {
	if s.onDelete != nil {
		s.onDelete(id, value)
	}
}

func (s *fullSolver) OnSourceToNode(_ *hierarchy.Scene, id hierarchy.NodeID, value string) {
	if s.onSource != nil {
		s.onSource(id, value)
	}
}

func (s *fullSolver) OnTargetFromNode(_ *hierarchy.Scene, id hierarchy.NodeID, value string) {
	if s.onTarget != nil {
		s.onTarget(id, value)
	}
}

func (s *fullSolver) OnInterpolationFromNode(_ *hierarchy.Scene, id hierarchy.NodeID, value string, fraction float64) {
	if s.onInterp != nil {
		s.onInterp(id, value, fraction)
	}
}

// sourceOnlySolver implements only SourceBindingSolver, so a binding driven by
// it has no lifecycle solver.
type sourceOnlySolver struct {
	onSource func(id hierarchy.NodeID, value string)
}

func (s *sourceOnlySolver) OnSourceToNode(_ *hierarchy.Scene, id hierarchy.NodeID, value string) {
	if s.onSource != nil {
		s.onSource(id, value)
	}
}

// inertSolver implements none of the binding solver interfaces.
type inertSolver struct{}

var (
	_ hierarchy.LifecycleBindingSolver[string]     = (*fullSolver)(nil)
	_ hierarchy.SourceBindingSolver[string]        = (*fullSolver)(nil)
	_ hierarchy.TargetBindingSolver[string]        = (*fullSolver)(nil)
	_ hierarchy.InterpolationBindingSolver[string] = (*fullSolver)(nil)
	_ hierarchy.SourceBindingSolver[string]        = (*sourceOnlySolver)(nil)
)

var _ = Describe("Binding", func() {
	var (
		scene *hierarchy.Scene
		nodes hierarchy.NodeView
	)

	BeforeEach(func() {
		scene = hierarchy.NewScene()
		nodes = scene.Nodes()
	})

	Describe("value storage", func() {
		It("stores and retrieves bound values", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			id := nodes.Create()

			Expect(binding.Has(id)).To(BeFalse())

			binding.Bind(id, "hello")
			Expect(binding.Has(id)).To(BeTrue())
			Expect(binding.Get(id)).To(Equal("hello"))
		})

		It("replaces the value when a node is rebound", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			id := nodes.Create()

			binding.Bind(id, "first")
			binding.Bind(id, "second")
			Expect(binding.Get(id)).To(Equal("second"))
		})

		It("removes the value when a node is unbound", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			id := nodes.Create()

			binding.Bind(id, "value")
			binding.Unbind(id)
			Expect(binding.Has(id)).To(BeFalse())
		})

		It("returns the zero value for an unbound but valid node", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			id := nodes.Create()

			Expect(binding.Get(id)).To(BeZero())
		})

		It("panics when binding an invalid node", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			Expect(func() {
				binding.Bind(hierarchy.NilNodeID, "value")
			}).To(Panic())
		})

		It("panics when getting from an invalid node", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			Expect(func() {
				binding.Get(hierarchy.NilNodeID)
			}).To(Panic())
		})

		It("panics when checking an invalid node", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			Expect(func() {
				binding.Has(hierarchy.NilNodeID)
			}).To(Panic())
		})

		It("panics when unbinding an invalid node", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			Expect(func() {
				binding.Unbind(hierarchy.NilNodeID)
			}).To(Panic())
		})
	})

	Describe("source transfer", func() {
		It("applies to every bound node", func() {
			var seen []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onSource: func(_ hierarchy.NodeID, value string) {
					seen = append(seen, value)
				},
			})
			first := nodes.Create()
			second := nodes.Create()
			binding.Bind(first, "a")
			binding.Bind(second, "b")

			scene.ApplySourcesToNodes()

			Expect(seen).To(ConsistOf("a", "b"))
		})

		It("applies to a single node only", func() {
			var seen []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onSource: func(_ hierarchy.NodeID, value string) {
					seen = append(seen, value)
				},
			})
			first := nodes.Create()
			second := nodes.Create()
			binding.Bind(first, "a")
			binding.Bind(second, "b")

			scene.ApplySourceToNode(first, false)

			Expect(seen).To(ConsistOf("a"))
		})

		It("applies recursively to a subtree", func() {
			var seen []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onSource: func(_ hierarchy.NodeID, value string) {
					seen = append(seen, value)
				},
			})
			parent := nodes.Create()
			child := nodes.Create()
			grandChild := nodes.Create()
			nodes.AttachChild(parent, child, false)
			nodes.AttachChild(child, grandChild, false)
			binding.Bind(parent, "p")
			binding.Bind(child, "c")
			binding.Bind(grandChild, "g")

			scene.ApplySourceToNode(parent, true)

			Expect(seen).To(ConsistOf("p", "c", "g"))
		})

		It("does nothing for a solver that is not a source solver", func() {
			binding := hierarchy.NewBinding[string](scene, &inertSolver{})
			id := nodes.Create()
			binding.Bind(id, "a")

			Expect(func() {
				scene.ApplySourcesToNodes()
			}).ToNot(Panic())
		})
	})

	Describe("target transfer", func() {
		It("applies to every bound node", func() {
			var seen []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onTarget: func(_ hierarchy.NodeID, value string) {
					seen = append(seen, value)
				},
			})
			first := nodes.Create()
			second := nodes.Create()
			binding.Bind(first, "a")
			binding.Bind(second, "b")

			scene.ApplyTargetsFromNodes()

			Expect(seen).To(ConsistOf("a", "b"))
		})

		It("applies to a single node only", func() {
			var seen []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onTarget: func(_ hierarchy.NodeID, value string) {
					seen = append(seen, value)
				},
			})
			first := nodes.Create()
			second := nodes.Create()
			binding.Bind(first, "a")
			binding.Bind(second, "b")

			scene.ApplyTargetFromNode(first, false)

			Expect(seen).To(ConsistOf("a"))
		})
	})

	Describe("interpolation transfer", func() {
		It("applies to every bound node and forwards the fraction", func() {
			type record struct {
				value    string
				fraction float64
			}
			var seen []record
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onInterp: func(_ hierarchy.NodeID, value string, fraction float64) {
					seen = append(seen, record{value: value, fraction: fraction})
				},
			})
			id := nodes.Create()
			binding.Bind(id, "a")

			scene.ApplyInterpolationsFromNodes(0.25)

			Expect(seen).To(ConsistOf(record{value: "a", fraction: 0.25}))
		})

		It("applies recursively to a subtree", func() {
			var seen []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onInterp: func(_ hierarchy.NodeID, value string, _ float64) {
					seen = append(seen, value)
				},
			})
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)
			binding.Bind(parent, "p")
			binding.Bind(child, "c")

			scene.ApplyInterpolationFromNode(parent, 1.0, true)

			Expect(seen).To(ConsistOf("p", "c"))
		})
	})

	Describe("priority", func() {
		It("processes bindings in ascending priority order", func() {
			var order []string
			high := hierarchy.NewBinding[string](scene, &fullSolver{
				onSource: func(_ hierarchy.NodeID, _ string) {
					order = append(order, "high")
				},
			})
			low := hierarchy.NewBinding[string](scene, &fullSolver{
				onSource: func(_ hierarchy.NodeID, _ string) {
					order = append(order, "low")
				},
			})
			high.SetPriority(10)
			low.SetPriority(1)

			id := nodes.Create()
			high.Bind(id, "x")
			low.Bind(id, "x")

			scene.ApplySourcesToNodes()

			Expect(order).To(Equal([]string{"low", "high"}))
		})
	})

	Describe("lifecycle", func() {
		It("notifies the lifecycle solver when a bound node is deleted", func() {
			var deleted []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onDelete: func(_ hierarchy.NodeID, value string) {
					deleted = append(deleted, value)
				},
			})
			id := nodes.Create()
			binding.Bind(id, "a")

			nodes.Delete(id)

			Expect(deleted).To(ConsistOf("a"))
		})

		It("notifies the lifecycle solver for a deleted descendant", func() {
			var deleted []string
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onDelete: func(_ hierarchy.NodeID, value string) {
					deleted = append(deleted, value)
				},
			})
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)
			binding.Bind(parent, "p")
			binding.Bind(child, "c")

			nodes.Delete(parent)

			Expect(deleted).To(ConsistOf("p", "c"))
		})

		It("stops applying transfers once the binding is deleted", func() {
			count := 0
			binding := hierarchy.NewBinding[string](scene, &fullSolver{
				onSource: func(_ hierarchy.NodeID, _ string) {
					count++
				},
			})
			id := nodes.Create()
			binding.Bind(id, "a")

			binding.Delete()
			scene.ApplySourcesToNodes()

			Expect(count).To(BeZero())
		})

		It("purges the entry of a deleted node even without a lifecycle solver", func() {
			// A source-only binding has no lifecycle solver, yet deleting a bound
			// node must still remove its entry so that no stale, now-invalid node
			// is visited on the next transfer pass.
			count := 0
			binding := hierarchy.NewBinding[string](scene, &sourceOnlySolver{
				onSource: func(_ hierarchy.NodeID, _ string) {
					count++
				},
			})
			id := nodes.Create()
			binding.Bind(id, "a")

			nodes.Delete(id)
			scene.ApplySourcesToNodes()

			Expect(count).To(BeZero())
		})
	})
})
