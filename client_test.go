package seomoz

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMozApi struct {
	metrics map[string]*URLMetrics
	err     error
}

func (m *mockMozApi) GetURLMetrics(link string, cols int) (*URLMetrics, error) {
	return m.metrics[link], m.err
}

func (m *mockMozApi) GetBatchURLMetrics(urls []string, cols int) (map[string]*URLMetrics, error) {
	if m.err != nil {
		return nil, m.err
	}

	results := map[string]*URLMetrics{}
	for _, link := range urls {
		results[link] = m.metrics[link]
	}
	return results, nil
}

func TestEnvClient(t *testing.T) {
	os.Setenv("SEOMOZ_ACCESS_ID", "my_id")
	os.Setenv("SEOMOZ_SECRET_KEY", "my_secret")
	client := NewEnvClient()
	moz := client.moz.(*MOZ)
	assert.Equal(t, moz.AccessID, "my_id")
	assert.Equal(t, moz.SecretKey, "my_secret")
}

func TestClientGetSingleMetrics(t *testing.T) {
	expectedMetrics := &URLMetrics{DomainAuthority: 45, PageAuthority: 68, Links: 123, URL: "https://example.com"}
	mockMoz := &mockMozApi{metrics: map[string]*URLMetrics{expectedMetrics.URL: expectedMetrics}}
	client := &Client{moz: mockMoz}
	m, err := client.GetURLMetrics("https://example.com", 0)
	assert.Nil(t, err)
	assert.Equal(t, expectedMetrics, m)

	client = &Client{moz: &mockMozApi{err: errors.New("mock error")}}
	m, err = client.GetURLMetrics("https://example.com", 0)
	assert.Nil(t, m)
	assert.NotNil(t, err)
}

func TestClientGetBatchMetrics(t *testing.T) {
	expectedMetrics := &URLMetrics{DomainAuthority: 45, PageAuthority: 68, Links: 123, URL: "https://example.com"}
	mockMoz := &mockMozApi{metrics: map[string]*URLMetrics{expectedMetrics.URL: expectedMetrics}}
	client := &Client{moz: mockMoz}
	m, err := client.GetBatchURLMetrics([]string{"https://example.com"}, 0)
	assert.Nil(t, err)
	assert.Equal(t, mockMoz.metrics, m)

	client = &Client{moz: &mockMozApi{err: errors.New("mock error")}}
	m, err = client.GetBatchURLMetrics([]string{"https://example.com"}, 0)
	assert.Nil(t, m)
	assert.NotNil(t, err)
}

func TestClientGetBulkMetrics(t *testing.T) {
	expectedMetrics := map[string]*URLMetrics{
		"https://example.com/1": {DomainAuthority: 45, PageAuthority: 68, Links: 123, URL: "https://example.com/1"},
		"https://example.com/2": {DomainAuthority: 45, PageAuthority: 68, Links: 123, URL: "https://example.com/2"},
		"https://example.com/3": {DomainAuthority: 45, PageAuthority: 68, Links: 123, URL: "https://example.com/3"},
		"https://example.com/4": {DomainAuthority: 45, PageAuthority: 68, Links: 123, URL: "https://example.com/4"},
	}
	mockMoz := &mockMozApi{metrics: expectedMetrics}
	client := &Client{moz: mockMoz, MaxBatchURLs: 2}
	m, err := client.GetBulkURLMetrics([]string{"https://example.com/1", "https://example.com/2", "https://example.com/3", "https://example.com/4"}, 0)
	assert.Nil(t, err)
	assert.Equal(t, mockMoz.metrics, m)

	client = &Client{moz: &mockMozApi{err: errors.New("mock error")}, MaxBatchURLs: 0}
	m, err = client.GetBulkURLMetrics([]string{"https://example.com/1", "https://example.com/2", "https://example.com/3", "https://example.com/4"}, 0)
	assert.Nil(t, m)
	assert.NotNil(t, err)
}
