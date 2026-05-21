---
name: gpc-testers-orchestration
description: Inspect tracks and manage Google Groups configured as testers for internal, closed, and open testing tracks with gpc.
---

# gpc Testers Orchestration

Use this skill for Google Play testing track and tester workflows with `gpc`: inspecting which tracks exist on an app, reading the Google Groups currently configured as testers on a track, and replacing the tester configuration.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Tracks, Testing, And Distribution" --output json
```

Google Play has three testing track families: `internal` (up to 100 testers, no review), `alpha` (closed testing), `beta` (open or closed testing). Production is not a testing track. Testers are configured per-track via Google Groups; individual email lists are not exposed through this API.

## List Tracks

```sh
gpc tracks list --package com.example.app --output json
```

`tracks list` works through a temporary edit and returns every track on the app, including release status. Always inspect before mutating tester configuration so an in-flight release is not perturbed.

## Read Testers

```sh
gpc testers get --package com.example.app --track internal --output json
```

`testers get` returns the Google Groups currently configured for that track. The response is the source of truth — Play Console UI lags edit operations.

## Replace Testers

`testers update` replaces the full Google Group list on a track (it is not additive). Dry-run first:

```sh
gpc testers update \
  --package com.example.app \
  --track internal \
  --google-group qa@example.com \
  --google-group external-testers@example.com \
  --dry-run \
  --output json
```

Repeat `--google-group` for multiple groups. Add `--confirm` only after the user has reviewed the planned final state. Omitting all `--google-group` flags clears the list, which is destructive — never invoke without explicit user intent.

## Guardrails

- `testers update` replaces the entire Google Group set on a track; it does not merge. Always read with `testers get` first and pass the full desired set on update.
- Google Groups are app-level state. Renaming or deleting a group on the Google Workspace side breaks the tester configuration without warning from Play.
- `gpc tracks list` opens a temporary edit and deletes it on success. Failures may leave a dangling edit — `gpc validate --package ...` clears stale edits.
- Internal track testers do not go through a Google review; closed and open beta tracks do. Plan timing accordingly.
- Tester configuration is invisible to Play Console UI for some minutes after an `update`. Trust the API response, not the dashboard.
