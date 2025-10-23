package market

import (
	"errors"
	"strings"
	"sync"

	"cloud.google.com/go/civil"
)

// A Market represents an exchange where financial instruments are traded.
type Market struct {
	Agent
	q Quoter
}

// Search returns a Security with given symbol.
func (m *Market) Search(symbol string) (*Security, error) {
	p, err := m.Security(symbol)
	if err != nil {
		return nil, err
	}

	var l, d civil.Date
	if l, err = civil.ParseDate(p.Listed); err != nil {
		return nil, err
	}

	if p.Delisted != "" {
		if d, err = civil.ParseDate(p.Delisted); err != nil {
			return nil, err
		}
	}
	return &Security{*p, l, d, m.q}, nil
}

var (
	mu  sync.RWMutex
	drs = make(map[string]Driver)
)

// Register makes the named Market available for querying data.
func Register(name string, d Driver) {
	mu.Lock()
	defer mu.Unlock()

	if d == nil {
		panic("market: A nil Market is registered")
	}

	name = strings.ToLower(name)
	if _, ok := drs[name]; ok {
		panic("market: Market " + name + " had been registered twice")
	}

	drs[name] = d
}

// Open returns a registered Market by given name.
func Open(name string) (*Market, error) {
	mu.RLock()
	defer mu.RUnlock()

	d, ok := drs[name]
	if !ok {
		return nil, errors.New("market: Unknown driver " + name)
	}

	a, q, err := d.Open()
	if err != nil {
		return nil, err
	}
	return &Market{a, q}, nil
}

// All returns all registered Markets.
func All() ([]*Market, error) {
	mu.RLock()
	defer mu.RUnlock()

	var ms []*Market
	for name := range drs {
		m, err := Open(name)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}

	return ms, nil
}
