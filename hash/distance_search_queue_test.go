package hash

import (
	"math"
	"math/rand"
	"sort"
	"testing"
)

func TestQueue(t *testing.T) {
	n := 100
	numSources := 3
	sources := make([][]*Element, numSources)
	for j := range sources {
		d := &DistanceSearchQueue{}
		d.queue = make([]*Element, n)
		for i := 0; i < n; i++ {
			d.queue[i] = &Element{distance: math.Abs(rand.NormFloat64())}
		}
		sort.Sort(d)
		for i := 0; i < n; i++ {
			d.queue[i].coords = []int{i}
		}
		sources[j] = d.queue
	}

	d := NewDistanceSearchQueue(n, sources)
	res := d.Search()
	for i := 0; i < n-1; i++ {
		if res[i].distance > res[i+1].distance {
			t.Fail()
		}
	}
	ans := GetBruteForceAnswer(sources)
	for i := 0; i < n; i++ {
		if ans[i].distance != res[i].distance {
			t.Fail()
		}
	}
}

func GetBruteForceAnswer(sources [][]*Element) []*Element {
	d := &DistanceSearchQueue{}
	e := make([]int, 0)
	d.queue = ListAllCandidates(e, 0, sources, 0)
	sort.Sort(d)
	return d.queue
}

// A recursive method to list the exponential amount of candidate so the result can be checked using brute force
func ListAllCandidates(positions []int, distance float64, sources [][]*Element, i int) []*Element {
	candidates := make([]*Element, 0)
	nextCoords := make([]int, len(positions)+1)
	copy(nextCoords[:], positions[:])
	for j, n := range sources[i] {
		nextCoords[i] = j
		if i == len(sources)-1 {
			c := make([]int, len(nextCoords))
			copy(c[:], nextCoords[:])
			candidates = append(candidates, &Element{
				coords:   c,
				distance: distance + n.distance,
			})
		} else {
			candidates = append(candidates, ListAllCandidates(nextCoords, distance+n.distance, sources, i+1)...)
		}
	}
	return candidates
}
