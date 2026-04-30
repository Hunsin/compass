package statistics

import (
	"context"
	"errors"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
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

func (s *store) CreateMarginTransactions(ctx context.Context, exchange, symbol string, txs []*pb.MarginTransaction) error {
	exchange = strings.ToLower(exchange)
	symbol = strings.ToUpper(symbol)

	secID, err := s.securityID(ctx, exchange, symbol)
	if err != nil {
		return err
	}

	params := make([]model.InsertMarginTransactionsParams, len(txs))
	for i, tx := range txs {
		params[i] = model.InsertMarginTransactionsParams{
			SecID:                       secID,
			Date:                        civil.DateOf(tx.GetDate().AsTime()),
			MarginPurchaseBuy:           tx.GetMarginPurchaseBuy(),
			MarginPurchaseRedemption:    tx.GetMarginPurchaseRedemption(),
			MarginPurchaseCashRepayment: tx.GetMarginPurchaseCashRepayment(),
			MarginPurchaseBalance:       tx.GetMarginPurchaseBalance(),
			MarginPurchaseLimit:         tx.GetMarginPurchaseLimit(),
			ShortSale:                   tx.GetShortSale(),
			ShortSaleRedemption:         tx.GetShortSaleRedemption(),
			ShortSaleStockRepayment:     tx.GetShortSaleStockRepayment(),
			ShortSaleBalance:            tx.GetShortSaleBalance(),
			ShortSaleLimit:              tx.GetShortSaleLimit(),
			QuotaNextDay:                tx.GetQuotaNextDay(),
		}
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
		mpBuy := r.MarginPurchaseBuy
		mpRedemption := r.MarginPurchaseRedemption
		mpCashRepay := r.MarginPurchaseCashRepayment
		mpBalance := r.MarginPurchaseBalance
		mpLimit := r.MarginPurchaseLimit
		ss := r.ShortSale
		ssRedemption := r.ShortSaleRedemption
		ssStockRepay := r.ShortSaleStockRepayment
		ssBalance := r.ShortSaleBalance
		ssLimit := r.ShortSaleLimit
		quota := r.QuotaNextDay
		result[i] = &pb.MarginTransaction{
			Date:                        timestamppb.New(r.Date.In(time.UTC)),
			MarginPurchaseBuy:           &mpBuy,
			MarginPurchaseRedemption:    &mpRedemption,
			MarginPurchaseCashRepayment: &mpCashRepay,
			MarginPurchaseBalance:       &mpBalance,
			MarginPurchaseLimit:         &mpLimit,
			ShortSale:                   &ss,
			ShortSaleRedemption:         &ssRedemption,
			ShortSaleStockRepayment:     &ssStockRepay,
			ShortSaleBalance:            &ssBalance,
			ShortSaleLimit:              &ssLimit,
			QuotaNextDay:                &quota,
		}
	}
	return result, nil
}

func (s *store) Health(ctx context.Context) error {
	if err := s.db.Ping(ctx); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to ping database")
		return err
	}
	return nil
}
