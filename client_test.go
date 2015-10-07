package seomoz

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func mockHTTPHandler(body string, err error) func(*http.Request) (*http.Response, error) {
	return func(*http.Request) (*http.Response, error) {
		return &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(body)))}, err
	}
}

func mockBodyHandler(out []byte, err error) func(io.ReadCloser) ([]byte, error) {
	return func(io.ReadCloser) ([]byte, error) {
		return out, err
	}
}

func TestEnvClient(t *testing.T) {
	os.Setenv("SEOMOZ_ACCESS_ID", "my_id")
	os.Setenv("SEOMOZ_SECRET_KEY", "my_secret")
	client := NewEnvClient()
	assert.Equal(t, client.AccessID, "my_id")
	assert.Equal(t, client.SecretKey, "my_secret")
}

func TestClientSignature(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	assert.Equal(t, client.signature(300), "0Lb5oVSPnkN6KyZ2oDS6tPgTZNI=")
}

func TestClientQueryParams(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	v, _ := url.ParseQuery(client.rawQuery(10))
	assert.Equal(t, v.Get("Cols"), "10")
	assert.Equal(t, v.Get("AccessID"), "my_id")
}

func TestClientGetRequest(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	req := client.buildGetRequest("https://www.example.com", "hello=world")
	assert.Equal(t, req.URL.String(), "http://lsapi.seomoz.com/linkscape/url-metrics/https%3A%2F%2Fwww.example.com?hello=world")
}

func TestClientPostRequest(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	req := client.buildPostRequest([]string{"https://www.example.com"}, "hello=world")
	assert.Equal(t, req.URL.String(), "http://lsapi.seomoz.com/linkscape/url-metrics/?hello=world")
}

func TestClientUnmarshalSingleResponse(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	jsonResponse := `{"upa": 21, "pda": 42, "uid": 5, "uu": "https://example.com"}`

	resp := &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	m, err := client.unmarshalSingleResponse(resp)
	assert.Nil(t, err)
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com")

	defaultBodyHandler = mockBodyHandler(nil, errors.New("testing error"))
	_, err = client.unmarshalSingleResponse(resp)
	assert.NotNil(t, err)
	defaultBodyHandler = readAllCloser
}

func TestClientUnmarshalBulkResponse(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	jsonResponse := `[{"upa": 21, "pda": 42, "uid": 5, "uu": ""}]`

	link := "https://example.com"
	resp := &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	result, err := client.unmarshalBulkResponse([]string{link}, resp)
	assert.Nil(t, err)
	m := result[link]
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "example.com/")

	link = "https://example.com/#%fg"
	resp = &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	result, err = client.unmarshalBulkResponse([]string{link}, resp)
	assert.Nil(t, err)
	m = result[link]
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com/#%fg")

	resp = &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	_, err = client.unmarshalBulkResponse([]string{link, link}, resp)
	assert.NotNil(t, err)

	resp = &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(``)))}
	_, err = client.unmarshalBulkResponse([]string{link}, resp)
	assert.NotNil(t, err)

	defaultBodyHandler = mockBodyHandler(nil, errors.New("testing error"))
	_, err = client.unmarshalBulkResponse([]string{link}, resp)
	assert.NotNil(t, err)
	defaultBodyHandler = readAllCloser
}

func TestClientGetSingleMetrics(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	DefaultHTTPHandler = mockHTTPHandler(`{"upa": 21, "pda": 42, "uid": 5, "uu": "https://example.com"}`, nil)
	m, err := client.GetURLMetrics("", 0)
	assert.Nil(t, err)
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com")

	DefaultHTTPHandler = mockHTTPHandler("", errors.New("testing error"))
	_, err = client.GetURLMetrics("", 0)
	assert.NotNil(t, err)
}

func TestClientGetBulkMetrics(t *testing.T) {
	client := &Client{AccessID: "my_id", SecretKey: "my_secret"}
	DefaultHTTPHandler = mockHTTPHandler(`[{"upa": 21, "pda": 42, "uid": 5, "uu": "https://example.com"}]`, nil)
	results, err := client.GetBulkURLMetrics([]string{"https://example.com"}, 0)
	assert.Nil(t, err)
	m := results["https://example.com"]
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com")

	DefaultHTTPHandler = mockHTTPHandler("", errors.New("testing error"))
	_, err = client.GetBulkURLMetrics([]string{"https://example.com"}, 0)
	assert.NotNil(t, err)
}
