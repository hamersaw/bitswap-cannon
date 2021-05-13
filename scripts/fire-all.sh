#!/bin/bash

usage="USAGE: $(basename $0) [OPTIONS...]
Generalized script to execute bitswap-cannon on a localized
cluster over all combinations of seeders and leechers.

OPTIONS:
    -c <count>          number of nodes in ipfs cluster [default: 8]
    -d <portdelta>      delta for each ipfs node port [default: 100]
    -f <filename>       filename to test
    -h                  display this help menu
    -p <port>           starting ipfs node port [default: 5001]"

# identify project directory and configuration file locations
scriptdir="$(dirname $0)"
case $scriptdir in
    /*) 
        projectdir="$scriptdir/.."
        ;;
    *) 
        projectdir="$(pwd)/$scriptdir/.."
        ;;
esac

# parse opts
nodecount=8
portdelta=100
port=5001
while getopts "c:d:f:hp:" opt; do
    case $opt in
        c)
            nodecount=$OPTARG
            ;;
        d)
            portdelta=$OPTARG
            ;;
        f)
            filename=$OPTARG
            ;;
        h)
            echo "$usage"
            exit 0
            ;;
        p)
            port=$OPTARG
            ;;
        ?)
            echo "$usage"
            exit 1
            ;;
    esac
done

# test if filename is empty
[ -z "$filename" ] && echo "$usage" && exit 1

# iterate over all combinations of seeders and leechers
maxseeders=$(( nodecount - 1 ))
for seedercount in $(seq 1 $maxseeders); do
    maxleechers=$(( nodecount - seedercount ))
    for leechercount in $(seq 1 $maxleechers); do
        currentport=$port
        hostcount=0

        # create seeder, leecher, and unallocated args
        seederargs=""
        for i in $(seq 1 $seedercount); do
            seederargs+=" -s localhost:$currentport"

            currentport=$(( currentport + portdelta ))
            hostcount=$(( hostcount + 1 ))
        done

        leecherargs=""
        for i in $(seq 1 $leechercount); do
            leecherargs+=" -l localhost:$currentport"

            currentport=$(( currentport + portdelta ))
            hostcount=$(( hostcount + 1 ))
        done

        unallocatedargs=""
        while (( $hostcount < $nodecount )); do
            unallocatedargs+=" -u localhost:$currentport"

            currentport=$(( currentport + portdelta ))
            hostcount=$(( hostcount + 1 ))
        done

        # execute benchmark
        echo "seeder(s):$seedercount leecher(s):$leechercount"
        $projectdir/bin/bitswap-cannon \
            $seederargs $leecherargs $unallocatedargs -f $filename \
            > $seedercount-$leechercount.json

        sleep 8
    done
done
