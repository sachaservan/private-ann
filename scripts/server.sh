#!/bin/bash

usage() { 
    echo "Usage: $0 
    [--sid <0|1>] 
    [--dataset <dataset name>] 
    [--cachedir <cache directory>] 
    [--numtables <num tables>] 
    [--numprobes <num probes>] 
    [--bucketcap <max bucket size>]
    [--maxval <max coordinate value>] 
    [--pwm <projection width mean>] 
    [--pws <projection width std]
    [--procs <degree of parallelism>]"
    echo "Example: 
    bash server.sh --sid 0 --dataset mnist --cachedir cache --numtables 10 --numprobes 100 --bucketcap 1 --maxval 1000 --pwm 887.7 --pws 244.9 --procs 1"
    1>&2; exit 1; 
}

	
POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --sid)
    SERVID="$2"
    shift # past argument
    shift # past value
    ;;
    --dataset)
    DATASET="$2"
    shift # past argument
    shift # past value
    ;;
    --cachedir)
    CACHEDIR="$2"
    shift # past argument
    shift # past value
    ;;
    --numtables)
    NUMTABLES="$2"
    shift # past argument
    shift # past value
    ;;
    --numprobes)
    NUMPROBES="$2"
    shift # past argument
    shift # past value
    ;;
    --bucketcap)
    BUCKETCAP="$2"
    shift # past argument
    shift # past value
    ;;
    --maxval)
    MAXVAL="$2"
    shift # past argument
    shift # past value
    ;;
    --pwm)
    PWMEAN="$2"
    shift # past argument
    shift # past value
    ;;
    --pws)
    PWSTD="$2"
    shift # past argument
    shift # past value
    ;;
    --procs)
    PROCS="$2"
    shift # past argument
    shift # past value
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done

set -- "${POSITIONAL[@]}" # restore positional parameters
shift $((OPTIND-1))

if  [ -z "${SERVID}" ] || 
    [ -z "${DATASET}" ] || 
    [ -z "${CACHEDIR}" ] || 
    [ -z "${NUMTABLES}" ] || 
    [ -z "${NUMPROBES}" ] || 
    [ -z "${BUCKETCAP}" ] || 
    [ -z "${MAXVAL}" ] || 
    [ -z "${PWMEAN}" ] || 
    [ -z "${PWSTD}" ] || 
    [ -z "${PROCS}" ]; then
    usage
    exit
fi

echo 'Server ID:  ' ${SERVID}
echo 'Dataset:    ' ${DATASET}
echo 'Cache dir:  ' ${CACHEDIR}
echo 'Num Tables: ' ${NUMTABLES}
echo 'Num Probes: ' ${NUMPROBES}
echo 'Bucket cap: ' ${BUCKETCAP}
echo 'Max value:  ' ${MAXVAL}
echo 'Proj. Mean: ' ${PWMEAN}
echo 'Proj. Std:  ' ${PWSTD}
echo 'Num Procs:  ' ${PROCS}

# make sure cache directory exists 
mkdir -p ${CACHEDIR}

# build the server 
go build -o ../bin/server ../cmd/server/main.go 
../bin/server \
    --serverid ${SERVID} \
    --dataset ${DATASET} \
    --cachedir ${CACHEDIR} \
    --numtables ${NUMTABLES} \
    --numprobes ${NUMPROBES} \
    --projectionwidthmean ${PWMEAN} \
    --projectionwidthstddev ${PWSTD} \
    --maxcoordinatevalue ${MAXVAL} \
    --numprocs ${PROCS} \
    --bucketsize ${BUCKETCAP} \


