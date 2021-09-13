package hash

import (
	"math"
	"sort"
	"sync"

	"github.com/sachaservan/vec"
)

/*
This implements a hash function by finding the closest point of the leech lattice
See Appendix B of http://web.mit.edu/andoni/www/papers/cSquared.pdf
The lattice provides the densest sphere packing of 24 dimensional space.
Sloane proved that the lattice points are at most sqrt(2) away from any point
Which bounds the error accordingly
*/

var once = sync.Once{}

type LatticeHash struct {
	H     *HashCommon
	Scale float64
}

/*
This constructs a new leech lattice hash where the lattice points are the centers of spheres of radius sqrt(8)
A random rotation and translation are applied and the space is scaled for the desired "R" - LSH hash width
The JL-transform step is performed in the same matrix as rotation
*/
func NewLatticeHash(dim int, width, max float64) *LatticeHash {
	// alternatively this could be read from a file
	once.Do(Precompute)

	// lattice is scaled by sqrt(8)
	baseScale := 2 * math.Sqrt2

	// HashCommon projection with an orthogonal matrix is implicitly a JL transform
	// However, it needs to be normalized by column rather than row
	jlScale := math.Sqrt(float64(dim) / 24.0)

	// width scales the space down to fit within the lattice
	H := &LatticeHash{H: NewHashCommon(dim, 24, max, true), Scale: baseScale * jlScale / width}
	return H
}

// This computes the hash and the squared distance to the closest vector
func (l *LatticeHash) HashWithDist(v *vec.Vec) (*vec.Vec, float64) {
	// apply rotation and translation
	v = l.H.Project(v)
	// apply scaling
	v = v.Scale(l.Scale)
	// this always returns an integer coordinate of the leech lattice
	v, dist := LeechLatticeClosestVector(v)
	// lattice points are always the same set of keys
	// To distinguish between hashes add a random value to the vector
	// We could also unrotate to return the true closest point
	v, _ = v.Add(l.H.Offsets)
	return v, dist
}

// This returns the k closest hashes and squared distances
func (l *LatticeHash) MultiProbeHashWithDist(v *vec.Vec, probes int) ([]*vec.Vec, []float64) {
	// apply rotation and translation
	v = l.H.Project(v)
	// apply scaling
	v = v.Scale(l.Scale)
	vs, dists := LeechLatticeClosestVectors(v, probes)
	for i := range vs {
		vs[i], _ = vs[i].Add(l.H.Offsets)

	}
	return vs, dists
}

func (l *LatticeHash) Hash(v *vec.Vec) uint64 {
	H, _ := l.HashWithDist(v)
	return l.H.UHash.Hash(H)
}

func (l *LatticeHash) MultiHash(v *vec.Vec, probes int) []uint64 {
	H, _ := l.MultiProbeHashWithDist(v, probes)
	hashes := make([]uint64, probes)
	for i := range H {
		hashes[i] = l.H.UHash.Hash(H[i])
	}
	return hashes
}

func LeechLatticeClosestVector(v *vec.Vec) (*vec.Vec, float64) {
	p, dist := LeechLatticeClosestPoint(v.Coords)
	for i := range p {
		// ensure no floating point shenanigans
		p[i] = math.Round(p[i])
	}
	return vec.NewVec(p), dist
}

func LeechLatticeClosestVectors(v *vec.Vec, numPoints int) ([]*vec.Vec, []float64) {
	points, distances := LeechLatticeClosestPoints(v.Coords, numPoints)
	out := make([]*vec.Vec, numPoints)
	dists := make([]float64, numPoints)
	for i := range out {
		for j := range points[i] {
			points[i][j] = math.Round(points[i][j])
		}
		out[i] = vec.NewVec(points[i])
		dists[i] = distances[i]
	}
	return out, dists
}

/*
This algorithm is based on https://ieeexplore.ieee.org/stamp/stamp.jsp?tp=&arnumber=1057135
According to https://ieeexplore.ieee.org/stamp/stamp.jsp?tp=&arnumber=243466 this takes 56000 flops
While the more optimal algorithm takes only 3595, so this could be improved significantly
But it is much more understandable.

We build the Leech lattice out of 3 copies of the E_8 lattice, which are built out of 2 copies of the D_8 lattice.
*/

