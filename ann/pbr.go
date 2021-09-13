package ann

import (
	"sort"

	"github.com/sachaservan/private-ann/hash"
	"github.com/sachaservan/private-ann/pir/field"
	"github.com/sachaservan/vec"
)

type PBRBuckets struct {
	buckets    [][2]uint64
	size       uint64
	Max        uint64
	NumBuckets int
}

func NewPBRBuckets(max uint64, numBuckets uint64) *PBRBuckets {
	buckets := make([][2]uint64, numBuckets)
	start := uint64(0)
	skip := max / numBuckets
	extra := max % numBuckets
	for i := range buckets {
		end := start + skip
		if extra > 0 {
			end++
			extra--
		}
		buckets[i] = [2]uint64{start, end}
		start = end
	}
	return &PBRBuckets{
		buckets:    buckets,
		size:       skip,
		Max:        max,
		NumBuckets: int(numBuckets),
	}
}

func (p *PBRBuckets) FindBucket(hash uint64) uint64 {
	guess := hash / p.size
	if guess >= uint64(len(p.buckets)) || p.buckets[guess][0] > hash {
		guess--
	} else if p.buckets[guess][1] <= hash {
		guess++
	}
	return guess
}

type sorter struct {
	keys   []uint64
	values []field.FP
}

func ComputeBucketDivisions(numBuckets int, keys []uint64, values []field.FP) ([]int, []int) {
	// first sort data
	s := sorter{keys, values}
	sort.Sort(&s)
	p := NewPBRBuckets(hash.Prime, uint64(numBuckets))
	starts := make([]int, numBuckets)
	stops := make([]int, numBuckets)
	// technically we could use binary search but a linear scan suffices
	bucket := 0
	for i := 0; i < len(keys); i++ {
		if keys[i] >= p.buckets[bucket][1] {
			stops[bucket] = i
			starts[bucket+1] = i
			bucket++
		}
	}
	stops[numBuckets-1] = len(keys)
	return starts, stops
}

func (s *sorter) Len() int {
	return len(s.keys)
}

func (s *sorter) Less(i, j int) bool {
	return s.keys[i] < s.keys[j]
}

func (s *sorter) Swap(i, j int) {
	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
	s.values[i], s.values[j] = s.values[j], s.values[i]
}

func ComputeProbes(hashFunction hash.Hash, query *vec.Vec, numPartitions, numProbes int) []uint64 {
	output := make([]uint64, numPartitions)
	hashes := hashFunction.MultiHash(query, numProbes)
	buckets := NewPBRBuckets(hash.Prime, uint64(numPartitions))
	used := make([]bool, numPartitions)
	// hashes (should be) in optimal order so first come first serve
	for _, h := range hashes {
		bucket := buckets.FindBucket(h)
		if !used[bucket] {
			used[bucket] = true
			output[bucket] = h
		}
	}
	return output
}
