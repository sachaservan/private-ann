package main

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/sachaservan/private-ann/ann"
	"github.com/sachaservan/private-ann/client"
	"github.com/sachaservan/private-ann/hash"
)

// command-line arguments to run the server
var args struct {
	ServerAddrs         []string
	ServerPorts         []string
	SecurityBits        int    `default:"1024"`  // e.g., 1024 RSA security; 128 for secret-sharing security
	SingleServer        bool   `default:"false"` // use single server encrypted cPIR
	ExperimentNumTrials int    `default:"1"`     // number of times to run this experiment configuration
	ExperimentSaveFile  string `default:"output.json"`
	EvaluateProfileHash bool   `default:"false"` // run client server protocol to compute hash of client's profile
	EvaluatePrivateANN  bool   `default:"false"` // run ANN search protocol
	AutoCloseClient     bool   `default:"true"`  // close client when done
}

func main() {

	gob.Register(&hash.MultiLatticeHash{})

	arg.MustParse(&args)

	cli := &client.Client{}
	cli.ServerAddresses = args.ServerAddrs
	cli.ServerPorts = args.ServerPorts
	cli.Experiment = &client.RuntimeExperiment{}

	// init experiment
	cli.Experiment.QueryClientMS = make([]int64, 0)
	cli.Experiment.QueryServerMS = make([]int64, 0)
	cli.Experiment.QueryMaskingServerUS = make([]int64, 0)
	cli.Experiment.QueryUpBandwidthBytes = make([]int64, 0)
	cli.Experiment.QueryDownBandwidthBytes = make([]int64, 0)

	for i := 0; i < args.ExperimentNumTrials; i++ {

		log.Printf("[Client]: waiting for servers to initialize \n")

		// wait for the servers to finish initializing
		cli.WaitForExperimentStart()

		log.Printf("[Client]: starting experiment \n")

		start := time.Now()

		// Step 1: Initialize the session (returns hash functions and test queries)
		log.Printf("[Client]: initializing session \n")
		cli.InitSession()
		log.Printf("[Client]: session initialized (SID = %v) in %v seconds\n", cli.SessionParams.SessionID, time.Since(start).Seconds())

		// Step 2: Compute hash for the test query
		start = time.Now()
		q := cli.SessionParams.TestQuery

		keys := make([][]uint64, cli.SessionParams.NumTables)
		for i := range keys {
			// Returns numProbes values inserted into numPartition buckets
			// 0 is the value in slots without hashes
			// Here numPartitions = numProbes
			keys[i] = ann.ComputeProbes(cli.SessionParams.HashFunctions[i], q, cli.SessionParams.NumProbes, cli.SessionParams.NumProbes)
		}

		// Step 3: query the buckets using PIR
		log.Printf("[Client]: querying %v buckets in %v tables\n",
			cli.SessionParams.NumTables*cli.SessionParams.NumProbes,
			cli.SessionParams.NumTables)

		cli.PrivateANNQuery(keys)

		queryTime := time.Since(start).Milliseconds()
		cli.Experiment.QueryClientMS = append(cli.Experiment.QueryClientMS, queryTime)
		log.Printf("[Client]: ANN query took %v seconds\n", time.Since(start).Seconds())

		// Experiment completed
		log.Printf("[Client]: finished experiment trial %v of %v \n", i+1, args.ExperimentNumTrials)
	}

	// write the result of the evalaution to the specified file
	experimentJSON, _ := json.MarshalIndent(cli.Experiment, "", " ")
	ioutil.WriteFile(args.ExperimentSaveFile, experimentJSON, 0644)

	// prevent client from closing until user input
	if !args.AutoCloseClient {
		log.Printf("[Client]: press enter to close")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
	}

	// terminate the client's session on the server
	cli.TerminateSessions()
}
