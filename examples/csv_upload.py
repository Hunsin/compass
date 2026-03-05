"""
Upload 1-minute OHLCV data from a CSV file to the Quote gRPC service.

CSV format (header: ts,Open,Close,High,Volume,Low):
    2025-12-11 09:01:00,1510.0,1510.0,1515.0,2590,1505.0

The CSV filename must follow the pattern <symbol>_<date>.csv.
Timestamps in the CSV are treated as local exchange time and stored as-is.

Usage:
    python examples/csv_upload/upload.py [--server ADDR] <csv_file>

    # default server: localhost:50168
    python examples/csv_upload/upload.py examples/csv_upload/2330_2025-12-11.csv

Prerequisites:
    pip install grpcio protobuf
"""

import argparse
import csv
import sys
from datetime import datetime, timezone
from pathlib import Path

# Add the buf-generated Python output directory to sys.path so that the
# generated files can resolve their own internal imports (e.g. quote_pb2_grpc
# does "from quote import quote_pb2").
sys.path.insert(0, str(Path(__file__).parent.parent / "protocols" / "gen" / "python"))

import grpc
from google.protobuf import duration_pb2, timestamp_pb2
from quote import quote_pb2, quote_pb2_grpc

EXCHANGE = "twse"
INTERVAL_1M = duration_pb2.Duration(seconds=60)


def parse_csv(path: Path) -> list:
    """Read the CSV and return a list of OHLCV proto messages."""
    rows = []
    with open(path, newline="") as f:
        reader = csv.DictReader(f)
        for row in reader:
            # Timestamps in the CSV are local exchange time.
            # Replace tzinfo with UTC so that the proto Timestamp seconds
            # encode the wall-clock value directly; the server stores it as
            # a TIMESTAMP (no time zone) column, preserving the local time.
            dt = datetime.strptime(row["ts"], "%Y-%m-%d %H:%M:%S").replace(
                tzinfo=timezone.utc
            )
            ts = timestamp_pb2.Timestamp()
            ts.FromDatetime(dt)
            rows.append(
                quote_pb2.OHLCV(
                    ts=ts,
                    open=float(row["Open"]),
                    high=float(row["High"]),
                    low=float(row["Low"]),
                    close=float(row["Close"]),
                    volume=int(row["Volume"]),
                )
            )
    return rows


def main():
    parser = argparse.ArgumentParser(description="Upload OHLCV CSV to Quote service")
    parser.add_argument("csv_file", help="Path to the CSV file")
    parser.add_argument(
        "--server", default="localhost:50168", help="gRPC server address"
    )
    args = parser.parse_args()

    csv_path = Path(args.csv_file)
    if not csv_path.exists():
        print(f"error: file not found: {csv_path}", file=sys.stderr)
        sys.exit(1)

    # Derive symbol from filename: "2330_2025-12-11.csv" → "2330"
    symbol = csv_path.stem.split("_")[0]

    ohlcv = parse_csv(csv_path)
    print(f"Read {len(ohlcv)} rows from {csv_path.name}")

    with grpc.insecure_channel(args.server) as channel:
        stub = quote_pb2_grpc.QuoteServiceStub(channel)
        stub.CreateOHLCVs(
            quote_pb2.CreateOHLCVsRequest(
                exchange=EXCHANGE,
                symbol=symbol,
                interval=INTERVAL_1M,
                ohlcv=ohlcv,
            )
        )
    print(f"Uploaded {len(ohlcv)} 1m bars for {EXCHANGE}:{symbol}")


if __name__ == "__main__":
    try:
        main()
    except grpc.RpcError as e:
        print(f"RPC error [{e.code()}]: {e.details()}", file=sys.stderr)
        sys.exit(1)
