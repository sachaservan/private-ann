package server

import (
	"math/rand"
	"testing"

	"github.com/sachaservan/private-ann/pir"
	"github.com/sachaservan/private-ann/pir/field"
)

func TestObliviousMasking(t *testing.T) {

	nslots := 10

	for i := 0; i < 100; i++ {
		index := rand.Intn(nslots)

		slots := make([]field.FP, nslots)
		specialSlot := field.RandomFieldElement()
		empty := field.FP(0)

		// make sure special slot is non-zero
		for specialSlot == empty {
			specialSlot = field.RandomFieldElement()
		}

		// fmt.Printf("special slot = %v\n", specialSlot)

		for i := 0; i < nslots; i++ {
			// all-zero for i < index
			slots[i] = field.FP(0)

			// make everything after i = index either the special slot
			// or some combination of the special slot, random slots, and zero.
			if i == index {
				slots[i] = specialSlot
			} else if i > index && rand.Intn(2) == 0 {
				slots[i] = specialSlot // add the special slot again
			} else if i > index && rand.Intn(2) == 0 {
				slots[i] = field.RandomFieldElement() // add a new random slot
			}
		}

		// convert to correct format for oblivious masking (PIR result)s
		queryRes := make([]*pir.SecretSharedQueryResult, nslots)
		for i := 0; i < nslots; i++ {
			queryRes[i] = &pir.SecretSharedQueryResult{}
			queryRes[i].Share = slots[i]

			// fmt.Println(queryRes[i].Share.Data)
		}

		masked := obliviousMasking(queryRes)

		for i := 0; i < nslots; i++ {

			if i < index && masked[i].Share != 0 {
				t.Fatalf("non-zero slot at index %v < %v", i, index)
			}

			if i == index && masked[i].Share != specialSlot {
				t.Fatalf("wrong slot at index %v == %v", i, index)
			}

			if i > index && masked[i].Share == 0 {
				t.Fatalf("non-random slot at index %v > %v", i, index)
			}

			if i > index && masked[i].Share == specialSlot {
				t.Fatalf("non-random slot at index %v > %v", i, index)
			}
		}
	}

}

func BenchmarkObliviousMasking(b *testing.B) {
	nslots := 10000
	slots := make([]field.FP, nslots)
	index := rand.Intn(nslots)

	specialSlot := field.RandomFieldElement()
	empty := field.FP(0)

	// make sure special slot is non-zero
	for specialSlot == empty {
		specialSlot = field.RandomFieldElement()
	}

	// fmt.Printf("special slot = %v\n", specialSlot)

	for i := 0; i < nslots; i++ {
		// all-zero for i < index
		slots[i] = field.FP(0)

		// make everything after i = index either the special slot
		// or some combination of the special slot, random slots, and zero.
		if i == index {
			slots[i] = specialSlot
		} else if i > index && rand.Intn(2) == 0 {
			slots[i] = specialSlot // add the special slot again
		} else if i > index && rand.Intn(2) == 0 {
			slots[i] = field.RandomFieldElement() // add a new random slot
		}
	}

	// convert to correct format for oblivious masking (PIR result)s
	queryRes := make([]*pir.SecretSharedQueryResult, nslots)
	for i := 0; i < nslots; i++ {
		queryRes[i] = &pir.SecretSharedQueryResult{}
		queryRes[i].Share = slots[i]

		// fmt.Println(queryRes[i].Share.Data)
	}
	for i := 0; i < b.N; i++ {

		masked := obliviousMasking(queryRes)

		for i := 0; i < nslots; i++ {

			if i < index && masked[i].Share != 0 {
				b.Fatalf("non-zero slot at index %v < %v", i, index)
			}

			if i == index && masked[i].Share != specialSlot {
				b.Fatalf("wrong slot at index %v == %v", i, index)
			}

			if i > index && masked[i].Share == 0 {
				b.Fatalf("non-random slot at index %v > %v", i, index)
			}

			if i > index && masked[i].Share == specialSlot {
				b.Fatalf("non-random slot at index %v > %v", i, index)
			}
		}
	}

}
