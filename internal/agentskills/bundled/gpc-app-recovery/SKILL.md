---
name: gpc-app-recovery
description: List, draft, target, deploy, and cancel Google Play app recovery actions with gpc.
---

# gpc App Recovery

Use this skill for Google Play app recovery action workflows with `gpc`: inspecting existing recoveries, drafting new ones for a version code, applying targeting, deploying, or canceling.

App recovery actions push a remote in-app update to users on a specific bad version code. Mutations are guarded by `--dry-run` and `--confirm`.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Tracks, Testing, And Distribution" --output json
gpc app-recovery list --package com.example.app --version-code 123 --output json
```

`list` shows existing recovery actions for that version code. Always inspect before drafting a new one.

## Draft

`create` drafts a remote in-app update recovery action for an explicit version code. Dry-run first:

```sh
gpc app-recovery create \
  --package com.example.app \
  --version-code 123 \
  --region US \
  --dry-run \
  --output json
```

Repeat `--region` for multiple regions. Drafted actions exist as IDs but are not deployed.

## Targeting

Add or update targeting on an existing draft:

```sh
gpc app-recovery add-targeting \
  --package com.example.app \
  --id 7 \
  --region US \
  --dry-run \
  --output json
```

## Deploy

```sh
gpc app-recovery deploy \
  --package com.example.app \
  --id 7 \
  --dry-run \
  --output json
```

Add `--confirm` only when the user is sure the recovery should ship.

## Cancel

```sh
gpc app-recovery cancel \
  --package com.example.app \
  --id 7 \
  --dry-run \
  --output json
```

## Guardrails

- Deploying a recovery action pushes an update to users on a specific version code. Treat as a production-impacting mutation. Never use `--confirm` without explicit user intent.
- Recovery actions are tied to a single version code. Confirm the version code matches the one you want to remediate.
- Cancelation is the safe rollback if a draft was deployed in error.
- Use `gpc app-recovery list` to confirm current state before and after any mutation.
