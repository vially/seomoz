package main

import (
	"fmt"
	"github.com/vially/seomoz"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: seomoz URL [COLS]")
		os.Exit(1)
	}

	queryURL := os.Args[1]
	cols := seomoz.DefaultCols
	if len(os.Args) > 2 {
		if columns, err := strconv.Atoi(os.Args[2]); err != nil {
			log.Fatalln("Invalid COLS value: " + os.Args[2])
		} else {
			cols = columns
		}
	}

	seomoz := seomoz.NewEnvClient()
	metrics, err := seomoz.GetURLMetrics(queryURL, cols)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("URL: %s\nLinks: %.0f\nPage Authority: %.0f\nDomain Authority: %.0f\n", metrics.URL, metrics.Links, metrics.PageAuthority, metrics.DomainAuthority)
}
