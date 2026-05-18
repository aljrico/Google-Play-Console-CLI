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
gpc init --dry-run
```

The service account needs access to the target app in Play Console and the Google Play Android Developer API enabled in the linked Google Cloud project.

### Command Shape

```sh
gpc tracks list --package com.example.app
gpc capabilities --status tested
gpc docs parity --output markdown
gpc device-tier-configs list --package com.example.app --page-size 25
gpc device-tier-configs get --package com.example.app --id 7
gpc status --package com.example.app
gpc validate --package com.example.app
gpc releases list --package com.example.app --track internal
gpc releases upload --package com.example.app --track internal --aab ./app-release.aab --release-note "en-US=Bug fixes." --dry-run
gpc releases promote --package com.example.app --from internal --to production --version-code 123 --status draft --dry-run
gpc releases halt --package com.example.app --track production --version-code 123 --dry-run
gpc releases resume --package com.example.app --track production --version-code 123 --status inProgress --user-fraction 0.25 --dry-run
gpc internal-sharing upload --package com.example.app --aab ./app-release.aab --dry-run
gpc app-recovery list --package com.example.app --version-code 123
gpc generated-apks list --package com.example.app --version-code 123
gpc images list --package com.example.app --language en-US --type phoneScreenshots
gpc listings update --package com.example.app --language en-US --title "Example" --dry-run
gpc listings delete --package com.example.app --language en-US --dry-run
gpc details update --package com.example.app --contact-email support@example.com --dry-run
gpc data-safety update --package com.example.app --csv ./data-safety.csv --dry-run
gpc reviews list --package com.example.app --max-results 25
gpc reviews get --package com.example.app --review-id review-123
gpc reviews reply --package com.example.app --review-id review-123 --text "Thanks for the feedback." --dry-run
gpc in-app-products list --package com.example.app
gpc in-app-products get --package com.example.app --sku coins_100
gpc subscriptions list --package com.example.app --page-size 50
gpc subscriptions get --package com.example.app --product-id premium_monthly
gpc subscription-offers list --package com.example.app --product-id premium_monthly --base-plan-id monthly
gpc subscription-offers get --package com.example.app --product-id premium_monthly --base-plan-id monthly --offer-id intro
gpc purchases product --package com.example.app --token PURCHASE_TOKEN
gpc purchases subscription --package com.example.app --token PURCHASE_TOKEN
gpc purchases voided list --package com.example.app --max-results 25
gpc orders get --package com.example.app --order-id GPA.1234-5678-9012-34567
gpc orders batch-get --package com.example.app --order-id GPA.1234 --order-id GPA.5678
gpc pricing convert-region-prices --package com.example.app --currency USD --units 9 --nanos 990000000
gpc users list --developer 1234567890
```

Review APIs follow Google Play's limits: list responses are recent reviews with comments, reply text is capped at 350 characters, and live replies require a service account with review-reply access.

`in-app-products` uses Google's legacy `inappproducts` API. Use it for managed products and catalog inspection; `subscriptions` and `subscription-offers` use the newer monetization resources.

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

`gpc` defaults to JSON so scripts get stable output. Override it with:

```sh
gpc version --output json --pretty
gpc version --output table
gpc version --output markdown
```

## Status

Early but functional. Auth/profile storage, the command taxonomy, core release workflows, localized listings, app-level details, review reading/replies, read-only monetization catalog commands, purchase checks, and order lookup are in place.

See [docs/PARITY.md](docs/PARITY.md) for the working parity map against App Store Connect CLI.
