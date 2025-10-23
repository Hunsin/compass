package bucket

import (
	"testing"

	"github.com/Hunsin/compass/crawler"

	_ "github.com/Hunsin/compass/twse"
)

var tsmc, ttt crawler.Security

func init() {
	tw, _ := crawler.Open("twse")
	tsmc, _ = tw.Search("2330")
	ttt, _ = tw.Search("0050")
}

func diff(b *Security, c crawler.Security, t *testing.T, fn string) {
	if !equal(b, c) {
		t.Errorf("%s failed.\n"+
			"want: %s, %s, %s, %s, %s\n"+
			"got : %s, %s, %s, %s, %s", fn,
			c.Symbol(), c.Market(), c.Name(), c.Type(), c.Listed(),
			b.Symbol, b.Market, b.Name, b.Type, b.Listed,
		)
	}
}

func equal(b *Security, c crawler.Security) bool {
	return b.Listed.Compare(c.Listed()) == 0 &&
		b.Market == c.Market() &&
		b.Symbol == c.Symbol() &&
		b.Name == c.Name() &&
		b.Type == c.Type()
}

func TestNewSecurity(t *testing.T) {
	bk.InitTables()
	sec, err := bk.NewSecurity(tsmc)
	if err != nil {
		t.Fatalf("bucket.NewSecurity exits with error: %v", err)
	}
	diff(sec, tsmc, t, "bucket.NewSecurity")

	id := sec.id
	if id == "" {
		t.Error("bucket.NewSecurity failed: id not set")
	}

	// insert same crawler.Security should return same Security
	sec, err = bk.NewSecurity(tsmc)
	if err != nil {
		t.Fatalf("bucket.NewSecurity exits with error: %v", err)
	}
	diff(sec, tsmc, t, "bucket.NewSecurity")

	if sec.id != id {
		t.Errorf("bucket.NewSecurity failed: Insert same crawler.Security returns different id."+
			"\nwant: %s, got: %s", id, sec.id)
	}
}

func TestFind(t *testing.T) {
	bk.InitTables()
	if _, err := bk.NewSecurity(ttt); err != nil {
		t.Fatalf("bucket.NewSecurity exits with error: %v", err)
	}

	sec, err := bk.Find("0050", "TWSE")
	if err != nil {
		t.Fatalf("bucket.Find exits with error: %v", err)
	}
	diff(sec, ttt, t, "bucket.Find")
}

func TestSecurities(t *testing.T) {
	bk.InitTables()
	cs := []crawler.Security{tsmc, ttt}
	for i := range cs {
		if _, err := bk.NewSecurity(cs[i]); err != nil {
			t.Fatalf("bucket.NewSecurity exits with error: %v", err)
		}
	}

	list, err := bk.Securities("TWSE")
	if err != nil {
		t.Fatalf("bucket.Securities exits with error: %v", err)
	}
	if len(list) != len(cs) {
		t.Errorf("bucket.Securities failed. Num. of insert: %d, Num. of got: %d", len(cs), len(list))
	}
}
