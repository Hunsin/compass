"""
Read margin transactions back from the Statistics gRPC service for the
selected symbols (2317, 2330, 2454) over a date range and print them as
a table to verify what was stored.

Usage:
    python examples/statistics/get_margin_transactions.py [--server ADDR]

    # default server: localhost:50178
    python examples/statistics/get_margin_transactions.py --server localhost:50178

Prerequisites:
    pip install grpcio protobuf
"""

import argparse
import sys
from datetime import date, timedelta
from pathlib import Path

# Make `_common` importable regardless of the current working directory.
sys.path.insert(0, str(Path(__file__).resolve().parent))

from _common import (
    DEFAULT_SERVER,
    EXCHANGE,
    SYMBOLS,
    create_channel,
    statistics_pb2,
    statistics_pb2_grpc,
    to_timestamp,
)


def fetch_back(stub, symbol: str, start: date, end: date):
    # 'from' is a Python reserved word; pass it via dict-unpacking.
    req = statistics_pb2.GetMarginTransactionsRequest(
        exchange=EXCHANGE,
        symbol=symbol,
        before=to_timestamp(end + timedelta(days=1)),
        **{"from": to_timestamp(start)},
    )
    return list(stub.GetMarginTransactions(req))


HEADERS = (
    "date",
    "mgn_purchase",
    "mgn_sales",
    "cash_redem",
    "mgn_balance",
    "shrt_cover",
    "shrt_sale",
    "stk_redem",
    "shrt_balance",
    "offset",
)


def fmt_row(cells, widths):
    return "  " + " | ".join(
        str(c).rjust(w) if i > 0 else str(c).ljust(w)
        for i, (c, w) in enumerate(zip(cells, widths))
    )


def print_table(symbol: str, rows):
    print(f"\n=== {EXCHANGE}:{symbol} ===")
    if not rows:
        print("  (no rows)")
        return
    widths = [max(len(h), 11) for h in HEADERS]
    widths[0] = 10  # date column
    print(fmt_row(HEADERS, widths))
    print("  " + "-+-".join("-" * w for w in widths))
    for tx in rows:
        print(
            fmt_row(
                (
                    tx.date.ToDatetime().date().isoformat(),
                    tx.margin_purchase,
                    tx.margin_sales,
                    tx.cash_redemption,
                    tx.margin_balance,
                    tx.short_covering,
                    tx.short_sale,
                    tx.stock_redemption,
                    tx.short_balance,
                    tx.margin_short_offset,
                ),
                widths,
            )
        )


def main():
    parser = argparse.ArgumentParser(
        description="Read margin transactions from the Statistics service and print them"
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
        for symbol in SYMBOLS:
            print_table(symbol, fetch_back(stub, symbol, start, end))


if __name__ == "__main__":
    import grpc

    try:
        main()
    except grpc.RpcError as e:
        print(f"RPC error [{e.code()}]: {e.details()}", file=sys.stderr)
        sys.exit(1)
