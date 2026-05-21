---
name: gpc-revenuecat-catalog-sync
description: Reconcile Google Play subscriptions and in-app products with RevenueCat products, entitlements, offerings, and packages using gpc and the RevenueCat MCP.
---

# gpc RevenueCat Catalog Sync

Use this skill when setting up, validating, or syncing a Google Play monetization catalog against RevenueCat. The flow uses `gpc` for the Play side and the RevenueCat MCP for the RevenueCat side; product IDs are the join key and sync direction matters.

## First Checks

Verify the Play side:

```sh
gpc auth doctor --output json
gpc capabilities --section "Monetization" --output json
```

Confirm the RevenueCat MCP is connected and that the active RevenueCat project is the one you intend to reconcile. RevenueCat catalogs are scoped per project; a stray project pick will reconcile against the wrong catalog.

## Read The Play Side

```sh
gpc subscriptions list --package com.example.app --output json
gpc one-time-products list --package com.example.app --output json
gpc in-app-products list --package com.example.app --output json
```

`subscriptions` and `one-time-products` use the new monetization APIs; `in-app-products` is Google's legacy managed-product API. Most live catalogs have entries in more than one; combine the three lists into one snapshot of Play product IDs, base plans, purchase options, and regional pricing.

## Read The RevenueCat Side

Use the RevenueCat MCP to read the RevenueCat catalog for the active project:

- Products, filtered to `play_store` platform — the join layer; each carries the underlying Play product ID
- Entitlements — the permission a subscriber unlocks (lives only in RevenueCat)
- Offerings and packages — the UI-side grouping of products

Capture a snapshot mirroring the Play one.

## Compare

The join key is the Play product ID. Bucket each ID:

- Present on Play and on RevenueCat → check that the RevenueCat product points at the right Play identifier and that any entitlement attachment is intentional
- Present on Play only → likely missing from RevenueCat; create the RC product or stop the run if Play was the wrong source
- Present on RevenueCat only (with `play_store` platform) → either an orphaned RevenueCat product or a Play deletion that did not propagate

`gpc diff json` produces a stable diff of two saved JSON snapshots:

```sh
gpc diff json ./snapshot-play.json ./snapshot-rc.json --fail-on-change
```

## Apply Changes

Apply changes one side at a time, with `--dry-run` first.

Play side, guarded with `gpc`:

```sh
gpc subscriptions create --package com.example.app --product-id premium_monthly --from-json subscription.json --regions-version 2026/05 --dry-run --output json
gpc one-time-products patch --package com.example.app --product-id coins_100 --listing-language en-US --title "100 coins" --regions-version 2026/05 --dry-run --output json
```

RevenueCat side via the RevenueCat MCP: create products with the existing Play product ID, attach them to entitlements, and surface them in offerings. RevenueCat treats the store product ID as immutable, so the ID created on Play is the ID you bind on RevenueCat.

## Guardrails

- Treat each sync run as one-directional: either Play → RevenueCat or RevenueCat → Play. Mixing within one session produces unclear state.
- Product IDs are immutable on both sides. Never "rename" a product; delete and re-create only with explicit user intent.
- Entitlements live in RevenueCat, not Play. Do not project entitlement state into a Play patch.
- Subscription base-plan pricing is managed on Play. RevenueCat reads store-side prices, it does not write them.
- Never use `--confirm` on `gpc` mutations or apply RevenueCat catalog mutations without explicit user intent and a successful dry-run / preview first.
- Snapshot both sides to JSON before any apply step so a rollback comparison is always available.
