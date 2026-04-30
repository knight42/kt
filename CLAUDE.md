# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is kt

kt (Kubernetes Tail) is a CLI tool that streams logs from Kubernetes pods, similar to `kubectl logs -f`. It supports filtering by pod name/regexp, label selectors, or higher-level resources (Deployment, Service, StatefulSet, DaemonSet, HPA, Job, ReplicaSet, ReplicationController, CronJob). It auto-discovers new pods, discards deleted ones, retries on container restarts, and colorizes output.

## Build & Install

```bash
make install          # go install with version ldflags
go build .            # quick local build without version info
make darwin-amd64     # cross-compile for a specific platform
make all              # build darwin-amd64, linux-amd64, windows-amd64
make releases         # tar.gz + zip archives for distribution
```

Version info (Version, BuildDate, GitCommit) is injected via ldflags defined in the Makefile into `pkg/version`.

## Architecture

The codebase follows a pipeline: **CLI parsing -> resource resolution -> watch loop -> per-pod tailers -> log consumer**.

- **Root command** (`main.go`): Single cobra command. Parses flags, delegates to `Options.Complete` then `Options.Run`.
- **Options** (`options.go`): Resolves CLI args into either a pod name regexp (1 arg) or a label selector (2 args: resource type + name, or `-l` flag). For higher-level resources, it fetches the object and extracts the pod selector via `getPodsSelector` in `helpers.go`.
- **helpers.go**: Maps Kubernetes resource types to their pod label selectors. HPA is resolved recursively by following `scaleTargetRef` to the underlying workload. Uses `k8s.io/cli-runtime/pkg/resource.Builder` for server-side resource fetching.
- **Controller** (`pkg/controller/`): Watches pods matching the resolved selector/name. On Added/Modified/Deleted events, creates or manages per-pod `Tailer` instances. A single goroutine (`consumeLog`) drains the shared log channel to stdout with buffered writes and optional color prefixes. Configured via functional options (`option.go`).
- **Tailer** (`pkg/tailer/`): Per-pod log streamer. Creates a `Task` per container, each running `fetchLog` in its own goroutine. `fetchLog` uses the Kubernetes streaming logs API. Supports retry on container restart (triggered by the controller on Modified events). The tailer owns a root context; `Close()` cancels all container tasks.
- **api.Log** (`pkg/api/types.go`): The message struct passed through the log channel from tailers to the controller's consumer.
- **pkg/log**: Minimal leveled logger (V-style verbosity) writing to stderr. `-v` flag controls debug output.

## Commit conventions

Group related changes into separate commits (e.g., don't mix dependency updates with code changes, or feature work with refactoring).

## Key patterns

- Pod selection has two paths: by name regexp (watches all pods, filters client-side) vs. by label selector (server-side filtering). These are mutually exclusive.
- The controller uses a `map[types.UID]tailer.Tailer` to track active pods. Each tailer manages a `map[string]*Task` for its containers.
- Color assignment happens once per pod at tailer creation time via `pickColor()`.
- The `--previous` flag disables the watch loop entirely and uses synchronous `TailSync()` instead.
