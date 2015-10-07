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
	cols := 103079217156
	if len(os.Args) > 2 {
		if columns, err := strconv.Atoi(os.Args[2]); err != nil {
			log.Fatalln("Invalid COLS value: " + os.Args[2])
		} else {
			cols = columns
		}
	}

	seomoz := seomoz.NewEnvClient()
	m, err := seomoz.GetURLMetrics(queryURL, cols)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%s\t%.0f\t%.0f\t%.0f\n", m.URL, m.Links, m.PageAuthority, m.DomainAuthority)
}
