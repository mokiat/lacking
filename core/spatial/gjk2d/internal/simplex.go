package internal

type Simplex struct {
	Vertices    [3]MinkowskiVertex
	VertexCount uint32
}

func EmptySimplex() Simplex {
	return Simplex{}
}

func PointSimplex(point MinkowskiVertex) Simplex {
	return Simplex{
		Vertices:    [3]MinkowskiVertex{point},
		VertexCount: 1,
	}
}

func EdgeSimplex(a, b MinkowskiVertex) Simplex {
	return Simplex{
		Vertices:    [3]MinkowskiVertex{a, b},
		VertexCount: 2,
	}
}

func TriangleSimplex(a, b, c MinkowskiVertex) Simplex {
	return Simplex{
		Vertices:    [3]MinkowskiVertex{a, b, c},
		VertexCount: 3,
	}
}

func (s *Simplex) HasVertex(vertex MinkowskiVertex) bool {
	for i := range s.VertexCount {
		if vertex.Refs == s.Vertices[i].Refs {
			return true
		}
	}
	return false
}
