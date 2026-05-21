---
name: gpc-insights-pipeline
description: Download Google Play finance and statistics reports and compose them into derived KPIs with gpc.
---

# gpc Insights Pipeline

Use this skill for Google Play reporting workflows with `gpc`: downloading finance and statistics reports from the Cloud Storage bucket, summarizing each file, and composing them into derived KPIs.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Monetization" --output json
```

The bucket ID is shown in Play Console under "Download reports" and is typically shaped like `pubsite_prod_rev_0123456789`. The service account needs read access on the bucket.

## Download

Finance reports are ZIP; statistics reports are CSV. Use exact object paths.

```sh
gpc finance reports download \
  --bucket pubsite_prod_rev_0123456789 \
  --object earnings/earnings_202605.zip \
  --file ./earnings_202605.zip \
  --dry-run \
  --output json

gpc analytics stats download \
  --bucket pubsite_prod_rev_0123456789 \
  --object stats/store_performance/store_performance_com.example.app_202605_country.csv \
  --file ./store_performance.csv \
  --dry-run \
  --output json
```

Dry-run reports the planned destination. Remove it to fetch. Existing destination files are not overwritten without `--force`.

Finance archives are ZIPs and must be extracted before `summarize`:

```sh
unzip -o ./earnings_202605.zip -d ./earnings_202605
```

Earnings reports are typically a single CSV per month covering all apps; estimated-sales archives may contain multiple CSVs. Inspect the extracted contents and pass the relevant CSV to `summarize`.

## Summarize One File

```sh
gpc finance reports summarize --file ./earnings_202605.csv --output json
gpc analytics stats summarize --file ./store_performance.csv --output json
```

`finance reports summarize` groups rows by transaction type (or financial status for estimated-sales) and currency, summing the merchant-currency amount per group. `analytics stats summarize` sums additive metrics and averages rate-like metrics.

## Derived KPIs

`insights reports summarize` composes a downloaded finance CSV with a statistics CSV into one local JSON artifact with derived KPIs: net revenue by report type and currency, store listing acquisitions, acquisition rate, and revenue per acquisition.

```sh
gpc insights reports summarize \
  --finance-file ./earnings_202605.csv \
  --stats-file ./store_performance.csv \
  --output json
```

## Anomalies

```sh
gpc insights anomalies summarize --file ./vitals-anomalies.json --output json
```

This is a local summary of the JSON output of `gpc vitals anomalies list`. Paginated inputs are marked partial.

## Guardrails

- Reports are static historical files; these commands never mutate Play Console.
- Earnings reports contain financial data. Treat outputs as sensitive.
- Object paths are exact (no globbing) — confirm the path before fetching.
- When a CSV column is missing, derived KPIs that require it are dropped, not faked.
