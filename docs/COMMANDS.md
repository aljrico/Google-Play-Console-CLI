# Command Reference

## gpc

Fast, scriptable CLI for the Google Play Developer API

```sh
gpc [flags]
```

### Global Flags

- `--output` / `-o`: Output format: json, table, markdown (default `json`)
- `--pretty`: Pretty-print JSON output (default `false`)

### Commands

- `gpc account`: Inspect local Google Play account configuration
- `gpc analytics`: Summarize Google Play statistics reports
- `gpc app-recovery`: Inspect and manage Google Play app recovery actions
- `gpc apps`: Inspect Google Play apps
- `gpc auth`: Manage Google Play API authentication
- `gpc capabilities`: List gpc command parity and capability status
- `gpc completion`: Generate the autocompletion script for the specified shell
- `gpc data-safety`: Update Google Play data safety declarations
- `gpc details`: Manage app-level Google Play details
- `gpc device-tier-configs`: Inspect Google Play device tier configs
- `gpc diff`: Compare local Google Play payloads
- `gpc docs`: Print embedded gpc documentation
- `gpc finance`: Summarize Google Play financial reports
- `gpc generated-apks`: Inspect generated APK metadata for an App Bundle version
- `gpc grants`: Manage Google Play app access grants
- `gpc images`: Manage localized Google Play store images
- `gpc in-app-products`: Inspect legacy Google Play in-app products
- `gpc init`: Create a local gpc workspace
- `gpc insights`: Summarize Google Play data exports
- `gpc install-skills`: Install bundled gpc agent skills
- `gpc internal-sharing`: Upload artifacts to Google Play internal app sharing
- `gpc listings`: Manage localized Google Play store listings
- `gpc metadata`: Apply app details and localized listings from a file
- `gpc migrate`: Inspect local metadata for migration
- `gpc notifications`: Inspect Google Play notification payloads
- `gpc notify`: Send release workflow notifications
- `gpc one-time-product-offers`: Inspect Google Play one-time product offers
- `gpc one-time-products`: Inspect Google Play one-time products
- `gpc orders`: Inspect and refund Google Play orders
- `gpc pricing`: Inspect Google Play price conversions
- `gpc publish`: Run high-level Google Play publishing workflows
- `gpc purchases`: Inspect and manage Google Play purchase tokens
- `gpc releases`: Upload and manage Google Play releases
- `gpc reviews`: Read and reply to Google Play reviews
- `gpc schema`: Print the Google Play discovery schema
- `gpc search`: Search gpc commands and flags
- `gpc snitch`: Report gpc friction
- `gpc status`: Summarize Google Play release status
- `gpc subscription-offers`: Inspect Google Play subscription offers
- `gpc subscriptions`: Inspect Google Play monetization subscriptions
- `gpc system-apks`: Inspect Google Play system APK variants
- `gpc testers`: Manage Google Play track tester groups
- `gpc tracks`: Manage Google Play release tracks
- `gpc users`: Inspect and manage Google Play Console users
- `gpc validate`: Validate a temporary Google Play edit
- `gpc version`: Print version information
- `gpc vitals`: Inspect Google Play Developer Reporting vitals
- `gpc web`: Inspect Play Console browser automation support
- `gpc workflow`: Run repo-local gpc workflows

### gpc account

Inspect local Google Play account configuration

```sh
gpc account
```

#### Commands

- `gpc account status`: Summarize local account and service account metadata

### gpc analytics

Summarize Google Play statistics reports

```sh
gpc analytics
```

#### Commands

- `gpc analytics stats`: Summarize downloaded Play statistics CSVs

#### gpc analytics stats

Summarize downloaded Play statistics CSVs

```sh
gpc analytics stats
```

##### Commands

- `gpc analytics stats download`: Download a Play statistics report CSV from Google Cloud Storage
- `gpc analytics stats summarize`: Summarize a Play statistics report CSV

##### gpc analytics stats download

Download a Play statistics report CSV from Google Cloud Storage

```sh
gpc analytics stats download [flags]
```

###### Flags

- `--bucket`: Google Play reports bucket, for example pubsite_prod_rev_0123456789
- `--dry-run`: Print the planned report download without calling Google Cloud Storage (default `false`)
- `--file`: Destination .csv or .zip path
- `--force`: Overwrite the destination file (default `false`)
- `--object`: Cloud Storage object path for the report

##### gpc analytics stats summarize

Summarize a Play statistics report CSV

```sh
gpc analytics stats summarize [flags]
```

###### Flags

- `--file`: Downloaded Google Play statistics report CSV

### gpc app-recovery

Inspect and manage Google Play app recovery actions

```sh
gpc app-recovery
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc app-recovery add-targeting`: Add targeting to an app recovery action
- `gpc app-recovery cancel`: Cancel an app recovery action
- `gpc app-recovery create`: Create a draft remote in-app update recovery action
- `gpc app-recovery deploy`: Deploy a draft app recovery action
- `gpc app-recovery list`: List app recovery actions for a version code

#### gpc app-recovery add-targeting

Add targeting to an app recovery action

```sh
gpc app-recovery add-targeting [flags]
```

##### Flags

- `--all-users`: Target all users (default `false`)
- `--confirm`: Apply the app recovery targeting update (default `false`)
- `--dry-run`: Print the planned app recovery targeting update without calling Google Play (default `false`)
- `--id`: App recovery action ID
- `--region`: ISO 3166-1 alpha-2 region code to add, repeatable (default `[]`)
- `--sdk-level`: Android SDK level to add, repeatable (default `[]`)

#### gpc app-recovery cancel

Cancel an app recovery action

```sh
gpc app-recovery cancel [flags]
```

##### Flags

- `--confirm`: Apply the app recovery mutation (default `false`)
- `--dry-run`: Print the planned app recovery mutation without calling Google Play (default `false`)
- `--id`: App recovery action ID

#### gpc app-recovery create

Create a draft remote in-app update recovery action

```sh
gpc app-recovery create [flags]
```

##### Flags

- `--all-users`: Target all users (default `false`)
- `--confirm`: Create the draft app recovery action (default `false`)
- `--dry-run`: Print the planned app recovery creation without calling Google Play (default `false`)
- `--region`: ISO 3166-1 alpha-2 region code to target, repeatable (default `[]`)
- `--sdk-level`: Android SDK level to target, repeatable (default `[]`)
- `--version-code`: App version code to target, repeatable (default `[]`)
- `--version-code-end`: Highest app version code to target, inclusive (default `0`)
- `--version-code-start`: Lowest app version code to target, inclusive (default `0`)

#### gpc app-recovery deploy

Deploy a draft app recovery action

```sh
gpc app-recovery deploy [flags]
```

##### Flags

- `--confirm`: Apply the app recovery mutation (default `false`)
- `--dry-run`: Print the planned app recovery mutation without calling Google Play (default `false`)
- `--id`: App recovery action ID

#### gpc app-recovery list

List app recovery actions for a version code

```sh
gpc app-recovery list [flags]
```

##### Flags

- `--version-code`: Version code targeted by recovery actions (default `0`)

### gpc apps

Inspect Google Play apps

```sh
gpc apps
```

#### Commands

- `gpc apps list`: List apps visible to the active service account

### gpc auth

Manage Google Play API authentication

```sh
gpc auth
```

#### Commands

- `gpc auth doctor`: Validate the active auth profile
- `gpc auth login`: Store a service account profile
- `gpc auth status`: Show the active auth profile

#### gpc auth login

Store a service account profile

```sh
gpc auth login [flags]
```

##### Flags

- `--name`: Profile name
- `--service-account`: Path to a Google service account JSON key

### gpc capabilities

List gpc command parity and capability status

```sh
gpc capabilities [flags]
```

