package ann

import (
	"sort"

	"github.com/sachaservan/private-ann/hash"
	"github.com/sachaservan/private-ann/pir/field"
	"github.com/sachaservan/vec"
)

type PBRBuckets struct {
	Buckets    [][2]uint64
	Size       uint64
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
		Buckets:    buckets,
		Size:       skip,
		Max:        max,
		NumBuckets: int(numBuckets),
	}
}

func (p *PBRBuckets) FindBucket(hash uint64) uint64 {
	guess := hash / p.Size
	if guess >= uint64(len(p.Buckets)) || p.Buckets[guess][0] > hash {
		guess--
	} else if p.Buckets[guess][1] <= hash {
		guess++
	}
	return guess
}

type sorter struct {
	keys   []uint64
	values []field.FP
}

func ComputeBucketDivisions(numBuckets int, keys []uint64, values []field.FP, hashKeyBits int) ([]int, []int) {
	// first sort data
	s := sorter{keys, values}
	sort.Sort(&s)

	mod := uint64(1) << hashKeyBits
	p := NewPBRBuckets(mod, uint64(numBuckets))
	starts := make([]int, numBuckets)
	stops := make([]int, numBuckets)
	// technically we could use binary search but a linear scan suffices
	bucket := 0
	for i := 0; i < len(keys); i++ {
		if keys[i] >= p.Buckets[bucket][1] {
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
