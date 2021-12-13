package client

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/rpc"
	"sync"

	"github.com/sachaservan/private-ann/pir"

	"github.com/sachaservan/private-ann/cmd/api"
)

// RuntimeExperiment captures all the information needed to
// evaluate a two-server deployment
type RuntimeExperiment struct {
	DatasetName             string  `json:"dataset_name"`
	DatasetSize             int     `json:"dataset_size"`
	NumFeatures             int     `json:"num_features"`
	NumTables               int     `json:"num_tables"`
	NumProbes               int     `json:"num_probes"`
	HashFunctionRange       int     `json:"hash_function_range"`
	NumServerProcs          int     `json:"num_server_procs"`
	ServerPreprocessingMS   int64   `json:"server_preprocessing_ms"`
	QueryUpBandwidthBytes   []int64 `json:"query_up_bandwidth_bytes"`
	QueryDownBandwidthBytes []int64 `json:"query_down_bandwidth_bytes"`
	QueryServerMS           []int64 `json:"dpf_server_ms"`
	QueryMaskingServerUS    []int64 `json:"masking_server_us"`
	QueryClientMS           []int64 `json:"query_client_ms"`
}

// ServerA is the ID (index) of the first server
const ServerA int = 0

// ServerB is the ID (index) of the second server (only used for secret-sharing based API calls)
const ServerB int = 1

// Client is used to store all relevant client information
type Client struct {
	ServerAddresses []string
	ServerPorts     []string
	SessionParams   *api.SessionParameters

	// all timing information collected during protocol execution
	Experiment *RuntimeExperiment
}

// WaitForExperimentStart completes once the servers are ready
// to start the experiment
func (client *Client) WaitForExperimentStart() {
	args := api.WaitForExperimentArgs{}
	res := api.WaitForExperimentResponse{}

	// wait for server A
	if !client.call(ServerA, "Server.WaitForExperiment", &args, &res) {
		panic("failed to make RPC call")
	}

	// wait for server B
	if !client.call(ServerB, "Server.WaitForExperiment", &args, &res) {
		panic("failed to make RPC call")
	}
}

// InitSession creates a new API session with the server
func (client *Client) InitSession() {

	args := &api.InitSessionArgs{}
	res := &api.InitSessionResponse{}

	if !client.call(ServerA, "Server.InitSession", &args, &res) {
		panic("failed to make RPC call")
	}

	// TODO: don't manually copy these? Is there a cleaner way?
	client.SessionParams = &api.SessionParameters{
		SessionID:           res.SessionID,
		NumTables:           res.NumTables,
		NumProbes:           res.NumProbes,
		TestQuery:           res.TestQuery,
		HashFunctions:       res.HashFunctions,
		HashFunctionRange:   res.HashFunctionRange,
		TableBucketMetadata: res.TableBucketMetadata,
	}

	client.Experiment.NumProbes = res.NumProbes
	client.Experiment.HashFunctionRange = res.HashFunctionRange
	client.Experiment.NumTables = res.NumTables
	client.Experiment.DatasetName = res.StatsDatasetName
	client.Experiment.DatasetSize = res.StatsDatasetSize
	client.Experiment.NumFeatures = res.StatsNumFeatures
	client.Experiment.NumServerProcs = res.StatsDatasetSize

}

