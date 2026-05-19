# gpc Release Flow

Use this skill for Google Play release, track, rollout, and validation workflows with `gpc`.

## First Checks

Confirm auth and command support:

```sh
gpc auth doctor --output json
gpc capabilities --section "Apps And Releases" --output json
gpc tracks list --package com.example.app --output json
```

## Upload

Dry-run first:

```sh
gpc releases upload \
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
gpc releases promote \
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
gpc releases halt --package com.example.app --track production --version-code 123 --dry-run
gpc releases resume --package com.example.app --track production --version-code 123 --status inProgress --user-fraction 0.25 --dry-run
gpc status --package com.example.app --include-drafts --output json
```

## Guardrails

- Never commit an edit without explicit user intent.
- Keep release notes localized with `language=text`.
- Validate with `gpc validate --package ...` when troubleshooting edit state.
- Use JSON output for handoffs and CI logs.
