package seomoz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Client is the main object used to interact with the SeoMoz API
type Client struct {
	AccessID  string
	SecretKey string
}

// NewEnvClient instantiates a new Client configured from the environment variables SEOMOZ_ACCESS_ID and SEOMOZ_SECRET_KEY
func NewEnvClient() *Client {
	return &Client{os.Getenv("SEOMOZ_ACCESS_ID"), os.Getenv("SEOMOZ_SECRET_KEY")}
}

func (s *Client) signature(expires int64) string {
	message := fmt.Sprintf("%s\n%d", s.AccessID, expires)
	return computeHmac(message, s.SecretKey)
}

func (s *Client) rawQuery(cols int) string {
	expires := time.Now().Unix() + 300
	v := url.Values{}
	v.Set("AccessID", s.AccessID)
	v.Set("Expires", strconv.FormatInt(expires, 10))
	v.Set("Signature", s.signature(expires))
	v.Set("Cols", strconv.Itoa(cols))
	return v.Encode()
}

func (s *Client) buildRequest(urls []string, cols int) (*http.Request, error) {
	urlsCount := len(urls)
	fullAPIURL := "http://lsapi.seomoz.com/linkscape/url-metrics/"
	if urlsCount == 1 {
		fullAPIURL = fmt.Sprintf("%s/%s", fullAPIURL, url.QueryEscape(urls[0]))
	}

	apiURL, err := url.Parse(fullAPIURL)
	if err != nil {
		return nil, err
	}
	apiURL.RawQuery = s.rawQuery(cols)

	var req *http.Request
	if urlsCount == 1 {
		req, err = http.NewRequest("GET", apiURL.String(), nil)
	} else {
		urlsJSON, _ := json.Marshal(urls)
		req, err = http.NewRequest("POST", apiURL.String(), bytes.NewReader(urlsJSON))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	}
	return req, err
}

func (s *Client) unmarshalResponse(urls []string, resp *http.Response) (map[string]*URLMetrics, error) {
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var metrics []*URLMetrics

	urlsCount := len(urls)
	if urlsCount == 1 {
		metric := &URLMetrics{}
		err = json.Unmarshal(content, &metric)
		metrics = []*URLMetrics{metric}
	} else {
		err = json.Unmarshal(content, &metrics)
	}

	if err != nil {
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
func (s *Client) GetURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error) {
	req, err := s.buildRequest(urls, cols)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return s.unmarshalResponse(urls, resp)
}
