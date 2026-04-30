CREATE TABLE margin_transactions (
    sec_id                         uuid   NOT NULL,
    date                           date   NOT NULL,
    margin_purchase_buy            bigint NOT NULL,
    margin_purchase_redemption     bigint NOT NULL,
    margin_purchase_cash_repayment bigint NOT NULL,
    margin_purchase_balance        bigint NOT NULL,
    margin_purchase_limit          bigint NOT NULL,
    short_sale                     bigint NOT NULL,
    short_sale_redemption          bigint NOT NULL,
    short_sale_stock_repayment     bigint NOT NULL,
    short_sale_balance             bigint NOT NULL,
    short_sale_limit               bigint NOT NULL,
    quota_next_day                 bigint NOT NULL,
    UNIQUE (sec_id, date),
    FOREIGN KEY (sec_id) REFERENCES securities(id) ON DELETE CASCADE
);

CREATE INDEX idx_margin_transactions_sec_id_date ON margin_transactions (sec_id, date);
