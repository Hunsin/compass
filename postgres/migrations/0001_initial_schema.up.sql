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

CREATE TABLE ohlcva_per_min (
    sec_id    UUID NOT NULL,
    ts   TIMESTAMP NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume  BIGINT NOT NULL,
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
    volume  BIGINT NOT NULL,
    amount NUMERIC NOT NULL,
    UNIQUE (sec_id, ts),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (ts);

CREATE INDEX idx_ohlcva_per_30min_sec_id_ts ON ohlcva_per_30min (sec_id, ts);

CREATE TABLE ohlcva_per_day (
    sec_id    UUID NOT NULL,
    date      DATE NOT NULL,
    open   NUMERIC NOT NULL,
    high   NUMERIC NOT NULL,
    low    NUMERIC NOT NULL,
    close  NUMERIC NOT NULL,
    volume  BIGINT NOT NULL,
    amount NUMERIC NOT NULL,
    UNIQUE (sec_id, date),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
) PARTITION BY RANGE (date);

CREATE INDEX idx_ohlcva_per_day_sec_id_date ON ohlcva_per_day (sec_id, date);
