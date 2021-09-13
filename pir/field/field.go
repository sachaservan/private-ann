package field

import (
	"math/big"
	"math/rand"
)

type FP uint64

// need prime > n so that every id can be represented
// need prime > nL so no overflow during oblivious masking
const fieldPrime = 2147483647 // 2^31-1, 31 bits
var fieldPrimeBigInt *big.Int

func init() {
	fieldPrimeBigInt = big.NewInt(fieldPrime)
}

// [0, p) + [0, p) -> [0, p)
func Add(a, b FP) FP {
	out := a + b
	if out >= fieldPrime {
		out = out - fieldPrime
	}
	return out
}

func Negate(a FP) FP {
	if a != 0 {
		return fieldPrime - a
	}
	return 0
}

func Multiply(a, b FP) FP {
	return fieldMod(a * b)
}

// Reduce [0, p^2) to [0, p)
func fieldMod(a FP) FP {
	// in general
	// return FP(a % fieldPrime)

	// go compiler might be smart enough to optimize this
	// One optimization could be
	return Add(FP(a>>31), FP(a&fieldPrime))
}

func RandomFieldElement() FP {
	r := rand.Intn(fieldPrime)
	return FP(r)
}
