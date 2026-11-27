# AGENTS.md

This file defines the working rules for coding agents in this repository.

## Scope

- Applies to the entire repository.
- If a deeper `AGENTS.md` is added later, the deeper file overrides this one for its subtree.

## Repository overview

- This is a Go CLI/server project for controlling home devices and related integrations.
- Primary entrypoints are `main.go`, `server/`, `subcommand/`, and `internal/`.
- The canonical project documentation is `README.md`.
- `docs/` should stay minimal. Add durable architecture notes there only when `README.md` is no longer the right place.

## Important directories

- `subcommand/`: user-facing CLI commands, config loading, and command wiring
- `internal/`: internal packages and implementation details
- `server/`: server-side behavior such as Slack/socket mode handling
- `config/`: sample configuration files
- `docs/`: supplementary design/architecture notes
- `tools/dspy-resolver/`: production HTTP resolver
- `tools/dspy/`: offline resolver optimization and evaluation
- `tools/dspy_common/`: shared DSPy program, model settings, and artifact format

## Working rules

- Keep changes narrowly scoped to the user request.
- Prefer existing patterns over introducing new abstractions.
- Follow the [Action.Run tracing convention](docs/architecture.md#action-のトレース方針) when adding an Action.
- Do not treat legacy prompt documents as source of truth; follow `README.md` and current code.
- Preserve uncommitted user changes outside the files you need to touch.
- When updating behavior that affects documentation or configuration, update `README.md` in the same change when appropriate.

## Build, test, and lint

Use the same quality gates as CI where practical:

- `go build ./...`
- `go test ./...`
- `golangci-lint run`

CI also runs:

- `gosec -exclude=G404 ./...`
- `govulncheck ./...`

Run the heavier checks when your change touches security-sensitive or dependency-sensitive areas.

For Python changes, follow the [CI workflow](.github/workflows/build.yml). The `dspy-resolver-test` matrix runs `pip check` and pytest with the batch and service dependencies. The `resolver-tools-test` job covers extraction, review, and synthetic data generation.

## Go conventions

- Run `gofmt` on changed Go files.
- Keep package structure simple; prefer small focused functions over speculative abstractions.
- Add or update tests for behavior changes when feasible.
- Avoid introducing new dependencies unless they are justified by the task.

## Configuration and secrets

- Never commit real tokens or secrets.
- Prefer environment-variable overrides for sensitive values such as:
  - `SMARTHOME_SWITCHBOT_TOKEN`
  - `SMARTHOME_SWITCHBOT_SECRET`
  - `SMARTHOME_LLM_API_KEY`
  - `SMARTHOME_INFLUXDB_TOKEN`
  - `SMARTHOME_SLACK_APP_TOKEN`
  - `SMARTHOME_SLACK_BOT_TOKEN`
- Treat `config/*.sample` as templates only; keep examples safe and generic.
- Keep real utterances, reviewed production data, and reports containing them under the ignored `tmp/` directory. Use shareable fixtures for version-controlled tests.

## Documentation rules

- `README.md` is the primary source of truth for setup and usage.
- Keep examples consistent with actual flags, env vars, and current command names.
- Follow the [README documentation policy](README.md#ドキュメント方針) for architecture notes and ADRs.

## Agent handoff expectations

When finishing a task, report:

- what changed
- whether tests/lint were run
- any follow-up risk, limitation, or manual step still required
