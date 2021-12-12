package server

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/sachaservan/private-ann/cmd/api"
	"github.com/sachaservan/private-ann/hash"
	"github.com/sachaservan/private-ann/pir"
	"github.com/sachaservan/private-ann/pir/field"
	"github.com/sachaservan/vec"
)

// Server maintains all the necessary server state
type Server struct {
	DatasetName string
	DBSize      int

	// PIR databases containing the LSH tables
	TableDBs          []*pir.Database
	NumTables         int         // number of tables in total
	NumProbes         int         // number of probes performed per table
	TestQuery         *vec.Vec    // query that the client can use to test
	HashFunctions     []hash.Hash // LSH hash functions used to make the tables
	HashFunctionRange int         // range size of the universal hash function (in bits)

	NumProcs int // num processors to use
	Listener net.Listener
	Ready    bool // true when server has initialized
	Killed   bool // true if server killed

	CacheDir string // cache directory for storing pre-built hash tables

	StatsTotalPreprocessingTime int64 // time taken to build the hash tables
	StatsDatasetNumFeatures     int
}

// WaitForExperiment is used to signal to a waiting client that the server has finishied initializing
func (server *Server) WaitForExperiment(args *api.WaitForExperimentArgs, reply *api.WaitForExperimentResponse) error {

	for !server.Ready {
		time.Sleep(1 * time.Second)
	}

	return nil
}

// PrivateANNQuery performs PIR queries for buckets in the hash tables of the ANN data structure
func (server *Server) PrivateANNQuery(args *api.ANNQueryArgs, reply *api.ANNQueryResponse) error {

	log.Printf("[Server]: received request to PrivateANNQuery")

	start := time.Now()

	// numPartitions * numTables candidates
	candidates := make([]*pir.SecretSharedQueryResult, server.TableDBs[0].BatchSize*server.NumTables)

	wg := sync.WaitGroup{}
	wg.Add(server.NumTables)
	for t := 0; t < server.NumTables; t++ {
		go func(t int) {
			db := server.TableDBs[t]

			// results is a batch of results, one for each batch
			res, err := db.PrivateSecretSharedBatchQuery(args.SecretShared[t])
			if err != nil {
				panic(err)
			}

			// optional: rand.Shuffle(res)

			copy(candidates[t*db.BatchSize:(t+1)*db.BatchSize], res)
			wg.Done()
		}(t)
	}
	wg.Wait()

	reply.StatsQueryTimeInMS = time.Since(start).Milliseconds()

	start = time.Now()
	masked := obliviousMasking(candidates)
	reply.StatsMaskingTimeInUS = time.Since(start).Microseconds()

	reply.SessionID = args.SessionID
	reply.ResSecretShared = masked
	reply.StatsMaskingTimeInUS = int64(time.Since(start).Microseconds())

	log.Printf("[Server]: processed PrivateANNQuery request in %v ms", reply.StatsQueryTimeInMS)

	return nil
}

// TODO(sss): figure where this should live, not great to have it as a function in server
func obliviousMasking(slots []*pir.SecretSharedQueryResult) []*pir.SecretSharedQueryResult {

	// init the results
	res := make([]*pir.SecretSharedQueryResult, len(slots))
	for i := 0; i < len(slots); i++ {
		res[i] = &pir.SecretSharedQueryResult{}
	}

	sum := field.FP(0)
	for i := 0; i < len(slots); i++ {
		rand := field.RandomFieldElement()
		randSum := field.Multiply(rand, sum)
		res[i].Share = field.Add(slots[i].Share, randSum)
		sum = field.Add(sum, slots[i].Share)
	}

	return res
}
