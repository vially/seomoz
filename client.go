package seomoz

import (
	"math"
	"os"
	"sync"
)

var (
	// DefaultCols is the sum of the numeric values for columns: Page Authority, Domain Authority, Links and Canonical URL
	// See https://moz.com/help/guides/moz-api/mozscape/api-reference/url-metrics for more info
	DefaultCols = 103079217156

	defaultMaxBatchURLs = 10
)

// URLMetrics is the basic structure returned by the API
type URLMetrics struct {
	PageAuthority   float64 `json:"upa"`
	DomainAuthority float64 `json:"pda"`
	Links           float64 `json:"uid"`
	URL             string  `json:"uu"`
}

type mozAPI interface {
	GetURLMetrics(link string, cols int) (*URLMetrics, error)
	GetBatchURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error)
}

// Client is the main object used to interact with the SeoMoz API
type Client struct {
	moz          mozAPI
	MaxBatchURLs int
}

// NewEnvClient instantiates a new Client configured from the environment variables SEOMOZ_ACCESS_ID and SEOMOZ_SECRET_KEY
func NewEnvClient() *Client {
	return NewClient(os.Getenv("SEOMOZ_ACCESS_ID"), os.Getenv("SEOMOZ_SECRET_KEY"))
}

// NewClient instantiates a new Client
func NewClient(accessID, secretKey string) *Client {
	return &Client{moz: &MOZ{AccessID: accessID, SecretKey: secretKey}}
}

// GetURLMetrics fetches the metrics for the given urls
func (s *Client) GetURLMetrics(link string, cols int) (*URLMetrics, error) {
	return s.moz.GetURLMetrics(link, cols)
}

// GetBatchURLMetrics executes a batch API call. At most 10 URLs are allowed in a batch call.
func (s *Client) GetBatchURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error) {
	return s.moz.GetBatchURLMetrics(urls, cols)
}

// GetBulkURLMetrics executes as many batch calls as required in order to analyze all the URLs.
func (s *Client) GetBulkURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error) {
	maxBatchURLs := s.MaxBatchURLs
	if maxBatchURLs == 0 {
		maxBatchURLs = defaultMaxBatchURLs
	}

	batchesNo := int(math.Ceil(float64(len(urls)) / float64(maxBatchURLs)))
	responses := make(chan batchResponse, batchesNo)

	var wg sync.WaitGroup
	for i := 0; i < len(urls); i += maxBatchURLs {
		var batchURLs []string
		if i+maxBatchURLs >= len(urls) {
			batchURLs = urls[i:]
		} else {
			batchURLs = urls[i : i+maxBatchURLs]
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			batch, err := s.moz.GetBatchURLMetrics(batchURLs, cols)
			responses <- batchResponse{batch, err}
		}()
	}

	wg.Wait()
	close(responses)

	results := map[string]*URLMetrics{}
	for resp := range responses {
		if resp.err != nil {
			return nil, resp.err
		}

		for link, metrics := range resp.batch {
			results[link] = metrics
		}
	}

	return results, nil
}

type batchResponse struct {
	batch map[string]*URLMetrics
	err   error
}
