# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
make build                             # local build with version ldflags
make install                           # go install with version ldflags
go test ./...                          # run all tests
go test ./pkg/query/ -v                # run a specific package
go test ./pkg/controller/ -run TestFoo # run a specific test
```

Releases are handled by goreleaser (`.goreleaser.yml`), triggered on tag push.

## Conventions

- Group related changes into separate commits — don't mix deps, code, and docs.
- Table-driven tests use `map[string]struct{}` style, not `[]struct{name string}`.

## Gotchas

- **onPodAdded ordering**: the pod must be added to `podsTailer` and `updatePrefixState` called BEFORE `t.Tail()`. This prevents a race where tailer goroutines send logs before prefix state is set. There is a regression test for this.
- Pod selection has two mutually exclusive paths: name regexp (client-side filter) vs. label selector (server-side). Don't mix them.

## Architecture

kt is a CLI that streams logs from Kubernetes pods. Pipeline: **CLI parsing -> resource resolution -> watch loop -> per-pod tailers -> log consumer**.

- `main.go` / `options.go` / `helpers.go` — cobra command, flag parsing, resource-to-pod-selector resolution. HPA resolves recursively via `scaleTargetRef`.
- `pkg/controller/` — watches pods, manages per-pod `Tailer` instances, consumes logs via a shared channel. Tailer creation injectable via `newTailerFn` for testing.
- `pkg/tailer/` — per-pod log streamer. One goroutine per container, retry on restart, root context for cleanup.
- `pkg/query/` — recursive descent parser for `-q/--query` log filter DSL (`and`/`or`, parens, quoted strings).
- `pkg/api/` — `Log` struct passed through the channel.
- `pkg/log/` — minimal V-style leveled logger to stderr.
- `docs/solutions/` — documented solutions to past problems, organized by category with YAML frontmatter (`module`, `tags`, `problem_type`).
