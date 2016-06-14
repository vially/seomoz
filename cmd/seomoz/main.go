package main

import (
	"fmt"
	"github.com/jawher/mow.cli"
	"github.com/vially/seomoz"
	"log"
	"os"
)

func main() {
	app := cli.App("seomoz", "Analyze URLs using SEOmoz")

	app.Spec = "[--cols=<SEOmoz COLS>] URL..."

	var (
		cols = app.IntOpt("c cols", seomoz.DefaultCols, "SEOmoz COLS")
		urls = app.StringsArg("URL", nil, "URLs to analyze")
	)

	app.Action = func() {
		client := seomoz.NewEnvClient()
		response, err := client.GetBulkURLMetrics(*urls, *cols)
		if err != nil {
			log.Fatalln(err)
		}

		for _, metrics := range response {
			fmt.Printf("%s\tLinks: %.0f\tPage Authority: %.0f\tDomain Authority: %.0f\n", metrics.URL, metrics.Links, metrics.PageAuthority, metrics.DomainAuthority)
		}
	}

	app.Run(os.Args)
}
