package quote

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
)

// Service implements the gRPC QuoteService.
type Service struct {
	pb.UnimplementedQuoteServiceServer
	db *model.Queries
}

// New creates a new Service backed by the given database queries.
func New(db *model.Queries) *Service {
	return &Service{db: db}
}

// floatToNumeric converts a float64 to pgtype.Numeric.
func floatToNumeric(f float64) pgtype.Numeric {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	dotIdx := strings.IndexByte(s, '.')
	var intStr string
	var exp int32
	if dotIdx == -1 {
		intStr = s
	} else {
		intStr = s[:dotIdx] + s[dotIdx+1:]
		exp = -int32(len(s) - dotIdx - 1)
	}
	bigInt, _ := new(big.Int).SetString(intStr, 10)
	return pgtype.Numeric{Int: bigInt, Exp: exp, Valid: true}
}

// numericToFloat converts a pgtype.Numeric to float64.
func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid || n.NaN || n.Int == nil {
		return 0
	}
	f, _ := new(big.Float).SetInt(n.Int).Float64()
	if n.Exp != 0 {
		f *= math.Pow10(int(n.Exp))
	}
	return f
}

// uint64ToNumeric converts a uint64 to pgtype.Numeric.
func uint64ToNumeric(u uint64) pgtype.Numeric {
	return pgtype.Numeric{Int: new(big.Int).SetUint64(u), Exp: 0, Valid: true}
}

// numericToUint64 converts a pgtype.Numeric to uint64.
func numericToUint64(n pgtype.Numeric) uint64 {
	return uint64(math.Round(numericToFloat(n)))
}
