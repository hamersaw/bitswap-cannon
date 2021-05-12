package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "strings"
    "time"
    "os"

    shell "github.com/ipfs/go-ipfs-api"
)

type FlagList []string

func (i *FlagList) Set(value string) error {
    *i = append(*i, value)
    return nil
}

func (i *FlagList) String() string {
    return strings.Join(*i, ",")
}

type BitswapStat struct {
    BlocksReceived uint64
    BlocksSent uint64
    DataReceived uint64
    DataSent uint64
    DupBlksReceived uint64
    DupDataReceived uint64
    MessagesReceived uint64
}

type Host struct {
    Type string
    Addr string
    DurationNs time.Duration `json:",omitempty"`
    BitswapStat BitswapStat
}

type HostDuration struct {
    Host string
    Duration time.Duration
}

func main() {
    // parse command line arguments
    var bufferSize int
    flag.IntVar(&bufferSize, "b", 4096, "size of reader buffer (bytes)")

    var filenames, leechers, seeders, unallocateds FlagList
    flag.Var(&filenames, "f", "filename(s) to process")
    flag.Var(&leechers, "l", "host to leech file(s)")
    flag.Var(&seeders, "s", "host to seed file(s)")
    flag.Var(&unallocateds, "u", "unallocated hosts (to include in output)")

    flag.Parse()

    // validate arguments
    if len(filenames) == 0 || len(leechers) == 0 || len(seeders) == 0 {
        fmt.Println("invalid arguments") // TODO - better message
        flag.Usage()
        os.Exit(1)
    }

    hostAddrsMap := map[string][]string{
        "Leecher": leechers,
        "Seeder": seeders,
        "Unallocated": unallocateds,
    }

    // gather base statistics
    baseBitswapStats := make(map[string]BitswapStat)
    for _, hostAddrsList := range hostAddrsMap {
        for _, hostAddr := range hostAddrsList {
            // initialize host HTTP API shell
            sh := shell.NewShell(hostAddr)

            // query bitswap stats
            var bitswapStat BitswapStat
            err := sh.Request("bitswap/stat").
                Exec(context.Background(), &bitswapStat)
            if err != nil {
                fmt.Println("failed to retrieve stats:", err)
                continue
            }

            baseBitswapStats[hostAddr] = bitswapStat
        }
    }

    // add file(s) to seeders
    cids := make([]string, len(filenames))

    // iterate over seeders
    for _, seeder := range seeders {
        // initialize seeder HTTP API shell
        sh := shell.NewShell(seeder)

        // iterate over filenames
        for fileIndex, filename := range filenames {
            // open file reader
            f, err := os.Open(filename)
            if err != nil {
                fmt.Println("failed to open file:", err)
                continue
            }

            // add file to seeder
            cid, err := sh.Add(f)
            if err != nil {
                fmt.Println("failed to add file:", err)
                continue
            }

            // capture file cid
            cids[fileIndex] = cid
        }
    }

    // cat file(s) (to nowhere) from each leecher
    ch := make(chan HostDuration)

    // iterate over cids and leechers
    for _, cid := range cids {
        for _, leecher := range leechers {
            // start new go routine to execute in parallel
            go func(cid string, leecher string) {
                // initialize leecher HTTP API shell
                sh := shell.NewShell(leecher)
                start := time.Now()

                /*err := sh.Get(cid, "/dev/null")
                if err != nil {
                    fmt.Println("failed to read cid:", err)
                    return
                }*/

                // open file reader with "cat" call
                r, err := sh.Cat(cid)
                if err != nil {
                    fmt.Println("failed to read cid:", err)
                    return
                }

                // read until empty
                buf := make([]byte, bufferSize)
                for {
                    _, err := r.Read(buf)
                    if err == io.EOF {
                        break
                    } else if err != nil {
                        fmt.Println("failed to read bytes:", err)
                        continue
                    }
                }

                // return leecher host duration
                elapsed := time.Since(start)
                ch <- HostDuration{leecher, elapsed}

                // close reader
                err = r.Close()
                if err != nil {
                    fmt.Println("failed to close reader:", err)
                }
            }(cid, leecher)
        }
    }

    // wait for all retrievals to complete
    leecherDurations := make(map[string]time.Duration)
    for i := 0; i < len(cids) * len(leechers); i++ {
        leecherDuration := <- ch
        //fmt.Printf("RECV COMPLETED %d\n", i)

        // TODO - max of Host? (will receive one for each cid)
        leecherDurations[leecherDuration.Host] = leecherDuration.Duration
    }

    // output stats
    hosts := make([]Host, len(leechers) + len(seeders) + len(unallocateds))
    hostsIndex := 0

    // iterate over host types and host addresses
    for hostType, hostAddrsList := range hostAddrsMap {
        for _, hostAddr := range hostAddrsList {
            // initialize host HTTP API shell
            sh := shell.NewShell(hostAddr)

            // query bitswap stats
            var bitswapStat BitswapStat
            err := sh.Request("bitswap/stat").
                Exec(context.Background(), &bitswapStat)
            if err != nil {
                fmt.Println("failed to retrieve stats:", err)
                continue
            }

            // substract base bitswap stats
            baseBitswapStat := baseBitswapStats[hostAddr]
            bitswapStat.BlocksReceived -= baseBitswapStat.BlocksReceived
            bitswapStat.BlocksSent -= baseBitswapStat.BlocksSent
            bitswapStat.DataReceived -= baseBitswapStat.DataReceived
            bitswapStat.DataReceived -= baseBitswapStat.DataReceived
            bitswapStat.DataSent -= baseBitswapStat.DataSent
            bitswapStat.DupBlksReceived -= baseBitswapStat.DupBlksReceived
            bitswapStat.DupDataReceived -= baseBitswapStat.DupDataReceived
            bitswapStat.MessagesReceived -= baseBitswapStat.MessagesReceived

            // retrieve leech durations (only exists for leechers)
            var t time.Duration
            leecherDuration, exists := leecherDurations[hostAddr]
            if exists {
                t = leecherDuration
            }

            // add host to hosts
            hosts[hostsIndex] = Host{hostType, hostAddr, t, bitswapStat}
            hostsIndex += 1
        }
    }

    // encode hosts array as json (with indenting) and print
    hostsJSON, err := json.MarshalIndent(hosts, "", "  ")
    if err != nil {
        fmt.Println("failed to marshal hosts as json:", err)
    }

    fmt.Println(string(hostsJSON))

    // cleanup - iterate over host types and host addresses
    for hostType, hostAddrsList := range hostAddrsMap {
        for _, hostAddr := range hostAddrsList {
            // initialize host HTTP API shell
            sh := shell.NewShell(hostAddr)

            // iterate over cids
            if hostType == "Seeder" {
                for _, cid := range cids {
                    err := sh.Unpin(cid)
                    if err != nil {
                        fmt.Println("failed to unpin cid:", err)
                        continue
                    }
                }
            }

            // garbage collect repository
            err := sh.Request("repo/gc").
                Exec(context.Background(), nil)
            if err != nil {
                fmt.Println("failed to garbage collect repo:", err)
                continue
            }
        }
    }
}
