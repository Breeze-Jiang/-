# Local Goose Migration Image Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make P0 Docker Compose migrations work without GHCR credentials while preserving the migration-before-API startup contract.

**Architecture:** Build a pinned Goose executable in a dedicated Dockerfile stage based on the existing Go builder image. Compose builds that stage locally for the one-shot migration service; it retains the existing migration command, read-only SQL mount, and PostGIS health dependency.

**Tech Stack:** Docker multi-stage builds, Docker Compose, Go 1.24 Alpine, `github.com/pressly/goose/v3` v3.24.1, PostgreSQL/PostGIS, Redis.

## Global Constraints

- P0 scope only; do not add forum, chat, annotation, plan, or real-provider features.
- Do not add GitHub Container Registry or other credentials.
- Pin Goose to `v3.24.1`.
- Preserve the existing migration SQL and API startup dependency on successful migration.
- Do not commit `backend/.env` or real secrets.

---

### Task 1: Add a deterministic local migration image

**Files:**
- Modify: `backend/Dockerfile`
- Modify: `backend/docker-compose.yml`

**Interfaces:**
- Consumes: `db/migrations/*.sql` mounted at `/migrations`.
- Produces: Docker target `migrate` whose entrypoint is `/goose`.
- Consumed by: Compose service `migrate`, with its existing `goose -dir /migrations postgres <url> up` arguments.

- [x] **Step 1: Add the failing build check**

Run:

```powershell
Set-Location backend
docker compose build migrate
```

Expected before implementation: FAIL because Compose attempts to pull `ghcr.io/pressly/goose:3.24.1` and the registry denies anonymous access.

- [x] **Step 2: Add the pinned Goose build stage and minimal runtime target**

Modify `backend/Dockerfile` so the top of the file becomes:

```dockerfile
FROM golang:1.24-alpine AS goose-build
RUN GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

FROM alpine:3.21 AS migrate
COPY --from=goose-build /out/goose /goose
ENTRYPOINT ["/goose"]

FROM golang:1.24-alpine AS build
```

Keep the existing API build and final `distroless` stages unchanged after the new stages.

Change the Compose migration service image field to this build definition:

```yaml
  migrate:
    build:
      context: .
      target: migrate
```

Keep the existing `command`, `volumes`, and `depends_on` fields exactly as they are.

- [x] **Step 3: Verify the migration image builds locally**

Run:

```powershell
Set-Location backend
docker compose build migrate
docker compose run --rm migrate --version
```

Expected: both commands exit 0 and the version output begins with `goose version: v3.24.1`.

- [ ] **Step 4: Commit the isolated configuration change**

```powershell
git add backend/Dockerfile backend/docker-compose.yml
git commit -m "build: construct local goose migration image"
```

Only commit if the repository has an initialized branch and the user has authorized commits; otherwise leave changes unstaged and report that fact.

### Task 2: Run the complete Compose startup and database acceptance checks

**Files:**
- Uses: `backend/docker-compose.yml`
- Uses: `backend/db/migrations/00001_p0_schema.sql`
- Uses: `backend/db/migrations/00002_p0_place_feedback.sql`

**Interfaces:**
- Consumes: the Task 1 `migrate` Docker target and local ignored `backend/.env`.
- Produces: running `api`, `postgres`, `redis`, and `fake-provider` containers; successfully completed `migrate` container.

- [x] **Step 1: Start the stack**

Run:

```powershell
Set-Location backend
docker compose up -d --build
docker compose ps
```

Expected: `postgres` and `redis` are healthy; `migrate` has exit code 0; `api` and `fake-provider` are running.

- [x] **Step 2: Verify migration history and required extensions**

Run:

```powershell
docker compose exec -T postgres psql -U alongtu -d alongtu -c "SELECT version_id, is_applied FROM goose_db_version ORDER BY id;"
docker compose exec -T postgres psql -U alongtu -d alongtu -c "SELECT extname FROM pg_extension WHERE extname IN ('postgis','pgcrypto') ORDER BY extname;"
```

Expected: migration versions `1` and `2` are applied; both `postgis` and `pgcrypto` are returned.

- [x] **Step 3: Verify API process and dependency readiness**

Run:

```powershell
Invoke-WebRequest http://127.0.0.1:8080/health/live -UseBasicParsing
Invoke-WebRequest http://127.0.0.1:8080/health/ready -UseBasicParsing
```

Expected: both return HTTP 200. Record only status codes; do not log tokens, phone numbers, or exact user locations.

- [ ] **Step 4: Verify teardown is reversible**

Run:

```powershell
docker compose down
```

Expected: application containers and network stop; named database and Redis volumes remain intact unless explicitly removed by the user.

Status: deliberately not run; the validated stack remains available for continuing frontend/backend development.

- [ ] **Step 5: Commit documentation updates only if required**

If image-source behavior needs an acceptance-audit correction, update only the relevant backend deployment note and commit it separately after user authorization. Do not modify the product master plan for an implementation-only image-source change.

## Self-review

- Spec coverage: Task 1 removes GHCR dependency while preserving target version and Compose behavior; Task 2 verifies build, migration ordering, schema state, readiness, and reversible cleanup.
- Placeholder scan: no unfinished implementation placeholders are present.
- Interface consistency: the `migrate` target provides `/goose`, matching Compose's existing Goose command arguments.

## Execution Handoff

Plan saved to `docs/superpowers/plans/2026-07-16-local-goose-migration-image.md`.

Execution will remain inline in this session: the task is two tightly coupled configuration and verification steps, and the current collaboration rules do not authorize subagent delegation.
