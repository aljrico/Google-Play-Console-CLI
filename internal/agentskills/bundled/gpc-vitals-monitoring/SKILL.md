---
name: gpc-vitals-monitoring
description: Query Android vitals metric sets, search error issues and reports, and list anomalies with gpc.
---

# gpc Vitals Monitoring

Use this skill for Android vitals workflows with `gpc`: inspecting metric-set metadata, querying metric rows over a date range, searching crash/ANR/non-fatal issues, and listing detected anomalies.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Review, Quality, And Feedback" --output json
```

Vitals uses the Play Developer Reporting API, which is separate from the Android Publisher API. The Reporting API must be enabled on the GCP project and the service account needs app-level "View app information (read-only)" access.

## Metric Sets

A metric set is a named collection of metrics (e.g. `crash-rate`, `error-count`, `anr-rate`). Inspect freshness windows and supported metrics before querying:

```sh
gpc vitals metric-set get --package com.example.app --metric-set crash-rate --output json
```

## Query

Query metric rows with explicit metrics, dimensions, filters, and date range. Never rely on API defaults:

```sh
gpc vitals metric-set query \
  --package com.example.app \
  --metric-set crash-rate \
  --metric crashRate \
  --dimension versionCode \
  --aggregation DAILY \
  --start-date 2026-05-01 \
  --end-date 2026-05-19 \
  --output json
```

`DAILY` aggregation requires the America/Los_Angeles time zone (the API default). `HOURLY` aggregation only supports UTC. Set `--time-zone` explicitly when the time-zone semantics matter.

## Error Issues

Search grouped issues (a CRASH or ANR signature):

```sh
gpc vitals errors issues search \
  --package com.example.app \
  --filter "errorIssueType = CRASH" \
  --start-date 2026-05-01 \
  --end-date 2026-05-19 \
  --order-by "errorReportCount desc" \
  --output json
```

`--filter` accepts the Reporting API filter grammar (`errorIssueType`, `versionCode`, etc.). Order issues by `errorReportCount desc` to triage by impact.

## Error Reports

Search individual reports inside an issue:

```sh
gpc vitals errors reports search \
  --package com.example.app \
  --filter "errorIssueId = issue-123" \
  --start-date 2026-05-01 \
  --end-date 2026-05-19 \
  --time-zone UTC \
  --output json
```

`--time-zone` defaults to UTC when omitted. Set it to UTC explicitly when comparing report counts across boundaries.

## Anomalies

```sh
gpc vitals anomalies list \
  --package com.example.app \
  --filter 'activeBetween("2026-05-01T00:00:00Z", "2026-05-19T00:00:00Z")' \
  --output json
```

`activeBetween(start, end)` is the documented filter for time-window anomalies. The result is a list of detected statistical anomalies on monitored metrics.

## Summarize Locally

Combine the JSON output with the insights skill for derived summaries:

```sh
gpc vitals anomalies list --package com.example.app --output json > anomalies.json
gpc insights anomalies summarize --file anomalies.json --output json
```

## Guardrails

- The Reporting API requires its own enablement and IAM grant. `gpc auth doctor` will not catch a missing Reporting API permission until a vitals call returns 403.
- Always pass explicit `--start-date` and `--end-date`. Implicit defaults change with the API and produce unstable results in CI.
- Issue filters can return very large result sets without a date range. Always scope.
- All vitals commands are read-only. There are no mutations to guard with `--confirm`.
- Crash/ANR data is sensitive in some jurisdictions. Treat exports as confidential.
