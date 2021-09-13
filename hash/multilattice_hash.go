package hash

import (
	"math/rand"

	"github.com/sachaservan/vec"
)

/*

 Construct a higher dimensional lattice as a direct product of copies of the leech lattice
 For example, with 2 copies we have:

 1) view a 48-dimensional vector as the concatenation of 2 24-dimensional vectors
 2) Find the closest leech lattice point to each 24-dimensional vector, and then concatenate the result.

 The error increases, but only by a factor of sqrt(2).
 This reduces the error from dimensionality reduction, at the cost of error in the lattice
 Adding more tables/multiprobes can reduce lattice error, but not dimensionality reduction error
 So this is a valuable trade off

 Permuting coordinates does not change distances and this can make sure the coordinates are chosen at random
*/

type MultiLatticeHash struct {
	Hashes      []*LatticeHash
	Permutation []int
	Spans       [][2]int
	UHash       *UniversalHash
}

func NewMultiLatticeHash(dim, copies int, width, max float64) *MultiLatticeHash {
	m := &MultiLatticeHash{}
	m.Permutation = rand.Perm(dim)
	m.Hashes = make([]*LatticeHash, copies)
	m.Spans = Spans(dim, copies)
	for i := 0; i < copies; i++ {
		m.Hashes[i] = NewLatticeHash(m.Spans[i][1]-m.Spans[i][0], width, max)
	}
	m.UHash = NewUniversalHash(copies * 24)
	return m
}

// This function divides the input dimension into copies parts as evenly as possible
func Spans(total int, numSpans int) [][2]int {
	Spans := make([][2]int, numSpans)
	start := 0
	skip := total / numSpans
	extra := total % numSpans
	for i := range Spans {
		end := start + skip
		if extra > 0 {
			end++
			extra--
		}
		Spans[i] = [2]int{start, end}
		start = end
	}
	return Spans
}

func (m *MultiLatticeHash) Hash(v *vec.Vec) uint64 {
	h, _ := m.HashWithDist(v)
	return m.UHash.Hash(h)
}

func (m *MultiLatticeHash) HashWithDist(v *vec.Vec) (*vec.Vec, float64) {
	permuted := make([]float64, v.Size())
	for i := range permuted {
		permuted[i] = v.Coord(m.Permutation[i])
	}
	totalDist := 0.0
	totalHash := make([]float64, 0)
	for i := range m.Hashes {
		hash, dist := m.Hashes[i].HashWithDist(vec.NewVec(permuted[m.Spans[i][0]:m.Spans[i][1]]))
		totalHash = append(totalHash, hash.Coords...)
		totalDist += dist
	}
	return vec.NewVec(totalHash), totalDist
}

// We have to iterate through each of the closest points of the sublattices to find the closest point
func (m *MultiLatticeHash) MultiProbeHashWithDist(v *vec.Vec, probes int) ([]*vec.Vec, []float64) {
	permuted := make([]float64, v.Size())
	for i := range permuted {
		permuted[i] = v.Coord(m.Permutation[i])
	}
	Hashes := make([][]*vec.Vec, len(m.Hashes))
	sources := make([][]*Element, len(m.Hashes))
	for i := range m.Hashes {
		var distances []float64
		Hashes[i], distances = m.Hashes[i].MultiProbeHashWithDist(vec.NewVec(permuted[m.Spans[i][0]:m.Spans[i][1]]), probes)
		sources[i] = make([]*Element, len(distances))
		for j, d := range distances {
			sources[i][j] = &Element{coords: []int{j}, distance: d}
		}
	}
	d := NewDistanceSearchQueue(probes, sources)
	c := d.Search()
	output := make([]*vec.Vec, probes)
	distances := make([]float64, probes)
	for k, e := range c {
		hash := make([]float64, 0)
		for j := range e.coords {
			hash = append(hash, Hashes[j][e.coords[j]].Coords...)
		}
		distances[k] = e.distance
		output[k] = vec.NewVec(hash)
	}
	return output, distances
}

func (m *MultiLatticeHash) MultiHash(v *vec.Vec, probes int) []uint64 {
	vs, _ := m.MultiProbeHashWithDist(v, probes)
	Hashes := make([]uint64, probes)
	for i := range vs {
		Hashes[i] = m.UHash.Hash(vs[i])
	}
	return Hashes
}
