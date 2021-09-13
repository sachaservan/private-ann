package server

import (
	"log"

	"github.com/sachaservan/private-ann/cmd/api"
	"github.com/sachaservan/private-ann/pir"
)

// ClientSession is a KNN query issued by a client over multiple rounds
// and keeps the state until the client is done
type ClientSession struct {
	SessionID int64
}

// InitSession initializes a new KNN query session for the client
func (server *Server) InitSession(args api.InitSessionArgs, reply *api.InitSessionResponse) error {

	log.Printf("[Server]: received request to InitSession")

	// session ID
	sessionID := int64(0)

	dbmd := make([]*pir.DBMetadata, len(server.TableDBs))
	for i := 0; i < len(server.TableDBs); i++ {
		dbmd[i] = &server.TableDBs[i].DBMetadata
	}

	reply.SessionID = sessionID
	reply.HashFunctions = server.HashFunctions
	reply.TableBucketMetadata = dbmd
	reply.NumProbes = server.NumProbes
	reply.NumTables = server.NumTables
	reply.TestQuery = server.TestQuery
	reply.StatsDatasetName = server.DatasetName
	reply.StatsDatasetSize = server.DBSize
	reply.StatsPreprocessingTimeInMS = server.StatsTotalPreprocessingTime
	reply.StatsNumFeatures = server.StatsDatasetNumFeatures
	reply.StatsNumServerProcs = server.NumProcs

	return nil
}

// TerminateSession kills the server
func (server *Server) TerminateSession(args *api.TerminateSessionArgs, reply *api.TerminateSessionResponse) error {
	server.Killed = true
	return nil
}
