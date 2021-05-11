package main

import (
    "flag"
    "fmt"
    "io"
    "strings"
    "sync"
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

func main() {
    // parse command line arguments
    var bufferSize int
    flag.IntVar(&bufferSize, "b", 4096, "size of reader buffer (bytes)")

    var filenames, leechers, seeders flagList
    flag.Var(&filenames, "f", "filename(s) to process")
    flag.Var(&leechers, "l", "host to leech file(s)")
    flag.Var(&seeders, "s", "host to seed file(s)")

    flag.Parse()

    // validate arguments
    if len(filenames) == 0 || len(leechers) == 0 || len(seeders) == 0 {
        fmt.Println("invalid arguments") // TODO - better message
        flag.Usage()
        os.Exit(1)
    }

    //fmt.Println("filenames:", filenames)
    //fmt.Println("leechers:", leechers)
    //fmt.Println("seeders:", seeders)

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

            fmt.Println(filename + " " + cid)
            cids[fileIndex] = cid
        }
    }

    // cat file(s) (to nowhere) from each leecher
    var wg sync.WaitGroup
    for _, leecher := range leechers {
        sh := shell.NewShell(leecher)

        for _, cid := range cids {
            wg.Add(1)
            go func() {
                defer wg.Done()
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
                fmt.Println("retrieved file in", elapsed)

                err = r.Close()
                if err != nil {
                    fmt.Println("failed to close reader:", err)
                }
            }()
        }
    }

    // wait for all retrievals to complete
    wg.Wait()

    // TODO - output stats

    // TODO - cleanup
}
