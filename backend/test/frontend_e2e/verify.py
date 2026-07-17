"""Black-box API check for the Android P0 authentication and confirmation path.

Only test credentials are used. The script deliberately never prints phone
numbers, verification codes, access tokens, or refresh tokens.
"""

import json
import os
import sys
from urllib.parse import urlencode
from urllib.request import Request, urlopen
from urllib.error import HTTPError


BASE_URL = os.environ.get("FRONTEND_E2E_BASE_URL", "http://127.0.0.1:18080") + "/api/v1"
SMS_CODE = os.environ.get("E2E_SMS_CODE", "123456")
DEVICE_ID = "frontend-e2e-device-0001"
IDEMPOTENCY_KEY = "frontend-e2e-confirmation-0001"


def request(method, path, body=None, headers=None, expected=200):
    encoded = None if body is None else json.dumps(body).encode("utf-8")
    merged_headers = {"Accept": "application/json"}
    if encoded is not None:
        merged_headers["Content-Type"] = "application/json"
    if headers:
        merged_headers.update(headers)
    req = Request(BASE_URL + path, data=encoded, headers=merged_headers, method=method)
    try:
        with urlopen(req, timeout=10) as response:
            payload = None if response.status == 204 else json.load(response)
            if response.status != expected:
                raise AssertionError(f"unexpected status {response.status} for {method} {path}")
            return payload
    except HTTPError as error:
        payload = json.load(error)
        if error.code != expected:
            raise AssertionError(f"unexpected status {error.code} for {method} {path}") from error
        return payload


def main():
    send = request("POST", "/auth/sms/send", {"phone": "13000000000", "deviceId": DEVICE_ID}, expected=202)
    challenge_id = send["data"]["challengeId"]
    verified = request("POST", "/auth/sms/verify", {"challengeId": challenge_id, "code": SMS_CODE, "deviceId": DEVICE_ID})
    tokens = verified["data"]
    access = tokens["accessToken"]
    refresh = tokens["refreshToken"]

    query = urlencode({
        "latitude": "39.908823", "longitude": "116.397470", "coordinateSystem": "gcj02",
        "category": "toilet", "radiusMeters": "5000", "limit": "20",
    })
    places = request("GET", f"/places/nearby?{query}")
    place_id = places["data"]["items"][0]["id"]
    authorization = {"Authorization": f"Bearer {access}", "Idempotency-Key": IDEMPOTENCY_KEY}
    first = request("POST", f"/places/{place_id}/confirmations", {"kind": "exists", "result": "confirmed"}, authorization, expected=201)
    repeated = request("POST", f"/places/{place_id}/confirmations", {"kind": "exists", "result": "confirmed"}, authorization, expected=201)
    if first["data"]["id"] != repeated["data"]["id"]:
        raise AssertionError("idempotent confirmation did not return its original record")

    conflict = request("POST", f"/places/{place_id}/confirmations", {"kind": "opening_hours", "result": "confirmed"}, authorization, expected=409)
    if conflict.get("error", {}).get("code") != "idempotency_conflict":
        raise AssertionError("different confirmation payload must report idempotency_conflict")

    request("POST", "/auth/logout", {"refreshToken": refresh}, expected=204)
    denied = request("POST", f"/places/{place_id}/confirmations", {"kind": "exists", "result": "confirmed"}, authorization, expected=401)
    if denied.get("error", {}).get("code") not in {"invalid_session", "invalid_token"}:
        raise AssertionError("revoked session write must be rejected")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"frontend P0 e2e failed: {exc}", file=sys.stderr)
        sys.exit(1)
    print("frontend P0 e2e passed")
