"""
Shared helpers for the Statistics service example scripts:
loads the buf-generated `statistics.v1` Python package (working around
the stdlib `statistics` shadowing) and exposes constants/utilities.
"""

import collections
import importlib.util
import sys
import types
from datetime import date, datetime, timezone
from pathlib import Path

_GEN_DIR = (
    Path(__file__).resolve().parent.parent.parent / "protocols" / "gen" / "python"
)


def _load_proto_module(qualname: str, path: Path):
    spec = importlib.util.spec_from_file_location(qualname, path)
    mod = importlib.util.module_from_spec(spec)
    sys.modules[qualname] = mod
    spec.loader.exec_module(mod)
    return mod


# Pre-register `statistics` and `statistics.v1` as namespace packages so
# statistics_pb2_grpc's `from statistics.v1 import statistics_pb2` resolves.
_stats_pkg = types.ModuleType("statistics")
_stats_pkg.__path__ = [str(_GEN_DIR / "statistics")]
sys.modules["statistics"] = _stats_pkg
_stats_v1_pkg = types.ModuleType("statistics.v1")
_stats_v1_pkg.__path__ = [str(_GEN_DIR / "statistics" / "v1")]
sys.modules["statistics.v1"] = _stats_v1_pkg

statistics_pb2 = _load_proto_module(
    "statistics.v1.statistics_pb2",
    _GEN_DIR / "statistics" / "v1" / "statistics_pb2.py",
)
_stats_v1_pkg.statistics_pb2 = statistics_pb2
statistics_pb2_grpc = _load_proto_module(
    "statistics.v1.statistics_pb2_grpc",
    _GEN_DIR / "statistics" / "v1" / "statistics_pb2_grpc.py",
)

import grpc  # noqa: E402

from google.protobuf import timestamp_pb2  # noqa: E402

EXCHANGE = "twse"
SYMBOLS = ("2317", "2330", "2454")
DEFAULT_SERVER = "localhost:50178"


def to_timestamp(d: date) -> timestamp_pb2.Timestamp:
    """Convert a date to a UTC midnight Timestamp."""
    ts = timestamp_pb2.Timestamp()
    ts.FromDatetime(datetime(d.year, d.month, d.day, tzinfo=timezone.utc))
    return ts


_ClientCallDetails = collections.namedtuple(
    "_ClientCallDetails",
    ("method", "timeout", "metadata", "credentials", "wait_for_ready", "compression"),
)


class _AuthInterceptor(
    grpc.UnaryUnaryClientInterceptor, grpc.UnaryStreamClientInterceptor
):
    """Injects an Authorization header into every gRPC call."""

    def __init__(self, token: str):
        self._metadata = (("authorization", f"Bearer {token}"),)

    def _attach(self, client_call_details):
        metadata = list(client_call_details.metadata or [])
        metadata.extend(self._metadata)
        new_details = _ClientCallDetails(
            client_call_details.method,
            client_call_details.timeout,
            metadata,
            getattr(client_call_details, "credentials", None),
            getattr(client_call_details, "wait_for_ready", None),
            getattr(client_call_details, "compression", None),
        )
        return new_details

    def intercept_unary_unary(self, continuation, client_call_details, request):
        return continuation(self._attach(client_call_details), request)

    def intercept_unary_stream(self, continuation, client_call_details, request):
        return continuation(self._attach(client_call_details), request)


def create_channel(server: str, token: str | None = None) -> grpc.Channel:
    """Create a gRPC channel, optionally attaching a Bearer token to every call."""
    channel = grpc.insecure_channel(server)
    if token is None:
        return channel
    return grpc.intercept_channel(channel, _AuthInterceptor(token))
