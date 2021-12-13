package dpfc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/sachaservan/private-ann/pir/field"
)

const numTrials = 1000

func TestCorrectPointFunctionTwoServer(t *testing.T) {

	for trial := 0; trial < numTrials; trial++ {
		num := 20

		specialIndex := uint64(rand.Intn(num))

		// generate fss Keys on client
		client := ClientDPFInitialize()

		// fmt.Printf("index  %v\n", specialIndex)
		keyA, keyB := client.GenDPFKeys(specialIndex, 64)

		// fmt.Printf("keyA = %v\n", keyA)
		// fmt.Printf("keyB = %v\n", keyB)

		// simulate the server
		server := ServerDPFInitialize(client.PrfKey)

		indices := make([]uint64, num)
		for i := 0; i < num; i++ {
			indices[i] = uint64(rand.Intn(num))
		}
		ans0 := server.BatchEval(keyA, indices)
		ans1 := server.BatchEval(keyB, indices)

		// fmt.Printf("ans0 = %v\n", ans0)
		// fmt.Printf("ans1 = %v\n", ans1)
		for i := 0; i < num; i++ {

			fmt.Printf("ans0 = %v\n", ans0[i])
			fmt.Printf("ans1 = %v\n", ans1[i])

			sum := field.Add(field.FP(ans0[i]), field.FP(ans1[i]))

			if uint64(indices[i]) == specialIndex && uint(sum) != 1 {
				t.Fatalf("Expected: %v Got: %v", 1, sum)
			}

			if uint64(indices[i]) != specialIndex && sum != 0 {
				t.Fatalf("Expected: 0 Got: %v", sum)
			}
		}
	}
}

func Benchmark2PartyServerInit(b *testing.B) {

	fClient := ClientDPFInitialize()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ServerDPFInitialize(fClient.PrfKey)
	}
}

func Benchmark2Party64BitKeywordEval(b *testing.B) {

	client := ClientDPFInitialize()
	keyA, _ := client.GenDPFKeys(1, 64)
	server := ServerDPFInitialize(client.PrfKey)

	indices := make([]uint64, 1)
	indices[0] = 1

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		server.BatchEval(keyA, indices)
	}
}

func BenchmarkDPFGen(b *testing.B) {

	client := ClientDPFInitialize()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.GenDPFKeys(1, 256)

	}
}
