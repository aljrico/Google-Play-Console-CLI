---
name: gpc-rtdn-pipeline
description: Set up, pull, and decode Google Play real-time developer notifications with gpc.
---

# gpc RTDN Pipeline

Use this skill for Google Play real-time developer notifications (RTDN) workflows with `gpc`: Pub/Sub topic and subscription provisioning, pulling messages, and decoding payloads.

## First Checks

```sh
gpc auth doctor --output json
gpc capabilities --section "Automation And Utility" --output json
```

The service account needs Pub/Sub Admin (or equivalent) on the GCP project that owns the topic.

## Pub/Sub Setup

`gpc notifications pubsub setup` creates the Google Cloud topic, the subscription, and the Pub/Sub Publisher IAM binding for `google-play-developer-notifications@system.gserviceaccount.com`. It does not select the topic in Play Console — that is an operator step.

Dry-run first to inspect what would be created:

```sh
gpc notifications pubsub setup \
  --project play-project \
  --topic play-rtdn \
  --subscription play-rtdn-sub \
  --dry-run \
  --output json
```

Add `--confirm` to commit. After commit, configure the topic on the app in Play Console (Monetization setup → Real-time developer notifications) and send a test notification from there.

## Pull And Decode

`gpc notifications pubsub pull` reads messages from a pull subscription. Messages are not acknowledged unless both `--ack` and `--confirm` are passed.

Read first without acking:

```sh
gpc notifications pubsub pull \
  --project play-project \
  --subscription play-rtdn-sub \
  --decode-rtdn \
  --output json
```

Ack only after handling succeeds:

```sh
gpc notifications pubsub pull \
  --project play-project \
  --subscription play-rtdn-sub \
  --decode-rtdn \
  --ack \
  --confirm \
  --output json
```

## Decode A Single Payload

For one-shot decoding of a captured payload:

```sh
gpc notifications rtdn decode --file ./pubsub-rtdn.json --output json
gpc notifications rtdn decode --file ./unwrapped-rtdn.json --unwrapped --output json
```

Pass `--unwrapped` when Pub/Sub push delivery sends the developer notification JSON directly. Pass exactly one of `--file` or `--data`.

## Guardrails

- Never ack messages without `--ack --confirm`. Failure to handle a notification correctly should leave it un-acked so Pub/Sub redelivers.
- Setup is a live mutation. Use `--confirm` only with explicit user intent.
- Topic selection in Play Console is manual. Do not promise full end-to-end automation.
- RTDN payloads contain purchase tokens. Treat decoded output as sensitive.
