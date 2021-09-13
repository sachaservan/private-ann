package field

import "testing"

func TestAdd(t *testing.T) {
	if Add(3, 5) != 8 {
		t.Fail()
	}

	// test overflow
	for x := 0; x < 10; x++ {
		p := fieldPrime - x
		if Add(FP(x), FP(p)) != 0 {
			t.Fail()
		}
	}

	for x := 0; x < 10; x++ {
		p := fieldPrime + 5 - x
		if Add(FP(x), FP(p)) != 5 {
			t.Fail()
		}
	}
	if Add(FP(fieldPrime-1), FP(fieldPrime-1)) != fieldPrime-2 {
		t.Fail()
	}
}

func TestModulus(t *testing.T) {
	if fieldMod(fieldPrime) != 0 {
		t.Fail()
	}
	if fieldMod(fieldPrime+1) != 1 {
		t.Fail()
	}
	if fieldMod(0) != 0 {
		t.Fail()
	}
	if fieldMod(1) != 1 {
		t.Fail()
	}
	if fieldMod(fieldPrime*fieldPrime-1) != fieldPrime-1 {
		t.Fail()
	}
	if fieldMod(fieldPrime*fieldPrime) != 0 {
		t.Fail()
	}
}
