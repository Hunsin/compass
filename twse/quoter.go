package twse

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/civil"

	"github.com/Hunsin/compass/market"
	"github.com/Hunsin/compass/trade"
	"github.com/Hunsin/compass/trade/pb"
)

var day = newAgent("http://www.twse.com.tw/en/exchangeReport/STOCK_DAY?response=json&date=%4d%02d%02d&stockNo=%s")

type apiDay struct {
	Data   [][]string `json:"data"`
	Date   string     `json:"date"`
	Fields []string   `json:"fields"`
	Stat   string     `json:"stat"`
}

// A quoter is an instance which implements market.Quoter interface.
type quoter struct{}

// Month returns a list of trade.Quote by given year and month.
func (q *quoter) Month(symbol string, year int, month time.Month) ([]*trade.Quote, error) {
	// check values
	if month < 1 || month > 12 {
		return nil, market.Error(pb.Status_BAD_REQUEST, fmt.Sprintf("twse: Invalid month %d", month))
	}
	if year < 1992 || year > time.Now().Year() {
		return nil, fmt.Errorf("twse: Invalid year %d", year)
	}

	// the first date available is 1992/01/04
	d := 1
	if year == 1992 {
		d = 4
	}

	st := apiDay{}
	err := day.do(func(r io.Reader) error {
		return json.NewDecoder(r).Decode(&st)
	}, year, month, d, symbol)
	if err != nil {
		return nil, err
	}

	if st.Stat != "OK" {
		return nil, fmt.Errorf("twse: %s", st.Stat)
	}

	var qs []*trade.Quote
	for i := range st.Data {
		v, _ := strconv.ParseUint(formatNum(st.Data[i][1]), 10, 64) // volume
		s, _ := strconv.Atoi(formatNum(st.Data[i][2]))              // value
		o, _ := strconv.ParseFloat(formatNum(st.Data[i][3]), 64)    // open
		h, _ := strconv.ParseFloat(formatNum(st.Data[i][4]), 64)    // highest
		l, _ := strconv.ParseFloat(formatNum(st.Data[i][5]), 64)    // lowest
		c, _ := strconv.ParseFloat(formatNum(st.Data[i][6]), 64)    // close
		qs = append(qs, &trade.Quote{
			Date:   formatDate(st.Data[i][0]),
			Open:   o,
			High:   h,
			Low:    l,
			Close:  c,
			Volume: v,
			Avg:    float64(s) / float64(v)})
	}
	return qs, nil
}

// Year returns a list of trade.Quote in given year.
func (q *quoter) Year(symbol string, year int) ([]*trade.Quote, error) {
	start, end := 1, 12
	now := time.Now()
	if year == now.Year() {
		end = int(now.Month())
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	yr := make(map[int][]*trade.Quote)
	ch := make(chan error)
	defer close(ch)

	for i := start; i < end+1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m, err := q.Month(symbol, year, time.Month(i))
			if err != nil && err.Error() != "twse: No Data!" {
				ch <- err
				return
			}

			mu.Lock()
			defer mu.Unlock()
			yr[i] = m
		}(i)
	}

	wg.Wait()
	select {
	case err := <-ch:
		return nil, err
	default:
		for i := start + 1; i < end+1; i++ {
			yr[start] = append(yr[start], yr[i]...)
		}
		if len(yr[start]) == 0 {
			return nil, market.Unlisted("twse: No data found")
		}
		return yr[start], nil
	}
}

func (q *quoter) Range(symbol string, start, end civil.Date) ([]*trade.Quote, error) {
	return nil, market.Unimplemented("twse: range method not supported")
}
