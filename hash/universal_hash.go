package hash

import (
	"math"
	"math/rand"

	"github.com/ncw/gmp"
	"github.com/sachaservan/vec"
)

// Basic vector universal hash function for the field F_p
// Output a_0 + a_1 x_1 + a_2 x_2 + \ldots  mod p.

type UniversalHash struct {
	Coefficients []*gmp.Int
	Modulus      *gmp.Int
}

// largest 64 bit prime
const Prime = 18446744073709551557

func NewUniversalHash(dim int) *UniversalHash {
	u := &UniversalHash{Coefficients: make([]*gmp.Int, dim+1)}
	u.Modulus = new(gmp.Int).SetUint64(Prime)
	for i := range u.Coefficients {
		rBytes := make([]byte, 8)
		rand.Read(rBytes)
		u.Coefficients[i] = new(gmp.Int).SetBytes(rBytes)

		// keep sampling until we get a suitable element from the field
		// to avoid biased universal hashing
		for u.Coefficients[i].Cmp(u.Modulus) >= 0 {
			rand.Read(rBytes)
			u.Coefficients[i] = new(gmp.Int).SetBytes(rBytes)
		}
	}
	return u
}

func (u *UniversalHash) Hash(v *vec.Vec) uint64 {
	if v.Size()+1 != len(u.Coefficients) {
		panic("Universal hash size mismatch")
	}
	t := new(gmp.Int)
	s := new(gmp.Int).Set(u.Coefficients[0])
	for i, f := range v.Coords {
		t.SetUint64(math.Float64bits(f))
		s = s.AddMul(u.Coefficients[i+1], t)
	}
	s = s.Mod(s, u.Modulus)
	return s.Uint64()
}
