CREATE TABLE exchanges (
    abbr TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    timezone TEXT NOT NULL
);

CREATE TABLE securities (
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange TEXT NOT NULL,
    symbol   TEXT NOT NULL,
    name     TEXT NOT NULL,
    UNIQUE (exchange, symbol),
    FOREIGN KEY (exchange) REFERENCES exchanges(abbr) ON DELETE CASCADE
);

CREATE TABLE ohlcv_per_min (
    sec_id    UUID NOT NULL,
    ts   TIMESTAMP NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume  BIGINT NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE INDEX idx_ohlcv_per_min_sec_id_ts ON ohlcv_per_min (sec_id, ts);

-- The 30-minute table is designed to be upserted with the latest data from the
-- per-minute table, so we use ON CONFLICT DO UPDATE to handle both insertions and updates.
-- The timestamps are always aligned to 30-minute boundaries.
CREATE TABLE ohlcv_per_30min (
    sec_id    UUID NOT NULL,
    ts   TIMESTAMP NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume  BIGINT NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE INDEX idx_ohlcv_per_30min_sec_id_ts ON ohlcv_per_30min (sec_id, ts);

CREATE TABLE ohlcv_per_day (
    sec_id    UUID NOT NULL,
    date      DATE NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume  BIGINT NOT NULL,
    UNIQUE (sec_id, date),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (date);

CREATE INDEX idx_ohlcv_per_day_sec_id_date ON ohlcv_per_day (sec_id, date);