#### Flags

- `--section`: Filter by parity matrix section
- `--status`: Filter by status: planned, implemented, tested, documented, blocked, not applicable

### gpc completion

Generate the autocompletion script for the specified shell

```sh
gpc completion
```

#### Commands

- `gpc completion bash`: Generate the autocompletion script for bash
- `gpc completion fish`: Generate the autocompletion script for fish
- `gpc completion powershell`: Generate the autocompletion script for powershell
- `gpc completion zsh`: Generate the autocompletion script for zsh

#### gpc completion bash

Generate the autocompletion script for bash

```sh
gpc completion bash
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

#### gpc completion fish

Generate the autocompletion script for fish

```sh
gpc completion fish [flags]
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

#### gpc completion powershell

Generate the autocompletion script for powershell

```sh
gpc completion powershell [flags]
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

#### gpc completion zsh

Generate the autocompletion script for zsh

```sh
gpc completion zsh [flags]
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

### gpc data-safety

Update Google Play data safety declarations

```sh
gpc data-safety
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc data-safety update`: Upload a data safety CSV declaration

#### gpc data-safety update

Upload a data safety CSV declaration

```sh
gpc data-safety update [flags]
```

##### Flags

- `--confirm`: Apply the data safety update (default `false`)
- `--csv`: Path to the data safety CSV export
- `--dry-run`: Print the planned data safety update without calling Google Play (default `false`)

### gpc details

Manage app-level Google Play details

```sh
gpc details
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc details get`: Get app-level details
- `gpc details update`: Patch app-level details

#### gpc details update

Patch app-level details

```sh
gpc details update [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--contact-email`: User-visible support email
- `--contact-phone`: User-visible support phone
- `--contact-website`: User-visible support website
- `--default-language`: Default BCP-47 language, for example en-US
- `--dry-run`: Print the planned details update without calling Google Play (default `false`)

### gpc device-tier-configs

Inspect Google Play device tier configs

```sh
gpc device-tier-configs
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc device-tier-configs get`: Get one device tier config
- `gpc device-tier-configs list`: List device tier configs

#### gpc device-tier-configs get

Get one device tier config

```sh
gpc device-tier-configs get [flags]
```

##### Flags

- `--id`: Device tier config ID (default `0`)

#### gpc device-tier-configs list

List device tier configs

```sh
gpc device-tier-configs list [flags]
```

##### Flags

- `--page-size`: Maximum configs to return, 0 uses the Google default (default `0`)
- `--page-token`: Pagination token from a previous response

### gpc diff

Compare local Google Play payloads

```sh
gpc diff
```

#### Commands

- `gpc diff json`: Compare two JSON files with stable JSON Pointer paths

#### gpc diff json

Compare two JSON files with stable JSON Pointer paths

```sh
gpc diff json FROM TO [flags]
```

##### Flags

- `--fail-on-change`: Exit nonzero when the JSON files differ (default `false`)

### gpc docs

Print embedded gpc documentation

```sh
gpc docs [flags]
```

#### Commands

- `gpc docs commands`: Print generated command reference
- `gpc docs parity`: Print the asc-to-gpc parity matrix

### gpc finance

Summarize Google Play financial reports

```sh
gpc finance
```

#### Commands

- `gpc finance reports`: Summarize downloaded Play financial report CSVs

#### gpc finance reports

Summarize downloaded Play financial report CSVs

```sh
gpc finance reports
```

##### Commands

- `gpc finance reports download`: Download a Play financial report ZIP from Google Cloud Storage
- `gpc finance reports summarize`: Summarize a Play financial report CSV

##### gpc finance reports download

Download a Play financial report ZIP from Google Cloud Storage

```sh
gpc finance reports download [flags]
```

###### Flags

- `--bucket`: Google Play reports bucket, for example pubsite_prod_rev_0123456789
- `--dry-run`: Print the planned report download without calling Google Cloud Storage (default `false`)
- `--file`: Destination .csv or .zip path
- `--force`: Overwrite the destination file (default `false`)
- `--object`: Cloud Storage object path for the report

##### gpc finance reports summarize

Summarize a Play financial report CSV

```sh
gpc finance reports summarize [flags]
```

###### Flags

- `--file`: Downloaded Google Play earnings or estimated-sales CSV

### gpc generated-apks

Inspect generated APK metadata for an App Bundle version

```sh
gpc generated-apks
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc generated-apks download`: Download one generated APK by download ID
- `gpc generated-apks list`: List generated APKs for a version code

#### gpc generated-apks download

Download one generated APK by download ID

```sh
gpc generated-apks download [flags]
```

##### Flags

- `--download-id`: Generated APK download ID from generated-apks list
- `--dry-run`: Print the planned download without calling Google Play (default `false`)
- `--file`: Destination .apk path
- `--force`: Overwrite the destination file (default `false`)
- `--version-code`: App Bundle version code (default `0`)

#### gpc generated-apks list

List generated APKs for a version code

```sh
gpc generated-apks list [flags]
```

##### Flags

- `--version-code`: App Bundle version code (default `0`)

### gpc grants

Manage Google Play app access grants

```sh
gpc grants
```

#### Commands

- `gpc grants create`: Create app-level access for a Play Console user
- `gpc grants delete`: Delete an app-level access grant
- `gpc grants patch`: Replace app-level permissions for an access grant

#### gpc grants create

Create app-level access for a Play Console user

```sh
gpc grants create [flags]
```

##### Flags

