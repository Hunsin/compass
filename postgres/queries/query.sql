-- name: GetSecurities :many
SELECT * FROM securities WHERE market = @market;

-- name: GetSecuritiesBySymbols :many
SELECT * FROM securities WHERE market = @market AND symbol IN (@symbols);

-- name: InsertSecurities :copyfrom
INSERT INTO securities (id, market, symbol, name, isin) VALUES (@id, @market, @symbol, @name, @isin);

-- name: GetOHLCVAsPerMin :many
SELECT * FROM ohlcva_per_min WHERE sec_id = @sec_id AND ts >= @start AND ts < @before ORDER BY ts;

-- name: InsertOHLCVAsPerMin :copyfrom
INSERT INTO ohlcva_per_min (
    sec_id, ts, open, high, low, close, volume, amount
) VALUES (
    @sec_id, @ts, @open, @high, @low, @close, @volume, @amount
);

-- name: GetOHLCVAsPer30Min :many
SELECT * FROM ohlcva_per_30min WHERE sec_id = @sec_id AND ts >= @start AND ts < @before ORDER BY ts;

-- name: InsertOHLCVAsPer30Min :copyfrom
INSERT INTO ohlcva_per_30min (
    sec_id, ts, open, high, low, close, volume, amount
) VALUES (
    @sec_id, @ts, @open, @high, @low, @close, @volume, @amount
);

-- name: GetOHLCVAsPerDay :many
SELECT * FROM ohlcva_per_day WHERE sec_id = @sec_id AND ts >= @start AND ts < @before ORDER BY ts;

-- name: InsertOHLCVAsPerDay :copyfrom
INSERT INTO ohlcva_per_day (
    sec_id, ts, open, high, low, close, volume, amount
) VALUES (
    @sec_id, @ts, @open, @high, @low, @close, @volume, @amount
);