package crawler

import (
	"errors"
	"sync"

	"cloud.google.com/go/civil"
)

// A Daily represents the daily trading data of a Security.
type Daily struct {
	Date   civil.Date
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int
	Avg    float64
}

// A Security represents a financial instrument in a market.
type Security interface {
	Symbol() string
	Market() string
	Name() string
	Type() string
	Listed() civil.Date
	Date(year, month, date int) (Daily, error)
	Month(year, month int) ([]Daily, error)
	Year(int) ([]Daily, error)
}

// A Market represents an exchange where financial instruments are traded.
type Market interface {
	Search(string) (Security, error)
}

// An ErrNotListed represents an error when trying to get trading data of
// a Security at the date before it is listed.
type ErrNotListed struct {
	Err string
}

func (nl *ErrNotListed) Error() string {
	return nl.Err
}

var (
	mksMu sync.Mutex
	mks   = make(map[string]Market)
)

// Register makes the named Market available for querying data.
func Register(name string, m Market) {
	mksMu.Lock()
	defer mksMu.Unlock()

	if m == nil {
		panic("crawler: A nil Market is registered")
	}
	if _, ok := mks[name]; ok {
		panic("crawler: Market " + name + " had been registered twice")
	}

	mks[name] = m
}

// Open returns a registered Market by given name.
func Open(name string) (Market, error) {
	mksMu.Lock()
	defer mksMu.Unlock()

	m, ok := mks[name]
	if !ok {
		return nil, errors.New("crawler: Unknown driver " + name)
	}

	return m, nil
}
