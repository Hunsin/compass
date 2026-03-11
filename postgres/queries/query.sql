-- name: InsertExchange :exec
INSERT INTO exchanges (abbr, name, timezone) VALUES (@abbr, @name, @timezone);

-- name: GetExchange :one
SELECT abbr, name, timezone FROM exchanges WHERE abbr = @abbr;

-- name: GetExchanges :many
SELECT * FROM exchanges;

-- name: InsertSecurities :copyfrom
INSERT INTO securities (exchange, symbol, name) VALUES (@exchange, @symbol, @name);

-- name: InsertSecurity :one
INSERT INTO securities (exchange, symbol, name) VALUES (@exchange, @symbol, @name) RETURNING id;

-- name: GetSecurities :many
SELECT * FROM securities WHERE exchange = @exchange;

-- name: GetSecuritiesBySymbols :many
SELECT * FROM securities WHERE exchange = @exchange AND symbol = ANY(@symbols::text[]);

-- name: InsertOHLCVsPerMin :copyfrom
INSERT INTO ohlcv_per_min (
    sec_id, ts, open, high, low, close, volume
) VALUES (
    @sec_id, @ts, @open, @high, @low, @close, @volume
);

-- name: GetOHLCVsPerMin :many
SELECT * FROM ohlcv_per_min WHERE sec_id = @sec_id AND ts >= @start AND ts < @before ORDER BY ts;

-- name: UpsertOHLCVPer30Min :exec
INSERT INTO ohlcv_per_30min (
    sec_id, ts, open, high, low, close, volume
) VALUES (
    @sec_id, @ts, @open, @high, @low, @close, @volume
) ON CONFLICT (sec_id, ts) DO UPDATE SET
    open   = CASE WHEN @is_first::boolean THEN EXCLUDED.open ELSE ohlcv_per_30min.open END,
    high   = GREATEST(ohlcv_per_30min.high, EXCLUDED.high),
    low    = LEAST(ohlcv_per_30min.low, EXCLUDED.low),
    close  = CASE WHEN @is_last::boolean THEN EXCLUDED.close ELSE ohlcv_per_30min.close END,
    volume = ohlcv_per_30min.volume + EXCLUDED.volume;

-- name: GetOHLCVsPer30Min :many
SELECT * FROM ohlcv_per_30min WHERE sec_id = @sec_id AND ts >= @start AND ts < @before ORDER BY ts;

-- name: InsertOHLCVsPerDay :copyfrom
INSERT INTO ohlcv_per_day (
    sec_id, date, open, high, low, close, volume
) VALUES (
    @sec_id, @date, @open, @high, @low, @close, @volume
);

-- name: GetOHLCVsPerDay :many
SELECT * FROM ohlcv_per_day WHERE sec_id = @sec_id AND date >= @start AND date < @before ORDER BY date;
