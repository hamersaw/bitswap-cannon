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
All results are evaluated on a cluster of 8 IPFS nodes using go-ipfs v0.8.0. Data is presented in tables where columns are the number of leechers and rows are the number of seeders.
#### bandwidth overhead
Profiling cumulative bandwidth (sent and received) used by bitswap at seeders and leechers. All experiments were performed on a 62.873MB (65927375 bytes) file.

    # average data (MB) received at leecher
    cat 1-1.json | jq '.[] | select(.Type=="Leecher") | .BitswapStat.DataReceived' | awk '{ sum += $1 / 1048576; n++ } END { if (n > 0) print sum / n; }'

        s
    l       1       2       3       4       5       6       7
        1   63.8883 89.5205 83.3473 81.2096 87.3923 91.6023 89.3049
        2   77.1391 81.8895 81.6884 85.0176 83.796  84.5656
        3   79.1393 75.1392 83.5603 87.1516 87.3256
        4   79.3895 70.5141 94.8994 87.4037
        5   77.7627 76.1453 79.7269
        6   70.8892 72.8934
        7   77.8897

    # average duplicate data (MB) received at leecher
    cat 1-1.json | jq '.[] | select(.Type=="Leecher") | .BitswapStat.DupDataReceived' | awk '{ sum += $1 / 1048576; n++ } END { if (n > 0) print sum / n; }'

        s
    l       1       2       3       4       5       6       7
        1   1.00005 25.2513 19.2923 17.0011 22.7516 27.2958 24.806
        2   13.2507 17.7511 15.9588 21.0012 19.9528 20.4627
        3   12.5007 11.2506 19.7512 23.6889 23.4824
        4   14.0007 4.87531 28.1734 23.0092
        5   11.8739 11.5007 16.1676
        6   7.25039 8.12947
        7   11.7506

    # average data (MB) sent at seeder
    cat 1-1.json | jq '.[] | select(.Type=="Seeder") | .BitswapStat.DataSent' | awk '{ sum += $1 / 1048576; n++ } END { if (n > 0) print sum / n; }'

        s
    l       1       2       3       4       5       6       7
        1   63.8883 126.777 180.414 237.552 301.317 350.078 416.967
        2   38.5696 77.2641 113.77  148.153 174.472 214.792
        3   26.3798 49.5927 79.4727 104.352 123.065
        4   19.8474 34.8195 65.7924 79.702
        5   15.5525 30.1057 46.2337
        6   11.8149 24.0048
        7   11.1271

    # average data (MB) sent at leecher
    cat 1-1.json | jq '.[] | select(.Type=="Leecher") | .BitswapStat.DataSent' | awk '{ sum += $1 / 1048576; n++ } END { if (n > 0) print sum / n; }'

        s
    l       1       2       3       4       5       6       7
        1   0       26.1322 23.2092 21.8215 27.1288 33.256  29.7383
        2   0       4.62535 5.84156 10.9412 14.0073 12.9684
        3   0       0.75009 4.08761 8.88726 13.4865
        4   0       0.87515 7.17615 7.70169
        5   0       0.88092 2.6708
        6   0       0.87913
        7   0

#### latency
Profiling average latency (reported in seconds) of leecher(s) retrieving files from a collection of seeder(s). 

    # print latency (in seconds) of experiement in 1-1.json
    cat 1-1.json | jq '.[] | select(.Type=="Leecher") | .DurationNs' | awk '{ sum += $1 / 1000000000; n++ } END { if (n > 0) print sum / n; }'

    # latency with 63MB file / 32mbit rate / 5ms latency
        s
    l       1       2       3       4       5       6       7
        1   12.905  18.7637 26.3404 34.7096 42.653  49.0131 58.4388
        2   9.73273 13.8473 18.343  22.2418 25.2442 30.3976
        3   9.65849 10.3919 12.4982 15.6277 18.8755
        4   9.3676  10.4656 11.1182 14.7428
        5   9.4887  9.66059 10.1685
        6   9.45899 10.125
        7   9.70365

    # latency with 63MB file / 64mbit rate / 5ms latency
        s
    l       1       2       3       4       5       6       7
        1   6.13508 9.16534 12.9417 16.5825 20.6987 24.8331 28.5406
        2   4.97399 6.35153 8.48773 10.7971 12.8222 15.4159
        3   4.975   5.19275 6.72779 7.50169 9.54685
        4   4.6895  5.13987 5.30471 6.73876
        5   4.89948 4.74645 5.13683
        6   5.37335 4.76758
        7   4.78973

    # latency with 63MB file / 1024mbit rate / 5ms latency
        s
    l       1       2       3       4       5       6       7
        1   1.88801 2.0403  2.32527 2.87491 3.30684 3.54146 4.27279
        2   1.62424 1.91057 2.02899 2.42672 2.89661 3.35631
        3   1.32481 1.78345 2.257   2.37563 3.07995
        4   1.43134 1.88306 2.18223 2.75982
        5   1.18861 1.74551 3.58629
        6   1.5619  1.68282
        7   1.39399

    # latency with 127MB file / 1024mbit rate / 5ms latency
        s
    l       1       2       3       4       5       6       7
        1   4.06483 4.00246 4.79805 7.42693 6.50621 8.06386 11.7059
        2   2.51439 3.45292 4.349   5.42584 6.38147 10.1541
        3   3.73258 3.89758 4.39107 5.6271  6.63559
        4   2.54475 4.6712  4.24312 5.50618
        5   3.19705 5.3354  5.35932
        6   2.72534 3.44556
        7   2.47749
    
## todo
- clean up error reporting
