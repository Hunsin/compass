"""
Read trading history of a TWSE security from the Quote gRPC service.

Timestamps in the DB are stored as local exchange time represented as UTC
(matching the convention used by csv_upload.py), so the query window is
built the same way: treat the desired local date boundaries as UTC.

Usage:
    # fetch 1m bars for 2025-12-11 (the date in the sample CSV)
    python examples/get_history.py --symbol 2330 --date 2025-12-11

    # fetch 1m bars of 2330 for today (default)
    python examples/get_history.py

    # fetch 30m bars for a date range
    python examples/get_history.py --symbol 2330 --date 2025-12-11 --interval 30m

Prerequisites:
    pip install grpcio protobuf
"""

import argparse
import sys
from datetime import date, datetime, timedelta, timezone
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent / "protocols" / "gen" / "python"))

import grpc
from google.protobuf import duration_pb2, timestamp_pb2
from quote import quote_pb2, quote_pb2_grpc

EXCHANGE = "twse"
SYMBOL = "2330"
SERVER_ADDR = "localhost:50168"

INTERVALS = {
    "1m":  duration_pb2.Duration(seconds=60),
    "5m":  duration_pb2.Duration(seconds=300),
    "30m": duration_pb2.Duration(seconds=1800),
    "1h":  duration_pb2.Duration(seconds=3600),
    "1d":  duration_pb2.Duration(seconds=86400),
    "1w":  duration_pb2.Duration(seconds=604800),
    "1M":  duration_pb2.Duration(seconds=2592000),
}


def date_to_ts(d: date) -> timestamp_pb2.Timestamp:
    """Encode a date's midnight as a Timestamp using the local-as-UTC convention."""
    dt = datetime(d.year, d.month, d.day, tzinfo=timezone.utc)
    ts = timestamp_pb2.Timestamp()
    ts.FromDatetime(dt)
    return ts


def main():
    parser = argparse.ArgumentParser(
        description=f"Fetch trading data for {EXCHANGE.upper()}"
    )
    parser.add_argument(
        "--symbol",
        default=SYMBOL,
        help="Trading symbol to query, default: 2330",
    )
    parser.add_argument(
        "--date",
        default=date.today().isoformat(),
        help="Trading date to query (YYYY-MM-DD), default: today",
    )
    parser.add_argument(
        "--interval",
        default="1m",
        choices=INTERVALS.keys(),
        help="Bar interval, default: 1m",
    )
    parser.add_argument("--server", default=SERVER_ADDR)
    args = parser.parse_args()

    trading_date = date.fromisoformat(args.date)
    from_ts = date_to_ts(trading_date)
    before_ts = date_to_ts(trading_date + timedelta(days=1))

    with grpc.insecure_channel(args.server) as channel:
        stub = quote_pb2_grpc.QuoteServiceStub(channel)
        # "from" is a Python keyword; use **{} unpacking to pass it by name.
        bars = list(
            stub.GetOHLCVAs(
                quote_pb2.GetOHLCVAsRequest(**{
                    "exchange": EXCHANGE,
                    "symbol": args.symbol,
                    "interval": INTERVALS[args.interval],
                    "from": from_ts,
                    "before": before_ts,
                })
            )
        )

    if not bars:
        print(f"No {args.interval} data for {EXCHANGE}:{SYMBOL} on {trading_date}")
        return

    print(f"{len(bars)} {args.interval} bars for {EXCHANGE}:{SYMBOL} on {trading_date}\n")
    print(f"{'Time':<8}  {'Open':>8}  {'High':>8}  {'Low':>8}  {'Close':>8}  {'Volume':>8}  {'Amount':>16}")
    print("-" * 76)
    for bar in bars:
        ts_str = datetime.fromtimestamp(bar.ts.seconds, tz=timezone.utc).strftime("%H:%M:%S")
        print(
            f"{ts_str:<8}  {bar.open:>8.1f}  {bar.high:>8.1f}  {bar.low:>8.1f}"
            f"  {bar.close:>8.1f}  {bar.volume:>8}  {bar.amount:>16,}"
        )


if __name__ == "__main__":
    try:
        main()
    except grpc.RpcError as e:
        print(f"RPC error [{e.code()}]: {e.details()}", file=sys.stderr)
        sys.exit(1)
