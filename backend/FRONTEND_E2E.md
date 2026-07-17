# Android P0 frontend API verification

This check validates the contract the Android client relies on without changing
the normal local `.env` or the existing Compose stack.

```powershell
docker compose -p alongtu-frontend-e2e -f docker-compose.yml -f docker-compose.frontend-test.yml up -d --build
$env:FRONTEND_E2E_BASE_URL = 'http://127.0.0.1:18080'
$env:E2E_SMS_CODE = '123456'
python test/frontend_e2e/verify.py
docker compose -p alongtu-frontend-e2e -f docker-compose.yml -f docker-compose.frontend-test.yml down -v
```

The override has its own Docker project, volumes, and host ports: API 18080,
PostGIS 15432, and Redis 16379. It replaces the base service `env_file`, then
sets `APP_ENV=test`, a fake map key and a fixed test code only inside that API
container. It never reads or changes the standard `.env` file.

The black-box check covers the client-facing envelopes for SMS challenge
creation and verification, nearby place retrieval, confirmation idempotency,
idempotency conflicts, logout, and rejection of writes with a revoked session.
It never prints phone numbers, verification codes, access tokens, or refresh
tokens.
