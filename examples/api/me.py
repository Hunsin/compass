"""
Verify the current user's identity via the API service.

Calls ``GET /api/me`` with a Bearer token obtained from ``login.py``
(or any other means) and prints the authenticated user's ID.

Usage:
    # Pass the token directly
    python examples/api/me.py --token <ACCESS_TOKEN>

    # Or pipe the token from login.py via jq
    python examples/api/login.py | jq -r .access_token | xargs -I {} python examples/api/me.py --token {}

Prerequisites:
    pip install requests
"""

import argparse
import json
import sys

import requests

DEFAULT_SERVER = "http://localhost:50189"


def main():
    parser = argparse.ArgumentParser(
        description="Verify the current user identity via GET /api/me"
    )
    parser.add_argument(
        "--server", default=DEFAULT_SERVER, help="HTTP gateway address"
    )
    parser.add_argument(
        "--token", required=True, help="Bearer access token from /api/login"
    )
    args = parser.parse_args()

    url = f"{args.server}/api/me"
    headers = {"Authorization": f"Bearer {args.token}"}

    resp = requests.get(url, headers=headers, timeout=10)
    resp.raise_for_status()

    data = resp.json()
    print(json.dumps(data, indent=2))


if __name__ == "__main__":
    try:
        main()
    except requests.RequestException as e:
        print(f"Request error: {e}", file=sys.stderr)
        if getattr(e, "response", None) is not None:
            print(e.response.text, file=sys.stderr)
        sys.exit(1)
