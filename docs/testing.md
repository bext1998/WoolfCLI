# Go Testing

This repository targets Go 1.22 or newer.

## Local test commands

On Windows PowerShell:

```powershell
.\scripts\test.ps1
```

On Windows, the script sets project-local Go workspace paths when they are not already configured:

- `GOCACHE` defaults to `.gocache`
- `GOTMPDIR` defaults to `.gotmp`

This keeps test runs independent from the user profile Go build cache, which can fail on some Windows machines because of local permissions. The directories are ignored by Git. If you need a custom cache or temporary directory, set `GOCACHE` or `GOTMPDIR` before running the script; explicit environment variables are preserved.

The script also initializes the PowerShell process to UTF-8 before running tests. This avoids mojibake when reading or printing UTF-8 project files such as `README.md` and `docs/woolf-spec.md` on Windows PowerShell. For ad hoc inspection commands in the same terminal, use:

```powershell
[Console]::InputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
Get-Content docs/woolf-spec.md -Encoding UTF8
```

Optional checks:

```powershell
.\scripts\test.ps1 -Vet
.\scripts\test.ps1 -Race
.\scripts\test.ps1 -Coverage
```

On shells with `make`:

```sh
make test
make test-vet
make test-race
make test-cover
```

## Continuous integration

GitHub Actions runs `go mod download`, `go test ./...`, and `go vet ./...` on Ubuntu and Windows for pushes to `main` or `master` and for pull requests.

## Focused smoke coverage

The smoke test in `internal/cli/root_test.go` verifies that the root CLI command still exposes the expected command surface without requiring an OpenRouter API key or user runtime data.

The `agents` command tests cover listing and showing built-in roles, adding a custom YAML role into the configured agents directory, loading it through the registry, and deleting it again.

The `start` command tests use a fake chat client so the orchestration path can be exercised without calling OpenRouter. These tests cover both the successful session path and the error path where agent/API failures must surface as command errors.

The config tests verify environment overrides for `OPENROUTER_API_KEY` and `WOOLF_SESSIONS_DIR`, plus API key masking behavior used by user-facing configuration output.

The OpenRouter client tests verify SSE parsing, missing API key handling, HTTP status to Woolf error-code mapping, `Retry-After` handling for 429 responses, and bounded retry behavior for 5xx responses. These tests use local HTTP test servers and do not call OpenRouter.

The orchestrator tests verify session persistence, cancellation behavior, stream error handling, and context propagation between agents. Stream errors should persist the skipped response, mark the session as `error`, and emit an error event instead of being reported as a completed run. Context builder tests verify that draft content, session summaries, user interventions, focus ranges, role prompt metadata, stance tags, and previous agent responses are included in the messages sent to the chat client.
