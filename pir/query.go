package pir

import (
	"github.com/sachaservan/private-ann/pir/dpfc"
	"github.com/sachaservan/private-ann/pir/field"
)

// QueryShare is a secret share of a query over the database
// to retrieve a row
type QueryShare struct {
	DPFKey         *dpfc.DPFKey
	PrfKey         dpfc.PrfKey
	ShareNumber    uint
	IsKeywordBased bool
}

// BatchQueryShare is a secret share of a batch query over the database
// to retrieve a row
type BatchQueryShare struct {
	Queries []*QueryShare
}

// NewIndexQueryShares generates PIR query shares for the index
func (dbmd *DBMetadata) NewIndexQueryShares(index uint64, numShares uint, rangeBits uint) []*QueryShare {
	return dbmd.newQueryShares(index, numShares, true, rangeBits)
}

// NewKeywordQueryShares generates keyword-based PIR query shares for keyword
func (dbmd *DBMetadata) NewKeywordQueryShares(keyword uint64, numShares uint, rangeBits uint) []*QueryShare {
	return dbmd.newQueryShares(keyword, numShares, false, rangeBits)
}

// NewQueryShares generates random PIR query shares for the index
func (dbmd *DBMetadata) newQueryShares(key uint64, numShares uint, isIndexQuery bool, rangeBits uint) []*QueryShare {

	if numShares != 2 {
		panic("only two-server DPF supported")
	}

	client := dpfc.ClientDPFInitialize()

	keyA, keyB := client.GenDPFKeys(key, rangeBits)

	shares := make([]*QueryShare, numShares)
	for i := 0; i < int(numShares); i++ {
		shares[i] = &QueryShare{}
		shares[i].ShareNumber = uint(i)
		shares[i].PrfKey = client.PrfKey
		shares[i].IsKeywordBased = !isIndexQuery

		if i == 0 {
			shares[i].DPFKey = keyA
		} else {
			shares[i].DPFKey = keyB
		}
	}

	client.Free()

	return shares
}

// Recover combines shares of slots to recover the data
func Recover(resShares []*SecretSharedQueryResult) field.FP {

	res := field.FP(0)
	for _, s := range resShares {
		res = field.Add(res, s.Share)
	}

	return res
}
