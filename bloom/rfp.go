package bloom

import (
	"sync"

	"github.com/goware/urlx"

	"github.com/tylertreat/BoomFilters"
)

type RFPFilter struct {
	//IncludeHeaders []string

	mu   sync.Mutex
	boom boom.Filter
}

func NewRFPFilter() *RFPFilter {
	return &RFPFilter{
		boom: boom.NewDefaultScalableBloomFilter(0.01),
	}
}

func (c *RFPFilter) Add(u string) (bool, error) {
	url, err := urlx.Parse(u)
	if err != nil {
		return false, err
	}
	normalized, err := urlx.Normalize(url)
	if err != nil {
		return false, err
	}

	c.mu.Lock()
	ok := c.boom.TestAndAdd([]byte(normalized))
	c.mu.Unlock()
	return ok, nil
}

func (c *RFPFilter) Exist(u string) (bool, error) {
	url, err := urlx.Parse(u)
	if err != nil {
		return false, err
	}
	normalized, err := urlx.Normalize(url)
	if err != nil {
		return false, err
	}

	c.mu.Lock()
	ok := c.boom.Test([]byte(normalized))
	c.mu.Unlock()
	return ok, nil
}
