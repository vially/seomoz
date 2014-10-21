package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/vially/seomoz"
)

func parseFlags() (url string, cols int) {
	if len(os.Args) < 2 {
		log.Fatal("URL not specified")
	}

	url = os.Args[1]
	if len(os.Args) > 2 {
		cols, _ = strconv.Atoi(os.Args[2])
	} else {
		cols = 103079217156
	}
	return
}

func main() {
	godotenv.Load()
	queryURL, cols := parseFlags()
	accessID := os.Getenv("SEOMOZ_ACCESS_ID")
	secretKey := os.Getenv("SEOMOZ_SECRET_KEY")
	seomoz := seomoz.NewClient(accessID, secretKey)
	metrics, err := seomoz.GetURLMetrics([]string{queryURL, "example.com"}, cols)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range metrics {
		fmt.Printf("%s\t%.0f\t%.0f\t%.0f\n", m.URL, m.Links, m.PageAuthority, m.DomainAuthority)
	}
}