- `--confirm`: Apply the grant creation (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned grant creation without calling Google Play (default `false`)
- `--package`: Android package name, for example com.example.app
- `--permission`: App-level grant permission, repeatable (default `[]`)
- `--user-email`: Play Console user email

#### gpc grants delete

Delete an app-level access grant

```sh
gpc grants delete [flags]
```

##### Flags

- `--confirm`: Apply the grant deletion (default `false`)
- `--dry-run`: Print the planned grant deletion without calling Google Play (default `false`)
- `--name`: Grant resource name, developers/{developer}/users/{email}/grants/{package}

#### gpc grants patch

Replace app-level permissions for an access grant

```sh
gpc grants patch [flags]
```

##### Flags

- `--confirm`: Apply the grant patch (default `false`)
- `--dry-run`: Print the planned grant patch without calling Google Play (default `false`)
- `--name`: Grant resource name, developers/{developer}/users/{email}/grants/{package}
- `--permission`: App-level grant permission, repeatable (default `[]`)

### gpc images

Manage localized Google Play store images

```sh
gpc images
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc images delete`: Delete one store image
- `gpc images delete-all`: Delete all store images for one language and image type
- `gpc images list`: List store images for one language and image type
- `gpc images upload`: Upload one store image

#### gpc images delete

Delete one store image

```sh
gpc images delete [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned image deletion without calling Google Play (default `false`)
- `--image-id`: Google Play image ID to delete
- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

#### gpc images delete-all

Delete all store images for one language and image type

```sh
gpc images delete-all [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned image deletion without calling Google Play (default `false`)
- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

#### gpc images list

List store images for one language and image type

```sh
gpc images list [flags]
```

##### Flags

- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

#### gpc images upload

Upload one store image

```sh
gpc images upload [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned image upload without calling Google Play (default `false`)
- `--file`: Path to a .jpg, .jpeg, or .png image
- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

### gpc in-app-products

Inspect legacy Google Play in-app products

```sh
gpc in-app-products
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc in-app-products batch-delete`: Delete multiple legacy managed in-app products
- `gpc in-app-products batch-get`: Get multiple legacy in-app products
- `gpc in-app-products create`: Create a legacy managed in-app product
- `gpc in-app-products delete`: Delete a legacy managed in-app product
- `gpc in-app-products get`: Get one legacy in-app product
- `gpc in-app-products list`: List legacy in-app products
- `gpc in-app-products patch`: Patch a legacy managed in-app product

#### gpc in-app-products batch-delete

Delete multiple legacy managed in-app products

```sh
gpc in-app-products batch-delete [flags]
```

##### Flags

- `--confirm`: Delete the managed in-app products (default `false`)
- `--dry-run`: Print the planned managed in-app product batch deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--sku`: In-app product SKU; repeatable, up to 100 (default `[]`)

#### gpc in-app-products batch-get

Get multiple legacy in-app products

```sh
gpc in-app-products batch-get [flags]
```

##### Flags

- `--sku`: In-app product SKU; repeatable (default `[]`)

#### gpc in-app-products create

Create a legacy managed in-app product

```sh
gpc in-app-products create [flags]
```

##### Flags

- `--confirm`: Create the managed in-app product (default `false`)
- `--default-language`: Default BCP-47 listing language, for example en-US
- `--default-price`: Default checkout price as CURRENCY:MICROS, for example USD:1990000
- `--description`: Default listing description
- `--dry-run`: Print the planned managed in-app product creation without calling Google Play (default `false`)
- `--sku`: In-app product SKU
- `--status`: Initial product status: active or inactive (default `inactive`)
- `--title`: Default listing title

#### gpc in-app-products delete

Delete a legacy managed in-app product

```sh
gpc in-app-products delete [flags]
```

##### Flags

- `--confirm`: Delete the managed in-app product (default `false`)
- `--dry-run`: Print the planned managed in-app product deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--sku`: In-app product SKU

#### gpc in-app-products get

Get one legacy in-app product

```sh
gpc in-app-products get [flags]
```

##### Flags

- `--sku`: In-app product SKU

#### gpc in-app-products list

List legacy in-app products

```sh
gpc in-app-products list [flags]
```

##### Flags

- `--token`: Pagination token from a previous response

#### gpc in-app-products patch

Patch a legacy managed in-app product

```sh
gpc in-app-products patch [flags]
```

##### Flags

- `--confirm`: Apply the managed in-app product patch (default `false`)
- `--default-language`: Default BCP-47 listing language to set on the product
- `--default-price`: Default checkout price as CURRENCY:MICROS, for example USD:1990000
- `--description`: Default listing description
- `--dry-run`: Print the planned managed in-app product patch without calling Google Play (default `false`)
- `--eea-withdrawal-right-type`: EEA withdrawal right type: WITHDRAWAL_RIGHT_DIGITAL_CONTENT or WITHDRAWAL_RIGHT_SERVICE
- `--listing-language`: BCP-47 listing language to update when --title and --description are set
- `--regional-price`: Regional checkout price as REGION:CURRENCY:MICROS, for example US:USD:2990000; repeatable (default `[]`)
- `--regional-streaming-tax`: US streaming tax type as US:STREAMING_TAX_TYPE, for example US:STREAMING_TAX_TYPE_TELCO_VIDEO_SALES; repeatable (default `[]`)
- `--regional-tax-tier`: Regional reduced tax tier as REGION:TAX_TIER, for example FR:TAX_TIER_NEWS_1; repeatable (default `[]`)
- `--sku`: In-app product SKU
- `--status`: Product status: active or inactive
- `--title`: Default listing title
- `--tokenized-digital-asset`: Whether the managed product represents a tokenized digital asset: true or false

### gpc init

Create a local gpc workspace

```sh
gpc init [flags]
```

#### Flags

- `--directory`: Directory for gpc helper files (default `.gpc`)
- `--dry-run`: Print the planned init files without writing (default `false`)
- `--force`: Overwrite existing gpc helper files (default `false`)

### gpc insights

Summarize Google Play data exports

```sh
gpc insights
```

#### Commands

- `gpc insights anomalies`: Summarize Android vitals anomalies

#### gpc insights anomalies

Summarize Android vitals anomalies

```sh
gpc insights anomalies
```

##### Commands

- `gpc insights anomalies summarize`: Summarize a vitals anomalies JSON export

##### gpc insights anomalies summarize

Summarize a vitals anomalies JSON export

```sh
gpc insights anomalies summarize [flags]
```

###### Flags

- `--file`: JSON output from gpc vitals anomalies list

### gpc install-skills

Install bundled gpc agent skills

```sh
gpc install-skills [flags]
```

#### Flags

- `--directory`: Directory for installed agent skills, defaults to ~/.agents/skills
- `--dry-run`: Print planned skill installs without writing (default `false`)
- `--force`: Overwrite existing skill files (default `false`)
- `--skill`: Install one bundled skill by name; repeat to install multiple (default `[]`)

#### Commands

- `gpc install-skills list`: List bundled gpc agent skills

### gpc internal-sharing

Upload artifacts to Google Play internal app sharing

```sh
gpc internal-sharing
```

#### Commands

- `gpc internal-sharing upload`: Upload an APK or Android App Bundle to internal app sharing

#### gpc internal-sharing upload

Upload an APK or Android App Bundle to internal app sharing

```sh
gpc internal-sharing upload [flags]
```

##### Flags

- `--aab`: Path to the Android App Bundle to upload
- `--apk`: Path to the APK to upload
- `--dry-run`: Print the planned internal sharing upload without calling Google Play (default `false`)
- `--package`: Android package name, for example com.example.app

### gpc listings

Manage localized Google Play store listings

```sh
gpc listings
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc listings delete`: Delete one localized store listing
- `gpc listings delete-all`: Delete all localized store listings
- `gpc listings get`: Get one localized store listing
- `gpc listings list`: List localized store listings
- `gpc listings update`: Create or update one localized store listing

#### gpc listings delete

Delete one localized store listing

```sh
gpc listings delete [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned listing deletion without calling Google Play (default `false`)
- `--language`: BCP-47 listing language, for example en-US

#### gpc listings delete-all

Delete all localized store listings

```sh
gpc listings delete-all [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned listing deletion without calling Google Play (default `false`)

#### gpc listings get

Get one localized store listing

```sh
gpc listings get [flags]
```

##### Flags

- `--language`: BCP-47 listing language, for example en-US

#### gpc listings update

Create or update one localized store listing

```sh
gpc listings update [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned listing update without calling Google Play (default `false`)
- `--full-description`: Localized full description
- `--language`: BCP-47 listing language, for example en-US
- `--short-description`: Localized short description
- `--title`: Localized app title
- `--video`: Promotional YouTube video URL

### gpc metadata

Apply app details and localized listings from a file

```sh
gpc metadata
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc metadata apply`: Apply a JSON metadata file through one Google Play edit

#### gpc metadata apply

Apply a JSON metadata file through one Google Play edit

```sh
gpc metadata apply [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned metadata update without calling Google Play (default `false`)
- `--file`: Path to metadata JSON

### gpc migrate

Inspect local metadata for migration

```sh
gpc migrate
```

#### Commands

- `gpc migrate supply`: Inspect fastlane supply metadata

#### gpc migrate supply

Inspect fastlane supply metadata

```sh
gpc migrate supply
```

##### Commands

- `gpc migrate supply changelogs`: Convert fastlane supply changelogs to release-note payloads
- `gpc migrate supply convert`: Convert fastlane supply listings to gpc metadata JSON
- `gpc migrate supply images`: Convert fastlane supply images to image upload payloads
- `gpc migrate supply inspect`: Inventory a fastlane supply metadata directory

##### gpc migrate supply changelogs

Convert fastlane supply changelogs to release-note payloads

```sh
gpc migrate supply changelogs [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)
- `--version-code`: Only include changelogs for this version code (default `0`)

##### gpc migrate supply convert

Convert fastlane supply listings to gpc metadata JSON

```sh
gpc migrate supply convert [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)

##### gpc migrate supply images

Convert fastlane supply images to image upload payloads

```sh
gpc migrate supply images [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)
- `--language`: Only include images for this BCP-47 listing language
- `--type`: Only include this image type

##### gpc migrate supply inspect

Inventory a fastlane supply metadata directory

```sh
gpc migrate supply inspect [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)

### gpc notifications

Inspect Google Play notification payloads

```sh
gpc notifications
```

#### Commands

- `gpc notifications pubsub`: Set up Google Cloud Pub/Sub for Play notifications
- `gpc notifications rtdn`: Inspect real-time developer notifications

#### gpc notifications pubsub

Set up Google Cloud Pub/Sub for Play notifications

```sh
gpc notifications pubsub
```

##### Commands

- `gpc notifications pubsub pull`: Pull Pub/Sub messages for Play notifications
- `gpc notifications pubsub setup`: Create Pub/Sub resources for Play real-time developer notifications

##### gpc notifications pubsub pull

Pull Pub/Sub messages for Play notifications

```sh
gpc notifications pubsub pull [flags]
```

###### Flags

- `--ack`: Acknowledge pulled messages after output succeeds (default `false`)
- `--confirm`: Confirm acknowledgement when --ack is set (default `false`)
- `--decode-rtdn`: Decode each message data field as a Google Play RTDN payload (default `false`)
- `--max-messages`: Maximum Pub/Sub messages to pull (default `10`)
- `--project`: Google Cloud project ID that owns the Pub/Sub subscription
- `--subscription`: Pub/Sub subscription ID to pull from

##### gpc notifications pubsub setup

Create Pub/Sub resources for Play real-time developer notifications

```sh
gpc notifications pubsub setup [flags]
```

###### Flags

- `--ack-deadline`: Subscription acknowledgement deadline in seconds (default `10`)
- `--confirm`: Create Pub/Sub resources and grant the Google Play publisher role (default `false`)
- `--dry-run`: Print the planned Pub/Sub setup without calling Google Cloud (default `false`)
- `--project`: Google Cloud project ID that owns the Pub/Sub resources
- `--push-endpoint`: Optional HTTPS push endpoint; omit for pull subscriptions
- `--subscription`: Pub/Sub subscription ID to create
- `--topic`: Pub/Sub topic ID to create

#### gpc notifications rtdn

Inspect real-time developer notifications

```sh
gpc notifications rtdn
```

##### Commands

- `gpc notifications rtdn decode`: Decode a Pub/Sub RTDN push payload

##### gpc notifications rtdn decode

Decode a Pub/Sub RTDN push payload

```sh
gpc notifications rtdn decode [flags]
```

###### Flags

- `--data`: Inline RTDN JSON payload; required unless --file is set
- `--file`: RTDN JSON payload file; required unless --data is set
- `--unwrapped`: Decode an unwrapped push payload containing the developer notification directly (default `false`)

### gpc notify

Send release workflow notifications

```sh
gpc notify
```

#### Commands

- `gpc notify discord`: Send a Discord incoming webhook notification
- `gpc notify send`: Send a JSON notification webhook
- `gpc notify slack`: Send a Slack incoming webhook notification

#### gpc notify discord

Send a Discord incoming webhook notification

```sh
gpc notify discord [flags]
```

##### Flags

- `--confirm`: Send the Discord webhook (default `false`)
- `--dry-run`: Print the Discord payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS Discord incoming webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the Discord incoming webhook URL (default `GPC_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the Discord incoming webhook URL

#### gpc notify send

Send a JSON notification webhook

```sh
gpc notify send [flags]
```

##### Flags

- `--confirm`: Send the notification webhook (default `false`)
- `--dry-run`: Print the notification payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the webhook URL (default `GPC_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the webhook URL

#### gpc notify slack

Send a Slack incoming webhook notification

```sh
gpc notify slack [flags]
```

##### Flags

- `--confirm`: Send the Slack webhook (default `false`)
- `--dry-run`: Print the Slack payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS Slack incoming webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the Slack incoming webhook URL (default `GPC_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the Slack incoming webhook URL

### gpc one-time-product-offers

Inspect Google Play one-time product offers

```sh
gpc one-time-product-offers
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc one-time-product-offers activate`: activate a one-time product offer
- `gpc one-time-product-offers batch-activate`: activate multiple one-time product offers
- `gpc one-time-product-offers batch-cancel`: Cancel multiple pre-order one-time product offers and pending orders
- `gpc one-time-product-offers batch-deactivate`: deactivate multiple one-time product offers
- `gpc one-time-product-offers batch-delete`: Delete multiple one-time product offers
- `gpc one-time-product-offers batch-get`: Get multiple one-time product offers
- `gpc one-time-product-offers cancel`: cancel a one-time product offer
- `gpc one-time-product-offers deactivate`: deactivate a one-time product offer
- `gpc one-time-product-offers get`: Get a one-time product offer
- `gpc one-time-product-offers list`: List one-time product offers

#### gpc one-time-product-offers activate

activate a one-time product offer

```sh
gpc one-time-product-offers activate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer state update (default `false`)
- `--dry-run`: Print the planned one-time product offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### gpc one-time-product-offers batch-activate

activate multiple one-time product offers

```sh
gpc one-time-product-offers batch-activate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer batch state update (default `false`)
- `--dry-run`: Print the planned one-time product offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### gpc one-time-product-offers batch-cancel

Cancel multiple pre-order one-time product offers and pending orders

```sh
gpc one-time-product-offers batch-cancel [flags]
```

##### Flags

- `--confirm`: Cancel the pre-order offers and their pending orders (default `false`)
- `--dry-run`: Print the planned pre-order offer cancellation without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### gpc one-time-product-offers batch-deactivate

deactivate multiple one-time product offers

```sh
gpc one-time-product-offers batch-deactivate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer batch state update (default `false`)
- `--dry-run`: Print the planned one-time product offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### gpc one-time-product-offers batch-delete

Delete multiple one-time product offers

```sh
gpc one-time-product-offers batch-delete [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer batch deletion (default `false`)
- `--dry-run`: Print the planned one-time product offer batch deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to delete as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### gpc one-time-product-offers batch-get

Get multiple one-time product offers

```sh
gpc one-time-product-offers batch-get [flags]
```

##### Flags

- `--offer`: Offer to fetch as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options

#### gpc one-time-product-offers cancel

cancel a one-time product offer

```sh
gpc one-time-product-offers cancel [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer state update (default `false`)
- `--dry-run`: Print the planned one-time product offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### gpc one-time-product-offers deactivate

deactivate a one-time product offer

```sh
gpc one-time-product-offers deactivate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer state update (default `false`)
- `--dry-run`: Print the planned one-time product offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### gpc one-time-product-offers get

Get a one-time product offer

```sh
gpc one-time-product-offers get [flags]
```

##### Flags

- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### gpc one-time-product-offers list

List one-time product offers

```sh
gpc one-time-product-offers list [flags]
```

##### Flags

- `--page-size`: Maximum offers to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--product-id`: Parent one-time product ID, or - for all products
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for all purchase options

### gpc one-time-products

Inspect Google Play one-time products

```sh
gpc one-time-products
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc one-time-products batch-delete`: Delete multiple one-time products
- `gpc one-time-products batch-get`: Get multiple one-time products
- `gpc one-time-products batch-patch-listings`: Batch patch localized one-time product listings
- `gpc one-time-products delete`: Delete a one-time product
- `gpc one-time-products get`: Get a one-time product
- `gpc one-time-products list`: List one-time products
- `gpc one-time-products patch`: Patch a one-time product listing
- `gpc one-time-products purchase-option`: Manage one-time product purchase options

#### gpc one-time-products batch-delete

Delete multiple one-time products

```sh
gpc one-time-products batch-delete [flags]
```

##### Flags

- `--confirm`: Apply the one-time product batch deletion (default `false`)
- `--dry-run`: Print the planned one-time product batch deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID; repeatable, up to 100 (default `[]`)

#### gpc one-time-products batch-get

Get multiple one-time products

```sh
gpc one-time-products batch-get [flags]
```

##### Flags

- `--product-id`: One-time product ID; repeatable, up to 100 (default `[]`)

#### gpc one-time-products batch-patch-listings

Batch patch localized one-time product listings

```sh
gpc one-time-products batch-patch-listings [flags]
```

##### Flags

- `--confirm`: Apply the one-time product listing batch patch (default `false`)
- `--dry-run`: Print the planned one-time product listing batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--listing`: CSV listing patch productId,language,title,description; repeat for multiple localized listings (default `[]`)
- `--regions-version`: Google Play regions version required by oneTimeProducts.batchUpdate

#### gpc one-time-products delete

Delete a one-time product

```sh
gpc one-time-products delete [flags]
```

##### Flags

- `--confirm`: Apply the one-time product deletion (default `false`)
- `--dry-run`: Print the planned one-time product deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID

#### gpc one-time-products get

Get a one-time product

```sh
gpc one-time-products get [flags]
```

##### Flags

- `--product-id`: One-time product ID

#### gpc one-time-products list

List one-time products

```sh
gpc one-time-products list [flags]
```

##### Flags

- `--page-size`: Maximum one-time products to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response

#### gpc one-time-products patch

Patch a one-time product listing

```sh
gpc one-time-products patch [flags]
```

##### Flags

- `--confirm`: Apply the one-time product listing patch (default `false`)
- `--description`: Localized one-time product description
- `--dry-run`: Print the planned one-time product listing patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--listing-language`: BCP-47 language code for the listing to patch, for example en-US
- `--product-id`: One-time product ID
- `--regions-version`: Google Play regions version required by oneTimeProducts.patch
- `--title`: Localized one-time product title

#### gpc one-time-products purchase-option

Manage one-time product purchase options

```sh
gpc one-time-products purchase-option
```

##### Commands

- `gpc one-time-products purchase-option activate`: activate a one-time product purchase option
- `gpc one-time-products purchase-option batch-delete`: Delete one-time product purchase options
- `gpc one-time-products purchase-option deactivate`: deactivate a one-time product purchase option

##### gpc one-time-products purchase-option activate

activate a one-time product purchase option

```sh
gpc one-time-products purchase-option activate [flags]
```

###### Flags

- `--confirm`: Apply the purchase option state update (default `false`)
- `--dry-run`: Print the planned purchase option state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID
- `--purchase-option-id`: One-time product purchase option ID

##### gpc one-time-products purchase-option batch-delete

Delete one-time product purchase options

```sh
gpc one-time-products purchase-option batch-delete [flags]
```

###### Flags

- `--confirm`: Apply the purchase option batch deletion (default `false`)
- `--dry-run`: Print the planned purchase option batch deletion without calling Google Play (default `false`)
- `--force`: Also delete associated offers under each purchase option (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent one-time product ID, or - when deleting across products; inferred when omitted
- `--purchase-option`: Purchase option to delete as productId/purchaseOptionId; repeatable, up to 100 (default `[]`)

##### gpc one-time-products purchase-option deactivate

deactivate a one-time product purchase option

```sh
gpc one-time-products purchase-option deactivate [flags]
```

###### Flags

- `--confirm`: Apply the purchase option state update (default `false`)
- `--dry-run`: Print the planned purchase option state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID
- `--purchase-option-id`: One-time product purchase option ID

### gpc orders

Inspect and refund Google Play orders

```sh
gpc orders
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc orders batch-get`: Get multiple Google Play orders
- `gpc orders get`: Get one Google Play order
- `gpc orders refund`: Refund one Google Play order

#### gpc orders batch-get

Get multiple Google Play orders

```sh
gpc orders batch-get [flags]
```

##### Flags

- `--order-id`: Google Play order ID, repeatable (default `[]`)

#### gpc orders get

Get one Google Play order

```sh
gpc orders get [flags]
```

##### Flags

- `--order-id`: Google Play order ID

#### gpc orders refund

Refund one Google Play order

```sh
gpc orders refund [flags]
```

##### Flags

- `--confirm`: Apply the refund (default `false`)
- `--dry-run`: Print the planned refund without calling Google Play (default `false`)
- `--order-id`: Google Play order ID
- `--revoke`: Revoke the purchased item after refunding (default `false`)

### gpc pricing

Inspect Google Play price conversions

```sh
gpc pricing
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc pricing convert-region-prices`: Convert one source price into Play region prices

#### gpc pricing convert-region-prices

Convert one source price into Play region prices

```sh
gpc pricing convert-region-prices [flags]
```

##### Flags

- `--currency`: Source price currency code, for example USD
- `--nanos`: Fractional source price nanos, 0 to 999999999 (default `0`)
- `--units`: Whole source price units (default `0`)

### gpc publish

Run high-level Google Play publishing workflows

```sh
gpc publish
```

#### Commands

- `gpc publish internal`: Publish an Android App Bundle to the internal track

#### gpc publish internal

Publish an Android App Bundle to the internal track

```sh
gpc publish internal [flags]
```

##### Flags

- `--aab`: Path to the Android App Bundle to upload
- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned publishing workflow without calling Google Play (default `false`)
- `--package`: Android package name, for example com.example.app
- `--release-name`: Release name shown in Play Console
- `--release-note`: Localized release note as language=text, repeatable (default `[]`)
- `--status`: Release status: completed, draft, halted, inProgress (default `completed`)
- `--user-fraction`: Staged rollout fraction for inProgress or halted releases (default `0`)

### gpc purchases

Inspect and manage Google Play purchase tokens

```sh
gpc purchases
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc purchases product`: Get or mutate one in-app product purchase
- `gpc purchases subscription`: Get or mutate one subscription purchase
- `gpc purchases voided`: Inspect voided Google Play purchases

#### gpc purchases product

Get or mutate one in-app product purchase

```sh
gpc purchases product [flags]
```

##### Flags

- `--product-id`: Optional in-app product ID hint for stable output when Google omits line items
- `--token`: Purchase token

##### Commands

- `gpc purchases product acknowledge`: Acknowledge an in-app product purchase
- `gpc purchases product consume`: Consume an in-app product purchase

##### gpc purchases product acknowledge

Acknowledge an in-app product purchase

```sh
gpc purchases product acknowledge [flags]
```

###### Flags

- `--confirm`: Apply the product purchase mutation (default `false`)
- `--developer-payload`: Optional developer payload to attach to the acknowledgement
- `--dry-run`: Print the planned product purchase mutation without calling Google Play (default `false`)
- `--product-id`: In-app product ID
- `--token`: Purchase token

##### gpc purchases product consume

Consume an in-app product purchase

```sh
gpc purchases product consume [flags]
```

###### Flags

- `--confirm`: Apply the product purchase mutation (default `false`)
- `--dry-run`: Print the planned product purchase mutation without calling Google Play (default `false`)
- `--product-id`: In-app product ID
- `--token`: Purchase token

#### gpc purchases subscription

Get or mutate one subscription purchase

```sh
gpc purchases subscription [flags]
```

##### Flags

- `--token`: Purchase token

##### Commands

- `gpc purchases subscription acknowledge`: Acknowledge a subscription purchase through the legacy subscriptions API
- `gpc purchases subscription cancel`: Cancel a subscription purchase through the subscriptions v2 API
- `gpc purchases subscription revoke`: Revoke a subscription purchase

##### gpc purchases subscription acknowledge

Acknowledge a subscription purchase through the legacy subscriptions API

```sh
gpc purchases subscription acknowledge [flags]
```

###### Flags

- `--confirm`: Apply the subscription purchase mutation (default `false`)
- `--developer-payload`: Optional developer payload to attach to the acknowledgement
- `--dry-run`: Print the planned subscription purchase mutation without calling Google Play (default `false`)
- `--subscription-id`: Legacy subscription product ID
- `--token`: Purchase token

##### gpc purchases subscription cancel

Cancel a subscription purchase through the subscriptions v2 API

```sh
gpc purchases subscription cancel [flags]
```

###### Flags

- `--cancellation-type`: Cancellation type: userRequestedStopRenewals or developerRequestedStopPayments
- `--confirm`: Apply the subscription purchase mutation (default `false`)
- `--dry-run`: Print the planned subscription purchase mutation without calling Google Play (default `false`)
- `--token`: Purchase token

##### gpc purchases subscription revoke

Revoke a subscription purchase

```sh
gpc purchases subscription revoke [flags]
```

###### Flags

- `--confirm`: Apply the subscription revocation (default `false`)
- `--dry-run`: Print the planned subscription revocation without calling Google Play (default `false`)
- `--refund`: Refund type: full, prorated, or item
- `--refund-product-id`: Subscription product ID to refund when --refund item is used
- `--token`: Purchase token

#### gpc purchases voided

Inspect voided Google Play purchases

```sh
gpc purchases voided
```

##### Commands

- `gpc purchases voided list`: List voided purchases

##### gpc purchases voided list

List voided purchases

```sh
gpc purchases voided list [flags]
```

###### Flags

- `--end-time`: Newest seen-as-voided time in epoch milliseconds (default `0`)
- `--include-quantity-based-partial-refund`: Include quantity-based partial refunds (default `false`)
- `--max-results`: Maximum voided purchases to return (default `0`)
- `--start-index`: Zero-based voided purchase offset (default `0`)
- `--start-time`: Oldest seen-as-voided time in epoch milliseconds (default `0`)
- `--token`: Pagination token from a previous response
- `--type`: Voided purchase type: 0 for products, 1 for products and subscriptions (default `0`)

### gpc releases

Upload and manage Google Play releases

```sh
gpc releases
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc releases halt`: Halt a staged release
- `gpc releases list`: List releases for a track
- `gpc releases promote`: Promote a release from one track to another
- `gpc releases resume`: Resume a staged release
- `gpc releases upload`: Upload an APK or Android App Bundle to a track

#### gpc releases halt

Halt a staged release

```sh
gpc releases halt [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned halt workflow without calling Google Play (default `false`)
- `--track`: Track name (default `production`)
- `--version-code`: Version code to halt (default `0`)

#### gpc releases list

List releases for a track

```sh
gpc releases list [flags]
```

##### Flags

- `--track`: Track name, for example internal, alpha, beta, or production (default `internal`)

#### gpc releases promote

Promote a release from one track to another

```sh
gpc releases promote [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned promotion workflow without calling Google Play (default `false`)
- `--from`: Source track name (default `internal`)
- `--release-note`: Localized release note as language=text, repeatable (default `[]`)
- `--status`: Target release status: completed, draft, halted, inProgress (default `draft`)
- `--to`: Target track name (default `production`)
- `--user-fraction`: Staged rollout fraction for inProgress or halted releases (default `0`)
- `--version-code`: Version code to promote (default `0`)

#### gpc releases resume

Resume a staged release

```sh
gpc releases resume [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned resume workflow without calling Google Play (default `false`)
- `--status`: Target release status: completed or inProgress
- `--track`: Track name (default `production`)
- `--user-fraction`: Staged rollout fraction when status is inProgress (default `0`)
- `--version-code`: Version code to resume (default `0`)

#### gpc releases upload

Upload an APK or Android App Bundle to a track

```sh
gpc releases upload [flags]
```

##### Flags

- `--aab`: Path to the Android App Bundle to upload
- `--apk`: Path to the APK to upload
- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned release upload workflow without calling Google Play (default `false`)
- `--release-name`: Release name shown in Play Console
- `--release-note`: Localized release note as language=text, repeatable (default `[]`)
- `--status`: Release status: completed, draft, halted, inProgress (default `completed`)
- `--track`: Track name, for example internal, alpha, beta, or production (default `internal`)
- `--user-fraction`: Staged rollout fraction for inProgress or halted releases (default `0`)

### gpc reviews

Read and reply to Google Play reviews

```sh
gpc reviews
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc reviews get`: Get one Google Play review
- `gpc reviews list`: List Google Play reviews
- `gpc reviews reply`: Reply to a Google Play review

#### gpc reviews get

Get one Google Play review

```sh
gpc reviews get [flags]
```

##### Flags

- `--review-id`: Google Play review ID
- `--translation-language`: Language localization code for translated review text

#### gpc reviews list

List Google Play reviews

```sh
gpc reviews list [flags]
```

##### Flags

- `--max-results`: Maximum reviews to return (default `0`)
- `--start-index`: Zero-based review offset (default `0`)
- `--token`: Pagination token from a previous response
- `--translation-language`: Language localization code for translated reviews

#### gpc reviews reply

Reply to a Google Play review

```sh
gpc reviews reply [flags]
```

##### Flags

- `--confirm`: Apply the public review reply (default `false`)
- `--dry-run`: Print the planned review reply without calling Google Play (default `false`)
- `--review-id`: Google Play review ID
- `--text`: Public developer reply text

### gpc schema

Print the Google Play discovery schema

```sh
gpc schema [flags]
```

#### Flags

- `--method`: Filter by discovery method name or ID, for example list or androidpublisher.edits.tracks.list
- `--resource`: Filter by dotted discovery resource path, for example edits.tracks

### gpc search

Search gpc commands and flags

```sh
gpc search QUERY... [flags]
```

#### Flags

- `--limit`: Maximum number of matches; 0 returns all matches (default `20`)

### gpc snitch

Report gpc friction

```sh
gpc snitch
```

#### Commands

- `gpc snitch report`: Generate a GitHub issue URL for CLI friction

#### gpc snitch report

Generate a GitHub issue URL for CLI friction

```sh
gpc snitch report [flags]
```

##### Flags

- `--body`: Issue body
- `--command`: gpc command or workflow that caused friction
- `--label`: GitHub issue label; repeatable (default `[]`)
- `--repo`: GitHub repository as owner/name (default `aljrico/Google-Play-Console-CLI`)
- `--title`: Short issue title

### gpc status

Summarize Google Play release status

```sh
gpc status [flags]
```

#### Flags

- `--include-draft`: Include draft releases in the status summary (default `false`)
- `--package`: Android package name, for example com.example.app

### gpc subscription-offers

Inspect Google Play subscription offers

```sh
gpc subscription-offers
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc subscription-offers activate`: activate a subscription offer
- `gpc subscription-offers batch-activate`: activate multiple subscription offers
- `gpc subscription-offers batch-deactivate`: deactivate multiple subscription offers
- `gpc subscription-offers batch-get`: Get multiple subscription offers
- `gpc subscription-offers deactivate`: deactivate a subscription offer
- `gpc subscription-offers delete`: Delete a draft subscription offer
- `gpc subscription-offers get`: Get one subscription offer
- `gpc subscription-offers list`: List subscription offers

#### gpc subscription-offers activate

activate a subscription offer

```sh
gpc subscription-offers activate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--confirm`: Apply the subscription offer state update (default `false`)
- `--dry-run`: Print the planned subscription offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### gpc subscription-offers batch-activate

activate multiple subscription offers

```sh
gpc subscription-offers batch-activate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer batch state update (default `false`)
- `--dry-run`: Print the planned subscription offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/basePlanId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted

#### gpc subscription-offers batch-deactivate

deactivate multiple subscription offers

```sh
gpc subscription-offers batch-deactivate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer batch state update (default `false`)
- `--dry-run`: Print the planned subscription offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/basePlanId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted

#### gpc subscription-offers batch-get

Get multiple subscription offers

```sh
gpc subscription-offers batch-get [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans
- `--offer`: Offer to fetch as productId/basePlanId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent subscription product ID, or - for offers across products

#### gpc subscription-offers deactivate

deactivate a subscription offer

```sh
gpc subscription-offers deactivate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--confirm`: Apply the subscription offer state update (default `false`)
- `--dry-run`: Print the planned subscription offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### gpc subscription-offers delete

Delete a draft subscription offer

```sh
gpc subscription-offers delete [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--confirm`: Apply the subscription offer deletion (default `false`)
- `--dry-run`: Print the planned subscription offer deletion without calling Google Play (default `false`)
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### gpc subscription-offers get

Get one subscription offer

```sh
gpc subscription-offers get [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### gpc subscription-offers list

List subscription offers

```sh
gpc subscription-offers list [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for all base plans
- `--page-size`: Maximum offers to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--product-id`: Parent subscription product ID, or - for all products

### gpc subscriptions

Inspect Google Play monetization subscriptions

```sh
gpc subscriptions
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc subscriptions base-plan`: Manage subscription base plans
- `gpc subscriptions batch-get`: Get multiple monetization subscriptions
- `gpc subscriptions batch-patch-listings`: Batch patch localized subscription listings
- `gpc subscriptions delete`: Delete a draft-only monetization subscription
- `gpc subscriptions get`: Get one monetization subscription
- `gpc subscriptions list`: List monetization subscriptions
- `gpc subscriptions patch`: Patch a subscription listing

#### gpc subscriptions base-plan

Manage subscription base plans

```sh
gpc subscriptions base-plan
```

##### Commands

- `gpc subscriptions base-plan activate`: activate a subscription base plan
- `gpc subscriptions base-plan batch-activate`: Batch activate subscription base plans
- `gpc subscriptions base-plan batch-deactivate`: Batch deactivate subscription base plans
- `gpc subscriptions base-plan batch-migrate-prices`: Batch migrate subscription base plan prices
- `gpc subscriptions base-plan deactivate`: deactivate a subscription base plan

##### gpc subscriptions base-plan activate

activate a subscription base plan

```sh
gpc subscriptions base-plan activate [flags]
```

###### Flags

- `--base-plan-id`: Subscription base plan ID
- `--confirm`: Apply the base plan state update (default `false`)
- `--dry-run`: Print the planned base plan state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID

##### gpc subscriptions base-plan batch-activate

Batch activate subscription base plans

```sh
gpc subscriptions base-plan batch-activate [flags]
```

###### Flags

- `--base-plan`: Subscription base plan as productId/basePlanId; repeat for cross-subscription batches (default `[]`)
- `--base-plan-id`: Subscription base plan ID; repeat for multiple base plans (default `[]`)
- `--confirm`: Apply the base plan batch state update (default `false`)
- `--dry-run`: Print the planned base plan batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID, or - for base plans across subscriptions; inferred when --base-plan is used

##### gpc subscriptions base-plan batch-deactivate

Batch deactivate subscription base plans

```sh
gpc subscriptions base-plan batch-deactivate [flags]
```

###### Flags

- `--base-plan`: Subscription base plan as productId/basePlanId; repeat for cross-subscription batches (default `[]`)
- `--base-plan-id`: Subscription base plan ID; repeat for multiple base plans (default `[]`)
- `--confirm`: Apply the base plan batch state update (default `false`)
- `--dry-run`: Print the planned base plan batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID, or - for base plans across subscriptions; inferred when --base-plan is used

##### gpc subscriptions base-plan batch-migrate-prices

Batch migrate subscription base plan prices

```sh
gpc subscriptions base-plan batch-migrate-prices [flags]
```

###### Flags

- `--confirm`: Apply the base plan price migration (default `false`)
- `--dry-run`: Print the planned base plan price migration without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--migration`: Price migration as productId/basePlanId/REGION/RFC3339_TIME; repeat for multiple regions or base plans (default `[]`)
- `--price-increase-type`: Price increase type: optIn or optOut
- `--product-id`: Subscription product ID, or - for migrations across subscriptions; inferred from --migration values
- `--regions-version`: Google Play regions version required by batchMigratePrices

##### gpc subscriptions base-plan deactivate

deactivate a subscription base plan

```sh
gpc subscriptions base-plan deactivate [flags]
```

###### Flags

- `--base-plan-id`: Subscription base plan ID
- `--confirm`: Apply the base plan state update (default `false`)
- `--dry-run`: Print the planned base plan state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID

#### gpc subscriptions batch-get

Get multiple monetization subscriptions

```sh
gpc subscriptions batch-get [flags]
```

##### Flags

- `--product-id`: Subscription product ID; repeatable, up to 100 (default `[]`)

#### gpc subscriptions batch-patch-listings

Batch patch localized subscription listings

```sh
gpc subscriptions batch-patch-listings [flags]
```

##### Flags

- `--confirm`: Apply the subscription listing batch patch (default `false`)
- `--dry-run`: Print the planned subscription listing batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--listing`: CSV listing patch productId,language,title,description; repeat for multiple localized listings (default `[]`)
- `--regions-version`: Google Play regions version required by subscriptions.batchUpdate

#### gpc subscriptions delete

Delete a draft-only monetization subscription

```sh
gpc subscriptions delete [flags]
```

##### Flags

- `--confirm`: Apply the subscription deletion (default `false`)
- `--dry-run`: Print the planned subscription deletion without calling Google Play (default `false`)
- `--product-id`: Subscription product ID

#### gpc subscriptions get

Get one monetization subscription

```sh
gpc subscriptions get [flags]
```

##### Flags

- `--product-id`: Subscription product ID

#### gpc subscriptions list

List monetization subscriptions

```sh
gpc subscriptions list [flags]
```

##### Flags

- `--page-size`: Maximum subscriptions to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--show-archived`: Deprecated by Google; subscription archiving is no longer supported (default `false`)

#### gpc subscriptions patch

Patch a subscription listing

```sh
gpc subscriptions patch [flags]
```

##### Flags

- `--benefit`: Localized subscription benefit; repeatable, up to 4 (default `[]`)
- `--confirm`: Apply the subscription listing patch (default `false`)
- `--description`: Localized subscription description
- `--dry-run`: Print the planned subscription listing patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--listing-language`: BCP-47 language code for the listing to patch, for example en-US
- `--product-id`: Subscription product ID
- `--regions-version`: Google Play regions version required by subscriptions.patch
- `--title`: Localized subscription title

### gpc system-apks

Inspect Google Play system APK variants

```sh
gpc system-apks
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc system-apks variants`: Inspect generated system APK variants

#### gpc system-apks variants

Inspect generated system APK variants

```sh
gpc system-apks variants
```

##### Commands

- `gpc system-apks variants list`: List system APK variants for an App Bundle version

##### gpc system-apks variants list

List system APK variants for an App Bundle version

```sh
gpc system-apks variants list [flags]
```

###### Flags

- `--version-code`: App Bundle version code (default `0`)

### gpc testers

Manage Google Play track tester groups

```sh
gpc testers
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc testers get`: Get Google Groups configured as testers for a track
- `gpc testers update`: Replace Google Groups configured as testers for a track

#### gpc testers get

Get Google Groups configured as testers for a track

```sh
gpc testers get [flags]
```

##### Flags

- `--track`: Track name, for example internal, alpha, beta, or production (default `internal`)

#### gpc testers update

Replace Google Groups configured as testers for a track

```sh
gpc testers update [flags]
```

##### Flags

- `--clear`: Remove all testing Google Groups from the track (default `false`)
- `--confirm`: Commit the tester update (default `false`)
- `--dry-run`: Print the planned tester update without calling Google Play (default `false`)
- `--google-group`: Testing Google Group email address, repeatable (default `[]`)
- `--track`: Track name, for example internal, alpha, beta, or production (default `internal`)

### gpc tracks

Manage Google Play release tracks

```sh
gpc tracks
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc tracks list`: List release tracks

### gpc users

Inspect and manage Google Play Console users

```sh
gpc users
```

#### Commands

- `gpc users create`: Grant developer-account access to a user
- `gpc users delete`: Remove all developer-account access for a user
- `gpc users list`: List users with access to a developer account
- `gpc users patch`: Replace developer-account access fields for a user

#### gpc users create

Grant developer-account access to a user

```sh
gpc users create [flags]
```

##### Flags

- `--confirm`: Apply the user creation (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned user creation without calling Google Play (default `false`)
- `--expiration-time`: Optional RFC3339 access expiration time
- `--permission`: Developer-account permission, repeatable (default `[]`)
- `--user-email`: Play Console user email

#### gpc users delete

Remove all developer-account access for a user

```sh
gpc users delete [flags]
```

##### Flags

- `--confirm`: Apply the user deletion (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned user deletion without calling Google Play (default `false`)
- `--name`: User resource name, developers/{developer}/users/{email}
- `--user-email`: Play Console user email

#### gpc users list

List users with access to a developer account

```sh
gpc users list [flags]
```

##### Flags

- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--page-size`: Maximum users to return; use -1 to disable pagination (default `0`)
- `--page-token`: Pagination token from a previous response

#### gpc users patch

Replace developer-account access fields for a user

```sh
gpc users patch [flags]
```

##### Flags

- `--confirm`: Apply the user patch (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned user patch without calling Google Play (default `false`)
- `--expiration-time`: Optional RFC3339 access expiration time
- `--name`: User resource name, developers/{developer}/users/{email}
- `--permission`: Developer-account permission, repeatable; replaces the account-level permission list when provided (default `[]`)
- `--user-email`: Play Console user email

### gpc validate

Validate a temporary Google Play edit

```sh
gpc validate [flags]
```

#### Flags

- `--package`: Android package name, for example com.example.app

### gpc vitals

Inspect Google Play Developer Reporting vitals

```sh
gpc vitals
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `gpc vitals anomalies`: List Android vitals anomalies
- `gpc vitals errors`: Search Android vitals errors
- `gpc vitals metric-set`: Inspect Android vitals metric set metadata

#### gpc vitals anomalies

List Android vitals anomalies

```sh
gpc vitals anomalies
```

##### Commands

- `gpc vitals anomalies list`: List Android vitals anomalies

##### gpc vitals anomalies list

List Android vitals anomalies

```sh
gpc vitals anomalies list [flags]
```

###### Flags

- `--filter`: AIP-160 anomaly filter, for example activeBetween("2026-05-01T00:00:00Z", "2026-05-19T00:00:00Z")
- `--page-size`: Maximum anomalies to return, capped by Google at 100 (default `0`)
- `--page-token`: Pagination token from a previous response

#### gpc vitals errors

Search Android vitals errors

```sh
gpc vitals errors
```

##### Commands

- `gpc vitals errors issues`: Search grouped Android vitals error issues
- `gpc vitals errors reports`: Search Android vitals error reports

##### gpc vitals errors issues

Search grouped Android vitals error issues

```sh
gpc vitals errors issues
```

###### Commands

- `gpc vitals errors issues search`: Search grouped Android vitals error issues

###### gpc vitals errors issues search

Search grouped Android vitals error issues

```sh
gpc vitals errors issues search [flags]
```

###### Flags

- `--end-date`: End date, exclusive, in YYYY-MM-DD format
- `--filter`: AIP-160 filter expression for issue fields
- `--order-by`: Order issues by errorReportCount|distinctUsers asc|desc
- `--page-size`: Maximum issues to return, capped by Google at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--sample-error-report-limit`: Sample reports per issue; Google currently supports 0 or 1 (default `0`)
- `--start-date`: Start date, inclusive, in YYYY-MM-DD format
- `--time-zone`: Time zone for the interval; only UTC is supported when set

##### gpc vitals errors reports

Search Android vitals error reports

```sh
gpc vitals errors reports
```

###### Commands

- `gpc vitals errors reports search`: Search individual Android vitals error reports

###### gpc vitals errors reports search

Search individual Android vitals error reports

```sh
gpc vitals errors reports search [flags]
```

###### Flags

- `--end-date`: End date, exclusive, in YYYY-MM-DD format
- `--filter`: AIP-160 filter expression for report fields
- `--page-size`: Maximum reports to return, capped by Google at 100 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--start-date`: Start date, inclusive, in YYYY-MM-DD format
- `--time-zone`: Time zone for the interval; only UTC is supported when set

#### gpc vitals metric-set

Inspect Android vitals metric set metadata

```sh
gpc vitals metric-set
```

##### Commands

- `gpc vitals metric-set get`: Get Android vitals metric set freshness
- `gpc vitals metric-set query`: Query Android vitals metric rows

##### gpc vitals metric-set get

Get Android vitals metric set freshness

```sh
gpc vitals metric-set get [flags]
```

###### Flags

- `--metric-set`: Vitals metric set: anr-rate, crash-rate, error-count, excessive-wakeup-rate, lmk-rate, slow-rendering-rate, slow-start-rate, stuck-background-wakelock-rate

##### gpc vitals metric-set query

Query Android vitals metric rows

```sh
gpc vitals metric-set query [flags]
```

###### Flags

- `--aggregation`: Aggregation period: DAILY, or HOURLY where the metric set supports it
- `--dimension`: Dimension to break down by; repeat for multiple dimensions (default `[]`)
- `--end-date`: End date, exclusive, in YYYY-MM-DD format
- `--filter`: AIP-160 filter expression over supported dimensions
- `--metric`: Metric to request; repeat for multiple metrics (default `[]`)
- `--metric-set`: Vitals metric set: anr-rate, crash-rate, error-count, excessive-wakeup-rate, lmk-rate, slow-rendering-rate, slow-start-rate, stuck-background-wakelock-rate
- `--page-size`: Maximum rows to return, capped by Google at 100000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--start-date`: Start date, inclusive, in YYYY-MM-DD format
- `--time-zone`: IANA time zone for daily aggregation, for example America/Los_Angeles
- `--user-cohort`: User cohort where supported: OS_PUBLIC, OS_BETA, or APP_TESTERS

### gpc web

Inspect Play Console browser automation support

```sh
gpc web
```

#### Commands

- `gpc web status`: Explain the Play Console browser automation boundary

### gpc workflow

Run repo-local gpc workflows

```sh
gpc workflow
```

#### Flags

- `--file`: Workflow JSON file (default `.gpc/workflow.json`)

#### Commands

- `gpc workflow list`: List configured workflows
- `gpc workflow run`: Run one configured workflow

#### gpc workflow run

Run one configured workflow

```sh
gpc workflow run NAME [flags]
```

##### Flags

- `--confirm`: Execute the workflow shell steps (default `false`)
- `--dry-run`: Print the planned workflow steps without executing them (default `false`)
- `--workdir`: Working directory for shell steps; defaults to the workflow root

