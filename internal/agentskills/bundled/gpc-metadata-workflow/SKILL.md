# gpc Metadata Workflow

Use this skill for Google Play store listing, image, details, data-safety, and review workflows with `gpc`.

## Store Listings

Inspect before changing:

```sh
gpc listings list --package com.example.app --output json
gpc listings get --package com.example.app --language en-US --output json
```

Dry-run mutations first:

```sh
gpc listings update \
  --package com.example.app \
  --language en-US \
  --title "Example" \
  --short-description "Short copy" \
  --dry-run \
  --output json
```

## Images

```sh
gpc images list --package com.example.app --language en-US --type phoneScreenshots --output json
gpc images upload --package com.example.app --language en-US --type featureGraphic --file ./feature.png --dry-run --output json
```

Use `--confirm` only when the user wants the edit committed.

## App Details And Data Safety

```sh
gpc details get --package com.example.app --output json
gpc details update --package com.example.app --contact-email support@example.com --dry-run --output json
gpc data-safety update --package com.example.app --csv ./data-safety.csv --dry-run --output json
```

## Reviews

Review replies are live mutations. Keep replies concise and user-approved:

```sh
gpc reviews list --package com.example.app --max-results 25 --output json
gpc reviews reply --package com.example.app --review-id review-123 --text "Thanks for the feedback." --dry-run --output json
```
