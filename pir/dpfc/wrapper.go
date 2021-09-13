package dpfc

// Testing in C the time goes from 7 to 4 seconds for 1000000 with the O3 flag
// Since cgo removes all optimization flags we first compile a (optimized) static library and then link it

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: ${SRCDIR}/src/libdpf.a -lcrypto -lssl -lm
// #include "dpf.h"
import "C"
import (
	"unsafe"

	"github.com/sachaservan/private-ann/pir/field"
)

type PrfCtx *C.struct_evp_cipher_ctx_st

func NewDPFKey(bytes []byte, index int) *DPFKey {
	return &DPFKey{bytes, index}
}

const keySize = 18*64 + 18 + 8

func InitContext(prfKey []byte) PrfCtx {
	if len(prfKey) != 16 {
		panic("bad prf key size")
	}

	p := C.GetContext((*C.uchar)(unsafe.Pointer(&prfKey[0])))
	return p
}

func DestroyContext(ctx PrfCtx) {
	C.destroyContext(ctx)
}

func (dpf *Dpf) GenerateKeys(specialIndex uint64) (*DPFKey, *DPFKey) {

	k0 := make([]byte, keySize)
	k1 := make([]byte, keySize)

	C.genDPF(
		dpf.ctx, C.uint64_t(specialIndex),
		(*C.uchar)(unsafe.Pointer(&k0[0])),
		(*C.uchar)(unsafe.Pointer(&k1[0])),
	)

	return NewDPFKey(k0, 0), NewDPFKey(k1, 1)
}

func (dpf *Dpf) Evaluate(key *DPFKey, index uint64) field.FP {

	if len(key.Bytes) != keySize {
		panic("invalid key size")
	}

	res := C.evalDPF(
		dpf.ctx,
		C.bool(key.Index == 1),
		(*C.uchar)(unsafe.Pointer(&key.Bytes[0])),
		C.uint64_t(index),
	)

	return field.FP(res)
}
