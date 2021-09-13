package api

import (
	"github.com/sachaservan/private-ann/hash"
	"github.com/sachaservan/private-ann/pir"
	"github.com/sachaservan/vec"
)

// Error is provided as a response to API queries
type Error struct {
	Msg string
}

// ANNQueryArgs arguments for querying a collection of hash tables
type ANNQueryArgs struct {
	SessionID    int64
	MultiProbes  int
	SecretShared []*pir.BatchQueryShare // MultiProbes queries for each hash table
}

// ANNQueryResponse responds with a set of (masked) PIR query results
type ANNQueryResponse struct {
	Error                Error
	SessionID            int64
	ResSecretShared      []*pir.SecretSharedQueryResult
	StatsQueryTimeInMS   int64
	StatsMaskingTimeInUS int64
}

// InitSessionArgs arguments provided by client to initialize a new a PIR session
type InitSessionArgs struct {
}

// InitSessionResponse response to a client following session creation
type InitSessionResponse struct {
	SessionParameters
	Error                      Error
	StatsPreprocessingTimeInMS int64
	StatsDatasetSize           int
	StatsNumFeatures           int
	StatsDatasetName           string
	StatsNumServerProcs        int
}

// TerminateSessionArgs used by client to kill the server (useful for experiments)
type TerminateSessionArgs struct{}

// TerminateSessionResponse response to clients terminate session call
type TerminateSessionResponse struct{}

// WaitForExperimentArgs is used by the client to wait until the experiment starts
// before making API calls
type WaitForExperimentArgs struct{}

// WaitForExperimentResponse is used to signal to the client that server is ready
type WaitForExperimentResponse struct{}

// SessionParameters contains all the metadata information
// needed for a client to issue PIR queries
type SessionParameters struct {
	SessionID           int64
	NumTables           int               // number of hash tables
	NumProbes           int               // number of bucket probes per table
	TestQuery           *vec.Vec          // a test query to use in the evaluation
	HashFunctions       []hash.Hash       // hash functions the client uses to compute keys
	TableBucketMetadata []*pir.DBMetadata // PIR db metadata for table buckets
}
