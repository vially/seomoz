package seomoz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	defaultApiURL      = "http://lsapi.seomoz.com/linkscape/url-metrics/"
	defaultBodyHandler = readAllCloser
	DefaultHTTPHandler = http.DefaultClient.Do
)

// MOZ is the low lever http client used to interact with SEOmoz API
type MOZ struct {
	AccessID  string
	SecretKey string
}

func (s *MOZ) signature(expires int64) string {
	message := fmt.Sprintf("%s\n%d", s.AccessID, expires)
	return computeHmac(message, s.SecretKey)
}

func (s *MOZ) rawQuery(cols int) string {
	expires := time.Now().Unix() + 300
	v := url.Values{}
	v.Set("AccessID", s.AccessID)
	v.Set("Expires", strconv.FormatInt(expires, 10))
	v.Set("Signature", s.signature(expires))
	v.Set("Cols", strconv.Itoa(cols))
	return v.Encode()
}

func (s *MOZ) buildGetRequest(link string, params string) *http.Request {
	apiURL, _ := url.Parse(fmt.Sprintf("%s%s", defaultApiURL, url.QueryEscape(link)))
	apiURL.RawQuery = params
	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	return req
}

func (s *MOZ) buildPostRequest(urls []string, params string) *http.Request {
	urlsJSON, _ := json.Marshal(urls)
	apiURL, _ := url.Parse(defaultApiURL)
	apiURL.RawQuery = params
	req, _ := http.NewRequest("POST", apiURL.String(), bytes.NewReader(urlsJSON))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (s *MOZ) unmarshalSingleResponse(resp *http.Response) (*URLMetrics, error) {
	content, err := defaultBodyHandler(resp.Body)
	if err != nil {
		return nil, err
	}

	metric := &URLMetrics{}
	err = json.Unmarshal(content, &metric)
	return metric, err
}

func (s *MOZ) unmarshalBatchResponse(urls []string, resp *http.Response) (map[string]*URLMetrics, error) {
	content, err := defaultBodyHandler(resp.Body)
	if err != nil {
		return nil, err
	}

	var metrics []*URLMetrics
	if err = json.Unmarshal(content, &metrics); err != nil {
		return nil, err
	}

	if len(metrics) != len(urls) {
		return nil, errors.New("Invalid response: mismatch between number of urls requested and data received")
	}

	out := make(map[string]*URLMetrics)
	for i, m := range metrics {
		if m.URL == "" {
			if requestedURL, err := url.Parse(urls[i]); err != nil {
				m.URL = urls[i]
			} else {
				m.URL = fmt.Sprintf("%s%s", requestedURL.Host, requestedURL.RequestURI())
			}
		}
		out[urls[i]] = m
	}

	return out, nil
}

// GetURLMetrics fetches the metrics for the given urls
func (s *MOZ) GetURLMetrics(link string, cols int) (*URLMetrics, error) {
	resp, err := DefaultHTTPHandler(s.buildGetRequest(link, s.rawQuery(cols)))
	if err != nil {
		return nil, err
	}
	return s.unmarshalSingleResponse(resp)
}

// GetBatchURLMetrics executes a batch API call. At most 10 URLs are allowed in a batch call.
func (s *MOZ) GetBatchURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error) {
	resp, err := DefaultHTTPHandler(s.buildPostRequest(urls, s.rawQuery(cols)))
	if err != nil {
		return nil, err
	}
	return s.unmarshalBatchResponse(urls, resp)
}
