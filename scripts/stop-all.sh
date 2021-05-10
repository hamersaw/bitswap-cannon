#!/bin/bash

usage="USAGE: $(basename $0) [OPTIONS...]
OPTIONS:
    -h                  display this help menu"

# parse opts
while getopts ":h" opt; do
    case $opt in
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
while read pid; do
    kill $pid
done <<< $(ps -A -ww | grep ipfs | awk '{print $1}')