// PrivateANNQuery privately retrieves the values in buckets with associated keys
// keys from each table and returns the first non-zero candidate.
// keys: (NumTables, NumProbes) array keys to probe in each table
// keywordBits: size of each keyword (DPF bits)
func (client *Client) PrivateANNQuery(keys [][]uint64) int {

	var wg sync.WaitGroup

	allQueriesA := make([]*pir.BatchQueryShare, len(keys))
	allQueriesB := make([]*pir.BatchQueryShare, len(keys))

	if len(keys) != client.SessionParams.NumTables || len(keys[0]) != client.SessionParams.NumProbes {
		panic("keys should have shape (NumTables, NumProbes)")
	}

	for i := 0; i < client.SessionParams.NumTables; i++ {
		wg.Add(1)
		go func(tableIndex int) {
			defer wg.Done()

			// one query per "probe" in the ith table
			qA := make([]*pir.QueryShare, len(keys[tableIndex]))
			qB := make([]*pir.QueryShare, len(keys[tableIndex]))

			bucketDbmd := client.SessionParams.TableBucketMetadata[tableIndex]
			for _, k := range keys[tableIndex] {
				q := bucketDbmd.NewKeywordQueryShares(k, 2, uint(client.SessionParams.HashFunctionRange))
				qA[k] = q[0]
				qB[k] = q[1]
			}

			// batch query for server A
			batchQueryA := &pir.BatchQueryShare{}
			batchQueryB := &pir.BatchQueryShare{}
			batchQueryA.Queries = qA

			// batch query for server B
			batchQueryB.Queries = qB

			allQueriesA[tableIndex] = batchQueryA
			allQueriesB[tableIndex] = batchQueryB

		}(i)
	}

	// wait until all batch queries are constructed
	wg.Wait()

	// RPC both servers (in parallel)
	argsA := &api.ANNQueryArgs{}
	argsA.SessionID = client.SessionParams.SessionID
	argsA.SecretShared = allQueriesA

	argsB := &api.ANNQueryArgs{}
	argsB.SessionID = client.SessionParams.SessionID
	argsB.SecretShared = allQueriesA

	resA := &api.ANNQueryResponse{}
	resB := &api.ANNQueryResponse{}

	wg.Add(2)
	go func() {
		defer wg.Done()
		if !client.call(ServerA, "Server.PrivateANNQuery", &argsA, &resA) {
			panic("failed to make RPC call")
		}
	}()

	go func() {
		defer wg.Done()
		if !client.call(ServerB, "Server.PrivateANNQuery", &argsB, &resB) {
			panic("failed to make RPC call")
		}
	}()

	wg.Wait()

	// final candidate set (obliviously masked by the servers)
	candidate := 0

	// recover each slot and covert to a value (ID)
	total := client.SessionParams.NumTables * client.SessionParams.NumProbes
	for i := 0; i < total; i++ {
		shareA := resA.ResSecretShared[i]
		shareB := resB.ResSecretShared[i]
		val := int(pir.Recover([]*pir.SecretSharedQueryResult{shareA, shareB}))
		if val != 0 {
			candidate = val
			break
		}
	}

	// update the experiment statistics
	totalUploadBytes := getSizeInBytes(argsA) + getSizeInBytes(argsB)
	totalDownloadBytes := getSizeInBytes(resA) + getSizeInBytes(resB)
	servQuery := resA.StatsQueryTimeInMS
	servMasking := resA.StatsMaskingTimeInUS
	client.Experiment.QueryUpBandwidthBytes = append(client.Experiment.QueryUpBandwidthBytes, totalUploadBytes)
	client.Experiment.QueryDownBandwidthBytes = append(client.Experiment.QueryDownBandwidthBytes, totalDownloadBytes)
	client.Experiment.QueryServerMS = append(client.Experiment.QueryServerMS, servQuery)
	client.Experiment.QueryMaskingServerUS = append(client.Experiment.QueryMaskingServerUS, servMasking)

	return candidate
}

// TerminateSessions ends the client session on both servers
func (client *Client) TerminateSessions() {
	args := api.TerminateSessionArgs{}
	res := api.TerminateSessionResponse{}

	// kill server A
	if !client.call(ServerA, "Server.TerminateSession", &args, &res) {
		panic("failed to make RPC call")
	}

	// kill server B
	if !client.call(ServerB, "Server.TerminateSession", &args, &res) {
		panic("failed to make RPC call")
	}
}

// send an RPC request to the master, wait for the response
func (client *Client) call(serverID int, rpcname string, args interface{}, reply interface{}) bool {

	cli, err := rpc.DialHTTP("tcp", client.ServerAddresses[serverID]+":"+client.ServerPorts[serverID])
	if err != nil {
		log.Fatal("dialing:", err)
	}

	defer cli.Close()

	err = cli.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	log.Printf("[Client]: failed to call RPC with error %v", err)

	return false
}

func getSizeInBytes(s interface{}) int64 {
	var b bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&b) // Will write to network.
	err := enc.Encode(s)
	if err != nil {
		panic(err)
	}

	return int64(len(b.Bytes()))
}
