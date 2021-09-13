package hash

import (
	"github.com/sachaservan/vec"
)

// Abstraction for a LSH hash function

// Hash is an abstract hash function
type Hash interface {
	Hash(*vec.Vec) uint64
	// return k hashes (for multiprobing)
	MultiHash(*vec.Vec, int) []uint64
}
