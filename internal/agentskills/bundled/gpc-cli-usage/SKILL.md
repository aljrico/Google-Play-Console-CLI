---
name: gpc-cli-usage
description: Use gpc safely with explicit output formats, auth checks, command discovery, and dry-run-first behavior.
---

# gpc CLI Usage

Use this skill when working with `gpc`, the Google Play Console CLI.

## Defaults

- Prefer `--output json` for automation.
- Use `--pretty` only for human-readable JSON.
- Prefer explicit long flags in scripts.
- Run `gpc docs commands --output markdown` when command shape is uncertain.
- Run `gpc capabilities --output json` before assuming a feature exists.
- Treat live Google Play calls as opt-in. Use `--dry-run` when a mutation supports it.
- Use `--confirm` only when the user clearly wants the live mutation applied.

## Auth

Set up a service account profile first:

```sh
gpc auth login --name MyApp --service-account /path/to/service-account.json
gpc auth doctor --output json
```

Android Publisher commands require the Google Play Android Developer API. `gpc vitals` commands also require the Play Developer Reporting API.

## Discovery

Useful command discovery:

```sh
gpc docs commands --output markdown
gpc docs parity --output markdown
gpc capabilities --status tested --output json
gpc schema --resource edits.tracks --method list --output json
```

## Output

`gpc` defaults to JSON. Table and markdown are for humans:

```sh
gpc version --output json
gpc version --output table
gpc version --output markdown
```
