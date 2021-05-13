# bitswap-cannon
Benchmarking utility for the bitswap protocol in IPFS.

## usage
#### installation
    # build bitswap-cannon binary
    make
#### examples
    # execute on the file macav2-50k.csv with one seeder and one leecher
    ./bin/bitswap-cannon -s localhost:5001 -l localhost:5101 \
        -f ~/downloads/macav2-50k.csv

    # specify multiple seeders
    ./bin/bitswap-cannon -s localhost:5001 -s localhost:5101 \
        -s localhost:5201 -l localhost:5301 -f ~/downloads/macav2-50k.csv

    # specify multiple leechers
    ./bin/bitswap-cannon -s localhost:5001 -l localhost:5101 \
        -l localhost:5201 -l localhost:5301 -f ~/downloads/macav2-50k.csv

    # specify multiple seeders and leechers
    ./bin/bitswap-cannon -s localhost:5001 -s localhost:5101 \
        -l localhost:5201 -l localhost:5301 -f ~/downloads/macav2-50k.csv
#### scripts
    # use fire-all.sh script 
    ./scripts/fire-all.sh -f ~/downloads/macav2-1m.csv

## results
All results are evaluated on a cluster of 8 IPFS nodes using go-ipfs v0.8.0
#### latency
Profiling average latency of leecher(s) retrieving files from a collection of seeder(s). Data is presented in tables where columns are the number of leechers and rows are the number of seeders.

    # print latency (in seconds) of experiement
    cat 1-1.json | jq '.[] | select(.Type=="Leecher") | .DurationNs' | awk '{ sum += $1 / 1000000000; n++ } END { if (n > 0) print sum / n; }'
##### 63MB filesize - 32mbit rate / 5ms latency
|   | 1      | 2      | 3      | 4      | 5      | 6      | 7      |
| 1 | 12.905 | 18.763 | 26.340 | 34.709 | 42.653 | 49.013 | 58.438 |
| 2 | 9.7327 | 13.847 | 18.343 | 22.241 | 25.244 | 30.397 |        |
| 3 | 9.6584 | 10.391 | 12.498 | 15.627 | 18.875 |        |        |
| 4 | 9.3676 | 10.465 | 11.118 | 14.742 |        |        |        |
| 5 | 9.4887 | 9.6605 | 10.168 |        |        |        |        |
| 6 | 9.4589 | 10.125 |        |        |        |        |        |
| 7 | 9.7036 |        |        |        |        |        |        |
    

## todo
- clean up error reporting
