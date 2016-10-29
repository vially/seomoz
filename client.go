package seomoz

import "os"

var DefaultCols = 103079217156
var defaultMaxBatchURLs = 10

// URLMetrics is the basic structure returned by the API
type URLMetrics struct {
	PageAuthority   float64 `json:"upa"`
	DomainAuthority float64 `json:"pda"`
	Links           float64 `json:"uid"`
	URL             string  `json:"uu"`
}

type mozApi interface {
	GetURLMetrics(link string, cols int) (*URLMetrics, error)
	GetBatchURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error)
}

// Client is the main object used to interact with the SeoMoz API
type Client struct {
	moz          mozApi
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

	results := map[string]*URLMetrics{}
	for i := 0; i < len(urls); i += maxBatchURLs {
		var batchUrls []string
		if i+maxBatchURLs >= len(urls) {
			batchUrls = urls[i:]
		} else {
			batchUrls = urls[i : i+maxBatchURLs]
		}

		batch, err := s.moz.GetBatchURLMetrics(batchUrls, cols)
		if err != nil {
			return nil, err
		}

		for link, metrics := range batch {
			results[link] = metrics
		}
	}

	return results, nil
}
