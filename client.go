package seomoz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var defaultApiURL = "http://lsapi.seomoz.com/linkscape/url-metrics/"
var DefaultHTTPHandler = http.DefaultClient.Do

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

func (s *Client) buildGetRequest(link string, params string) (*http.Request, error) {
	apiURL, err := url.Parse(fmt.Sprintf("%s%s", defaultApiURL, url.QueryEscape(link)))
	if err != nil {
		return nil, err
	}
	apiURL.RawQuery = params
	return http.NewRequest("GET", apiURL.String(), nil)
}

func (s *Client) buildPostRequest(urls []string, params string) (*http.Request, error) {
	apiURL, err := url.Parse(defaultApiURL)
	if err != nil {
		return nil, err
	}
	apiURL.RawQuery = params

	urlsJSON, err := json.Marshal(urls)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL.String(), bytes.NewReader(urlsJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func readAllBody(rc io.ReadCloser) ([]byte, error) {
	content, err := ioutil.ReadAll(rc)
	defer rc.Close()
	return content, err
}

func (s *Client) unmarshalSingleResponse(resp *http.Response) (*URLMetrics, error) {
	content, err := readAllBody(resp.Body)
	if err != nil {
		return nil, err
	}

	metric := &URLMetrics{}
	err = json.Unmarshal(content, &metric)
	return metric, err
}

func (s *Client) unmarshalBulkResponse(urls []string, resp *http.Response) (map[string]*URLMetrics, error) {
	content, err := readAllBody(resp.Body)
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
	req, err := s.buildGetRequest(link, s.rawQuery(cols))
	if err != nil {
		return nil, err
	}

	resp, err := DefaultHTTPHandler(req)
	if err != nil {
		return nil, err
	}

	return s.unmarshalSingleResponse(resp)
}

func (s *Client) GetBulkURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error) {
	req, err := s.buildPostRequest(urls, s.rawQuery(cols))
	if err != nil {
		return nil, err
	}

	resp, err := DefaultHTTPHandler(req)
	if err != nil {
		return nil, err
	}

	return s.unmarshalBulkResponse(urls, resp)
}
