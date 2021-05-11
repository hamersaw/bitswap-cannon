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

type flagList []string

func (i *flagList) Set(value string) error {
    *i = append(*i, value)
    return nil
}

func (i *flagList) String() string {
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

    var filenames, leechers, seeders, unallocateds flagList
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

    // add file(s) to seeders
    cids := make([]string, len(filenames))
    for _, seeder := range seeders {
        sh := shell.NewShell(seeder)

        for fileIndex, filename := range filenames {
            f, err := os.Open(filename)
            if err != nil {
                fmt.Println("failed to open file:", err)
                continue
            }

            cid, err := sh.Add(f)
            if err != nil {
                fmt.Println("failed to add file:", err)
                continue
            }

            cids[fileIndex] = cid
        }
    }

    // cat file(s) (to nowhere) from each leecher
    ch := make(chan HostDuration)
    for _, cid := range cids {
        for _, leecher := range leechers {
            go func(cid string, leecher string) {
                sh := shell.NewShell(leecher)
                start := time.Now()

                r, err := sh.Cat(cid)
                if err != nil {
                    fmt.Println("failed to read cid:", err)
                    return
                }

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

                elapsed := time.Since(start)
                ch <- HostDuration{leecher, elapsed}

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

        // TODO - max of Host? (will receive one for each cid)
        leecherDurations[leecherDuration.Host] = leecherDuration.Duration
    }

    // output stats
    hosts := make([]Host, len(leechers) + len(seeders) + len(unallocateds))
    hostsIndex := 0

    hostAddrsMap := map[string][]string{
        "Leecher": leechers,
        "Seeder": seeders,
        "Unallocated": unallocateds,
    }

    for hostType, hostAddrsList := range hostAddrsMap {
        for _, hostAddr := range hostAddrsList {
            sh := shell.NewShell(hostAddr)

            var bitswapStat BitswapStat
            err := sh.Request("bitswap/stat").
                Exec(context.Background(), &bitswapStat)
            if err != nil {
                fmt.Println("failed to retrieve stats:", err)
                continue
            }

            var t time.Duration
            leecherDuration, exists := leecherDurations[hostAddr]
            if exists {
                t = leecherDuration
            }

            hosts[hostsIndex] = Host{hostType, hostAddr, t, bitswapStat}
            hostsIndex += 1
        }
    }

    hostsJSON, err := json.MarshalIndent(hosts, "", "  ")
    if err != nil {
        fmt.Println("failed to marshal hosts as json:", err)
    }

    fmt.Println(string(hostsJSON))

    // TODO - cleanup
}
