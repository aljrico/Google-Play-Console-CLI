---
name: gpc-purchase-verification
description: Inspect and act on Google Play purchase tokens with gpc — product and subscription status, acknowledge, consume, cancel, revoke, and voided purchases.
---

# gpc Purchase Verification

Use this skill for Google Play purchase token workflows with `gpc`: reading purchase state, acknowledging product or subscription purchases after entitlement is granted, consuming consumable products, canceling subscriptions, revoking purchases with refunds, and pulling voided purchase reports.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Monetization" --output json
```

Purchase tokens are sensitive. Treat token strings as secrets — they grant programmatic access to a user's purchase. Never paste them into shared logs or public artifacts.

## Inspect

Read-only token status:

```sh
gpc purchases product --package com.example.app --token PURCHASE_TOKEN --output json
gpc purchases subscription --package com.example.app --token PURCHASE_TOKEN --output json
```

`purchases product` reports purchase state and acknowledgement state for a one-time or managed product token. `purchases subscription` returns subscription v2 status — line items, billing period, account hold, entitlement window.

## Acknowledge

Acknowledge after the entitlement has actually been granted in your backend. Unacknowledged purchases auto-refund.

```sh
gpc purchases product acknowledge \
  --package com.example.app \
  --product-id coins_100 \
  --token PURCHASE_TOKEN \
  --dry-run \
  --output json

gpc purchases subscription acknowledge \
  --package com.example.app \
  --subscription-id premium_monthly \
  --token PURCHASE_TOKEN \
  --dry-run \
  --output json
```

Add `--confirm` only after the grant succeeded. Never acknowledge speculatively.

## Consume

For consumable products (coins, lives), call `consume` only after the in-game inventory has been credited:

```sh
gpc purchases product consume \
  --package com.example.app \
  --product-id coins_100 \
  --token PURCHASE_TOKEN \
  --dry-run \
  --output json
```

Consumption is irreversible from Google's side; the user can repurchase but the token cannot be reused.

## Cancel A Subscription

```sh
gpc purchases subscription cancel \
  --package com.example.app \
  --token PURCHASE_TOKEN \
  --cancellation-type userRequestedStopRenewals \
  --dry-run \
  --output json
```

`--cancellation-type` is required. Cancel stops auto-renewal at the end of the current period without refunding.

## Revoke With Refund

Revoke immediately ends entitlement and issues a refund:

```sh
gpc purchases subscription revoke \
  --package com.example.app \
  --token PURCHASE_TOKEN \
  --refund full \
  --dry-run \
  --output json

gpc purchases subscription revoke \
  --package com.example.app \
  --token PURCHASE_TOKEN \
  --refund item \
  --refund-product-id premium_addon \
  --dry-run \
  --output json
```

`--refund` accepts `full`, `prorated`, or `item`; `item` requires `--refund-product-id`. Revoke is the strongest mutation in this surface — only confirm with explicit user intent.

## Voided Purchases

```sh
gpc purchases voided list --package com.example.app --max-results 25 --output json
```

Pages reports of refunded or revoked purchases. Useful for reconciling backend entitlement state against Play-side refunds.

## Guardrails

- Token strings are sensitive. Redact them in any output that leaves the local environment.
- Always acknowledge after the grant, never before.
- Revoke is the strongest mutation: cancels the subscription, refunds the user, and ends entitlement immediately. Require explicit user intent and a successful dry-run preview.
- Consume is irreversible; the user can repurchase but the same token cannot be reused.
- `purchases voided list` is read-only and safe to run without confirm guards.
