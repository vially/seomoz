seomoz
======

SEOmoz golang client

# Usage

```go
package main

import (
	"fmt"
	"github.com/vially/seomoz"
	"log"
)

func main() {
    client := seomoz.NewEnvClient()
    metrics, err := client.GetURLMetrics(queryURL, seomoz.DefaultCols)
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Printf("Page Authority: %.0f\n", metrics.PageAuthority)
}
```

# Command Line Tool

```
$ seomoz wikipedia.org
URL: wikipedia.org/
Links: 1064773
Page Authority: 94
Domain Authority: 100
```

# License

The MIT License (MIT)
