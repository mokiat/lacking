package hierarchy_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/game/hierarchy"
)

var _ = Describe("Scene", func() {
	var (
		scene *hierarchy.Scene
		nodes hierarchy.NodeView
	)

	BeforeEach(func() {
		scene = hierarchy.NewScene()
		nodes = scene.Nodes()
	})

	// collectChildren returns the direct children of the specified node in
	// FirstChild -> NextSibling order.
	collectChildren := func(id hierarchy.NodeID) []hierarchy.NodeID {
		var result []hierarchy.NodeID
		for child := nodes.FirstChild(id); child != hierarchy.NilNodeID; child = nodes.NextSibling(child) {
			result = append(result, child)
		}
		return result
	}

	// collectAll returns all valid nodes reachable through Each.
	collectAll := func() []hierarchy.NodeID {
		var result []hierarchy.NodeID
		nodes.Each(func(id hierarchy.NodeID) bool {
			result = append(result, id)
			return true
		})
		return result
	}

	// expectMatrix asserts that the actual matrix matches the expected one.
	expectMatrix := func(actual, expected dprec.Mat4) {
		GinkgoHelper()
		Expect(actual).To(dprectest.HaveMat4Elements(
			expected.M11, expected.M12, expected.M13, expected.M14,
			expected.M21, expected.M22, expected.M23, expected.M24,
			expected.M31, expected.M32, expected.M33, expected.M34,
			expected.M41, expected.M42, expected.M43, expected.M44,
		))
	}

	// translation builds a pure translation matrix.
	translation := func(x, y, z float64) dprec.Mat4 {
		return dprec.TRSMat4(
			dprec.NewVec3(x, y, z),
			dprec.IdentityQuat(),
			dprec.NewVec3(1.0, 1.0, 1.0),
		)
	}

	Describe("NewScene", func() {
		It("creates a scene with no nodes", func() {
			Expect(collectAll()).To(BeEmpty())
		})
	})

	Describe("node creation", func() {
		It("creates a valid node", func() {
			id := nodes.Create()
			Expect(nodes.IsValid(id)).To(BeTrue())
		})

		It("creates a node that is a root", func() {
			id := nodes.Create()
			Expect(nodes.IsRoot(id)).To(BeTrue())
			Expect(nodes.Parent(id)).To(Equal(hierarchy.NilNodeID))
		})

		It("creates a node without children or siblings", func() {
			id := nodes.Create()
			Expect(nodes.FirstChild(id)).To(Equal(hierarchy.NilNodeID))
			Expect(nodes.NextSibling(id)).To(Equal(hierarchy.NilNodeID))
		})

		It("creates a node with an empty name", func() {
			id := nodes.Create()
			Expect(nodes.Name(id)).To(BeEmpty())
		})

		It("creates distinct nodes", func() {
			first := nodes.Create()
			second := nodes.Create()
			Expect(first).ToNot(Equal(second))
			Expect(collectAll()).To(ConsistOf(first, second))
		})
	})

	Describe("NilNodeID", func() {
		It("is not valid", func() {
			Expect(nodes.IsValid(hierarchy.NilNodeID)).To(BeFalse())
		})
	})

	Describe("name management", func() {
		It("stores and returns the name", func() {
			id := nodes.Create()
			nodes.SetName(id, "player")
			Expect(nodes.Name(id)).To(Equal("player"))
		})

		It("overwrites a previously set name", func() {
			id := nodes.Create()
			nodes.SetName(id, "first")
			nodes.SetName(id, "second")
			Expect(nodes.Name(id)).To(Equal("second"))
		})
	})

	Describe("FindNode", func() {
		It("finds a node by its name", func() {
			id := nodes.Create()
			nodes.SetName(id, "player")
			Expect(nodes.FindNode("player")).To(Equal(id))
		})

		It("finds a node regardless of its place in the hierarchy", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)
			nodes.SetName(child, "target")

			Expect(nodes.FindNode("target")).To(Equal(child))
		})

		It("returns NilNodeID when no node has the name", func() {
			id := nodes.Create()
			nodes.SetName(id, "player")
			Expect(nodes.FindNode("enemy")).To(Equal(hierarchy.NilNodeID))
		})

		It("returns NilNodeID for an empty scene", func() {
			Expect(nodes.FindNode("player")).To(Equal(hierarchy.NilNodeID))
		})

		It("matches the name exactly", func() {
			id := nodes.Create()
			nodes.SetName(id, "player")
			Expect(nodes.FindNode("play")).To(Equal(hierarchy.NilNodeID))
			Expect(nodes.FindNode("Player")).To(Equal(hierarchy.NilNodeID))
		})

		It("does not find a deleted node", func() {
			id := nodes.Create()
			nodes.SetName(id, "player")
			nodes.Delete(id)
			Expect(nodes.FindNode("player")).To(Equal(hierarchy.NilNodeID))
		})

		It("finds a valid node that reuses a deleted node's slot", func() {
			first := nodes.Create()
			nodes.SetName(first, "player")
			nodes.Delete(first)

			second := nodes.Create()
			nodes.SetName(second, "player")

			found := nodes.FindNode("player")
			Expect(found).To(Equal(second))
			Expect(nodes.IsValid(found)).To(BeTrue())
		})
	})

	Describe("FindNodeInSubtree", func() {
		// Builds the tree:
		//   root ("root")
		//     child ("child")
		//       grandChild ("grandChild")
		//   outsider ("outsider")
		var root, child, grandChild, outsider hierarchy.NodeID

		BeforeEach(func() {
			root = nodes.Create()
			child = nodes.Create()
			grandChild = nodes.Create()
			outsider = nodes.Create()
			nodes.SetName(root, "root")
			nodes.SetName(child, "child")
			nodes.SetName(grandChild, "grandChild")
			nodes.SetName(outsider, "outsider")
			nodes.AttachChild(root, child, false)
			nodes.AttachChild(child, grandChild, false)
		})

		It("finds the root node of the subtree itself", func() {
			Expect(nodes.FindNodeInSubtree(root, "root")).To(Equal(root))
		})

		It("finds a direct child within the subtree", func() {
			Expect(nodes.FindNodeInSubtree(root, "child")).To(Equal(child))
		})

		It("finds a deeply nested descendant within the subtree", func() {
			Expect(nodes.FindNodeInSubtree(root, "grandChild")).To(Equal(grandChild))
		})

		It("finds a descendant when starting from an intermediate node", func() {
			Expect(nodes.FindNodeInSubtree(child, "grandChild")).To(Equal(grandChild))
		})

		It("does not find a node outside the subtree", func() {
			Expect(nodes.FindNodeInSubtree(root, "outsider")).To(Equal(hierarchy.NilNodeID))
		})

		It("does not find an ancestor of the subtree root", func() {
			Expect(nodes.FindNodeInSubtree(child, "root")).To(Equal(hierarchy.NilNodeID))
		})

		It("returns NilNodeID when no node in the subtree has the name", func() {
			Expect(nodes.FindNodeInSubtree(root, "missing")).To(Equal(hierarchy.NilNodeID))
		})

		It("searches the whole scene when the root is NilNodeID", func() {
			Expect(nodes.FindNodeInSubtree(hierarchy.NilNodeID, "outsider")).To(Equal(outsider))
		})

		It("returns NilNodeID when the root ID is invalid", func() {
			nodes.Delete(root)
			Expect(nodes.FindNodeInSubtree(root, "grandChild")).To(Equal(hierarchy.NilNodeID))
		})

		It("matches the name exactly", func() {
			Expect(nodes.FindNodeInSubtree(root, "Child")).To(Equal(hierarchy.NilNodeID))
		})
	})

	Describe("AttachChild", func() {
		var parent, child hierarchy.NodeID

		BeforeEach(func() {
			parent = nodes.Create()
			child = nodes.Create()
			nodes.AttachChild(parent, child, false)
		})

		It("establishes the parent-child relationship", func() {
			Expect(nodes.Parent(child)).To(Equal(parent))
			Expect(nodes.FirstChild(parent)).To(Equal(child))
		})

		It("makes the child a non-root", func() {
			Expect(nodes.IsRoot(child)).To(BeFalse())
		})

		It("keeps the parent a root", func() {
			Expect(nodes.IsRoot(parent)).To(BeTrue())
		})

		It("is a no-op when re-attaching to the same parent", func() {
			nodes.AttachChild(parent, child, false)
			Expect(collectChildren(parent)).To(ConsistOf(child))
			Expect(nodes.Parent(child)).To(Equal(parent))
		})

		It("registers multiple children of the same parent", func() {
			second := nodes.Create()
			nodes.AttachChild(parent, second, false)
			Expect(collectChildren(parent)).To(ConsistOf(child, second))
			Expect(nodes.Parent(second)).To(Equal(parent))
		})

		It("re-parents a child that already has a parent", func() {
			newParent := nodes.Create()
			nodes.AttachChild(newParent, child, false)
			Expect(nodes.Parent(child)).To(Equal(newParent))
			Expect(collectChildren(parent)).To(BeEmpty())
			Expect(collectChildren(newParent)).To(ConsistOf(child))
		})
	})

	Describe("Detach", func() {
		It("detaches a child from its parent", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)

			nodes.Detach(child, false)

			Expect(nodes.IsRoot(child)).To(BeTrue())
			Expect(nodes.Parent(child)).To(Equal(hierarchy.NilNodeID))
			Expect(collectChildren(parent)).To(BeEmpty())
		})

		It("detaches a middle child while keeping its siblings", func() {
			parent := nodes.Create()
			first := nodes.Create()
			second := nodes.Create()
			third := nodes.Create()
			nodes.AttachChild(parent, first, false)
			nodes.AttachChild(parent, second, false)
			nodes.AttachChild(parent, third, false)

			nodes.Detach(second, false)

			Expect(collectChildren(parent)).To(ConsistOf(first, third))
			Expect(nodes.Parent(second)).To(Equal(hierarchy.NilNodeID))
		})

		It("is a no-op on a root node", func() {
			id := nodes.Create()
			nodes.Detach(id, false)
			Expect(nodes.IsRoot(id)).To(BeTrue())
			Expect(nodes.IsValid(id)).To(BeTrue())
		})
	})

	Describe("Delete", func() {
		It("invalidates the deleted node", func() {
			id := nodes.Create()
			nodes.Delete(id)
			Expect(nodes.IsValid(id)).To(BeFalse())
		})

		It("removes the node from iteration", func() {
			first := nodes.Create()
			second := nodes.Create()
			nodes.Delete(first)
			Expect(collectAll()).To(ConsistOf(second))
		})

		It("detaches the node from its parent", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)

			nodes.Delete(child)

			Expect(collectChildren(parent)).To(BeEmpty())
			Expect(nodes.IsValid(parent)).To(BeTrue())
		})

		It("recursively deletes descendants", func() {
			parent := nodes.Create()
			child := nodes.Create()
			grandChild := nodes.Create()
			nodes.AttachChild(parent, child, false)
			nodes.AttachChild(child, grandChild, false)

			nodes.Delete(parent)

			Expect(nodes.IsValid(parent)).To(BeFalse())
			Expect(nodes.IsValid(child)).To(BeFalse())
			Expect(nodes.IsValid(grandChild)).To(BeFalse())
			Expect(collectAll()).To(BeEmpty())
		})

		It("invalidates the old ID after the index is reused", func() {
			first := nodes.Create()
			nodes.Delete(first)
			second := nodes.Create()
			Expect(nodes.IsValid(first)).To(BeFalse())
			Expect(nodes.IsValid(second)).To(BeTrue())
			Expect(first).ToNot(Equal(second))
		})
	})

	Describe("Each", func() {
		It("visits every valid node", func() {
			first := nodes.Create()
			second := nodes.Create()
			third := nodes.Create()
			Expect(collectAll()).To(ConsistOf(first, second, third))
		})

		It("stops iterating when the callback returns false", func() {
			nodes.Create()
			nodes.Create()
			nodes.Create()

			var visited int
			nodes.Each(func(hierarchy.NodeID) bool {
				visited++
				return false
			})
			Expect(visited).To(Equal(1))
		})
	})

	Describe("Iter", func() {
		It("yields every valid node", func() {
			first := nodes.Create()
			second := nodes.Create()

			var visited []hierarchy.NodeID
			for id := range nodes.Iter() {
				visited = append(visited, id)
			}
			Expect(visited).To(ConsistOf(first, second))
		})

		It("stops yielding when the loop breaks early", func() {
			nodes.Create()
			nodes.Create()
			nodes.Create()

			var visited int
			for range nodes.Iter() {
				visited++
				break
			}
			Expect(visited).To(Equal(1))
		})
	})

	Describe("EachRoot", func() {
		collectRoots := func() []hierarchy.NodeID {
			var result []hierarchy.NodeID
			nodes.EachRoot(func(id hierarchy.NodeID) bool {
				result = append(result, id)
				return true
			})
			return result
		}

		It("visits every root node", func() {
			first := nodes.Create()
			second := nodes.Create()
			Expect(collectRoots()).To(ConsistOf(first, second))
		})

		It("does not visit non-root nodes", func() {
			root := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(root, child, false)
			Expect(collectRoots()).To(ConsistOf(root))
		})

		It("visits a node once it becomes a root again", func() {
			root := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(root, child, false)
			nodes.Detach(child, false)
			Expect(collectRoots()).To(ConsistOf(root, child))
		})

		It("visits nothing in an empty scene", func() {
			Expect(collectRoots()).To(BeEmpty())
		})

		It("does not visit deleted nodes", func() {
			first := nodes.Create()
			second := nodes.Create()
			nodes.Delete(first)
			Expect(collectRoots()).To(ConsistOf(second))
		})

		It("stops iterating when the callback returns false", func() {
			nodes.Create()
			nodes.Create()
			nodes.Create()

			var visited int
			nodes.EachRoot(func(hierarchy.NodeID) bool {
				visited++
				return false
			})
			Expect(visited).To(Equal(1))
		})
	})

	Describe("RootIter", func() {
		It("yields every root node", func() {
			first := nodes.Create()
			second := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(first, child, false)

			var visited []hierarchy.NodeID
			for id := range nodes.RootIter() {
				visited = append(visited, id)
			}
			Expect(visited).To(ConsistOf(first, second))
		})

		It("stops yielding when the loop breaks early", func() {
			nodes.Create()
			nodes.Create()
			nodes.Create()

			var visited int
			for range nodes.RootIter() {
				visited++
				break
			}
			Expect(visited).To(Equal(1))
		})
	})

	Describe("walking", func() {
		// Builds the tree:
		//   root
		//     branch
		//       leaf
		//     sibling
		var root, branch, leaf, sibling hierarchy.NodeID

		BeforeEach(func() {
			root = nodes.Create()
			branch = nodes.Create()
			leaf = nodes.Create()
			sibling = nodes.Create()
			nodes.AttachChild(root, branch, false)
			nodes.AttachChild(branch, leaf, false)
			nodes.AttachChild(root, sibling, false)
		})

		// indexOf returns the position of id in the visited slice, or -1.
		indexOf := func(visited []hierarchy.NodeID, id hierarchy.NodeID) int {
			for i, cur := range visited {
				if cur == id {
					return i
				}
			}
			return -1
		}

		// expectPreOrder asserts that every node appears after its parent.
		expectPreOrder := func(visited []hierarchy.NodeID) {
			GinkgoHelper()
			Expect(indexOf(visited, root)).To(BeNumerically("<", indexOf(visited, branch)))
			Expect(indexOf(visited, root)).To(BeNumerically("<", indexOf(visited, sibling)))
			Expect(indexOf(visited, branch)).To(BeNumerically("<", indexOf(visited, leaf)))
		}

		walkAll := func() []hierarchy.NodeID {
			var visited []hierarchy.NodeID
			nodes.Walk(func(id hierarchy.NodeID) bool {
				visited = append(visited, id)
				return true
			})
			return visited
		}

		walkSubtree := func(rootID hierarchy.NodeID) []hierarchy.NodeID {
			var visited []hierarchy.NodeID
			nodes.WalkSubtree(rootID, func(id hierarchy.NodeID) bool {
				visited = append(visited, id)
				return true
			})
			return visited
		}

		Describe("Walk", func() {
			It("visits every node in the scene", func() {
				Expect(walkAll()).To(ConsistOf(root, branch, leaf, sibling))
			})

			It("visits each node before its children", func() {
				expectPreOrder(walkAll())
			})

			It("visits the subtrees of all roots", func() {
				otherRoot := nodes.Create()
				otherChild := nodes.Create()
				nodes.AttachChild(otherRoot, otherChild, false)

				Expect(walkAll()).To(ConsistOf(root, branch, leaf, sibling, otherRoot, otherChild))
			})

			It("stops walking when the callback returns false", func() {
				var visited int
				nodes.Walk(func(hierarchy.NodeID) bool {
					visited++
					return false
				})
				Expect(visited).To(Equal(1))
			})

			It("visits nothing in an empty scene", func() {
				emptyScene := hierarchy.NewScene()
				var visited int
				emptyScene.Nodes().Walk(func(hierarchy.NodeID) bool {
					visited++
					return true
				})
				Expect(visited).To(BeZero())
			})
		})

		Describe("WalkSubtree", func() {
			It("visits the root of the subtree first", func() {
				visited := walkSubtree(branch)
				Expect(visited).ToNot(BeEmpty())
				Expect(visited[0]).To(Equal(branch))
			})

			It("visits the root and all its descendants", func() {
				Expect(walkSubtree(root)).To(ConsistOf(root, branch, leaf, sibling))
			})

			It("visits only the specified subtree", func() {
				Expect(walkSubtree(branch)).To(ConsistOf(branch, leaf))
			})

			It("visits each node before its children", func() {
				expectPreOrder(walkSubtree(root))
			})

			It("is a no-op for NilNodeID", func() {
				Expect(walkSubtree(hierarchy.NilNodeID)).To(BeEmpty())
			})

			It("is a no-op for a deleted node", func() {
				nodes.Delete(branch)
				Expect(walkSubtree(branch)).To(BeEmpty())
			})

			It("stops walking when the callback returns false", func() {
				var visited int
				nodes.WalkSubtree(root, func(hierarchy.NodeID) bool {
					visited++
					return false
				})
				Expect(visited).To(Equal(1))
			})
		})

		Describe("WalkIter", func() {
			It("yields every node in the scene", func() {
				var visited []hierarchy.NodeID
				for id := range nodes.WalkIter() {
					visited = append(visited, id)
				}
				Expect(visited).To(ConsistOf(root, branch, leaf, sibling))
			})

			It("stops yielding when the loop breaks early", func() {
				var visited int
				for range nodes.WalkIter() {
					visited++
					break
				}
				Expect(visited).To(Equal(1))
			})
		})

		Describe("WalkSubtreeIter", func() {
			It("yields the root and all its descendants", func() {
				var visited []hierarchy.NodeID
				for id := range nodes.WalkSubtreeIter(branch) {
					visited = append(visited, id)
				}
				Expect(visited).To(ConsistOf(branch, leaf))
			})

			It("yields nothing for an invalid root", func() {
				var visited int
				for range nodes.WalkSubtreeIter(hierarchy.NilNodeID) {
					visited++
				}
				Expect(visited).To(BeZero())
			})

			It("stops yielding when the loop breaks early", func() {
				var visited int
				for range nodes.WalkSubtreeIter(root) {
					visited++
					break
				}
				Expect(visited).To(Equal(1))
			})
		})

		Describe("SubtreeContains", func() {
			It("reports a node as containing itself", func() {
				Expect(nodes.SubtreeContains(branch, branch)).To(BeTrue())
			})

			It("reports a direct child as contained", func() {
				Expect(nodes.SubtreeContains(root, branch)).To(BeTrue())
			})

			It("reports a deeply nested descendant as contained", func() {
				Expect(nodes.SubtreeContains(root, leaf)).To(BeTrue())
			})

			It("does not report a node outside the subtree", func() {
				Expect(nodes.SubtreeContains(branch, sibling)).To(BeFalse())
			})

			It("does not report an ancestor as contained", func() {
				Expect(nodes.SubtreeContains(branch, root)).To(BeFalse())
			})

			It("returns false when the root is invalid", func() {
				Expect(nodes.SubtreeContains(hierarchy.NilNodeID, leaf)).To(BeFalse())
			})

			It("returns false when the searched node is invalid", func() {
				Expect(nodes.SubtreeContains(root, hierarchy.NilNodeID)).To(BeFalse())
			})

			It("returns false for a deleted searched node", func() {
				nodes.Delete(leaf)
				Expect(nodes.SubtreeContains(root, leaf)).To(BeFalse())
			})
		})
	})

	Describe("NodeHandle", func() {
		It("wraps a created node", func() {
			handle := nodes.CreateHandle()
			Expect(handle.IsValid()).To(BeTrue())
			Expect(handle.IsRoot()).To(BeTrue())
			Expect(nodes.IsValid(handle.ID())).To(BeTrue())
		})

		It("stores and returns the name", func() {
			handle := nodes.CreateHandle()
			handle.SetName("camera")
			Expect(handle.Name()).To(Equal("camera"))
		})

		It("navigates the hierarchy", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			nodes.AttachChild(parent.ID(), child.ID(), false)

			Expect(parent.FirstChild().ID()).To(Equal(child.ID()))
			Expect(child.Parent().ID()).To(Equal(parent.ID()))
			Expect(child.IsRoot()).To(BeFalse())
		})

		It("returns a nil handle when navigating past the hierarchy edges", func() {
			handle := nodes.CreateHandle()
			Expect(handle.Parent().ID()).To(Equal(hierarchy.NilNodeID))
			Expect(handle.FirstChild().ID()).To(Equal(hierarchy.NilNodeID))
			Expect(handle.NextSibling().ID()).To(Equal(hierarchy.NilNodeID))
		})

		It("attaches a child through the parent handle", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()

			parent.AttachChild(child, false)

			Expect(child.Parent().ID()).To(Equal(parent.ID()))
			Expect(parent.FirstChild().ID()).To(Equal(child.ID()))
			Expect(child.IsRoot()).To(BeFalse())
		})

		It("preserves the world transform when attaching through the parent handle", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			parent.SetPosition(dprec.NewVec3(10.0, 0.0, 0.0))
			child.SetPosition(dprec.NewVec3(1.0, 0.0, 0.0))

			parent.AttachChild(child, true)

			Expect(child.Position()).To(dprectest.HaveVec3Coords(-9.0, 0.0, 0.0))
		})

		It("detaches the wrapped node through the handle", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			parent.AttachChild(child, false)

			child.Detach(false)

			Expect(child.IsRoot()).To(BeTrue())
			Expect(child.Parent().ID()).To(Equal(hierarchy.NilNodeID))
			Expect(parent.FirstChild().ID()).To(Equal(hierarchy.NilNodeID))
		})

		It("preserves the world transform when detaching through the handle", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			parent.SetPosition(dprec.NewVec3(10.0, 0.0, 0.0))
			child.SetPosition(dprec.NewVec3(1.0, 0.0, 0.0))
			parent.AttachChild(child, false)
			// world position is now (11, 0, 0)

			child.Detach(true)

			expectMatrix(child.AbsoluteMatrix(), translation(11.0, 0.0, 0.0))
		})

		It("deletes the wrapped node", func() {
			handle := nodes.CreateHandle()
			handle.Delete()
			Expect(handle.IsValid()).To(BeFalse())
		})

		It("stores and returns the hidden state", func() {
			handle := nodes.CreateHandle()
			Expect(handle.IsHidden()).To(BeFalse())
			Expect(handle.IsAbsoluteHidden()).To(BeFalse())

			handle.SetHidden(true)
			Expect(handle.IsHidden()).To(BeTrue())
			Expect(handle.IsAbsoluteHidden()).To(BeTrue())
			Expect(nodes.IsHidden(handle.ID())).To(BeTrue())
		})

		It("reports a child as absolutely hidden when its parent is hidden", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			nodes.AttachChild(parent.ID(), child.ID(), false)

			parent.SetHidden(true)

			Expect(child.IsHidden()).To(BeFalse())
			Expect(child.IsAbsoluteHidden()).To(BeTrue())
		})

		It("mirrors the hidden state through the visibility accessors", func() {
			handle := nodes.CreateHandle()
			Expect(handle.IsVisible()).To(BeTrue())
			Expect(handle.IsAbsoluteVisible()).To(BeTrue())

			handle.SetVisible(false)
			Expect(handle.IsHidden()).To(BeTrue())
			Expect(handle.IsVisible()).To(BeFalse())
			Expect(handle.IsAbsoluteVisible()).To(BeFalse())
			Expect(nodes.IsVisible(handle.ID())).To(BeFalse())

			handle.SetVisible(true)
			Expect(handle.IsVisible()).To(BeTrue())
			Expect(handle.IsAbsoluteVisible()).To(BeTrue())
		})

		It("stores and returns the position", func() {
			handle := nodes.CreateHandle()
			handle.SetPosition(dprec.NewVec3(1.0, 2.0, 3.0))
			Expect(handle.Position()).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
			Expect(nodes.Position(handle.ID())).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
		})

		It("stores and returns the rotation", func() {
			handle := nodes.CreateHandle()
			rotation := dprec.RotationQuat(dprec.Degrees(90.0), dprec.NewVec3(0.0, 1.0, 0.0))
			handle.SetRotation(rotation)
			Expect(handle.Rotation()).To(dprectest.HaveQuatCoords(rotation.W, rotation.X, rotation.Y, rotation.Z))
		})

		It("stores and returns the scale", func() {
			handle := nodes.CreateHandle()
			handle.SetScale(dprec.NewVec3(2.0, 3.0, 4.0))
			Expect(handle.Scale()).To(dprectest.HaveVec3Coords(2.0, 3.0, 4.0))
		})

		It("stores and returns the translation, rotation, and scale together", func() {
			handle := nodes.CreateHandle()
			rotation := dprec.RotationQuat(dprec.Degrees(90.0), dprec.NewVec3(0.0, 1.0, 0.0))
			handle.SetTRS(dprec.NewVec3(1.0, 2.0, 3.0), rotation, dprec.NewVec3(4.0, 5.0, 6.0))

			position, storedRotation, scale := handle.TRS()
			Expect(position).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
			Expect(storedRotation).To(dprectest.HaveQuatCoords(rotation.W, rotation.X, rotation.Y, rotation.Z))
			Expect(scale).To(dprectest.HaveVec3Coords(4.0, 5.0, 6.0))
		})

		It("stores and returns the local matrix", func() {
			handle := nodes.CreateHandle()
			matrix := dprec.TRSMat4(
				dprec.NewVec3(1.0, 2.0, 3.0),
				dprec.IdentityQuat(),
				dprec.NewVec3(2.0, 2.0, 2.0),
			)
			handle.SetMatrix(matrix)

			Expect(handle.Position()).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
			Expect(handle.Scale()).To(dprectest.HaveVec3Coords(2.0, 2.0, 2.0))
			Expect(handle.Matrix()).To(dprectest.HaveMat4Elements(
				matrix.M11, matrix.M12, matrix.M13, matrix.M14,
				matrix.M21, matrix.M22, matrix.M23, matrix.M24,
				matrix.M31, matrix.M32, matrix.M33, matrix.M34,
				matrix.M41, matrix.M42, matrix.M43, matrix.M44,
			))
		})

		It("finds the wrapped node itself by name", func() {
			handle := nodes.CreateHandle()
			handle.SetName("root")
			Expect(handle.FindNode("root").ID()).To(Equal(handle.ID()))
		})

		It("finds a descendant by name", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			parent.AttachChild(child, false)
			child.SetName("target")

			Expect(parent.FindNode("target").ID()).To(Equal(child.ID()))
		})

		It("does not find a node outside its subtree", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			parent.AttachChild(child, false)
			child.SetName("target")

			// Searching from the child cannot reach its parent.
			Expect(child.FindNode("target").ID()).To(Equal(child.ID()))
			outsider := nodes.CreateHandle()
			outsider.SetName("outsider")
			Expect(parent.FindNode("outsider").ID()).To(Equal(hierarchy.NilNodeID))
		})

		It("returns a nil handle when no node in the subtree has the name", func() {
			handle := nodes.CreateHandle()
			handle.SetName("root")
			result := handle.FindNode("missing")
			Expect(result.ID()).To(Equal(hierarchy.NilNodeID))
			Expect(result.IsValid()).To(BeFalse())
		})

		It("walks the wrapped node and its descendants", func() {
			root := nodes.CreateHandle()
			branch := nodes.CreateHandle()
			leaf := nodes.CreateHandle()
			root.AttachChild(branch, false)
			branch.AttachChild(leaf, false)

			var visited []hierarchy.NodeID
			root.Walk(func(h hierarchy.NodeHandle) bool {
				visited = append(visited, h.ID())
				return true
			})
			Expect(visited).To(ConsistOf(root.ID(), branch.ID(), leaf.ID()))
			Expect(visited[0]).To(Equal(root.ID()))
		})

		It("stops walking when the callback returns false", func() {
			root := nodes.CreateHandle()
			child := nodes.CreateHandle()
			root.AttachChild(child, false)

			var visited int
			root.Walk(func(hierarchy.NodeHandle) bool {
				visited++
				return false
			})
			Expect(visited).To(Equal(1))
		})

		It("yields the wrapped node and its descendants through the iterator", func() {
			root := nodes.CreateHandle()
			child := nodes.CreateHandle()
			root.AttachChild(child, false)

			var visited []hierarchy.NodeID
			for h := range root.WalkIter() {
				visited = append(visited, h.ID())
			}
			Expect(visited).To(ConsistOf(root.ID(), child.ID()))
		})

		It("stops yielding when the loop breaks early", func() {
			root := nodes.CreateHandle()
			child := nodes.CreateHandle()
			root.AttachChild(child, false)

			var visited int
			for range root.WalkIter() {
				visited++
				break
			}
			Expect(visited).To(Equal(1))
		})

		It("reports a descendant as contained in its subtree", func() {
			root := nodes.CreateHandle()
			child := nodes.CreateHandle()
			root.AttachChild(child, false)

			Expect(root.SubtreeContains(child)).To(BeTrue())
			Expect(root.SubtreeContains(root)).To(BeTrue())
		})

		It("does not report an unrelated node as contained", func() {
			root := nodes.CreateHandle()
			child := nodes.CreateHandle()
			root.AttachChild(child, false)
			outsider := nodes.CreateHandle()

			Expect(root.SubtreeContains(outsider)).To(BeFalse())
			Expect(child.SubtreeContains(root)).To(BeFalse())
		})
	})

	Describe("hidden state", func() {
		It("is not hidden by default", func() {
			id := nodes.Create()
			Expect(nodes.IsHidden(id)).To(BeFalse())
			Expect(nodes.IsAbsoluteHidden(id)).To(BeFalse())
		})

		It("becomes hidden when explicitly hidden", func() {
			id := nodes.Create()
			nodes.SetHidden(id, true)
			Expect(nodes.IsHidden(id)).To(BeTrue())
			Expect(nodes.IsAbsoluteHidden(id)).To(BeTrue())
		})

		It("becomes visible again when unhidden", func() {
			id := nodes.Create()
			nodes.SetHidden(id, true)
			nodes.SetHidden(id, false)
			Expect(nodes.IsHidden(id)).To(BeFalse())
			Expect(nodes.IsAbsoluteHidden(id)).To(BeFalse())
		})

		It("is idempotent when hiding an already hidden node", func() {
			id := nodes.Create()
			nodes.SetHidden(id, true)
			nodes.SetHidden(id, true)
			Expect(nodes.IsHidden(id)).To(BeTrue())
		})

		It("does not report a visible child as explicitly hidden when its parent is hidden", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)

			nodes.SetHidden(parent, true)

			Expect(nodes.IsHidden(child)).To(BeFalse())
		})

		It("marks a child as absolutely hidden when its parent is hidden", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)

			nodes.SetHidden(parent, true)

			Expect(nodes.IsAbsoluteHidden(child)).To(BeTrue())
		})

		It("marks a whole subtree as absolutely hidden when an ancestor is hidden", func() {
			parent := nodes.Create()
			child := nodes.Create()
			grandChild := nodes.Create()
			nodes.AttachChild(parent, child, false)
			nodes.AttachChild(child, grandChild, false)

			nodes.SetHidden(parent, true)

			Expect(nodes.IsAbsoluteHidden(grandChild)).To(BeTrue())
		})

		It("restores a child's visibility when its parent is unhidden", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)

			nodes.SetHidden(parent, true)
			nodes.SetHidden(parent, false)

			Expect(nodes.IsAbsoluteHidden(child)).To(BeFalse())
		})

		It("keeps a child absolutely hidden while it is hidden itself and its parent is unhidden", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)

			nodes.SetHidden(parent, true)
			nodes.SetHidden(child, true)
			nodes.SetHidden(parent, false)

			Expect(nodes.IsAbsoluteHidden(child)).To(BeTrue())
		})

		It("marks a child as absolutely hidden when attached to an already hidden parent", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetHidden(parent, true)

			nodes.AttachChild(parent, child, false)

			Expect(nodes.IsAbsoluteHidden(child)).To(BeTrue())
		})

		It("clears a child's absolute hidden state when detached from a hidden parent", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)
			nodes.SetHidden(parent, true)

			nodes.Detach(child, false)

			Expect(nodes.IsAbsoluteHidden(child)).To(BeFalse())
		})

		It("mirrors the hidden state through the visibility accessors", func() {
			id := nodes.Create()
			Expect(nodes.IsVisible(id)).To(BeTrue())
			Expect(nodes.IsAbsoluteVisible(id)).To(BeTrue())

			nodes.SetVisible(id, false)
			Expect(nodes.IsHidden(id)).To(BeTrue())
			Expect(nodes.IsVisible(id)).To(BeFalse())
			Expect(nodes.IsAbsoluteVisible(id)).To(BeFalse())

			nodes.SetVisible(id, true)
			Expect(nodes.IsHidden(id)).To(BeFalse())
			Expect(nodes.IsVisible(id)).To(BeTrue())
			Expect(nodes.IsAbsoluteVisible(id)).To(BeTrue())
		})

		It("reports a child of a hidden parent as not absolutely visible while still visible itself", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.AttachChild(parent, child, false)

			nodes.SetVisible(parent, false)

			Expect(nodes.IsVisible(child)).To(BeTrue())
			Expect(nodes.IsAbsoluteVisible(child)).To(BeFalse())
		})
	})

	Describe("local transformation", func() {
		var id hierarchy.NodeID

		BeforeEach(func() {
			id = nodes.Create()
		})

		It("has an identity transformation by default", func() {
			Expect(nodes.Position(id)).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(nodes.Rotation(id)).To(dprectest.HaveQuatCoords(1.0, 0.0, 0.0, 0.0))
			Expect(nodes.Scale(id)).To(dprectest.HaveVec3Coords(1.0, 1.0, 1.0))
		})

		It("stores and returns the position", func() {
			nodes.SetPosition(id, dprec.NewVec3(1.0, 2.0, 3.0))
			Expect(nodes.Position(id)).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
		})

		It("stores and returns the scale", func() {
			nodes.SetScale(id, dprec.NewVec3(2.0, 3.0, 4.0))
			Expect(nodes.Scale(id)).To(dprectest.HaveVec3Coords(2.0, 3.0, 4.0))
		})

		It("stores and returns the translation, rotation, and scale together", func() {
			rotation := dprec.RotationQuat(dprec.Degrees(90.0), dprec.NewVec3(0.0, 1.0, 0.0))
			nodes.SetTRS(id, dprec.NewVec3(1.0, 2.0, 3.0), rotation, dprec.NewVec3(4.0, 5.0, 6.0))

			position, storedRotation, scale := nodes.TRS(id)
			Expect(position).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
			Expect(storedRotation).To(dprectest.HaveQuatCoords(rotation.W, rotation.X, rotation.Y, rotation.Z))
			Expect(scale).To(dprectest.HaveVec3Coords(4.0, 5.0, 6.0))
		})

		It("composes the local matrix from the transformation", func() {
			nodes.SetPosition(id, dprec.NewVec3(1.0, 2.0, 3.0))

			expected := dprec.TRSMat4(
				dprec.NewVec3(1.0, 2.0, 3.0),
				dprec.IdentityQuat(),
				dprec.NewVec3(1.0, 1.0, 1.0),
			)
			Expect(nodes.Matrix(id)).To(dprectest.HaveMat4Elements(
				expected.M11, expected.M12, expected.M13, expected.M14,
				expected.M21, expected.M22, expected.M23, expected.M24,
				expected.M31, expected.M32, expected.M33, expected.M34,
				expected.M41, expected.M42, expected.M43, expected.M44,
			))
		})

		It("decomposes a matrix into the local transformation", func() {
			matrix := dprec.TRSMat4(
				dprec.NewVec3(1.0, 2.0, 3.0),
				dprec.IdentityQuat(),
				dprec.NewVec3(2.0, 2.0, 2.0),
			)
			nodes.SetMatrix(id, matrix)

			Expect(nodes.Position(id)).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
			Expect(nodes.Scale(id)).To(dprectest.HaveVec3Coords(2.0, 2.0, 2.0))
		})

		It("stores and returns the rotation", func() {
			rotation := dprec.RotationQuat(dprec.Degrees(90.0), dprec.NewVec3(0.0, 1.0, 0.0))
			nodes.SetRotation(id, rotation)
			Expect(nodes.Rotation(id)).To(dprectest.HaveQuatCoords(rotation.W, rotation.X, rotation.Y, rotation.Z))
		})
	})

	Describe("absolute matrix", func() {
		It("equals the local matrix for a root node", func() {
			id := nodes.Create()
			nodes.SetPosition(id, dprec.NewVec3(1.0, 2.0, 3.0))
			expectMatrix(nodes.AbsoluteMatrix(id), translation(1.0, 2.0, 3.0))
		})

		It("composes the ancestor transformations for a child node", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.SetPosition(child, dprec.NewVec3(1.0, 2.0, 3.0))
			nodes.AttachChild(parent, child, false)

			expectMatrix(nodes.AbsoluteMatrix(child), translation(11.0, 2.0, 3.0))
		})

		It("composes the transformations across multiple levels", func() {
			parent := nodes.Create()
			child := nodes.Create()
			grandChild := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.SetPosition(child, dprec.NewVec3(0.0, 20.0, 0.0))
			nodes.SetPosition(grandChild, dprec.NewVec3(0.0, 0.0, 30.0))
			nodes.AttachChild(parent, child, false)
			nodes.AttachChild(child, grandChild, false)

			expectMatrix(nodes.AbsoluteMatrix(grandChild), translation(10.0, 20.0, 30.0))
		})

		It("updates when an ancestor is moved", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(child, dprec.NewVec3(1.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)

			nodes.SetPosition(parent, dprec.NewVec3(5.0, 0.0, 0.0))

			expectMatrix(nodes.AbsoluteMatrix(child), translation(6.0, 0.0, 0.0))
		})
	})

	Describe("ReferenceMatrix", func() {
		It("is the identity matrix for a root node", func() {
			id := nodes.Create()
			nodes.SetPosition(id, dprec.NewVec3(1.0, 2.0, 3.0))
			expectMatrix(nodes.ReferenceMatrix(id), dprec.IdentityMat4())
		})

		It("is the parent's absolute matrix for a child node", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)

			expectMatrix(nodes.ReferenceMatrix(child), translation(10.0, 0.0, 0.0))
		})
	})

	Describe("SetAbsoluteMatrix", func() {
		It("sets the absolute matrix of a root node", func() {
			id := nodes.Create()
			nodes.SetAbsoluteMatrix(id, translation(1.0, 2.0, 3.0))

			expectMatrix(nodes.AbsoluteMatrix(id), translation(1.0, 2.0, 3.0))
			Expect(nodes.Position(id)).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
		})

		It("derives the local transform relative to the parent", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)

			nodes.SetAbsoluteMatrix(child, translation(5.0, 0.0, 0.0))

			expectMatrix(nodes.AbsoluteMatrix(child), translation(5.0, 0.0, 0.0))
			Expect(nodes.Position(child)).To(dprectest.HaveVec3Coords(-5.0, 0.0, 0.0))
		})

		It("keeps the node with its parent when the parent later moves", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)
			nodes.SetAbsoluteMatrix(child, translation(5.0, 0.0, 0.0))

			nodes.SetPosition(parent, dprec.NewVec3(20.0, 0.0, 0.0))

			expectMatrix(nodes.AbsoluteMatrix(child), translation(15.0, 0.0, 0.0))
		})

		It("updates the absolute matrix of descendants", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(child, dprec.NewVec3(1.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)

			nodes.SetAbsoluteMatrix(parent, translation(10.0, 0.0, 0.0))

			expectMatrix(nodes.AbsoluteMatrix(child), translation(11.0, 0.0, 0.0))
		})
	})

	Describe("InterpolatedAbsoluteMatrix", func() {
		It("returns the current absolute matrix at fraction one", func() {
			id := nodes.Create()
			nodes.SetPosition(id, dprec.NewVec3(10.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 1.0), translation(10.0, 0.0, 0.0))
		})

		It("returns the previous absolute matrix at fraction zero", func() {
			id := nodes.Create()
			nodes.SetPosition(id, dprec.NewVec3(10.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 0.0), dprec.IdentityMat4())
		})

		It("interpolates halfway between the previous and current matrix", func() {
			id := nodes.Create()
			nodes.SetPosition(id, dprec.NewVec3(10.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 0.5), translation(5.0, 0.0, 0.0))
		})

		It("returns the absolute matrix when it has not changed", func() {
			id := nodes.Create()
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 0.5), dprec.IdentityMat4())
		})
	})

	Describe("preserving the world transform", func() {
		It("keeps the world transform when detaching from a moved parent", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.SetPosition(child, dprec.NewVec3(1.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)
			// world position is now (11, 0, 0)

			nodes.Detach(child, true)

			expectMatrix(nodes.AbsoluteMatrix(child), translation(11.0, 0.0, 0.0))
		})

		It("keeps the world transform when attaching to a moved parent", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.SetPosition(child, dprec.NewVec3(1.0, 0.0, 0.0))

			nodes.AttachChild(parent, child, true)

			// The child was at world (1,0,0); relative to a parent at (10,0,0)
			// its local position must become (-9,0,0) to preserve that.
			Expect(nodes.Position(child)).To(dprectest.HaveVec3Coords(-9.0, 0.0, 0.0))
		})
	})

	Describe("NodeHandle absolute matrix", func() {
		It("returns the absolute matrix", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			parent.SetPosition(dprec.NewVec3(10.0, 0.0, 0.0))
			child.SetPosition(dprec.NewVec3(1.0, 0.0, 0.0))
			nodes.AttachChild(parent.ID(), child.ID(), false)

			expectMatrix(child.AbsoluteMatrix(), translation(11.0, 0.0, 0.0))
		})

		It("returns the reference matrix", func() {
			parent := nodes.CreateHandle()
			child := nodes.CreateHandle()
			parent.SetPosition(dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.AttachChild(parent.ID(), child.ID(), false)

			expectMatrix(child.ReferenceMatrix(), translation(10.0, 0.0, 0.0))
		})

		It("sets the absolute matrix", func() {
			handle := nodes.CreateHandle()
			handle.SetAbsoluteMatrix(translation(1.0, 2.0, 3.0))
			expectMatrix(handle.AbsoluteMatrix(), translation(1.0, 2.0, 3.0))
		})

		It("returns the interpolated absolute matrix", func() {
			handle := nodes.CreateHandle()
			handle.SetPosition(dprec.NewVec3(10.0, 0.0, 0.0))
			expectMatrix(handle.InterpolatedAbsoluteMatrix(0.5), translation(5.0, 0.0, 0.0))
		})
	})

	Describe("independence", func() {
		var parent, child hierarchy.NodeID

		BeforeEach(func() {
			parent = nodes.Create()
			child = nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.SetPosition(child, dprec.NewVec3(1.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)
		})

		It("is not independent by default", func() {
			Expect(nodes.IsIndependent(child)).To(BeFalse())
		})

		It("marks the node as independent", func() {
			nodes.SetIndependent(child, true, false)
			Expect(nodes.IsIndependent(child)).To(BeTrue())
		})

		It("ignores the parent transform for the absolute matrix when independent", func() {
			nodes.SetIndependent(child, true, false)
			expectMatrix(nodes.AbsoluteMatrix(child), translation(1.0, 0.0, 0.0))
			expectMatrix(nodes.ReferenceMatrix(child), dprec.IdentityMat4())
		})

		It("follows the parent transform again when dependence is restored", func() {
			nodes.SetIndependent(child, true, false)
			nodes.SetIndependent(child, false, false)
			expectMatrix(nodes.AbsoluteMatrix(child), translation(11.0, 0.0, 0.0))
			expectMatrix(nodes.ReferenceMatrix(child), translation(10.0, 0.0, 0.0))
		})

		It("preserves the world transform when becoming independent", func() {
			nodes.SetIndependent(child, true, true)
			expectMatrix(nodes.AbsoluteMatrix(child), translation(11.0, 0.0, 0.0))
			Expect(nodes.Position(child)).To(dprectest.HaveVec3Coords(11.0, 0.0, 0.0))
		})

		It("preserves the world transform when dependence is restored", func() {
			nodes.SetIndependent(child, true, true)
			nodes.SetIndependent(child, false, true)
			expectMatrix(nodes.AbsoluteMatrix(child), translation(11.0, 0.0, 0.0))
			Expect(nodes.Position(child)).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
		})

		It("does not move an independent node when its parent moves", func() {
			nodes.SetIndependent(child, true, false)
			nodes.SetPosition(parent, dprec.NewVec3(100.0, 0.0, 0.0))
			expectMatrix(nodes.AbsoluteMatrix(child), translation(1.0, 0.0, 0.0))
		})

		It("keeps the node in the hierarchy while independent", func() {
			nodes.SetIndependent(child, true, false)
			Expect(nodes.IsRoot(child)).To(BeFalse())
			Expect(nodes.Parent(child)).To(Equal(parent))
			Expect(collectChildren(parent)).To(ConsistOf(child))
		})

		It("still inherits the hidden state while independent", func() {
			nodes.SetIndependent(child, true, false)
			nodes.SetHidden(parent, true)
			Expect(nodes.IsHidden(child)).To(BeFalse())
			Expect(nodes.IsAbsoluteHidden(child)).To(BeTrue())
		})

		It("is a no-op when the independence state is unchanged", func() {
			nodes.SetIndependent(child, false, false)
			Expect(nodes.IsIndependent(child)).To(BeFalse())
			expectMatrix(nodes.AbsoluteMatrix(child), translation(11.0, 0.0, 0.0))
		})

		It("exposes independence through a NodeHandle", func() {
			handle := nodes.Handle(child)
			Expect(handle.IsIndependent()).To(BeFalse())

			handle.SetIndependent(true, false)
			Expect(handle.IsIndependent()).To(BeTrue())
			Expect(nodes.IsIndependent(child)).To(BeTrue())
		})
	})

	Describe("snapping", func() {
		It("captures the current absolute matrix as the interpolation origin", func() {
			id := nodes.Create()
			nodes.SetPosition(id, dprec.NewVec3(10.0, 0.0, 0.0))

			nodes.Snap(id)

			// After snapping, previous == current, so interpolation is stable.
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 0.0), translation(10.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 1.0), translation(10.0, 0.0, 0.0))
		})

		It("interpolates from the snapped pose towards the new pose", func() {
			id := nodes.Create()
			nodes.SetPosition(id, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.Snap(id)

			nodes.SetPosition(id, dprec.NewVec3(20.0, 0.0, 0.0))

			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 0.0), translation(10.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 0.5), translation(15.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(id, 1.0), translation(20.0, 0.0, 0.0))
		})

		It("snaps descendants as well", func() {
			parent := nodes.Create()
			child := nodes.Create()
			nodes.SetPosition(parent, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.SetPosition(child, dprec.NewVec3(1.0, 0.0, 0.0))
			nodes.AttachChild(parent, child, false)
			// child world is now (11, 0, 0)

			nodes.Snap(parent)
			nodes.SetPosition(parent, dprec.NewVec3(20.0, 0.0, 0.0))
			// child world is now (21, 0, 0)

			expectMatrix(nodes.InterpolatedAbsoluteMatrix(child, 0.0), translation(11.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(child, 1.0), translation(21.0, 0.0, 0.0))
		})

		It("snaps the whole scene via AdvanceStep", func() {
			first := nodes.Create()
			second := nodes.Create()
			nodes.SetPosition(first, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.SetPosition(second, dprec.NewVec3(30.0, 0.0, 0.0))

			scene.AdvanceStep()

			nodes.SetPosition(first, dprec.NewVec3(20.0, 0.0, 0.0))
			nodes.SetPosition(second, dprec.NewVec3(50.0, 0.0, 0.0))

			expectMatrix(nodes.InterpolatedAbsoluteMatrix(first, 0.5), translation(15.0, 0.0, 0.0))
			expectMatrix(nodes.InterpolatedAbsoluteMatrix(second, 0.5), translation(40.0, 0.0, 0.0))
		})

		It("snaps the live nodes when the scene contains deleted ones", func() {
			stale := nodes.Create()
			live := nodes.Create()
			nodes.SetPosition(live, dprec.NewVec3(10.0, 0.0, 0.0))
			nodes.Delete(stale) // leaves a freed slot for AdvanceStep to skip

			scene.AdvanceStep()

			nodes.SetPosition(live, dprec.NewVec3(20.0, 0.0, 0.0))

			expectMatrix(nodes.InterpolatedAbsoluteMatrix(live, 0.5), translation(15.0, 0.0, 0.0))
		})

		It("snaps through a NodeHandle", func() {
			handle := nodes.CreateHandle()
			handle.SetPosition(dprec.NewVec3(10.0, 0.0, 0.0))
			handle.Snap()

			handle.SetPosition(dprec.NewVec3(20.0, 0.0, 0.0))

			expectMatrix(handle.InterpolatedAbsoluteMatrix(0.0), translation(10.0, 0.0, 0.0))
			expectMatrix(handle.InterpolatedAbsoluteMatrix(1.0), translation(20.0, 0.0, 0.0))
		})
	})
})
