---
name: gpc-ppp-pricing
description: Compute Google Play regional prices from a source amount and turn them into deterministic patch arguments for products, purchase options, base plans, and offer phases.
---

# gpc PPP Pricing

Use this skill for purchasing power parity (PPP) and regional pricing workflows with `gpc`. The two commands work as a pipeline: `gpc pricing convert-region-prices` produces the regional prices, and `gpc pricing build-price-patches` turns that output into ready-to-run patch commands for any priced surface in Play.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Monetization" --output json
```

## Convert

Convert a source `Money` value into Google's recommended regional prices:

```sh
gpc pricing convert-region-prices \
  --package com.example.app \
  --currency USD \
  --units 9 \
  --nanos 990000000 \
  --output json --pretty > conversion.json
```

`units` and `nanos` follow the Google Money shape: `9` units and `990_000_000` nanos is `$9.99`. The result is a `RegionPriceConversionResult` with per-region prices. Save it to disk — `build-price-patches` consumes it.

## Build Patches

Turn the converted prices into patch arguments and a suggested dry-run command for the target surface:

```sh
gpc pricing build-price-patches \
  --from-json conversion.json \
  --target in-app-product \
  --sku coins_100 \
  --output json

gpc pricing build-price-patches \
  --from-json conversion.json \
  --target one-time-product \
  --product-id coins_100 \
  --purchase-option-id buy \
  --output json

gpc pricing build-price-patches \
  --from-json conversion.json \
  --target subscription-base-plan \
  --product-id premium_monthly \
  --base-plan-id monthly \
  --output json

gpc pricing build-price-patches \
  --from-json conversion.json \
  --target subscription-offer-phase \
  --product-id premium_monthly \
  --base-plan-id monthly \
  --offer-id intro \
  --phase-index 0 \
  --output json
```

`--target` accepts: `in-app-product`, `one-time-product`, `subscription-base-plan`, `subscription-offer-phase`. The output includes both the `priceArgs` array and a `suggestedCommand` you can pipe to a shell after review. The suggested command always contains `--dry-run`; remove it and add `--confirm` only after explicit user intent.

## Workflow

```sh
gpc pricing convert-region-prices --package com.example.app --currency USD --units 4 --nanos 990000000 --output json --pretty > /tmp/conv.json
gpc pricing build-price-patches --from-json /tmp/conv.json --target subscription-base-plan --product-id premium_monthly --base-plan-id monthly --output json --pretty
```

Read the `suggestedCommand`, audit the regions and amounts, then run the command.

## Guardrails

- The conversion JSON file is the authority. Do not hand-edit individual region prices — re-run `convert-region-prices` with a different source value if the result is wrong.
- Duplicate region keys in hand-edited conversion JSON are rejected by `build-price-patches`. This is intentional fail-closed behavior.
- `--phase-index` is zero-based and corresponds to the offer's `phases[]` ordering. `0` is the first phase.
- `convert-region-prices` is read-only; `build-price-patches` is also read-only (it emits commands, it does not apply them).
- The suggested command targets the live API. Run it with `--dry-run` first; only `--confirm` with explicit user intent.