/*
The D_8 lattice consists of integer points in 8 dimensional space where the sum of the coordinates is even
The simplest algorithm for decoding (finding the closest point) is given in neilsloane.com/doc/Me83.pdf
First we find the closest integer point and check if the sum of the coordinates is even.
If it is, then we are done.
Otherwise, we find the second closest point, and note that this can be found by taking the coordinate that
was rounded the most, and then rounding it the other direction.  This changes the sum by 1 and makes it even.
*/

func D8Decode(f []float64) [8]float64 {
	v := [8]float64{}
	sum := 0
	farthestDist := -1.0
	farthestPos := 0
	otherDirection := 0.0
	for i := range f {
		v[i] = math.Round(f[i])
		sum += int(v[i])
		diff := f[i] - v[i]
		dist := math.Abs(diff)
		if dist > farthestDist {
			farthestDist = dist
			farthestPos = i
			if diff > 0 {
				// we rounded down, so the other direction is up
				otherDirection = v[i] + 1
			} else {
				otherDirection = v[i] - 1
			}
		}
	}
	if sum%2 == 0 {
		return v
	}
	v[farthestPos] = otherDirection
	return v
}

/*
The E_8 lattice consists of two copies of D_8, offset by the vector (1/2,1/2,1/2,1/2,1/2,1/2,1/2,1/2)
See https://en.wikipedia.org/wiki/E8_lattice#Lattice_points for details
We follow the algorithm in neilsloane.com/doc/Me83.pdf
We first find the closest point in each of the two copies of D_8
Then we return the closer of the two
*/
func E8Decode(f []float64) ([8]float64, float64) {
	y0 := D8Decode(f)
	t := [8]float64{}
	for k := range f {
		t[k] = f[k] - 0.5
	}
	y1 := D8Decode(t[:])
	for k := range y1 {
		y1[k] += 0.5
	}
	d0 := DistSquared(f, y0[:])
	d1 := DistSquared(f, y1[:])
	if d0 < d1 {
		return y0, d0
	} else {
		return y1, d1
	}
}

/*
Now we construct the Leech lattice based on the Turyn code as in https://ieeexplore.ieee.org/stamp/stamp.jsp?tp=&arnumber=1057135
There are 4096 possible ways for the three copies of E_8 to be arranged, and then these are all stuck together
We simply iterate over each possible construction.
The E_8 lattices are reused, so we first find the closest point in each possible E_8 lattice in the following function
*/

/*
Out lattice is scaled by sqrt8, so we are actually using lambda_8 = E8 * 4 - where all coordinates are multiplied by 4
Thus all lattice points are integers instead of half integers (in fact they are only even integers)
This function iterates through the possible arrangements as given in table 6 and stores the closest points to each in p
It also stores the (squared) distances to those points in d
*/
func LeechLatticeClosest(f []float64) ([256][3][8]float64, [256][3]float64) {
	p := [256][3][8]float64{}
	d := [256][3]float64{}
	for j := range TableVi {
		t := [24]float64{}
		pj := [8]float64{}
		for k := range pj {
			pj[k] = float64(TableVi[j][k])
		}
		for k := range t {
			t[k] = f[k] + pj[k%8]
			// Scale for E8
			t[k] = t[k] / 4
		}
		p[j][0], d[j][0] = E8Decode(t[0:8])
		p[j][1], d[j][1] = E8Decode(t[8:16])
		p[j][2], d[j][2] = E8Decode(t[16:24])
		for k := 0; k < 8; k++ {
			// Unscale
			// These should all be integers after this multiplication
			p[j][0][k] *= 4
			p[j][1][k] *= 4
			p[j][2][k] *= 4
			// technically each d (square of distance) should also be unscaled by 16
			// but it doesn't effect which one is the minimum
		}
	}
	return p, d
}

/*
This function iterates through each of the possible arrangements as given in Table 7
And finds the distance given by the arrangement
It then returns the closest point and its corresponding distance
*/
func LeechLatticeClosestPoint(f []float64) ([]float64, float64) {
	p, d := LeechLatticeClosest(f)
	best := math.MaxFloat64
	bestIndex := 0
	for j := range TableVii {
		dist := d[TableVii[j][0]][0] + d[TableVii[j][1]][1] + d[TableVii[j][2]][2]
		if dist < best {
			best = dist
			bestIndex = j
		}
	}
	bestPoint := make([]float64, 24)
	copy(bestPoint[0:8], p[TableVii[bestIndex][0]][0][:])
	copy(bestPoint[8:16], p[TableVii[bestIndex][1]][1][:])
	copy(bestPoint[16:24], p[TableVii[bestIndex][2]][2][:])
	// here we unscale d in case someone wants the accurate distance information
	return bestPoint, best * 16
}

