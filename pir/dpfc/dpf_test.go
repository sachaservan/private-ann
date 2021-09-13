package dpfc

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/sachaservan/private-ann/pir/field"
)

const numTrials = 1000

func TestDPF(t *testing.T) {
	client := ClientInitialize()
	keyA, keyB := client.GenerateKeys(5)
	server := ServerInitialize(client.PrfKey)

	share0 := server.Evaluate(keyA, 5)
	share1 := server.Evaluate(keyB, 5)
	if field.Add(share0, share1) != 1 {
		fmt.Printf("Other share 0 %v\n", share0)
		fmt.Printf("Other share 1 %v\n", share1)
		t.Fail()
	}
	share0 = server.Evaluate(keyA, 26)
	share1 = server.Evaluate(keyB, 26)
	if field.Add(share0, share1) != 0 {
		fmt.Printf("Other share 0 %v\n", share0)
		fmt.Printf("Other share 1 %v\n", share1)
		t.Fail()
	}
}

func TestCorrectPointFunctionTwoServer(t *testing.T) {

	for trial := 0; trial < numTrials; trial++ {
		num := rand.Intn(1<<10) + 100

		specialIndex := uint64(rand.Intn(num))

		// generate fss Keys on client
		client := ClientInitialize()

		// fmt.Printf("index  %v\n", specialIndex)
		keyA, keyB := client.GenerateKeys(specialIndex)

		// fmt.Printf("keyA = %v\n", keyA)
		// fmt.Printf("keyB = %v\n", keyB)

		// simulate the server
		server := ServerInitialize(client.PrfKey)

		for i := 0; i < num; i++ {
			ans0 := server.Evaluate(keyA, uint64(i))
			ans1 := server.Evaluate(keyB, uint64(i))

			// fmt.Printf("ans0 = %v\n", ans0)
			// fmt.Printf("ans1 = %v\n", ans1)

			sum := field.Add(ans0, ans1)

			if uint64(i) == specialIndex && uint(sum) != 1 {
				t.Fatalf("Expected: %v Got: %v", 1, sum)
			}

			if uint64(i) != specialIndex && sum != 0 {
				t.Fatalf("Expected: 0 Got: %v", sum)
			}
		}
	}
}

func Benchmark2PartyServerInit(b *testing.B) {

	fClient := ClientInitialize()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ServerInitialize(fClient.PrfKey)
	}
}

func Benchmark2Party64BitKeywordEval(b *testing.B) {

	client := ClientInitialize()
	keyA, _ := client.GenerateKeys(1)
	server := ServerInitialize(client.PrfKey)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		server.Evaluate(keyA, uint64(i))
	}
}
