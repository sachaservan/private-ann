package dpfc

// Testing in C the time goes from 7 to 4 seconds for 1000000 with the O3 flag
// Since cgo removes all optimization flags we first compile a (optimized) static library and then link it

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: ${SRCDIR}/src/libdpf.a -lcrypto -lssl -lm
// #include "dpf.h"
import "C"
import (
	"unsafe"
)

var HASH1BLOCKOUT uint = 4
var HASH2BLOCKOUT uint = 2

type PrfCtx *C.struct_evp_cipher_ctx_st
type Hash *C.struct_Hash

func NewDPFKey(bytes []byte, rangeSize uint, index uint64) *DPFKey {
	return &DPFKey{bytes, rangeSize, index}
}

func getRequiredKeySize(rangeSize uint) uint {
	// this is the required key size for the VDPF
	// so we overallocate for the DPF
	// TODO: this is all super hacky. Would be nice
	// to patch this up.
	return 18*rangeSize + 18 + 16 + 16*4
}

func InitDPFContext(prfKey []byte) PrfCtx {
	if len(prfKey) != 16 {
		panic("bad prf key size")
	}

	p := C.getDPFContext((*C.uchar)(unsafe.Pointer(&prfKey[0])))
	return p
}

func DestroyDPFContext(ctx PrfCtx) {
	C.destroyContext(ctx)
}

func (dpf *Dpf) GenDPFKeys(specialIndex uint64, rangeSize uint) (*DPFKey, *DPFKey) {

	keySize := getRequiredKeySize(rangeSize)
	k0 := make([]byte, keySize)
	k1 := make([]byte, keySize)

	C.genDPF(
		dpf.ctx,
		C.int(rangeSize),
		C.uint64_t(specialIndex),
		(*C.uchar)(unsafe.Pointer(&k0[0])),
		(*C.uchar)(unsafe.Pointer(&k1[0])),
	)

	return NewDPFKey(k0, rangeSize, 0), NewDPFKey(k1, rangeSize, 1)
}

func (dpf *Dpf) BatchEval(key *DPFKey, indices []uint64) []uint64 {

	keySize := getRequiredKeySize(key.RangeSize)
	if len(key.Bytes) != int(keySize) {
		panic("invalid key size")
	}

	res := make([]uint64, len(indices)*2) // returned output is uint128_t
	resTrunc := make([]uint64, len(indices))

	C.batchEvalDPF(
		dpf.ctx,
		C.int(key.RangeSize),
		C.bool(key.Index == 1),
		(*C.uchar)(unsafe.Pointer(&key.Bytes[0])),
		(*C.uint64_t)(unsafe.Pointer(&indices[0])),
		C.uint64_t(len(indices)),
		(*C.uint8_t)(unsafe.Pointer(&res[0])),
	)

	// skip two uint64 blocks at a time
	b := 0
	for i := 0; i < len(res); i += 2 {
		resTrunc[b] = res[i]
		b++
	}

	return resTrunc
}
