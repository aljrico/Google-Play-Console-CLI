# Google Play Console CLI

`gpc` is a fast, scriptable CLI for the Google Play Developer API.

The goal is the Android-side sibling to `asc`: predictable commands, JSON-friendly output, CI-first behavior, and no interactive prompts unless a command explicitly opts in.

## Quick Start

```sh
gpc version
gpc --help
```

### Authenticate

```sh
gpc auth login \
  --name "MyApp" \
  --service-account /path/to/service-account.json

gpc auth status
gpc auth doctor
```

The service account needs access to the target app in Play Console and the Google Play Android Developer API enabled in the linked Google Cloud project.

### Planned Command Shape

```sh
gpc tracks list --package com.example.app
gpc releases upload --package com.example.app --aab ./app-release.aab --track internal
gpc releases promote --package com.example.app --from internal --to production
```

### First Publish Flow

```sh
gpc publish internal \
  --package com.example.app \
  --aab ./app-release.aab \
  --release-name "1.2.3" \
  --dry-run \
  --output json \
  --pretty
```

Remove `--dry-run` to validate the edit against Google Play. Add `--confirm` to commit the edit after validation.

## Output

`gpc` defaults to table output in an interactive terminal and JSON when stdout is piped. Override it with:

```sh
gpc version --output json --pretty
gpc version --output table
gpc version --output markdown
```

## Status

Early scaffold. Auth/profile storage and the command taxonomy are in place; API-backed Play Console commands come next.

See [docs/PARITY.md](docs/PARITY.md) for the working parity map against App Store Connect CLI.
