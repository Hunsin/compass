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
SELECT * FROM securities WHERE exchange = @exchange AND symbol IN (@symbols);

-- name: InsertOHLCVAsPerMin :copyfrom
INSERT INTO ohlcva_per_min (
    sec_id, ts, open, high, low, close, volume, amount
) VALUES (
    @sec_id, @ts, @open, @high, @low, @close, @volume, @amount
);

-- name: GetOHLCVAsPerMin :many
SELECT * FROM ohlcva_per_min WHERE sec_id = @sec_id AND ts >= @start AND ts < @before ORDER BY ts;

-- name: InsertOHLCVAsPer30Min :copyfrom
INSERT INTO ohlcva_per_30min (
    sec_id, ts, open, high, low, close, volume, amount
) VALUES (
    @sec_id, @ts, @open, @high, @low, @close, @volume, @amount
);

-- name: GetOHLCVAsPer30Min :many
SELECT * FROM ohlcva_per_30min WHERE sec_id = @sec_id AND ts >= @start AND ts < @before ORDER BY ts;

-- name: InsertOHLCVAsPerDay :copyfrom
INSERT INTO ohlcva_per_day (
    sec_id, date, open, high, low, close, volume, amount
) VALUES (
    @sec_id, @date, @open, @high, @low, @close, @volume, @amount
);

-- name: GetOHLCVAsPerDay :many
SELECT * FROM ohlcva_per_day WHERE sec_id = @sec_id AND date >= @start AND date < @before ORDER BY date;
