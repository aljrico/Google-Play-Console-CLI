---
name: playpub-release-flow
description: Run Google Play release, track, rollout, promotion, and validation workflows with playpub.
---

# playpub Release Flow

Use this skill for Google Play release, track, rollout, and validation workflows with `playpub`.

## First Checks

Confirm auth and command support:

```sh
playpub auth doctor --output json
playpub capabilities --section "Apps And Releases" --output json
playpub tracks list --package com.example.app --output json
```

## Upload

Dry-run first:

```sh
playpub releases upload \
  --package com.example.app \
  --track internal \
  --aab ./app-release.aab \
  --release-name "1.2.3" \
  --release-note "en-US=Bug fixes." \
  --dry-run \
  --output json
```

Use `--confirm` only when the user wants the edit committed.

## Promote

Promotions require an explicit version code:

```sh
playpub releases promote \
  --package com.example.app \
  --from internal \
  --to production \
  --version-code 123 \
  --status draft \
  --dry-run \
  --output json
```

## Rollout Control

```sh
playpub releases halt --package com.example.app --track production --version-code 123 --dry-run
playpub releases resume --package com.example.app --track production --version-code 123 --status inProgress --user-fraction 0.25 --dry-run
playpub status --package com.example.app --include-draft --output json
```

## Guardrails

- Never commit an edit without explicit user intent.
- Keep release notes localized with `language=text`.
- Validate with `playpub validate --package ...` when troubleshooting edit state.
- Use JSON output for handoffs and CI logs.
