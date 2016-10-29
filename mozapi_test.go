package seomoz

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMOZSignature(t *testing.T) {
	moz := &MOZ{AccessID: "my_id", SecretKey: "my_secret"}
	assert.Equal(t, moz.signature(300), "0Lb5oVSPnkN6KyZ2oDS6tPgTZNI=")
}

func TestMOZQueryParams(t *testing.T) {
	moz := &MOZ{AccessID: "my_id", SecretKey: "my_secret"}
	v, _ := url.ParseQuery(moz.rawQuery(10))
	assert.Equal(t, v.Get("Cols"), "10")
	assert.Equal(t, v.Get("AccessID"), "my_id")
}

func TestMOZGetRequest(t *testing.T) {
	moz := &MOZ{}
	req := moz.buildGetRequest("https://www.example.com", "hello=world")
	assert.Equal(t, req.URL.String(), "http://lsapi.seomoz.com/linkscape/url-metrics/https%3A%2F%2Fwww.example.com?hello=world")
}

func TestMOZPostRequest(t *testing.T) {
	moz := &MOZ{}
	req := moz.buildPostRequest([]string{"https://www.example.com"}, "hello=world")
	assert.Equal(t, req.URL.String(), "http://lsapi.seomoz.com/linkscape/url-metrics/?hello=world")
}

func TestMOZUnmarshalSingleResponse(t *testing.T) {
	moz := &MOZ{}
	jsonResponse := `{"upa": 21, "pda": 42, "uid": 5, "uu": "https://example.com"}`

	resp := &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	m, err := moz.unmarshalSingleResponse(resp)
	assert.Nil(t, err)
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com")

	defaultBodyHandler = mockBodyHandler(nil, errors.New("testing error"))
	_, err = moz.unmarshalSingleResponse(resp)
	assert.NotNil(t, err)
	defaultBodyHandler = readAllCloser
}

func TestClientUnmarshalBatchResponse(t *testing.T) {
	moz := &MOZ{}
	jsonResponse := `[{"upa": 21, "pda": 42, "uid": 5, "uu": ""}]`

	link := "https://example.com"
	resp := &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	result, err := moz.unmarshalBatchResponse([]string{link}, resp)
	assert.Nil(t, err)
	m := result[link]
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "example.com/")

	link = "https://example.com/#%fg"
	resp = &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	result, err = moz.unmarshalBatchResponse([]string{link}, resp)
	assert.Nil(t, err)
	m = result[link]
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com/#%fg")

	resp = &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))}
	_, err = moz.unmarshalBatchResponse([]string{link, link}, resp)
	assert.NotNil(t, err)

	resp = &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(``)))}
	_, err = moz.unmarshalBatchResponse([]string{link}, resp)
	assert.NotNil(t, err)

	defaultBodyHandler = mockBodyHandler(nil, errors.New("testing error"))
	_, err = moz.unmarshalBatchResponse([]string{link}, resp)
	assert.NotNil(t, err)
	defaultBodyHandler = readAllCloser
}

func TestMOZGetSingleMetrics(t *testing.T) {
	moz := &MOZ{}
	DefaultHTTPHandler = mockHTTPHandler(`{"upa": 21, "pda": 42, "uid": 5, "uu": "https://example.com"}`, nil)
	m, err := moz.GetURLMetrics("", 0)
	assert.Nil(t, err)
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com")

	DefaultHTTPHandler = mockHTTPHandler("", errors.New("testing error"))
	_, err = moz.GetURLMetrics("", 0)
	assert.NotNil(t, err)
}

func TestMOZGetBatchMetrics(t *testing.T) {
	moz := &MOZ{}
	DefaultHTTPHandler = mockHTTPHandler(`[{"upa": 21, "pda": 42, "uid": 5, "uu": "https://example.com"}]`, nil)
	results, err := moz.GetBatchURLMetrics([]string{"https://example.com"}, 0)
	assert.Nil(t, err)
	m := results["https://example.com"]
	assert.Equal(t, m.DomainAuthority, float64(42))
	assert.Equal(t, m.PageAuthority, float64(21))
	assert.Equal(t, m.Links, float64(5))
	assert.Equal(t, m.URL, "https://example.com")

	DefaultHTTPHandler = mockHTTPHandler("", errors.New("testing error"))
	_, err = moz.GetBatchURLMetrics([]string{"https://example.com"}, 0)
	assert.NotNil(t, err)
}
