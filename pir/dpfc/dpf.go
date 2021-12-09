package dpfc

import "crypto/rand"

type PrfKey [16]byte

type DPFKey struct {
	Bytes     []byte
	RangeSize uint
	Index     uint64
}

type Dpf struct {
	PrfKey PrfKey
	ctx    PrfCtx
}

func ClientDPFInitialize() *Dpf {
	randKey := PrfKey{}
	_, err := rand.Read(randKey[:])
	if err != nil {
		panic("Error generating prf randomness")
	}
	return &Dpf{randKey, InitDPFContext(randKey[:])}
}

func ServerDPFInitialize(key PrfKey) *Dpf {
	return &Dpf{key, InitDPFContext(key[:])}
}

func (dpf *Dpf) Free() {
	DestroyDPFContext(dpf.ctx)
}
