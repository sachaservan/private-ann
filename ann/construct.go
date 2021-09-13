package ann

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/gonum/stat/distuv"
	"github.com/sachaservan/private-ann/hash"
	"github.com/sachaservan/private-ann/pir/field"
	"github.com/sachaservan/vec"
)

/*
With small numbers of tables, choosing radii by sampling from the distribution can lead to large variance in results
We use the quantiles instead to pick the points for consistency

However, the 0 and 100th quantiles are +- infinity for a normal distribution
The first function shrinks the quantiles towards the center
The second function shifts the quantiles by 1/2

Both seem to perform the same for large number of tables (as expected), but for small numbers of tables the second is better
*/

func GetNormalSequence(mean, stddev float64, numTables int) []float64 {
	n := distuv.Normal{Mu: mean, Sigma: stddev}
	radii := make([]float64, 0)
	for i := 0; i < numTables; i++ {
		q := n.Quantile(float64(i+1) / float64(numTables+1))
		// can't have negatives
		for q <= 0 {
			// replace with random normal value
			q = n.Rand()
		}
		radii = append(radii, q)
	}
	sort.Float64s(radii)
	return radii
}

func GetNormalSequence2(mean, stddev float64, numTables int) []float64 {
	n := distuv.Normal{Mu: mean, Sigma: stddev}
	radii := make([]float64, 0)
	for i := 0; i < numTables; i++ {
		q := n.Quantile((float64(i) + 0.5) / float64(numTables))
		// can't have negatives
		for q <= 0 {
			// replace with random normal value
			q = n.Rand()
		}
		radii = append(radii, q)
	}
	sort.Float64s(radii)
	return radii
}

func GetGeometricSequence(min, max float64, numTables int) []float64 {
	// Let R_0 = closest, R_i = R_0 * c^i where c is chosen such that R_{numTables - 1} = farthest
	// c = farthest/closest ^(1/(numtables-1))
	r := max / min
	radii := make([]float64, 0)
	for i := 0; i < numTables; i++ {
		radii = append(radii, min*math.Pow(r, float64(i)/float64(numTables-1)))
	}
	return radii
}

// choose from a normal distribution such that min and max match the specified values
func GetNormalSequence3(min, max float64, numTables int) ([]float64, float64, float64) {
	n := distuv.UnitNormal
	standardQuantiles := make([]float64, numTables)
	for i := range standardQuantiles {
		standardQuantiles[i] = n.Quantile(float64(i+1) / float64(numTables+1))
	}
	if numTables == 1 {
		avg := (min + max) / 2
		return []float64{avg}, avg, 0
	}
	// linear transform standardQuantiles[0] to min and standardQuantiles[numTables-1] to max
	a := (max - min) / (standardQuantiles[numTables-1] - standardQuantiles[0])
	b := min - a*standardQuantiles[0] // should just be the average of min and max

	n2 := distuv.Normal{Mu: b, Sigma: a}
	for i := range standardQuantiles {
		standardQuantiles[i] = a*standardQuantiles[i] + b
		if standardQuantiles[i] <= 0 {
			standardQuantiles[i] = n2.Rand()
		}
	}
	sort.Float64s(standardQuantiles)
	return standardQuantiles, b, a
}

type HashTable struct {
	table  int
	hashes map[uint64][]uint32
	mu     sync.Mutex
}

func NewHashTable(table int) *HashTable {
	return &HashTable{table: table, hashes: make(map[uint64][]uint32)}
}

func ComputeHashes(n int, h hash.Hash, data []*vec.Vec) ([]uint64, []field.FP) {
	table := NewHashTable(n)
	table.AddAll(h, data)
	return convertAndCap(table.hashes)
}

func (t *HashTable) AddAll(h hash.Hash, data []*vec.Vec) {
	numThreads := runtime.GOMAXPROCS(0)
	sections := hash.Spans(len(data), numThreads)
	errs := make(chan error)
	for i := 0; i < numThreads; i++ {
		go func(i int) {
			myHashes := make(map[uint64][]uint32)
			for row := sections[i][0]; row < sections[i][1]; row++ {
				v := data[row]
				hash := h.Hash(v)
				cur := myHashes[hash]
				myHashes[hash] = append(cur, uint32(row))
				// to give some sense of progress
				if (row & 16383) == 10000 {
					log.Printf("[Server]: table %d, completed row %v of %v\n", t.table, row-sections[i][0], sections[i][1]-sections[i][0])
				}
			}
			t.mu.Lock()
			t.Merge(myHashes)
			t.mu.Unlock()
			errs <- nil
		}(i)
	}
	for i := 0; i < numThreads; i++ {
		<-errs
	}
}

func (t *HashTable) Merge(other map[uint64][]uint32) {
	for k, v := range other {
		cur := t.hashes[k]
		if len(cur) > 0 {
			t.hashes[k] = append(t.hashes[k], v...)
		} else {
			t.hashes[k] = v
		}
	}
}

func (t *HashTable) Len() int {
	return len(t.hashes)
}

// Choose one element to keep from each bucket with multiple values
func convertAndCap(hashTable map[uint64][]uint32) ([]uint64, []field.FP) {
	keys := make([]uint64, 0)
	values := make([]field.FP, 0)
	for k, v := range hashTable {
		r := 0
		if len(v) > 1 {
			r = rand.Intn(len(v))
		}
		keys = append(keys, k)
		values = append(values, field.FP(v[r]))
	}
	return keys, values
}

func (t *HashTable) Get(h uint64) []uint32 {
	return t.hashes[h]
}
