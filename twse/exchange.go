package twse

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"

	"github.com/Hunsin/compass/market"
	"github.com/Hunsin/compass/trade"
	hu "github.com/Hunsin/go-htmlutil"
)

var isin = newAgent("http://isin.twse.com.tw/isin/e_class_main.jsp?owncode=%s&market=1")

// parseSecurity extracts the Security from given <tr> node.
func parseSecurity(tr *html.Node) (*trade.Security, error) {
	var td []string

	// push text contents of each <td> to slice
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "td" {
			td = append(td, hu.Text(c))
		}
	}

	// prevent panic
	if len(td) < 8 {
		return nil, fmt.Errorf("twse: Could not parse data at %v", *tr)
	}

	return &trade.Security{
		Market: "twse",
		Isin:   strings.TrimSpace(td[1]),
		Symbol: strings.TrimSpace(td[2]),
		Name:   strings.TrimSpace(td[3]),
		Type:   strings.TrimSpace(td[5]),
		Listed: formatDate(td[7]),
	}, nil
}

// An exchange implements the market.Agent interface.
type exchange struct{}

func (e *exchange) Security(symbol string) (*trade.Security, error) {
	var s *trade.Security
	return s, isin.do(func(r io.Reader) error {
		n, err := html.Parse(r)
		if err != nil {
			return err
		}

		var tr *html.Node
		hu.Last(n, func(n *html.Node) (found bool) {
			if found = n.Data == "td" && hu.Text(n) == symbol; found {
				tr = n.Parent
			}
			return
		})

		// return error if no <td> node with symbol was found
		if tr == nil {
			msg := fmt.Sprintf("twse: Symbol %s not found", symbol)
			return market.Unlisted(msg)
		}

		s, err = parseSecurity(tr)
		return err
	}, symbol)
}

func (e *exchange) Listed() ([]*trade.Security, error) {
	var ss []*trade.Security
	return ss, isin.do(func(r io.Reader) error {
		n, err := html.Parse(r)
		if err != nil {
			return err
		}

		hu.Walk(n, func(n *html.Node) (found bool) {
			if found = n.Data == "tr"; found {
				s, er := parseSecurity(n)
				if er != nil {
					err = er // only the last error left; intended behavior
					return
				}
				ss = append(ss, s)
			}
			return
		})

		// remove table header
		if len(ss) > 0 && ss[0].Isin == "ISIN Code" {
			ss = ss[1:]
		}

		return err
	}, "")
}

func (e *exchange) Profile() *trade.Market {
	return &trade.Market{
		Code:     "twse",
		Name:     "Taiwan Stock Exchange",
		Currency: "twd",
	}
}