/*
This function returns the k closest points and distances instead
*/
func LeechLatticeClosestPoints(f []float64, numPoints int) ([][]float64, []float64) {
	p, d := LeechLatticeClosest(f)
	c := Candidates{make([]uint64, len(TableVii)), make([]float64, len(TableVii))}
	for j := range TableVii {
		c.Distances[j] = d[TableVii[j][0]][0] + d[TableVii[j][1]][1] + d[TableVii[j][2]][2]
		c.Indexes[j] = uint64(j)
	}
	// priority queue or heap would be faster
	sort.Sort(&c)
	bestPoints := make([][]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		bestPoints[i] = make([]float64, 24)
		copy(bestPoints[i][0:8], p[TableVii[c.Indexes[i]][0]][0][:])
		copy(bestPoints[i][8:16], p[TableVii[c.Indexes[i]][1]][1][:])
		copy(bestPoints[i][16:24], p[TableVii[c.Indexes[i]][2]][2][:])

	}
	return bestPoints, c.Distances[:numPoints]
}

/*
We use table IV instead of table V, it is slower but more understandable
*/
var TableIVa = [16][8]int8{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{4, 0, 0, 0, 0, 0, 0, 0},
	{2, 2, 2, 2, 0, 0, 0, 0},
	{-2, 2, 2, 2, 0, 0, 0, 0},
	{2, 2, 0, 0, 2, 2, 0, 0},
	{-2, 2, 0, 0, 2, 2, 0, 0},
	{2, 2, 0, 0, 0, 0, 2, 2},
	{-2, 2, 0, 0, 0, 0, 2, 2},
	{2, 0, 2, 0, 2, 0, 2, 0},
	{-2, 0, 2, 0, 2, 0, 2, 0},
	{2, 0, 2, 0, 0, 2, 0, 2},
	{-2, 0, 2, 0, 0, 2, 0, 2},
	{2, 0, 0, 2, 2, 0, 0, 2},
	{-2, 0, 0, 2, 2, 0, 0, 2},
	{2, 0, 0, 2, 0, 2, 2, 0},
	{-2, 0, 0, 2, 0, 2, 2, 0},
}

var TableIVt = [18][8]int8{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{2, 2, 2, 0, 0, 2, 0, 0},
	{2, 2, 0, 2, 0, 0, 0, 2},
	{2, 0, 2, 2, 0, 0, 2, 0},
	{0, 2, 2, 2, 2, 0, 0, 0},
	{2, 2, 0, 0, 2, 0, 2, 0},
	{2, 0, 2, 0, 2, 0, 0, 2},
	{2, 0, 0, 2, 2, 2, 0, 0},
	{-3, 1, 1, 1, 1, 1, 1, 1},
	{3, -1, -1, 1, 1, -1, 1, 1},
	{3, -1, 1, -1, 1, 1, 1, -1},
	{3, 1, -1, -1, 1, 1, -1, 1},
	{3, 1, 1, 1, 1, -1, -1, -1},
	{3, -1, 1, 1, -1, 1, -1, 1},
	{3, 1, -1, 1, -1, 1, 1, -1},
	{3, 1, 1, -1, -1, -1, 1, 1},
}

