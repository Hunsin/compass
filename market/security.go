package market

import (
	"context"
	"time"

	"cloud.google.com/go/civil"

	"github.com/Hunsin/compass/trade"
	"github.com/Hunsin/compass/trade/pb"
)

// A Security represents a financial instrument in a market.
type Security struct {
	profile          trade.Security
	listed, delisted civil.Date
	quote            Quoter
}

// Profile returns the information of the Security.
func (s *Security) Profile() trade.Security {
	return s.profile
}

// Quotes returns a list of trading quotes from the start to the end date.
func (s *Security) Quotes(ctx context.Context, start, end civil.Date) ([]*trade.Quote, error) {
	if start.After(end) {
		start, end = end, start
	}

	if s.listed.After(end) {
		return nil, Unlisted(s.profile.Symbol + "listed on date" + s.profile.Listed)
	}
	if s.listed.After(start) {
		start = s.listed
	}
	if s.delisted.Before(end) {
		end = s.delisted
	}

	qs, err := s.quote.Range(s.profile.Symbol, start, end)
	if e, ok := err.(*Err); ok && e.Status() == pb.Status_UNIMPLEMENTED {
		qs = []*trade.Quote{} // make sure it's empty
		q := []*trade.Quote{}

		for start.Before(end) {
			select {
			case <-ctx.Done():
				return nil, Cancelled("query cancelled")
			default:
				if start.Month == time.January && (end.Year > start.Year || end.Month == time.December) {
					q, err = s.quote.Year(s.profile.Symbol, start.Year)
					start = civil.Date{Year: start.Year + 1, Month: 1, Day: 1}
				} else {
					q, err = s.quote.Month(s.profile.Symbol, start.Year, start.Month)
					if start.Month == time.December {
						start = civil.Date{Year: start.Year + 1, Month: 1, Day: 1}
					} else {
						start = civil.Date{Year: start.Year, Month: start.Month + 1, Day: 1}
					}
				}

				if err != nil {
					return nil, err
				}
				qs = append(qs, q...)
			}
		}

		// trim pre and suf
		var i int
		for i = 0; i < len(qs); i++ {
			d, err := civil.ParseDate(qs[i].Date)
			if err != nil {
				return nil, err
			}

			if !d.Before(start) {
				break
			}
		}
		qs = qs[i:]

		for i = len(qs) - 1; i > 0; i-- {
			d, err := civil.ParseDate(qs[i].Date)
			if err != nil {
				return nil, err
			}

			if d.Before(end) {
				break
			}
		}
		qs = qs[:i+1]
	}
	return qs, err
}
