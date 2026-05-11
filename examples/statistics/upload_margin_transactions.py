"""
Fetch daily stock margin transactions from TWSE for selected symbols
(2317, 2330, 2454) over a date range and upload them to the Statistics
gRPC service.

Usage:
    python examples/statistics/upload_margin_transactions.py [--server ADDR]

    # default server: localhost:50178
    python examples/statistics/upload_margin_transactions.py --server localhost:50178

Prerequisites:
    pip install grpcio protobuf requests
"""

import argparse
import csv
import io
import sys
import time
from datetime import date, timedelta
from pathlib import Path

# Make `_common` importable regardless of the current working directory.
sys.path.insert(0, str(Path(__file__).resolve().parent))

import requests
import urllib3

from _common import (
    DEFAULT_SERVER,
    EXCHANGE,
    SYMBOLS,
    create_channel,
    statistics_pb2,
    statistics_pb2_grpc,
    to_timestamp,
)

# TWSE's TLS cert lacks the Subject Key Identifier extension, which newer
# Python (3.14+) SSL stacks reject. The endpoint serves public market data,
# so it's safe to skip cert verification here.
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

TWSE_URL = "https://www.twse.com.tw/rwd/en/marginTrading/MI_MARGN"
# Delay between consecutive TWSE requests to avoid rate-limit throttling.
TWSE_REQUEST_INTERVAL = 2.0


def to_int(s: str) -> int:
    return int(s.replace(",", "").strip())


def fetch_margin(d: date) -> dict:
    """Fetch the TWSE daily margin trading CSV and return a dict
    mapping symbol -> MarginTransaction for the symbols of interest."""
    params = {
        "date": d.strftime("%Y%m%d"),
        "selectType": "STOCK",
        "response": "csv",
    }
    resp = requests.get(TWSE_URL, params=params, timeout=30, verify=False)
    resp.raise_for_status()
    # The CSV header is encoded in MS-950 (Big5); the numeric body is ASCII.
    text = resp.content.decode("ms950", errors="replace")
    reader = csv.reader(io.StringIO(text))

    wanted = set(SYMBOLS)
    out = {}
    # CSV columns (0-indexed):
    #   0  Security Code
    #   1  Margin Purchase
    #   2  Margin Sales
    #   3  Cash Redemption
    #   4  Margin Balance of Previous Day
    #   5  Margin Balance of the Day
    #   6  Margin Quota for the Next Day  (unused)
    #   7  Short Covering
    #   8  Short Sale
    #   9  Stock Redemption
    #   10 Short Balance of Previous Day
    #   11 Short Balance of the Day
    #   12 Short Quota for the Next Day   (unused)
    #   13 Offsetting of Margin Purchases and Short Sales
    for row in reader:
        if len(row) < 14:
            continue
        code = row[0].strip().strip('"')
        if code not in wanted:
            continue
        try:
            mt = statistics_pb2.MarginTransaction(
                margin_purchase=to_int(row[1]),
                margin_sales=to_int(row[2]),
                cash_redemption=to_int(row[3]),
                margin_balance=to_int(row[5]),
                short_covering=to_int(row[7]),
                short_sale=to_int(row[8]),
                stock_redemption=to_int(row[9]),
                short_balance=to_int(row[11]),
                margin_short_offset=to_int(row[13]),
            )
        except ValueError:
            continue
        out[code] = mt
    return out


def upload(stub, d: date, txs: dict) -> None:
    req = statistics_pb2.CreateMarginTransactionsRequest(
        exchange=EXCHANGE,
        date=to_timestamp(d),
        margin_transactions=txs,
    )
    stub.CreateMarginTransactions(req)


def main():
    parser = argparse.ArgumentParser(
        description="Upload TWSE margin transactions to the Statistics service"
    )
    parser.add_argument("--server", default=DEFAULT_SERVER, help="gRPC server address")
    parser.add_argument("--token", default=None, help="Keycloak access token (Bearer)")
    parser.add_argument(
        "--start", default="2026-05-04", help="start date (YYYY-MM-DD, inclusive)"
    )
    parser.add_argument(
        "--end", default="2026-05-08", help="end date (YYYY-MM-DD, inclusive)"
    )
    args = parser.parse_args()

    start = date.fromisoformat(args.start)
    end = date.fromisoformat(args.end)

    with create_channel(args.server, args.token) as channel:
        stub = statistics_pb2_grpc.StatisticsServiceStub(channel)

        d = start
        first = True
        while d <= end:
            if not first:
                time.sleep(TWSE_REQUEST_INTERVAL)
            first = False
            txs = fetch_margin(d)
            if txs:
                upload(stub, d, txs)
                print(f"{d}: uploaded {len(txs)} symbols ({sorted(txs)})")
            else:
                print(f"{d}: no data for selected symbols (market closed?)")
            d += timedelta(days=1)


if __name__ == "__main__":
    import grpc

    try:
        main()
    except grpc.RpcError as e:
        print(f"RPC error [{e.code()}]: {e.details()}", file=sys.stderr)
        sys.exit(1)
