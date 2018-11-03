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

// SpamScore represents the spam score for a subdomain
// the fields of which are described at https://moz.com/help/guides/moz-api/mozscape/getting-started-with-mozscape/spam-score
// This will eventually, according to moz at least, be replaces with the same 0-100 score the browser shows.
// Until then, SpamScore is an empty interface. Partially because the columns it describes aren't documented, and also
// to allow for type switches to provide some backwards compatability in code for when Moz change their API
type SpamScore interface{}

// URLMetrics is the basic structure returned by the API
type URLMetrics struct {
	Title                                     string    `json:"ut"`
	URL                                       string    `json:"uu"`
	Subdomain                                 string    `json:"ufq"`
	RootDomain                                string    `json:"upl"`
	ExternalEquityLinks                       float64   `json:"ueid"`
	SubdomainExternalLinks                    float64   `json:"feid"`
	RootDomainExternalLinks                   float64   `json:"peid"`
	EquityLinksSubdomains                     float64   `json:"ujid"`
	SubdomainsLinking                         float64   `json:"uifq"`
	RootDomainsLinking                        float64   `json:"uipl"`
	Links                                     float64   `json:"uid"`
	SubdomainSubdomainsLinking                float64   `json:"fid"`
	RootDomainRootDomainsLinking              float64   `json:"pid"`
	MozRankURLNormalized                      float64   `json:"umrp"`
	MozRankURLRaw                             float64   `json:"umrr"`
	MozRankSubdomainNormalized                float64   `json:"fmrp"`
	MozRankSubdomainRaw                       float64   `json:"fmrr"`
	MozTrustNormalized                        float64   `json:"utrp"`
	MozTrustRaw                               float64   `json:"utrr"`
	MozTrustSubdomainNormalized               float64   `json:"ftrp"`
	MozTrustSubdomainRaw                      float64   `json:"ftrr"`
	MozTrustDomainNormalized                  float64   `json:"ptrp"`
	MozTrustDomainRaw                         float64   `json:"ptrr"`
	MozRankExternalEquityNormalized           float64   `json:"uemrp"`
	MozRankExternalEquityRaw                  float64   `json:"uemrr"`
	MozRankSubdomainExternalEquityNormalized  float64   `json:"fejp"`
	MozRankSubdomainExternalEquityRaw         float64   `json:"fejr"`
	MozRankRootDomainExternalEquityNormalized float64   `json:"pejp"`
	MozRankRootDomainExternalEquityRaw        float64   `json:"pejr"`
	MozRankSubdomainCombinedNormalized        float64   `json:"pjp"`
	MozRankSubdomainCombinedRaw               float64   `json:"pjr"`
	MozRankRootDomainCombinedNormalized       float64   `json:"fjp"`
	MozRankRootDomainCombinedRaw              float64   `json:"fjr"`
	SubdomainSpamScore                        SpamScore `json:"fspsc"`
	HTTPStatusCode                            float64   `json:"us"`
	LinksToSubdomain                          float64   `json:"fuid"`
	LinksToRootDomain                         float64   `json:"puid"`
	RootDomainsLinkingToSubdomain             float64   `json:"fipl"`
	PageAuthority                             float64   `json:"upa"`
	DomainAuthority                           float64   `json:"pda"`
	ExternalLinks                             float64   `json:"ued"`
	ExternalLinksToSubdomain                  float64   `json:"fed"`
	ExternalLinksToRootDomain                 float64   `json:"ped"`
	LinkingCBlocks                            float64   `json:"pid"`
	LastCrawlTime                             int64     `json:"ulc"`
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
