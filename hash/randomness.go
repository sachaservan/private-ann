package hash

import (
	"math"
	"math/rand"

	"github.com/sachaservan/vec"
)

// Return a random rotation matrix chosen uniformly
func RandomRotationMatrix(dim int) []*vec.Vec {
	return GetRandomRotation(dim)
}

// Return a random directional vector
func RandomVector(dim int) *vec.Vec {
	return Normals(dim)
}

// Return a random vector chosen uniformly from the cube
func RandomTranslationVector(dim int, max float64) *vec.Vec {
	v := vec.NewVec(make([]float64, dim))
	for i := range v.Coords {
		v.SetValueToCoord((rand.Float64()-0.5)*2*max, i)
	}
	return v
}

// Identity matrix
func Identity(dim int) []*vec.Vec {
	m := make([]*vec.Vec, dim)
	for i := range m {
		row := make([]float64, dim)
		row[i] = 1
		m[i] = vec.NewVec(row)
	}
	return m
}

// Gaussian matrix
func Normals(size int) *vec.Vec {
	d := make([]float64, size)
	for i := range d {
		d[i] = rand.NormFloat64()
	}
	return vec.NewVec(d)
}

// Signum, but not 0
func Sign(v float64) int {
	if v >= 0 {
		return 1
	} else {
		return -1
	}
}

// these lowercase functions utilizing aliasing
// because they are only built to work with the rotation function below

func scalarMultiply(v *vec.Vec, a float64) *vec.Vec {
	for i := range v.Coords {
		v.Coords[i] = v.Coord(i) * a
	}
	return v
}

func matrixVectorDot(m []*vec.Vec, v *vec.Vec) *vec.Vec {
	p := make([]float64, len(m))
	for i := range p {
		p[i], _ = m[i].Dot(v)
	}
	return vec.NewVec(p)
}

func slice(v *vec.Vec, start, end int) *vec.Vec {
	return vec.NewVec(v.Coords[start:end])
}

func matSlice(m []*vec.Vec, r1, r2, c1, c2 int) []*vec.Vec {
	m2 := make([]*vec.Vec, r2-r1)
	for i := range m2 {
		m2[i] = slice(m[i+r1], c1, c2)
	}
	return m2
}

func outerProduct(v1 *vec.Vec, v2 *vec.Vec) []*vec.Vec {
	o := make([]*vec.Vec, v1.Size())
	for i := range o {
		s := make([]float64, v2.Size())
		for j := range s {
			s[j] = v1.Coord(i) * v2.Coord(j)
		}
		o[i] = vec.NewVec(s)
	}
	return o
}

func matSub(minuend, subtrahend []*vec.Vec) {
	for i := 0; i < len(minuend); i++ {
		v := minuend[i]
		for j := 0; j < v.Size(); j++ {
			v.AddToCoord(-subtrahend[i].Coord(j), j)
		}
	}
}

func product(v []int) int {
	f := 1
	for i := range v {
		f *= v[i]
	}
	return f
}

func transpose(v []*vec.Vec) []*vec.Vec {
	t := make([]*vec.Vec, v[0].Size())
	for j := range t {
		d := make([]float64, len(v))
		for i := range v {
			d[i] = v[i].Coord(j)
		}
		t[j] = vec.NewVec(d)
	}
	return t
}

// part of householder transform
func op(v []int, m []*vec.Vec) []*vec.Vec {
	for i := range m {
		for j := range v {
			m[i].SetValueToCoord(m[i].Coord(j)*float64(v[j]), j)
		}
	}
	return m
}

func GetRandomRotation(dim int) []*vec.Vec {
	// scipy/stats/_multivariate.py:3418-3432
	H := Identity(dim)
	D := make([]int, dim)
	for n := 0; n < dim-1; n++ {
		x := Normals(dim - n)
		norm2, _ := x.Dot(x)
		x0 := x.Coord(0)
		D[n] = Sign(x0)
		x.AddToCoord(float64(D[n])*math.Sqrt(norm2), 0)
		// This line is reciprocated so it is a multiply on the next line
		f := math.Sqrt(2.0 / (norm2 - x0*x0 + x.Coord(0)*x.Coord(0)))
		x = scalarMultiply(x, f)
		// Householder transformation
		matSub(matSlice(H, 0, dim, n, dim),
			outerProduct(
				matrixVectorDot(
					matSlice(H, 0, dim, n, dim),
					x),
				x))
	}
	if (dim-1)%2 == 0 {
		D[dim-1] = product(D[0 : dim-1])
	} else {
		D[dim-1] = -product(D[0 : dim-1])
	}
	// apparently
	H = transpose(op(D, transpose(H)))
	return H
}
