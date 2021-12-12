package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/sachaservan/private-ann/ann"
	"github.com/sachaservan/private-ann/hash"
	"github.com/sachaservan/private-ann/pir"
	"github.com/sachaservan/private-ann/pir/field"
	"github.com/sachaservan/private-ann/server"
	"github.com/sachaservan/vec"
)

type CachedHashTable struct {
	Dimension int        `json:"dimension"`
	N         int        `json:"n"`
	TestQuery []float64  `json:"testQuery"`
	Keys      []uint64   `json:"keys"`
	Values    []field.FP `json:"values"`
}

type ServerArgs struct {
	ServerID              int     `default:"0"`
	Dataset               string  `default:"../datasets/mnist"`
	CacheDir              string  `default:"../cache"`
	NumTables             int     `default:"10"`
	NumProbes             int     `default:"100"`
	HashFunctionRange     int     `default:"64"`
	ProjectionWidthMean   float64 `default:"887.7"`
	ProjectionWidthStddev float64 `default:"244.9"`
	MaxCoordinateValue    int     `default:"1000"`
	NumProcs              int     `default:"40"`
	BucketSize            int     `default:"1"`
	DistanceMetric        string  `default:"euclidean"`

	// only for synthetic dataset
	DatasetSize int `default:"10000"`
	NumFeatures int `default:"50"`
}

func main() {

	// command-line arguments to the server
	var args ServerArgs

	arg.MustParse(&args)
	// log.Printf("%v", args)

	////////////////////////////////////////////////////////////////////////////
	// IMPORTANT: All servers need to have the same randomness to generate
	// consistent hash tables.
	// As a hack, we use math.Rand for randomness (THIS IS NOT SECURE!) which
	// allows us to set the seed deterministically on both servers.
	// While this works well for experiment purposes, it should never be
	// used in the wild.
	rand.Seed(0)
	////////////////////////////////////////////////////////////////////////////

	if args.BucketSize <= 0 {
		panic("bucket size must be at least 1")
	} else if args.BucketSize != 1 {
		panic("bucket size not implemented")
	}

	log.Printf("[Server]: starting server with args:\n%+v\n", args)

	// limit the number of concurrent processors that we use
	// runtime.GOMAXPROCS(args.NumProcs)

	// init the server
	serv := &server.Server{
		NumProcs:          args.NumProcs,
		Ready:             false,
		DatasetName:       filepath.Base(args.Dataset),
		NumTables:         args.NumTables,
		NumProbes:         args.NumProbes,
		CacheDir:          args.CacheDir,
		HashFunctionRange: args.HashFunctionRange,
	}

	serverPort := "8000"
	if args.ServerID == 1 {
		serverPort = "8001"
	}

	go func(serv *server.Server) {
		// hack to ensure server starts before this completes
		time.Sleep(100 * time.Millisecond)

		start := time.Now()

		tables, hashes := readOrConstructCache(serv, &args)

		serv.HashFunctions = hashes
		serv.TestQuery = vec.NewVec(tables[0].TestQuery)

		log.Printf("[Server]: number of tables = %v\n", serv.NumTables)
		log.Printf("[Server]: number of probes = %v\n", serv.NumProbes)

		// build PIR databases for each LSH table
		serv.TableDBs = make([]*pir.Database, serv.NumTables)

		for i := range serv.TableDBs {
			table := pir.NewDatabase()
			starts, stops := ann.ComputeBucketDivisions(serv.NumProbes, tables[i].Keys, tables[i].Values)

			var err error
			err = table.BuildForKeysAndValues(tables[i].Keys, tables[i].Values)
			if err != nil {
				panic(err)
			}
			err = table.SetBatchingParameters(serv.NumProbes, starts, stops)
			if err != nil {
				panic(err)
			}
			serv.TableDBs[i] = table
		}

		serv.StatsDatasetNumFeatures = serv.TestQuery.Size()
		serv.StatsTotalPreprocessingTime = time.Since(start).Milliseconds()

		log.Printf("[Server]: server is ready and waiting for client on port %v\n", serverPort)

		// limit *after* hash tables are contructed!
		runtime.GOMAXPROCS(args.NumProcs)
		serv.Ready = true
	}(serv)

	// start the server in the background
	// will set ready=true when ready to take API calls
	go killLoop(serv)
	startServer(serv, serverPort)
}

