# bitswap-cannon

## usage
#### installation
    # build bitswap-cannon binary
    make
#### examples
    # execute on the file macav2-50k.csv with one seeder and one leecher
    ./bin/bitswap-cannon -s localhost:5001 -l localhost:5101 -f ~/downloads/macav2-50k.csv

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

## todo
- clean up error reporting
