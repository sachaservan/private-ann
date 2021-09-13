package dpfc

import "crypto/rand"

type PrfKey [16]byte

type DPFKey struct {
	Bytes []byte
	Index int
}
type Dpf struct {
	PrfKey PrfKey
	ctx    PrfCtx
}

func ClientInitialize() *Dpf {
	randKey := PrfKey{}
	_, err := rand.Read(randKey[:])
	if err != nil {
		panic("Error generating prf randomness")
	}
	return &Dpf{randKey, InitContext(randKey[:])}
}

func ServerInitialize(key PrfKey) *Dpf {
	return &Dpf{key, InitContext(key[:])}
}
