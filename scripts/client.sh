#!/bin/bash

# Command line arguments to run the client: 
# ServerAddrs: array of server IP addresses; if only one addr provided, assumes single server setting 
# ServerPorts: array of server ports 
# AutoCloseClient: if YES then kills the client once all requests havve completed 
# ExperimentNumTrials: number of times to run each experiment 

# make sure the results directory exists
mkdir -p ../results

# build the client 
go build -o ../bin/client ../cmd/client/main.go 

# configure arguments 
ServerAddrs=("localhost" "localhost")
ServerPorts=(8000 8001)
AutoCloseClient=true
ExperimentNumTrials=10
ExperimentSaveFile="../results/experiment${RANDOM}${RANDOM}.json"

# add the boolean flags
boolargs=()
if [ "$AutoCloseClient" = true ]; then 
    boolargs+=('--autocloseclient')
fi 

# run the experiemnts with the specified parameters 
../bin/client \
    --experimentsavefile ${ExperimentSaveFile} \
    --serveraddrs ${ServerAddrs[@]} \
    --serverports ${ServerPorts[@]} \
    --experimentnumtrials ${ExperimentNumTrials} \
    ${boolargs[@]} \

