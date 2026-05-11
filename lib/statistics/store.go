package statistics

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/singleflight"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/lib/oops"
	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
)

// DBTX is the database interface required by the statistics store.
type DBTX interface {
	model.DBTX
	Ping(context.Context) error
}

// store is a PostgreSQL-backed implementation of Model.
type store struct {
	db      DBTX
	queries model.Querier
	sg      singleflight.Group
}

// Connect creates a new Model backed by the given database connection.
func Connect(db DBTX) Model {
	return &store{db: db, queries: model.New(db)}
}

// securityID resolves the UUID for the given exchange and symbol.
// It uses singleflight to deduplicate concurrent lookups for the same key.
func (s *store) securityID(ctx context.Context, exchange, symbol string) (uuid.UUID, error) {
	key := exchange + ":" + symbol
	v, err, _ := s.sg.Do(key, func() (any, error) {
		return s.queries.GetSecurity(ctx, exchange, symbol)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.UUID{}, oops.NotFound("security %s not found", symbol)
	}
	if err != nil {
		return uuid.UUID{}, oops.Internal(err)
	}
	return v.(model.Security).ID, nil
}

func (s *store) CreateMarginTransactions(ctx context.Context, exchange string, date civil.Date, txs map[string]*pb.MarginTransaction) error {
	exchange = strings.ToLower(exchange)
	idBySymbol := make(map[string]uuid.UUID, len(txs))

	symbols := slices.Collect(maps.Keys(txs))
	secs, err := s.queries.GetSecuritiesBySymbols(ctx, exchange, symbols)
	if err != nil {
		return oops.Internal(err)
	}
	for _, sec := range secs {
		idBySymbol[sec.Symbol] = sec.ID
	}

	params := make([]model.InsertMarginTransactionsParams, 0, len(txs))
	for symbol, mt := range txs {
		secID := idBySymbol[symbol]
		params = append(params, model.InsertMarginTransactionsParams{
			SecID:             secID,
			Date:              date,
			MarginPurchase:    mt.GetMarginPurchase(),
			MarginSales:       mt.GetMarginSales(),
			CashRedemption:    mt.GetCashRedemption(),
			MarginBalance:     mt.GetMarginBalance(),
			ShortCovering:     mt.GetShortCovering(),
			ShortSale:         mt.GetShortSale(),
			StockRedemption:   mt.GetStockRedemption(),
			ShortBalance:      mt.GetShortBalance(),
			MarginShortOffset: mt.GetMarginShortOffset(),
		})
	}

	if _, err := s.queries.InsertMarginTransactions(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return oops.AlreadyExists("one or more margin transactions already exist")
		}
		return oops.Internal(err)
	}
	return nil
}

func (s *store) GetMarginTransactions(ctx context.Context, exchange, symbol string, from, before time.Time) ([]*pb.MarginTransaction, error) {
	exchange = strings.ToLower(exchange)
	symbol = strings.ToUpper(symbol)

	secID, err := s.securityID(ctx, exchange, symbol)
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.GetMarginTransactions(ctx, secID, civil.DateOf(from), civil.DateOf(before))
	if err != nil {
		return nil, oops.Internal(err)
	}

	result := make([]*pb.MarginTransaction, len(rows))
	for i, r := range rows {
		mp := r.MarginPurchase
		ms := r.MarginSales
		cr := r.CashRedemption
		mb := r.MarginBalance
		sc := r.ShortCovering
		ss := r.ShortSale
		sr := r.StockRedemption
		sb := r.ShortBalance
		mso := r.MarginShortOffset
		result[i] = &pb.MarginTransaction{
			Date:              timestamppb.New(r.Date.In(time.UTC)),
			MarginPurchase:    &mp,
			MarginSales:       &ms,
			CashRedemption:    &cr,
			MarginBalance:     &mb,
			ShortCovering:     &sc,
			ShortSale:         &ss,
			StockRedemption:   &sr,
			ShortBalance:      &sb,
			MarginShortOffset: &mso,
		}
	}
	return result, nil
}

func (s *store) Health(ctx context.Context) error {
	if err := s.db.Ping(ctx); err != nil {
		return err
	}
	return nil
}
