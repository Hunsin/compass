CREATE TABLE securities (
    id     UUID PRIMARY KEY,
    market TEXT NOT NULL,
    symbol TEXT NOT NULL,
    name   TEXT NOT NULL,
    isin   TEXT,
    UNIQUE (market, symbol)
);

CREATE TABLE ohlcva_per_min (
    sec_id    UUID NOT NULL,
    ts   TIMESTAMP NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume BIGINT  NOT NULL,
    amount NUMERIC NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE INDEX idx_ohlcva_per_min_sec_id_ts ON ohlcva_per_min (sec_id, ts);

CREATE TABLE ohlcva_per_30min (
    sec_id    UUID NOT NULL,
    ts   TIMESTAMP NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume BIGINT  NOT NULL,
    amount NUMERIC NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE INDEX idx_ohlcva_per_30min_sec_id_ts ON ohlcva_per_30min (sec_id, ts);

CREATE TABLE ohlcva_per_day (
    sec_id    UUID NOT NULL,
    ts   TIMESTAMP NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume BIGINT  NOT NULL,
    amount NUMERIC NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE INDEX idx_ohlcva_per_day_sec_id_ts ON ohlcva_per_day (sec_id, ts);
