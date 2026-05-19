---
name: playpub-cli-usage
description: Use playpub safely with explicit output formats, auth checks, command discovery, and dry-run-first behavior.
---

# playpub CLI Usage

Use this skill when working with `playpub`, the Google Play Console CLI.

## Defaults

- Prefer `--output json` for automation.
- Use `--pretty` only for human-readable JSON.
- Prefer explicit long flags in scripts.
- Run `playpub docs commands --output markdown` when command shape is uncertain.
- Run `playpub capabilities --output json` before assuming a feature exists.
- Treat live Google Play calls as opt-in. Use `--dry-run` when a mutation supports it.
- Use `--confirm` only when the user clearly wants the live mutation applied.

## Auth

Set up a service account profile first:

```sh
playpub auth login --name MyApp --service-account /path/to/service-account.json
playpub auth doctor --output json
```

Android Publisher commands require the Google Play Android Developer API. `playpub vitals` commands also require the Play Developer Reporting API.

## Discovery

Useful command discovery:

```sh
playpub docs commands --output markdown
playpub docs parity --output markdown
playpub capabilities --status tested --output json
playpub schema --resource edits.tracks --method list --output json
```

## Output

`playpub` defaults to JSON. Table and markdown are for humans:

```sh
playpub version --output json
playpub version --output table
playpub version --output markdown
```
