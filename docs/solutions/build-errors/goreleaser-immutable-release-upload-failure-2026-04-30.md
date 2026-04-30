---
title: Goreleaser fails to upload assets to GitHub release
date: 2026-04-30
category: build-errors
module: ci-cd
problem_type: build_error
component: tooling
symptoms:
  - "422 Cannot upload assets to an immutable release"
  - "Goreleaser release step fails after successful builds"
root_cause: config_error
resolution_type: config_change
severity: high
tags:
  - goreleaser
  - github-actions
  - release-workflow
  - ci-cd
---

# Goreleaser fails to upload assets to GitHub release

## Problem

Goreleaser builds all binaries successfully but fails at the asset upload step with HTTP 422 errors, preventing any release artifacts from being published.

## Symptoms

- Goreleaser logs show all 6 platform builds complete successfully
- Upload phase fails with: `422 Cannot upload assets to an immutable release`
- Every asset upload fails (tar.gz, zip, checksums.txt)
- The GitHub release exists but has no attached binaries

## What Didn't Work

- **Retrying the workflow**: The release was already in a published (immutable) state, so re-running wouldn't help.

## Solution

Change the release workflow trigger from `release: created` to `push: tags`:

```yaml
# Before — broken
on:
  release:
    types: [created]

# After — works
on:
  push:
    tags:
      - v*
```

The workflow must **not** have a pre-existing release when goreleaser runs. Let goreleaser create the release itself.

When releasing:

```bash
git tag v0.1.1
git push origin v0.1.1
# Goreleaser creates the release and uploads assets automatically
```

Do **not** create the release manually with `gh release create` before goreleaser runs.

## Why This Works

Goreleaser's `release` pipe creates a GitHub release and uploads assets in one step. When the workflow triggers on `release: created`, a human or automation has already created and published the release. GitHub marks published releases as immutable — their assets cannot be modified via the API. Goreleaser then tries to upload to this immutable release and gets 422 errors.

Triggering on tag push means no release exists yet when goreleaser runs, so it can create and populate the release in a single atomic operation.

## Prevention

- Always let goreleaser own the full release lifecycle (create + upload)
- If you need custom release notes, use goreleaser's changelog configuration or edit the release after goreleaser completes
- If a release needs to be redone, note that GitHub may block re-creating a recently deleted tag (`GH013: Cannot create ref due to creations being restricted`) — use the next patch version instead

## Related Issues

- GitHub docs on immutable releases: published releases cannot have assets added via API
- Goreleaser expects to create its own release — mixing manual release creation with goreleaser is unsupported
