package seomoz

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func computeHmac(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// URLMetrics is the basic structure returned by the API
type URLMetrics struct {
	PageAuthority   float64 `json:"pda"`
	DomainAuthority float64 `json:"upa"`
	Links           float64 `json:"uid"`
	URL             string  `json:"uu"`
}

// Client is the main object used to interact with the SeoMoz API
type Client struct {
	AccessID  string
	SecretKey string
}

// NewClient instantiates a new Client
func NewClient(accessID, secretKey string) *Client {
	return &Client{accessID, secretKey}
}

func (s *Client) signature(expires int64) string {
	message := fmt.Sprintf("%s\n%d", s.AccessID, expires)
	return computeHmac(message, s.SecretKey)
}

// GetURLMetrics fetches the metrics for the given urls
func (s *Client) GetURLMetrics(urls []string, cols int) (metrics []URLMetrics, err error) {
	expires := time.Now().Unix() + 300

	v := url.Values{}
	v.Set("AccessID", s.AccessID)
	v.Set("Expires", strconv.Itoa(int(expires)))
	v.Set("Signature", s.signature(expires))
	v.Set("Cols", strconv.Itoa(cols))

	urlsCount := len(urls)
	fullAPIURL := "http://lsapi.seomoz.com/linkscape/url-metrics/"
	if urlsCount == 1 {
		fullAPIURL = fmt.Sprintf("http://lsapi.seomoz.com/linkscape/url-metrics/%s", url.QueryEscape(urls[0]))
	}

	apiURL, err := url.Parse(fullAPIURL)
	if err != nil {
		return
	}
	apiURL.RawQuery = v.Encode()

	var resp *http.Response
	if urlsCount == 1 {
		resp, err = http.Get(apiURL.String())
	} else {
		urlsJSON, jsonErr := json.Marshal(urls)
		if jsonErr != nil {
			err = jsonErr
			return
		}
		resp, err = http.Post(apiURL.String(), "application/json", bytes.NewReader(urlsJSON))
	}

	if err != nil {
		return
	}

	content, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}

	if urlsCount == 1 {
		metric := URLMetrics{}
		err = json.Unmarshal(content, &metric)
		metrics = []URLMetrics{metric}
	} else {
		err = json.Unmarshal(content, &metrics)
	}
	return
}
