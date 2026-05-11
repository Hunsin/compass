"""
Log in to the API service via Keycloak and print the returned tokens.

The API service exposes an HTTP gateway (grpc-gateway) with a public
``POST /api/login`` endpoint that proxies the Keycloak token endpoint.

Usage:
    python examples/api/login.py
    python examples/api/login.py --server http://localhost:50189
    python examples/api/login.py --username testuser --password 123456

Prerequisites:
    pip install requests
"""

import argparse
import json
import sys

import requests

DEFAULT_SERVER = "http://localhost:50189"
DEFAULT_USERNAME = "testuser"
DEFAULT_PASSWORD = "123456"


def main():
    parser = argparse.ArgumentParser(
        description="Log in to the API service and print OAuth2 tokens"
    )
    parser.add_argument(
        "--server", default=DEFAULT_SERVER, help="HTTP gateway address"
    )
    parser.add_argument(
        "--username", default=DEFAULT_USERNAME, help="Keycloak username"
    )
    parser.add_argument(
        "--password", default=DEFAULT_PASSWORD, help="Keycloak password"
    )
    args = parser.parse_args()

    url = f"{args.server}/api/login"
    payload = {"username": args.username, "password": args.password}

    resp = requests.post(url, json=payload, timeout=10)
    resp.raise_for_status()

    data = resp.json()
    print(json.dumps(data, indent=2))

    # Print a quick hint for using the token in subsequent requests.
    token = data.get("access_token", data.get("accessToken", ""))
    if token:
        short = token[:20] + "..." if len(token) > 20 else token
        print(f"\nAccess token (truncated): {short}")
        print(
            "\nUse the full access_token value as:\n"
            '  curl -H "Authorization: Bearer <ACCESS_TOKEN>" '
            f"{args.server}/api/me"
        )


if __name__ == "__main__":
    try:
        main()
    except requests.HTTPError as e:
        print(f"HTTP error: {e}", file=sys.stderr)
        if e.response is not None:
            print(e.response.text, file=sys.stderr)
        sys.exit(1)