// Table 5 requires a less understandable E_8 decoder as it rotates the lattice
// If speed is a concern, there are much better decoders
/*
var TableVa = [16][8]int8{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{2, 2, 0, 0, 0, 0, 0, 0},
	{2, 0, 2, 0, 0, 0, 0, 0},
	{2, 0, 0, 2, 0, 0, 0, 0},
	{2, 0, 0, 0, 2, 0, 0, 0},
	{2, 0, 0, 0, 0, 2, 0, 0},
	{2, 0, 0, 0, 0, 0, 2, 0},
	{2, 0, 0, 0, 0, 0, 0, 2},
	{1, 1, 1, 1, 1, 1, 1, 1},
	{-1, -1, 1, 1, 1, 1, 1, 1},
	{-1, 1, -1, 1, 1, 1, 1, 1},
	{-1, 1, 1, -1, 1, 1, 1, 1},
	{-1, 1, 1, 1, -1, 1, 1, 1},
	{-1, 1, 1, 1, 1, -1, 1, 1},
	{-1, 1, 1, 1, 1, 1, -1, 1},
	{-1, 1, 1, 1, 1, 1, 1, -1},
}

var TableVt = [16][8]int8{
	{0, 0, 0, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, -1, 1, 1, 1},
	{-1, 1, 1, 1, 2, 0, 0, 0},
	{2, 0, 0, 0, 1, 1, 1, 1},
	{1, 1, 0, 2, 1, 1, 0, 0},
	{2, 0, 1, 1, 0, 0, 1, -1},
	{1, 1, 0, 0, 2, 0, -1, 1},
	{2, 0, -1, 1, -1, 1, 0, 0},
	{1, 2, 1, 0, 1, 0, 1, 0},
	{2, 1, 0, 1, 0, -1, 0, 1},
	{1, 0, 1, 0, 2, 1, 0, -1},
	{2, 1, 0, -1, -1, 0, 1, 0},
	{1, 0, 2, 1, 1, 0, 0, 1},
	{2, 1, 1, 0, 0, 1, -1, 0},
	{1, 0, 0, 1, 2, -1, 1, 0},
	{2, -1, 1, 0, -1, 0, 0, 1},
}
*/

var TableVi = [256][8]int8{}
var TableVii = [4096][3]uint8{}

/*
This function implements the precomputation steps as described in https://ieeexplore.ieee.org/stamp/stamp.jsp?tp=&arnumber=1057135
*/
func Precompute() {
	index := 0
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			for k := 0; k < 8; k++ {
				TableVi[index][k] = TableIVa[i][k] + TableIVt[j][k]
			}
			index++
		}
	}
	if index != 256 {
		panic("expected 256 elements")
	}
	index = 0
	for ti := 0; ti < 16; ti++ {
		t := TableIVt[ti]
		for ai := 0; ai < 16; ai++ {
			a := TableIVa[ai]
			at := [8]int8{}
			for k := 0; k < 8; k++ {
				at[k] = a[k] + t[k]
			}
			atIndex := FindTableIndex(at)
			for bi := 0; bi < 16; bi++ {
				b := TableIVa[bi]
				bt := [8]int8{}
				sum := [8]int8{}
				for k := 0; k < 8; k++ {
					bt[k] = b[k] + t[k]
					sum[k] = a[k] + b[k]
				}
				btIndex := FindTableIndex(bt)
				for ci := 0; ci < 16; ci++ {
					c := TableIVa[ci]
					s := sum
					for k := 0; k < 8; k++ {
						s[k] += c[k]
					}
					if !Is4E8Point(s) {
						continue
					}
					ct := [8]int8{}
					for k := 0; k < 8; k++ {
						ct[k] = c[k] + t[k]
					}
					ctIndex := FindTableIndex(ct)
					TableVii[index][0] = atIndex
					TableVii[index][1] = btIndex
					TableVii[index][2] = ctIndex
					index++
					break
				}
			}
		}
	}
	if index != 4096 {
		panic("Expected 4096 elements")
	}
}

func FindTableIndex(v [8]int8) uint8 {
	for i := 0; i < 256; i++ {
		if TableVi[i] == v {
			return uint8(i)
		}
	}
	panic("No matching vector found")
}

// Determine if v is a point of the lambda_8 = E_8 * 4 lattice
func Is4E8Point(v [8]int8) bool {
	sum := 0
	for i := range v {
		sum += int(v[i])
	}
	// must be even in E8, so must be multiple of 8 in 4E8
	if sum%8 != 0 {
		return false
	}
	// all of the coordinates are integers or all of the coordinates are half integers
	// -> all of the coordinates are 0 mod 4 or all are 2 mod 4
	ok := true
	for i := range v {
		if v[i]%4 != 0 {
			ok = false
			break
		}
	}
	if ok {
		return true
	}
	for i := range v {
		m := v[i] % 4
		// because of how modulus operator works on negatives
		// (could just use bit operations instead of modulus)
		if m != 2 && m != -2 {
			return false
		}
	}
	return true
}

func DistSquared(f1 []float64, f2 []float64) float64 {
	d := 0.0
	for k := range f1 {
		diff := f1[k] - f2[k]
		d += diff * diff
	}
	return d
}
