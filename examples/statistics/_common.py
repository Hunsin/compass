"""
Shared helpers for the Statistics service example scripts:
loads the buf-generated `statistics.v1` Python package (working around
the stdlib `statistics` shadowing) and exposes constants/utilities.
"""

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

from google.protobuf import timestamp_pb2  # noqa: E402

EXCHANGE = "twse"
SYMBOLS = ("2317", "2330", "2454")
DEFAULT_SERVER = "localhost:50178"


def to_timestamp(d: date) -> timestamp_pb2.Timestamp:
    """Convert a date to a UTC midnight Timestamp."""
    ts = timestamp_pb2.Timestamp()
    ts.FromDatetime(datetime(d.year, d.month, d.day, tzinfo=timezone.utc))
    return ts
