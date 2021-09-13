package pir

import (
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/sachaservan/private-ann/pir/field"
)

// test configuration parameters
const TestDBSize = 1 << 14
const BenchmarkDBSize = 1 << 20

const SlotBytes = 3
const SlotBytesStep = 5
const NumQueries = 50 // number of queries to run

func setup() {
	rand.Seed(time.Now().Unix())
}

// run with 'go test -v -run TestSharedQuery' to see log outputs.
func TestSharedQuery(t *testing.T) {
	setup()
	prof, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(prof)
	defer pprof.StopCPUProfile()

	db := GenerateRandomDB(TestDBSize, SlotBytes)

	for i := 0; i < NumQueries; i++ {
		qIndex := rand.Intn(db.DBSize)
		shares := db.NewIndexQueryShares(qIndex, 2)

		resA, err := db.PrivateSecretSharedQuery(shares[0])
		if err != nil {
			t.Fatalf("%v", err)
		}

		resB, err := db.PrivateSecretSharedQuery(shares[1])
		if err != nil {
			t.Fatalf("%v", err)
		}

		resultShares := [...]*SecretSharedQueryResult{resA, resB}
		res := Recover(resultShares[:])

		if db.Data[qIndex] != res {
			t.Fatalf(
				"Query result is incorrect. %v != %v\n",
				db.Data[qIndex],
				res,
			)
		}

		t.Logf("Slot %v, is %v\n", qIndex, res)
	}

}

func BenchmarkBuildDB(b *testing.B) {
	setup()

	// benchmark index build time
	for i := 0; i < b.N; i++ {
		GenerateRandomDB(BenchmarkDBSize, SlotBytes)
	}
}

func BenchmarkQuerySecretShares(b *testing.B) {
	setup()

	db := GenerateRandomDB(BenchmarkDBSize, SlotBytes)
	queryA := db.NewIndexQueryShares(0, 2)[0]

	b.ResetTimer()

	// benchmark index build time
	for i := 0; i < b.N; i++ {
		_, err := db.PrivateSecretSharedQuery(queryA)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkQueryGen(b *testing.B) {
	setup()

	db := GenerateRandomDB(BenchmarkDBSize, SlotBytes)

	b.ResetTimer()

	// benchmark index build time
	for i := 0; i < b.N; i++ {
		db.NewIndexQueryShares(0, 2)
	}
}

// GenerateRandomDB generates a database of slots (where each slot is of size NumBytes)
// the width and height parameter specify the number of rows and columns in the database
func GenerateRandomDB(size, numBytes int) *Database {

	db := Database{}
	db.Data = make([]field.FP, size)
	db.DBSize = size
	for i := 0; i < size; i++ {
		db.Data[i] = field.RandomFieldElement()
	}

	return &db
}

// GenerateEmptyDB  generates an empty database
func GenerateEmptyDB(size, numBytes int) *Database {

	db := Database{}
	db.Data = make([]field.FP, size)
	db.DBSize = size

	return &db
}
