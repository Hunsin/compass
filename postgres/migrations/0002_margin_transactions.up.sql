CREATE TABLE margin_transactions (
    sec_id              uuid   NOT NULL,
    date                date   NOT NULL,
    margin_purchase     bigint NOT NULL,
    margin_sales        bigint NOT NULL,
    cash_redemption     bigint NOT NULL,
    margin_balance      bigint NOT NULL,
    short_covering      bigint NOT NULL,
    short_sale          bigint NOT NULL,
    stock_redemption    bigint NOT NULL,
    short_balance       bigint NOT NULL,
    margin_short_offset bigint NOT NULL,
    UNIQUE (sec_id, date),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE INDEX idx_margin_transactions_sec_id_date ON margin_transactions (sec_id, date);
