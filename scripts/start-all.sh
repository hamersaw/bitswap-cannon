#!/bin/bash

usage="USAGE: $(basename $0) [OPTIONS...]
OPTIONS:
    -c <configfile>     configuration file location
    -h                  display this help menu"

# identify project directory and configuration file locations
scriptdir="$(dirname $0)"
case $scriptdir in
    /*) 
        projectdir="$scriptdir"
        ;;
    *) 
        projectdir="$(pwd)/$scriptdir"
        ;;
esac

# parse opts
configfile="$projectdir/../configs/hosts.txt"
while getopts "c:h" opt; do
    case $opt in
        c)
            configfile=$OPTARG
            ;;
        h)
            echo "$usage"
            exit 0
            ;;
        ?)
            echo "$usage"
            exit 1
            ;;
    esac
done

# iterate over nodes
nodeid=0
while read line; do
    # parse input line
    export IPFS_PATH=$(echo $line | awk '{print $1}')
    swarmport=$(echo $line | awk '{print $2}')
    apiport=$(echo $line | awk '{print $3}')
    gatewayport=$(echo $line | awk '{print $4}')

    # ip IPFS_PATH doesn't exist -> initialize node
    if [ ! -d "$IPFS_PATH" ]; then
        echo "[node $nodeid] iniitalizing"

        # initialize ipfs node
        ipfs init       

        # configure node ports
        ipfs config Addresses.API "/ip4/127.0.0.1/tcp/$apiport"
        ipfs config Addresses.Gateway "/ip4/127.0.0.1/tcp/$gatewayport"
        ipfs config --json Addresses.Swarm "[\"/ip4/0.0.0.0/tcp/$swarmport\", \"/ip4/0.0.0.0/udp/$swarmport/quic\"]"

        # configure bootstrap information
        ipfs bootstrap rm --all
        if [[ "$nodeid" == 0 ]]; then
            # generate swam key
            swarmkey="$IPFS_PATH/swarm.key"
            echo -e "/key/swarm/psk/1.0.0/\n/base16/\n`tr -dc 'a-f0-9' < /dev/urandom | head -c64`" > $swarmkey

            # set bootstrapaddr
            peerid=$(ipfs config show | jq '.Identity.PeerID' | sed -e 's/^"//' -e 's/"$//')
            bootstrapaddr="/ip4/127.0.0.1/tcp/$swarmport/ipfs/$peerid"
        else
            cp $swarmkey $IPFS_PATH
            ipfs bootstrap add "$bootstrapaddr"
        fi
    fi

    # start ipfs node and write pid to file
    export LIBP2P_FORCE_PNET=1 && ipfs daemon &

    nodeid=$(( nodeid + 1 ))
done <$configfile
