CREATE TABLE exchanges (
    abbr     text PRIMARY KEY,
    name     text NOT NULL UNIQUE,
    timezone text NOT NULL
);

CREATE TABLE securities (
    id       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange text NOT NULL,
    symbol   text NOT NULL,
    name     text NOT NULL,
    UNIQUE (exchange, symbol),
    FOREIGN KEY (exchange) REFERENCES exchanges(abbr) ON DELETE CASCADE
);

CREATE TABLE ohlcv_per_min (
    sec_id    uuid NOT NULL,
    ts   timestamp NOT NULL,
    open   numeric NOT NULL,
    high   numeric NOT NULL,
    low    numeric NOT NULL,
    close  numeric NOT NULL,
    volume  bigint NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE TABLE ohlcv_per_min_default PARTITION OF ohlcv_per_min DEFAULT;

CREATE INDEX idx_ohlcv_per_min_sec_id_ts ON ohlcv_per_min (sec_id, ts);

-- The 30-minute table is designed to be upserted with the latest data from the
-- per-minute table, so we use ON CONFLICT DO UPDATE to handle both insertions and updates.
-- The timestamps are always aligned to 30-minute boundaries.
CREATE TABLE ohlcv_per_30min (
    sec_id    uuid NOT NULL,
    ts   timestamp NOT NULL,
    open   numeric NOT NULL,
    high   numeric NOT NULL,
    low    numeric NOT NULL,
    close  numeric NOT NULL,
    volume  bigint NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE TABLE ohlcv_per_30min_default PARTITION OF ohlcv_per_30min DEFAULT;

CREATE INDEX idx_ohlcv_per_30min_sec_id_ts ON ohlcv_per_30min (sec_id, ts);

CREATE TABLE ohlcv_per_day (
    sec_id    uuid NOT NULL,
    date      date NOT NULL,
    open   numeric NOT NULL,
    high   numeric NOT NULL,
    low    numeric NOT NULL,
    close  numeric NOT NULL,
    volume  bigint NOT NULL,
    UNIQUE (sec_id, date),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (date);

CREATE TABLE ohlcv_per_day_default PARTITION OF ohlcv_per_day DEFAULT;

CREATE INDEX idx_ohlcv_per_day_sec_id_date ON ohlcv_per_day (sec_id, date);
