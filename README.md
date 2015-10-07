seomoz
======

SEOmoz golang client

# Usage

```go
seomoz := seomoz.NewEnvClient()
metrics, err := seomoz.GetURLMetrics(queryURL, cols)
if err != nil {
    log.Fatalln(err)
}
fmt.Printf("URL: %s\nLinks: %.0f\nPage Authority: %.0f\nDomain Authority: %.0f\n", metrics.URL, metrics.Links, metrics.PageAuthority, metrics.DomainAuthority)
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
