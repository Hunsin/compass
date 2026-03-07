package quote

import (
	"context"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Hunsin/compass/postgres/gen/model"
)

// Querier mirrors all methods of *model.Queries except WithTx, allowing DB
// to be tested without a real database connection.
type Querier interface {
	GetExchange(ctx context.Context, abbr string) (model.Exchange, error)
	GetExchanges(ctx context.Context) ([]model.Exchange, error)
	GetOHLCVsPer30Min(ctx context.Context, secID uuid.UUID, start, before pgtype.Timestamp) ([]model.OHLCVper30Min, error)
	GetOHLCVsPerDay(ctx context.Context, secID uuid.UUID, start, before civil.Date) ([]model.OHLCVperDay, error)
	GetOHLCVsPerMin(ctx context.Context, secID uuid.UUID, start, before pgtype.Timestamp) ([]model.OHLCVperMin, error)
	GetSecurities(ctx context.Context, exchange string) ([]model.Security, error)
	GetSecuritiesBySymbols(ctx context.Context, exchange string, symbols string) ([]model.Security, error)
	InsertExchange(ctx context.Context, abbr string, name string, timezone string) error
	InsertOHLCVsPer30Min(ctx context.Context, arg []model.InsertOHLCVsPer30MinParams) (int64, error)
	InsertOHLCVsPerDay(ctx context.Context, arg []model.InsertOHLCVsPerDayParams) (int64, error)
	InsertOHLCVsPerMin(ctx context.Context, arg []model.InsertOHLCVsPerMinParams) (int64, error)
	InsertSecurities(ctx context.Context, arg []model.InsertSecuritiesParams) (int64, error)
	InsertSecurity(ctx context.Context, exchange string, symbol string, name string) (uuid.UUID, error)
}
