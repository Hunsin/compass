"""
Example: using the QuoteService gRPC client.

Prerequisites:
    pip install grpcio grpcio-tools protobuf

Run (assumes the server is listening on localhost:50168):
    python protocols/example.py
"""

import sys
from pathlib import Path

# Add the buf-generated Python output directory to sys.path so that the
# generated files can resolve their own internal imports (e.g. quote_pb2_grpc
# does "from quote import quote_pb2").
sys.path.insert(0, str(Path(__file__).parent.parent / "protocols" / "gen" / "python"))

import grpc
from datetime import datetime, timezone, timedelta
from google.protobuf import empty_pb2, timestamp_pb2, duration_pb2
from quote import quote_pb2, quote_pb2_grpc

SERVER_ADDR = "localhost:50168"


def make_duration(minutes: int = 0, days: int = 0) -> duration_pb2.Duration:
    total_seconds = minutes * 60 + days * 86400
    return duration_pb2.Duration(seconds=total_seconds)


def make_timestamp(dt: datetime) -> timestamp_pb2.Timestamp:
    ts = timestamp_pb2.Timestamp()
    ts.FromDatetime(dt)
    return ts


def main():
    with grpc.insecure_channel(SERVER_ADDR) as channel:
        stub = quote_pb2_grpc.QuoteServiceStub(channel)

        # --- CreateExchange ---
        exchange = quote_pb2.Exchange(
            abbr="twse",
            name="Taiwan Stock Exchange",
            timezone="Asia/Taipei",
        )
        stub.CreateExchange(exchange)
        print(f"Created exchange: {exchange.abbr}")

        # --- GetExchanges ---
        print("\nAll exchanges:")
        for ex in stub.GetExchanges(empty_pb2.Empty()):
            print(f"  {ex.abbr}: {ex.name} ({ex.timezone})")

        # --- CreateSecurities (client-streaming) ---
        def security_stream():
            for symbol, name in [("2330", "Taiwan Semiconductor Manufacturing"), ("2317", "Hon Hai"), ("2454", "MediaTek")]:
                yield quote_pb2.Security(exchange="twse", symbol=symbol, name=name)

        stub.CreateSecurities(security_stream())
        print("Created securities: 2330, 2317, 2454")

        # --- GetSecurities ---
        print("\nSecurities in twse:")
        for sec in stub.GetSecurities(quote_pb2.Exchange(abbr="twse")):
            print(f"  {sec.symbol}: {sec.name}")


if __name__ == "__main__":
    try:
        main()
    except grpc.RpcError as e:
        print(f"RPC error [{e.code()}]: {e.details()}", file=sys.stderr)
        sys.exit(1)
