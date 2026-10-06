# Redalert development guide

## Project layout

- `redalert.go` and `cmd/` define the executable and CLI commands.
- `core/` orchestrates checks and service lifecycle; `checks/` implements check types.
- `assertions/` evaluates check results and `backoffs/` schedules retries.
- `notifiers/` sends alerts; `storage/` stores events and `stats/` derives statistics.
- `config/` loads configuration, `web/` serves the dashboard and HTTP API, `ui/` contains its source, and `servicepb/` contains the RPC protocol and generated Go code.

## Toolchain and setup

Use Go 1.27.1 with Go modules. `go.mod` and `go.sum` are authoritative; `glide.yaml` and `glide.lock` are historical. Install Python 3 for local fixture and smoke scripts. Docker is only needed for optional integration checks. Ordinary builds use the committed dashboard assets and do not need Node. UI regeneration requires Node 20.19.0 and npm 10.8.2.

## Commands

- `go mod download` downloads Go dependencies.
- `make build` builds the versioned `redalert` executable.
- `make dev` rebuilds the current checkout and runs the local dashboard with a loopback demo fixture. Set `REDALERT_DEV_PORT`, `REDALERT_DEV_RPC_PORT`, and `REDALERT_DEV_FIXTURE_PORT` to override ports.
- `make check` verifies Go formatting without rewriting files, runs unit tests, builds packages and the metadata-bearing executable, then runs `scripts/smoke.py` against it. Generated `servicepb/service.pb.go` is excluded from formatting checks.
- `make test-integration` runs opt-in Docker tests separately; `make test-deps` fetches their fixtures.
- After UI source changes, run `make embed-static` and commit the updated `web/assets` files. `servicepb/service.pb.go` is generated with `make build-proto`.

## Completion criteria

Before completing Go changes, run `make check`. For UI changes, regenerate and commit embedded assets as well. Keep Docker-backed tests opt-in unless the change specifically requires integration coverage, and preserve the documented polling-race and SSH coverage caveats until those follow-up issues are addressed.
