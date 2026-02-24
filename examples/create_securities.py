"""
Create the TWSE exchange and a set of securities via the Quote gRPC service.

Usage:
    python examples/create_securities.py
    python examples/create_securities.py --server localhost:50168

Prerequisites:
    pip install grpcio protobuf
"""

import argparse
import sys
from pathlib import Path

# Add the buf-generated Python output directory to sys.path so that the
# generated files can resolve their own internal imports (e.g. quote_pb2_grpc
# does "from quote import quote_pb2").
sys.path.insert(0, str(Path(__file__).parent.parent / "protocols" / "gen" / "python"))

import grpc
from google.protobuf import empty_pb2
from quote import quote_pb2, quote_pb2_grpc

DEFAULT_SERVER = "localhost:50168"

def main():
    parser = argparse.ArgumentParser(description="Create TWSE exchange and securities")
    parser.add_argument("--server", default=DEFAULT_SERVER, help="gRPC server address")
    args = parser.parse_args()

    with grpc.insecure_channel(args.server) as channel:
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
