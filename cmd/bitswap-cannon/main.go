package main

import (
    "flag"
    "fmt"
    "strings"
    "os"
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

    fmt.Println("filenmaes:", filenames)
    fmt.Println("leechers:", leechers)
    fmt.Println("seeders:", seeders)

    // TODO - load file(s) into seeders

    // TODO - add file(s) to bitswap ledger want on leechers

    // TODO - wait to complete

    // TODO - output stats
}
