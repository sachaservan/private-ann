package hash

import "github.com/sachaservan/vec"

type HashCommon struct {
	ProjectionLines []*vec.Vec
	Offsets         *vec.Vec
	Orthogonal      bool
	UHash           *UniversalHash
}

/*
Rotation and translation and dimension scaling that are common to many LSH hash functions
Dimensionality reduction is a random rotation followed by a projection (e.g taking the first k coordinates).
*/
func NewHashCommon(dim, amplification int, max float64, Orthogonal bool) *HashCommon {
	h := &HashCommon{Orthogonal: Orthogonal}
	h.ProjectionLines = make([]*vec.Vec, amplification)

	if Orthogonal {
		// if amplification is > dim, then there is no possible Orthogonal configuration and this panics
		// Perhaps the rotation matrix is max(dim, amplification)
		copy(h.ProjectionLines[:], RandomRotationMatrix(dim)[:amplification])
	} else {
		for i := range h.ProjectionLines {
			h.ProjectionLines[i] = RandomVector(dim)
		}
	}
	// Normalize vectors
	for i := range h.ProjectionLines {
		h.ProjectionLines[i] = h.ProjectionLines[i].Normalize()
	}
	h.Offsets = RandomTranslationVector(amplification, max)
	h.UHash = NewUniversalHash(amplification)
	return h
}

// Apply the rotation and translation
func (h *HashCommon) Project(v *vec.Vec) *vec.Vec {
	projection := vec.NewVec(make([]float64, len(h.ProjectionLines)))
	for i := range projection.Coords {
		projection.Coords[i], _ = v.Dot(h.ProjectionLines[i])
	}
	projection, _ = projection.Add(h.Offsets)
	return projection
}

// helper to sort points because golang
// A heap or priority queue would be more efficient
type Candidates struct {
	Indexes   []uint64
	Distances []float64
}

func (c *Candidates) Len() int {
	return len(c.Indexes)
}
func (c *Candidates) Less(i, j int) bool {
	return c.Distances[i] < c.Distances[j]
}
func (c *Candidates) Swap(i, j int) {
	t := c.Indexes[i]
	c.Indexes[i] = c.Indexes[j]
	c.Indexes[j] = t
	d := c.Distances[i]
	c.Distances[i] = c.Distances[j]
	c.Distances[j] = d
}
