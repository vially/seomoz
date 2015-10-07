package seomoz

// URLMetrics is the basic structure returned by the API
type URLMetrics struct {
	PageAuthority   float64 `json:"upa"`
	DomainAuthority float64 `json:"pda"`
	Links           float64 `json:"uid"`
	URL             string  `json:"uu"`
}
