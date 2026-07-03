package internal

// Simplex is a simplex (point, edge, triangle or tetrahedron) made of
// vertices of the Minkowski difference.
type Simplex struct {
	// Vertices holds the simplex vertices. Only the first VertexCount
	// entries are meaningful.
	Vertices [4]MinkowskiVertex
	// VertexCount indicates how many of the entries in Vertices are in use.
	VertexCount uint32
}

// EmptySimplex returns a [Simplex] with no vertices.
func EmptySimplex() Simplex {
	return Simplex{}
}

// PointSimplex returns a [Simplex] made of the single specified vertex.
func PointSimplex(point MinkowskiVertex) Simplex {
	return Simplex{
		Vertices:    [4]MinkowskiVertex{point},
		VertexCount: 1,
	}
}

// EdgeSimplex returns a [Simplex] made of the two specified vertices.
func EdgeSimplex(a, b MinkowskiVertex) Simplex {
	return Simplex{
		Vertices:    [4]MinkowskiVertex{a, b},
		VertexCount: 2,
	}
}

// TriangleSimplex returns a [Simplex] made of the three specified vertices.
func TriangleSimplex(a, b, c MinkowskiVertex) Simplex {
	return Simplex{
		Vertices:    [4]MinkowskiVertex{a, b, c},
		VertexCount: 3,
	}
}

// TetrahedronSimplex returns a [Simplex] made of the four specified vertices.
func TetrahedronSimplex(a, b, c, d MinkowskiVertex) Simplex {
	return Simplex{
		Vertices:    [4]MinkowskiVertex{a, b, c, d},
		VertexCount: 4,
	}
}

// HasVertex reports whether the simplex already contains the given vertex.
// Vertices are compared through their refs, which is exact, unlike a
// floating-point position comparison.
func (s *Simplex) HasVertex(vertex MinkowskiVertex) bool {
	for i := range s.VertexCount {
		if vertex.Refs == s.Vertices[i].Refs {
			return true
		}
	}
	return false
}