// avoid recomputing hash tables if a cached hash table already exists
func readOrConstructCache(serv *server.Server, args *ServerArgs) ([]*CachedHashTable, []hash.Hash) {
	var trainingData, testQueries []*vec.Vec
	var inputDim int
	cachedTables := make([]*CachedHashTable, serv.NumTables)

	// test if we have a cache
	cachedFilename := getCachedHashTableFilename(serv.DatasetName, serv.NumTables, serv.CacheDir, 0)
	_, err := ioutil.ReadFile(cachedFilename)

	// read the cache
	if err == nil {
		for i := range cachedTables {
			cachedFilename = getCachedHashTableFilename(serv.DatasetName, serv.NumTables, serv.CacheDir, i)
			var cached []byte
			cached, err = ioutil.ReadFile(cachedFilename)
			if err != nil {
				panic(fmt.Sprintf("error occured when loading cached hash table %v", err))
			}
			cachedTables[i] = &CachedHashTable{}
			err = json.Unmarshal(cached, cachedTables[i])
			if err != nil {
				panic(fmt.Sprintf("error occured when loading cached hash table %v", err))
			}
			log.Printf("[Server]: loaded cached table %v \n", cachedFilename)
		}
		serv.DBSize = cachedTables[0].N
		inputDim = cachedTables[0].Dimension
	}

	// otherwise load data
	if err != nil {
		log.Printf("[Server]: loading %v dataset\n", args.Dataset)
		var err2 error
		trainingData, testQueries, _, err2 = ann.ReadDataset(args.Dataset)
		if err2 != nil {
			panic(err)
		}
		serv.DBSize = len(trainingData)
		inputDim = trainingData[0].Size()
	}

	// construct hash functions
	radii := ann.GetNormalSequence2(args.ProjectionWidthMean, args.ProjectionWidthStddev, serv.NumTables)
	hashes := make([]hash.Hash, serv.NumTables)
	for i := 0; i < len(hashes); i++ {
		hashes[i] = hash.NewMultiLatticeHash(inputDim, 2, radii[i], float64(args.MaxCoordinateValue))
	}

	// construct the hash tables if we did not read from the cache
	if err != nil {
		log.Printf("[Server]: building ANN data structure for %v items\n", serv.DBSize)
		for i := range cachedTables {
			keys := make([][]uint64, serv.NumTables)
			values := make([][]field.FP, serv.NumTables)
			cachedFilename = getCachedHashTableFilename(serv.DatasetName, serv.NumTables, serv.CacheDir, i)
			// cached table does not exist
			keys[i], values[i] = ann.ComputeHashes(i, hashes[i], trainingData, uint64(serv.HashFunctionRange))
			cachedTables[i] = &CachedHashTable{
				Dimension: trainingData[0].Size(),
				N:         len(trainingData),
				TestQuery: testQueries[0].Coords,
				Keys:      keys[i],
				Values:    values[i],
			}
			// write the hash table key/values to the cache
			cachedJSON, _ := json.MarshalIndent(cachedTables[i], "", " ")
			ioutil.WriteFile(cachedFilename, cachedJSON, 0644)
			log.Printf("[Server]: cached table %v to %v\n", i, cachedFilename)
		}
	}
	return cachedTables, hashes
}

func getCachedHashTableFilename(dataset string, numTables int, basedir string, table int) string {
	return basedir + "/" + dataset + "_cached_table_" + strconv.Itoa(numTables) + "-" + strconv.Itoa(table) + ".json"
}

// kill server when Killed flag set
func killLoop(server *server.Server) {
	for !server.Killed {
		time.Sleep(100 * time.Millisecond)
	}

	server.Listener.Close()
}

func startServer(server *server.Server, port string) {

	gob.Register(&hash.MultiLatticeHash{})

	rpc.HandleHTTP()
	rpc.RegisterName("Server", server)
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	log.Println("[Server]: waiting for clients on port " + port)

	server.Listener = listener
	http.Serve(listener, nil)
}
