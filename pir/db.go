package pir

import (
	"errors"

	"github.com/sachaservan/private-ann/pir/dpfc"
	"github.com/sachaservan/private-ann/pir/field"
)

// DBMetadata contains information on the layout
// and size information for a slot database type
type DBMetadata struct {
	DBSize int
}

// Database is a set of slots arranged in a grid of size width x height
// where each slot has size slotBytes
type Database struct {
	DBMetadata
	Data     []field.FP
	Keywords []uint64 // set of keywords (optional)

	BatchSize   int   // (for batch queries) number of batches (aka regions)
	BatchStarts []int // (for batch queries) start index of each key region
	BatchStops  []int // (for batch queries) end index of each key region
}

// SecretSharedQueryResult contains shares of the resulting slots
type SecretSharedQueryResult struct {
	Share field.FP
}

// NewDatabase returns an empty database
func NewDatabase() *Database {
	return &Database{}
}

func (db *Database) SetBatchingParameters(batchSize int, batchStarts []int, batchStops []int) error {
	if batchSize == 0 {
		return errors.New("no batching parameters specified")
	}

	if batchSize != len(batchStarts) || batchSize != len(batchStops) {
		return errors.New("invalid batching parameters")
	}

	// make sure that keywords are sorted (if specified)
	if db.Keywords != nil && len(db.Keywords) != 0 {
		for i := 0; i < len(db.Keywords)-1; i++ {
			if db.Keywords[i] > db.Keywords[i+1] {
				return errors.New("keywords not sorted")
			}
		}
	}

	db.BatchSize = batchSize
	db.BatchStarts = batchStarts
	db.BatchStops = batchStops

	return nil
}

// PrivateSecretSharedQuery uses the provided PIR query to retreive a slot row
func (db *Database) PrivateSecretSharedQuery(query *QueryShare) (*SecretSharedQueryResult, error) {

	bits := db.ExpandSharedQuery(query, 0, db.DBSize)
	return db.PrivateSecretSharedQueryWithExpandedBits(query, bits, 0, db.DBSize)
}

// PrivateSecretSharedBatchQuery uses the provided PIR query to retreive a slot row
func (db *Database) PrivateSecretSharedBatchQuery(batchQuery *BatchQueryShare) ([]*SecretSharedQueryResult, error) {

	if db == nil {
		panic("database is null")
	}

	if db.BatchSize == 0 {
		panic("no batching parameters specified")
	}

	if db.BatchSize != len(db.BatchStarts) || db.BatchSize != len(db.BatchStops) {
		panic("invalid batching parameters")
	}

	var err error
	results := make([]*SecretSharedQueryResult, db.BatchSize)
	for b := 0; b < len(batchQuery.Queries); b++ {
		start := db.BatchStarts[b]
		stop := db.BatchStops[b]

		bits := db.ExpandSharedQuery(batchQuery.Queries[b], start, stop)
		results[b], err = db.PrivateSecretSharedQueryWithExpandedBits(batchQuery.Queries[b], bits, start, stop)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// PrivateSecretSharedQueryWithExpandedBits returns the result without expanding the query DPF
// start: index of start key
// stop: index of end key
func (db *Database) PrivateSecretSharedQueryWithExpandedBits(query *QueryShare, bits []field.FP, start, stop int) (*SecretSharedQueryResult, error) {

	result := field.FP(0)

	i := 0
	for row := start; row < stop; row++ {
		result = field.Add(result, field.Multiply(db.Data[row], bits[i]))
		i++
	}

	return &SecretSharedQueryResult{result}, nil
}

// ExpandSharedQuery returns the expands the DPF and returns an array of bits
// start: index of start key
// stop: index of end key
func (db *Database) ExpandSharedQuery(query *QueryShare, start, stop int) []field.FP {
	// init server DPF
	pf := dpfc.ServerInitialize(query.PrfKey)

	bits := make([]field.FP, db.DBSize)

	// expand the DPF into the bits array
	for i := start; i < stop; i++ {
		// key (index or uint) depending on whether
		// the query is keyword based or index based
		// when keyword based use FSS
		var key uint64
		if query.IsKeywordBased {
			key = db.Keywords[i]
		} else {
			key = uint64(i)
		}

		bits[i] = pf.Evaluate(query.DPFKey, key)
	}

	return bits
}

func (db *Database) BuildForKeysAndValues(keys []uint64, data []field.FP) error {
	db.BuildForData(data)
	err := db.SetKeywords(keys)
	return err
}

// BuildForData constrcuts a PIR database
func (db *Database) BuildForData(data []field.FP) {
	db.Data = data
	db.DBSize = len(data)
}

// SetKeywords set the keywords (uint64) associated with each row of the database
func (db *Database) SetKeywords(keywords []uint64) error {
	if len(keywords) != db.DBSize {
		return errors.New("number of keywords should match database size")
	}

	db.Keywords = keywords

	return nil
}
