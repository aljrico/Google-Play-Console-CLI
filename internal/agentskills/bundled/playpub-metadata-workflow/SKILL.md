---
name: playpub-metadata-workflow
description: Manage Google Play listings, images, app details, data safety, and review replies with playpub.
---

# playpub Metadata Workflow

Use this skill for Google Play store listing, image, details, data-safety, and review workflows with `playpub`.

## Store Listings

Inspect before changing:

```sh
playpub listings list --package com.example.app --output json
playpub listings get --package com.example.app --language en-US --output json
```

Dry-run mutations first:

```sh
playpub listings update \
  --package com.example.app \
  --language en-US \
  --title "Example" \
  --short-description "Short copy" \
  --dry-run \
  --output json
```

## Images

```sh
playpub images list --package com.example.app --language en-US --type phoneScreenshots --output json
playpub images upload --package com.example.app --language en-US --type featureGraphic --file ./feature.png --dry-run --output json
```

Use `--confirm` only when the user wants the edit committed.

## App Details And Data Safety

```sh
playpub details get --package com.example.app --output json
playpub details update --package com.example.app --contact-email support@example.com --dry-run --output json
playpub data-safety update --package com.example.app --csv ./data-safety.csv --dry-run --output json
```

## Reviews

Review replies are live mutations. Keep replies concise and user-approved:

```sh
playpub reviews list --package com.example.app --max-results 25 --output json
playpub reviews reply --package com.example.app --review-id review-123 --text "Thanks for the feedback." --dry-run --output json
```
