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
    bufferSize := flag.Int("b", 4096, "size of reader buffer (bytes)")
    filename := flag.String("f", "", "filename to process")

    var leechers, seeders, unallocateds FlagList
    flag.Var(&leechers, "l", "host to leech file(s)")
    flag.Var(&seeders, "s", "host to seed file(s)")
    flag.Var(&unallocateds, "u", "unallocated host(s) (to include in output)")

    flag.Parse()

    // validate arguments
    if *filename == "" {
        fmt.Println("Must specify filename '-f value'")
        flag.Usage()
        os.Exit(1)
    } else if len(leechers) == 0 {
        fmt.Println("Must specify at least one leecher with '-l value'")
        flag.Usage()
        os.Exit(1)
    } else if len(seeders) == 0 {
        fmt.Println("Must specify at least one seeder with '-l value'")
        flag.Usage()
        os.Exit(1)
    }

    hostAddrsMap := map[string][]string{
        "Leecher": leechers,
        "Seeder": seeders,
        "Unallocated": unallocateds,
    }

    // add file to seeders
    var cid *string
    for _, seeder := range seeders {
        sh := shell.NewShell(seeder)

        // open file reader
        f, err := os.Open(*filename)
        if err != nil {
            fmt.Println("failed to open file:", err)
            continue
        }

        // add file to seeder
        fileCid, err := sh.Add(f)
        if err != nil {
            fmt.Println("failed to add file:", err)
            continue
        }

        // capture file cid
        cid = &fileCid
    }

    // sleep to wait for information to settle
    time.Sleep(3 * time.Second)

    // cat file (to nowhere) from each leecher
    ch := make(chan HostDuration)

    // iterate over leechers
    for _, leecher := range leechers {
        // start new go routine to execute in parallel
        go func(bufferSize int, cid string, leecher string) {
            // initialize leecher HTTP API shell
            sh := shell.NewShell(leecher)
            start := time.Now()

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
        }(*bufferSize, *cid, leecher)
    }

    // wait for all retrievals to complete
    leecherDurations := make(map[string]time.Duration)
    for i := 0; i < len(leechers); i++ {
        leecherDuration := <-ch
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

            // if seeder -> unpin cid
            if hostType == "Seeder" {
                err := sh.Unpin(*cid)
                if err != nil {
                    fmt.Println("failed to unpin cid:", err)
                    continue
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
