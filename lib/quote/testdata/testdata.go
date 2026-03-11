package testdata

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func loadCSV(filename string) ([]*pb.OHLCV, error) {
	_, currentFile, _, _ := runtime.Caller(1)
	p := filepath.Join(filepath.Dir(currentFile), filename)

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	rows := make([]*pb.OHLCV, 0, len(records)-1)
	for _, rec := range records[1:] { // skip header
		ts, err := time.Parse(time.DateTime, rec[0])
		if err != nil {
			if ts, err = time.Parse(time.DateOnly, rec[0]); err != nil {
				return nil, err
			}
		}
		o, _ := strconv.ParseFloat(rec[1], 64)
		h, _ := strconv.ParseFloat(rec[2], 64)
		l, _ := strconv.ParseFloat(rec[3], 64)
		c, _ := strconv.ParseFloat(rec[4], 64)
		v, _ := strconv.ParseUint(rec[5], 10, 64)
		rows = append(rows, &pb.OHLCV{
			Ts:     timestamppb.New(ts.UTC()),
			Open:   &o,
			High:   &h,
			Low:    &l,
			Close:  &c,
			Volume: &v,
		})
	}
	return rows, nil
}

// HonHai returns daily OHLCVs of Hon Hai (2317) from 2025-10-01 to 2025-12-31.
func HonHai() ([]*pb.OHLCV, error) {
	return loadCSV("2317_2025Q4.csv")
}

// Mediatek returns OHLCVs with 60-second intervals of Mediatek (2454) on 2026-02-26.
func Mediatek() ([]*pb.OHLCV, error) {
	return loadCSV("2454_2026-02-26.csv")
}
