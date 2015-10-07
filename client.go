package seomoz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var DefaultHTTPHandler = http.DefaultClient.Do
var defaultApiURL = "http://lsapi.seomoz.com/linkscape/url-metrics/"
var defaultBodyHandler = readAllCloser

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

func (s *Client) buildGetRequest(link string, params string) *http.Request {
	apiURL, _ := url.Parse(fmt.Sprintf("%s%s", defaultApiURL, url.QueryEscape(link)))
	apiURL.RawQuery = params
	req, _ := http.NewRequest("GET", apiURL.String(), nil)
	return req
}

func (s *Client) buildPostRequest(urls []string, params string) *http.Request {
	urlsJSON, _ := json.Marshal(urls)
	apiURL, _ := url.Parse(defaultApiURL)
	apiURL.RawQuery = params
	req, _ := http.NewRequest("POST", apiURL.String(), bytes.NewReader(urlsJSON))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (s *Client) unmarshalSingleResponse(resp *http.Response) (*URLMetrics, error) {
	content, err := defaultBodyHandler(resp.Body)
	if err != nil {
		return nil, err
	}

	metric := &URLMetrics{}
	err = json.Unmarshal(content, &metric)
	return metric, err
}

func (s *Client) unmarshalBulkResponse(urls []string, resp *http.Response) (map[string]*URLMetrics, error) {
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
func (s *Client) GetURLMetrics(link string, cols int) (*URLMetrics, error) {
	resp, err := DefaultHTTPHandler(s.buildGetRequest(link, s.rawQuery(cols)))
	if err != nil {
		return nil, err
	}

	return s.unmarshalSingleResponse(resp)
}

func (s *Client) GetBulkURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error) {
	resp, err := DefaultHTTPHandler(s.buildPostRequest(urls, s.rawQuery(cols)))
	if err != nil {
		return nil, err
	}

	return s.unmarshalBulkResponse(urls, resp)
}
