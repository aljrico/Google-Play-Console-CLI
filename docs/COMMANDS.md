# Command Reference

## playpub

Go CLI for Google Play Developer API workflows

```sh
playpub [flags]
```

### Global Flags

- `--output` / `-o`: Output format: json, table, markdown (default `json`)
- `--pretty`: Pretty-print JSON output (default `false`)

### Commands

- `playpub account`: Inspect local Google Play account configuration
- `playpub analytics`: Summarize Google Play statistics reports
- `playpub app-recovery`: Inspect and manage Google Play app recovery actions
- `playpub apps`: Inspect Google Play apps
- `playpub auth`: Manage Google Play API authentication
- `playpub capabilities`: List playpub command parity and capability status
- `playpub completion`: Generate the autocompletion script for the specified shell
- `playpub data-safety`: Update Google Play data safety declarations
- `playpub details`: Manage app-level Google Play details
- `playpub device-tier-configs`: Inspect Google Play device tier configs
- `playpub diff`: Compare local Google Play payloads
- `playpub docs`: Print embedded playpub documentation
- `playpub finance`: Summarize Google Play financial reports
- `playpub generated-apks`: Inspect generated APK metadata for an App Bundle version
- `playpub grants`: Manage Google Play app access grants
- `playpub images`: Manage localized Google Play store images
- `playpub in-app-products`: Inspect legacy Google Play in-app products
- `playpub init`: Create a local playpub workspace
- `playpub insights`: Summarize Google Play data exports
- `playpub install-skills`: Install bundled playpub agent skills
- `playpub internal-sharing`: Upload artifacts to Google Play internal app sharing
- `playpub listings`: Manage localized Google Play store listings
- `playpub metadata`: Apply app details and localized listings from a file
- `playpub migrate`: Inspect local metadata for migration
- `playpub notifications`: Inspect Google Play notification payloads
- `playpub notify`: Send release workflow notifications
- `playpub one-time-product-offers`: Inspect Google Play one-time product offers
- `playpub one-time-products`: Inspect Google Play one-time products
- `playpub orders`: Inspect and refund Google Play orders
- `playpub pricing`: Build and inspect Google Play price conversions
- `playpub publish`: Run high-level Google Play publishing workflows
- `playpub purchases`: Inspect and manage Google Play purchase tokens
- `playpub releases`: Upload and manage Google Play releases
- `playpub reviews`: Read and reply to Google Play reviews
- `playpub schema`: Print the Google Play discovery schema
- `playpub search`: Search playpub commands and flags
- `playpub snitch`: Report playpub friction
- `playpub status`: Summarize Google Play release status
- `playpub subscription-offers`: Inspect Google Play subscription offers
- `playpub subscriptions`: Inspect Google Play monetization subscriptions
- `playpub system-apks`: Inspect Google Play system APK variants
- `playpub testers`: Manage Google Play track tester groups
- `playpub tracks`: Manage Google Play release tracks
- `playpub users`: Inspect and manage Google Play Console users
- `playpub validate`: Validate a temporary Google Play edit
- `playpub version`: Print version information
- `playpub vitals`: Inspect Google Play Developer Reporting vitals
- `playpub web`: Inspect Play Console browser automation support
- `playpub workflow`: Run repo-local playpub workflows

### playpub account

Inspect local Google Play account configuration

```sh
playpub account
```

#### Commands

- `playpub account status`: Summarize local account and service account metadata

### playpub analytics

Summarize Google Play statistics reports

```sh
playpub analytics
```

#### Commands

- `playpub analytics stats`: Summarize downloaded Play statistics CSVs

#### playpub analytics stats

Summarize downloaded Play statistics CSVs

```sh
playpub analytics stats
```

##### Commands

- `playpub analytics stats download`: Download a Play statistics report CSV from Google Cloud Storage
- `playpub analytics stats summarize`: Summarize a Play statistics report CSV

##### playpub analytics stats download

Download a Play statistics report CSV from Google Cloud Storage

```sh
playpub analytics stats download [flags]
```

###### Flags

- `--bucket`: Google Play reports bucket, for example pubsite_prod_rev_0123456789
- `--dry-run`: Print the planned report download without calling Google Cloud Storage (default `false`)
- `--file`: Destination .csv or .zip path
- `--force`: Overwrite the destination file (default `false`)
- `--object`: Cloud Storage object path for the report

##### playpub analytics stats summarize

Summarize a Play statistics report CSV

```sh
playpub analytics stats summarize [flags]
```

###### Flags

- `--file`: Downloaded Google Play statistics report CSV

### playpub app-recovery

Inspect and manage Google Play app recovery actions

```sh
playpub app-recovery
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub app-recovery add-targeting`: Add targeting to an app recovery action
- `playpub app-recovery cancel`: Cancel an app recovery action
- `playpub app-recovery create`: Create a draft remote in-app update recovery action
- `playpub app-recovery deploy`: Deploy a draft app recovery action
- `playpub app-recovery list`: List app recovery actions for a version code

#### playpub app-recovery add-targeting

Add targeting to an app recovery action

```sh
playpub app-recovery add-targeting [flags]
```

##### Flags

- `--all-users`: Target all users (default `false`)
- `--confirm`: Apply the app recovery targeting update (default `false`)
- `--dry-run`: Print the planned app recovery targeting update without calling Google Play (default `false`)
- `--id`: App recovery action ID
- `--region`: ISO 3166-1 alpha-2 region code to add, repeatable (default `[]`)
- `--sdk-level`: Android SDK level to add, repeatable (default `[]`)

#### playpub app-recovery cancel

Cancel an app recovery action

```sh
playpub app-recovery cancel [flags]
```

##### Flags

- `--confirm`: Apply the app recovery mutation (default `false`)
- `--dry-run`: Print the planned app recovery mutation without calling Google Play (default `false`)
- `--id`: App recovery action ID

#### playpub app-recovery create

Create a draft remote in-app update recovery action

```sh
playpub app-recovery create [flags]
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

#### playpub app-recovery deploy

Deploy a draft app recovery action

```sh
playpub app-recovery deploy [flags]
```

##### Flags

- `--confirm`: Apply the app recovery mutation (default `false`)
- `--dry-run`: Print the planned app recovery mutation without calling Google Play (default `false`)
- `--id`: App recovery action ID

#### playpub app-recovery list

List app recovery actions for a version code

```sh
playpub app-recovery list [flags]
```

##### Flags

- `--version-code`: Version code targeted by recovery actions (default `0`)

### playpub apps

Inspect Google Play apps

```sh
playpub apps
```

#### Commands

- `playpub apps list`: List apps visible to the active service account

### playpub auth

Manage Google Play API authentication

```sh
playpub auth
```

#### Commands

- `playpub auth doctor`: Validate the active auth profile
- `playpub auth login`: Store a service account profile
- `playpub auth status`: Show the active auth profile

#### playpub auth login

Store a service account profile

```sh
playpub auth login [flags]
```

##### Flags

- `--name`: Profile name
- `--service-account`: Path to a Google service account JSON key

### playpub capabilities

List playpub command parity and capability status

```sh
playpub capabilities [flags]
```

#### Flags

- `--section`: Filter by parity matrix section
- `--status`: Filter by status: planned, implemented, tested, documented, blocked, not applicable

### playpub completion

Generate the autocompletion script for the specified shell

```sh
playpub completion
```

#### Commands

- `playpub completion bash`: Generate the autocompletion script for bash
- `playpub completion fish`: Generate the autocompletion script for fish
- `playpub completion powershell`: Generate the autocompletion script for powershell
- `playpub completion zsh`: Generate the autocompletion script for zsh

#### playpub completion bash

Generate the autocompletion script for bash

```sh
playpub completion bash
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

#### playpub completion fish

Generate the autocompletion script for fish

```sh
playpub completion fish [flags]
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

#### playpub completion powershell

Generate the autocompletion script for powershell

```sh
playpub completion powershell [flags]
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

#### playpub completion zsh

Generate the autocompletion script for zsh

```sh
playpub completion zsh [flags]
```

##### Flags

- `--no-descriptions`: disable completion descriptions (default `false`)

### playpub data-safety

Update Google Play data safety declarations

```sh
playpub data-safety
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub data-safety update`: Upload a data safety CSV declaration

#### playpub data-safety update

Upload a data safety CSV declaration

```sh
playpub data-safety update [flags]
```

##### Flags

- `--confirm`: Apply the data safety update (default `false`)
- `--csv`: Path to the data safety CSV export
- `--dry-run`: Print the planned data safety update without calling Google Play (default `false`)

### playpub details

Manage app-level Google Play details

```sh
playpub details
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub details get`: Get app-level details
- `playpub details update`: Patch app-level details

#### playpub details update

Patch app-level details

```sh
playpub details update [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--contact-email`: User-visible support email
- `--contact-phone`: User-visible support phone
- `--contact-website`: User-visible support website
- `--default-language`: Default BCP-47 language, for example en-US
- `--dry-run`: Print the planned details update without calling Google Play (default `false`)

### playpub device-tier-configs

Inspect Google Play device tier configs

```sh
playpub device-tier-configs
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub device-tier-configs get`: Get one device tier config
- `playpub device-tier-configs list`: List device tier configs

#### playpub device-tier-configs get

Get one device tier config

```sh
playpub device-tier-configs get [flags]
```

##### Flags

- `--id`: Device tier config ID (default `0`)

#### playpub device-tier-configs list

List device tier configs

```sh
playpub device-tier-configs list [flags]
```

##### Flags

- `--page-size`: Maximum configs to return, 0 uses the Google default (default `0`)
- `--page-token`: Pagination token from a previous response

### playpub diff

Compare local Google Play payloads

```sh
playpub diff
```

#### Commands

- `playpub diff json`: Compare two JSON files with stable JSON Pointer paths

#### playpub diff json

Compare two JSON files with stable JSON Pointer paths

```sh
playpub diff json FROM TO [flags]
```

##### Flags

- `--fail-on-change`: Exit nonzero when the JSON files differ (default `false`)

### playpub docs

Print embedded playpub documentation

```sh
playpub docs [flags]
```

#### Commands

- `playpub docs commands`: Print generated command reference
- `playpub docs parity`: Print the asc-to-playpub parity matrix

### playpub finance

Summarize Google Play financial reports

```sh
playpub finance
```

#### Commands

- `playpub finance reports`: Summarize downloaded Play financial report CSVs

#### playpub finance reports

Summarize downloaded Play financial report CSVs

```sh
playpub finance reports
```

##### Commands

- `playpub finance reports download`: Download a Play financial report ZIP from Google Cloud Storage
- `playpub finance reports summarize`: Summarize a Play financial report CSV

##### playpub finance reports download

Download a Play financial report ZIP from Google Cloud Storage

```sh
playpub finance reports download [flags]
```

###### Flags

- `--bucket`: Google Play reports bucket, for example pubsite_prod_rev_0123456789
- `--dry-run`: Print the planned report download without calling Google Cloud Storage (default `false`)
- `--file`: Destination .csv or .zip path
- `--force`: Overwrite the destination file (default `false`)
- `--object`: Cloud Storage object path for the report

##### playpub finance reports summarize

Summarize a Play financial report CSV

```sh
playpub finance reports summarize [flags]
```

###### Flags

- `--file`: Downloaded Google Play earnings or estimated-sales CSV

### playpub generated-apks

Inspect generated APK metadata for an App Bundle version

```sh
playpub generated-apks
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub generated-apks download`: Download one generated APK by download ID
- `playpub generated-apks list`: List generated APKs for a version code

#### playpub generated-apks download

Download one generated APK by download ID

```sh
playpub generated-apks download [flags]
```

##### Flags

- `--download-id`: Generated APK download ID from generated-apks list
- `--dry-run`: Print the planned download without calling Google Play (default `false`)
- `--file`: Destination .apk path
- `--force`: Overwrite the destination file (default `false`)
- `--version-code`: App Bundle version code (default `0`)

#### playpub generated-apks list

List generated APKs for a version code

```sh
playpub generated-apks list [flags]
```

##### Flags

- `--version-code`: App Bundle version code (default `0`)

### playpub grants

Manage Google Play app access grants

```sh
playpub grants
```

#### Commands

- `playpub grants create`: Create app-level access for a Play Console user
- `playpub grants delete`: Delete an app-level access grant
- `playpub grants patch`: Replace app-level permissions for an access grant

#### playpub grants create

Create app-level access for a Play Console user

```sh
playpub grants create [flags]
```

##### Flags

- `--confirm`: Apply the grant creation (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned grant creation without calling Google Play (default `false`)
- `--package`: Android package name, for example com.example.app
- `--permission`: App-level grant permission, repeatable (default `[]`)
- `--user-email`: Play Console user email

#### playpub grants delete

Delete an app-level access grant

```sh
playpub grants delete [flags]
```

##### Flags

- `--confirm`: Apply the grant deletion (default `false`)
- `--dry-run`: Print the planned grant deletion without calling Google Play (default `false`)
- `--name`: Grant resource name, developers/{developer}/users/{email}/grants/{package}

#### playpub grants patch

Replace app-level permissions for an access grant

```sh
playpub grants patch [flags]
```

##### Flags

- `--confirm`: Apply the grant patch (default `false`)
- `--dry-run`: Print the planned grant patch without calling Google Play (default `false`)
- `--name`: Grant resource name, developers/{developer}/users/{email}/grants/{package}
- `--permission`: App-level grant permission, repeatable (default `[]`)

### playpub images

Manage localized Google Play store images

```sh
playpub images
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub images delete`: Delete one store image
- `playpub images delete-all`: Delete all store images for one language and image type
- `playpub images list`: List store images for one language and image type
- `playpub images upload`: Upload one store image

#### playpub images delete

Delete one store image

```sh
playpub images delete [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned image deletion without calling Google Play (default `false`)
- `--image-id`: Google Play image ID to delete
- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

#### playpub images delete-all

Delete all store images for one language and image type

```sh
playpub images delete-all [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned image deletion without calling Google Play (default `false`)
- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

#### playpub images list

List store images for one language and image type

```sh
playpub images list [flags]
```

##### Flags

- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

#### playpub images upload

Upload one store image

```sh
playpub images upload [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned image upload without calling Google Play (default `false`)
- `--file`: Path to a .jpg, .jpeg, or .png image
- `--language`: BCP-47 listing language, for example en-US
- `--type`: Image type: icon, featureGraphic, phoneScreenshots, sevenInchScreenshots, tenInchScreenshots, tvBanner, tvScreenshots, wearScreenshots

### playpub in-app-products

Inspect legacy Google Play in-app products

```sh
playpub in-app-products
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub in-app-products batch-delete`: Delete multiple legacy managed in-app products
- `playpub in-app-products batch-get`: Get multiple legacy in-app products
- `playpub in-app-products create`: Create a legacy managed in-app product
- `playpub in-app-products delete`: Delete a legacy managed in-app product
- `playpub in-app-products get`: Get one legacy in-app product
- `playpub in-app-products list`: List legacy in-app products
- `playpub in-app-products patch`: Patch a legacy managed in-app product

#### playpub in-app-products batch-delete

Delete multiple legacy managed in-app products

```sh
playpub in-app-products batch-delete [flags]
```

##### Flags

- `--confirm`: Delete the managed in-app products (default `false`)
- `--dry-run`: Print the planned managed in-app product batch deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--sku`: In-app product SKU; repeatable, up to 100 (default `[]`)

#### playpub in-app-products batch-get

Get multiple legacy in-app products

```sh
playpub in-app-products batch-get [flags]
```

##### Flags

- `--sku`: In-app product SKU; repeatable (default `[]`)

#### playpub in-app-products create

Create a legacy managed in-app product

```sh
playpub in-app-products create [flags]
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

#### playpub in-app-products delete

Delete a legacy managed in-app product

```sh
playpub in-app-products delete [flags]
```

##### Flags

- `--confirm`: Delete the managed in-app product (default `false`)
- `--dry-run`: Print the planned managed in-app product deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--sku`: In-app product SKU

#### playpub in-app-products get

Get one legacy in-app product

```sh
playpub in-app-products get [flags]
```

##### Flags

- `--sku`: In-app product SKU

#### playpub in-app-products list

List legacy in-app products

```sh
playpub in-app-products list [flags]
```

##### Flags

- `--token`: Pagination token from a previous response

#### playpub in-app-products patch

Patch a legacy managed in-app product

```sh
playpub in-app-products patch [flags]
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

### playpub init

Create a local playpub workspace

```sh
playpub init [flags]
```

#### Flags

- `--directory`: Directory for playpub helper files (default `.playpub`)
- `--dry-run`: Print the planned init files without writing (default `false`)
- `--force`: Overwrite existing playpub helper files (default `false`)

### playpub insights

Summarize Google Play data exports

```sh
playpub insights
```

#### Commands

- `playpub insights anomalies`: Summarize Android vitals anomalies
- `playpub insights reports`: Summarize finance and analytics report insights

#### playpub insights anomalies

Summarize Android vitals anomalies

```sh
playpub insights anomalies
```

##### Commands

- `playpub insights anomalies summarize`: Summarize a vitals anomalies JSON export

##### playpub insights anomalies summarize

Summarize a vitals anomalies JSON export

```sh
playpub insights anomalies summarize [flags]
```

###### Flags

- `--file`: JSON output from playpub vitals anomalies list

#### playpub insights reports

Summarize finance and analytics report insights

```sh
playpub insights reports
```

##### Commands

- `playpub insights reports summarize`: Summarize Play finance and statistics report insights

##### playpub insights reports summarize

Summarize Play finance and statistics report insights

```sh
playpub insights reports summarize [flags]
```

###### Flags

- `--finance-file`: Downloaded Google Play earnings or estimated-sales CSV; repeatable (default `[]`)
- `--stats-file`: Downloaded Google Play statistics CSV; repeatable (default `[]`)

### playpub install-skills

Install bundled playpub agent skills

```sh
playpub install-skills [flags]
```

#### Flags

- `--directory`: Directory for installed agent skills, defaults to ~/.agents/skills
- `--dry-run`: Print planned skill installs without writing (default `false`)
- `--force`: Overwrite existing skill files (default `false`)
- `--skill`: Install one bundled skill by name; repeat to install multiple (default `[]`)

#### Commands

- `playpub install-skills list`: List bundled playpub agent skills

### playpub internal-sharing

Upload artifacts to Google Play internal app sharing

```sh
playpub internal-sharing
```

#### Commands

- `playpub internal-sharing upload`: Upload an APK or Android App Bundle to internal app sharing

#### playpub internal-sharing upload

Upload an APK or Android App Bundle to internal app sharing

```sh
playpub internal-sharing upload [flags]
```

##### Flags

- `--aab`: Path to the Android App Bundle to upload
- `--apk`: Path to the APK to upload
- `--dry-run`: Print the planned internal sharing upload without calling Google Play (default `false`)
- `--package`: Android package name, for example com.example.app

### playpub listings

Manage localized Google Play store listings

```sh
playpub listings
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub listings delete`: Delete one localized store listing
- `playpub listings delete-all`: Delete all localized store listings
- `playpub listings get`: Get one localized store listing
- `playpub listings list`: List localized store listings
- `playpub listings update`: Create or update one localized store listing

#### playpub listings delete

Delete one localized store listing

```sh
playpub listings delete [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned listing deletion without calling Google Play (default `false`)
- `--language`: BCP-47 listing language, for example en-US

#### playpub listings delete-all

Delete all localized store listings

```sh
playpub listings delete-all [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned listing deletion without calling Google Play (default `false`)

#### playpub listings get

Get one localized store listing

```sh
playpub listings get [flags]
```

##### Flags

- `--language`: BCP-47 listing language, for example en-US

#### playpub listings update

Create or update one localized store listing

```sh
playpub listings update [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned listing update without calling Google Play (default `false`)
- `--full-description`: Localized full description
- `--language`: BCP-47 listing language, for example en-US
- `--short-description`: Localized short description
- `--title`: Localized app title
- `--video`: Promotional YouTube video URL

### playpub metadata

Apply app details and localized listings from a file

```sh
playpub metadata
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub metadata apply`: Apply a JSON metadata file through one Google Play edit

#### playpub metadata apply

Apply a JSON metadata file through one Google Play edit

```sh
playpub metadata apply [flags]
```

##### Flags

- `--changes-not-sent-for-review`: Commit without sending changes for review (required when the app is in an enforcement-required state) (default `false`)
- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned metadata update without calling Google Play (default `false`)
- `--file`: Path to metadata JSON

### playpub migrate

Inspect local metadata for migration

```sh
playpub migrate
```

#### Commands

- `playpub migrate supply`: Inspect fastlane supply metadata

#### playpub migrate supply

Inspect fastlane supply metadata

```sh
playpub migrate supply
```

##### Commands

- `playpub migrate supply changelogs`: Convert fastlane supply changelogs to release-note payloads
- `playpub migrate supply convert`: Convert fastlane supply listings to playpub metadata JSON
- `playpub migrate supply images`: Convert fastlane supply images to image upload payloads
- `playpub migrate supply inspect`: Inventory a fastlane supply metadata directory

##### playpub migrate supply changelogs

Convert fastlane supply changelogs to release-note payloads

```sh
playpub migrate supply changelogs [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)
- `--version-code`: Only include changelogs for this version code (default `0`)

##### playpub migrate supply convert

Convert fastlane supply listings to playpub metadata JSON

```sh
playpub migrate supply convert [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)

##### playpub migrate supply images

Convert fastlane supply images to image upload payloads

```sh
playpub migrate supply images [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)
- `--language`: Only include images for this BCP-47 listing language
- `--type`: Only include this image type

##### playpub migrate supply inspect

Inventory a fastlane supply metadata directory

```sh
playpub migrate supply inspect [flags]
```

###### Flags

- `--directory`: fastlane supply metadata directory (default `fastlane/metadata/android`)

### playpub notifications

Inspect Google Play notification payloads

```sh
playpub notifications
```

#### Commands

- `playpub notifications pubsub`: Set up Google Cloud Pub/Sub for Play notifications
- `playpub notifications rtdn`: Inspect real-time developer notifications

#### playpub notifications pubsub

Set up Google Cloud Pub/Sub for Play notifications

```sh
playpub notifications pubsub
```

##### Commands

- `playpub notifications pubsub pull`: Pull Pub/Sub messages for Play notifications
- `playpub notifications pubsub setup`: Create Pub/Sub resources for Play real-time developer notifications

##### playpub notifications pubsub pull

Pull Pub/Sub messages for Play notifications

```sh
playpub notifications pubsub pull [flags]
```

###### Flags

- `--ack`: Acknowledge pulled messages after output succeeds (default `false`)
- `--confirm`: Confirm acknowledgement when --ack is set (default `false`)
- `--decode-rtdn`: Decode each message data field as a Google Play RTDN payload (default `false`)
- `--max-messages`: Maximum Pub/Sub messages to pull (default `10`)
- `--project`: Google Cloud project ID that owns the Pub/Sub subscription
- `--subscription`: Pub/Sub subscription ID to pull from

##### playpub notifications pubsub setup

Create Pub/Sub resources for Play real-time developer notifications

```sh
playpub notifications pubsub setup [flags]
```

###### Flags

- `--ack-deadline`: Subscription acknowledgement deadline in seconds (default `10`)
- `--confirm`: Create Pub/Sub resources and grant the Google Play publisher role (default `false`)
- `--dry-run`: Print the planned Pub/Sub setup without calling Google Cloud (default `false`)
- `--project`: Google Cloud project ID that owns the Pub/Sub resources
- `--push-endpoint`: Optional HTTPS push endpoint; omit for pull subscriptions
- `--subscription`: Pub/Sub subscription ID to create
- `--topic`: Pub/Sub topic ID to create

#### playpub notifications rtdn

Inspect real-time developer notifications

```sh
playpub notifications rtdn
```

##### Commands

- `playpub notifications rtdn decode`: Decode a Pub/Sub RTDN push payload

##### playpub notifications rtdn decode

Decode a Pub/Sub RTDN push payload

```sh
playpub notifications rtdn decode [flags]
```

###### Flags

- `--data`: Inline RTDN JSON payload; required unless --file is set
- `--file`: RTDN JSON payload file; required unless --data is set
- `--unwrapped`: Decode an unwrapped push payload containing the developer notification directly (default `false`)

### playpub notify

Send release workflow notifications

```sh
playpub notify
```

#### Commands

- `playpub notify discord`: Send a Discord incoming webhook notification
- `playpub notify github`: Send a GitHub repository dispatch-shaped webhook notification
- `playpub notify google-chat`: Send a Google Chat incoming webhook notification
- `playpub notify mattermost`: Send a Mattermost incoming webhook notification
- `playpub notify send`: Send a JSON notification webhook
- `playpub notify slack`: Send a Slack incoming webhook notification
- `playpub notify teams`: Send a Microsoft Teams Workflows webhook notification

#### playpub notify discord

Send a Discord incoming webhook notification

```sh
playpub notify discord [flags]
```

##### Flags

- `--confirm`: Send the Discord webhook (default `false`)
- `--dry-run`: Print the Discord payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS Discord incoming webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the Discord incoming webhook URL (default `PLAYPUB_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the Discord incoming webhook URL

#### playpub notify github

Send a GitHub repository dispatch-shaped webhook notification

```sh
playpub notify github [flags]
```

##### Flags

- `--confirm`: Send the GitHub webhook (default `false`)
- `--dry-run`: Print the GitHub payload without sending (default `false`)
- `--event-type`: GitHub repository dispatch event_type (default `playpub.notify`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS GitHub repository dispatch webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the GitHub repository dispatch webhook URL (default `PLAYPUB_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the GitHub repository dispatch webhook URL

#### playpub notify google-chat

Send a Google Chat incoming webhook notification

```sh
playpub notify google-chat [flags]
```

##### Flags

- `--confirm`: Send the Google Chat webhook (default `false`)
- `--dry-run`: Print the Google Chat payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS Google Chat incoming webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the Google Chat incoming webhook URL (default `PLAYPUB_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the Google Chat incoming webhook URL

#### playpub notify mattermost

Send a Mattermost incoming webhook notification

```sh
playpub notify mattermost [flags]
```

##### Flags

- `--confirm`: Send the Mattermost webhook (default `false`)
- `--dry-run`: Print the Mattermost payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS Mattermost incoming webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the Mattermost incoming webhook URL (default `PLAYPUB_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the Mattermost incoming webhook URL

#### playpub notify send

Send a JSON notification webhook

```sh
playpub notify send [flags]
```

##### Flags

- `--confirm`: Send the notification webhook (default `false`)
- `--dry-run`: Print the notification payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the webhook URL (default `PLAYPUB_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the webhook URL

#### playpub notify slack

Send a Slack incoming webhook notification

```sh
playpub notify slack [flags]
```

##### Flags

- `--confirm`: Send the Slack webhook (default `false`)
- `--dry-run`: Print the Slack payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS Slack incoming webhook URL; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the Slack incoming webhook URL (default `PLAYPUB_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the Slack incoming webhook URL

#### playpub notify teams

Send a Microsoft Teams Workflows webhook notification

```sh
playpub notify teams [flags]
```

##### Flags

- `--confirm`: Send the Microsoft Teams webhook (default `false`)
- `--dry-run`: Print the Microsoft Teams payload without sending (default `false`)
- `--field`: Notification field as name=value; repeatable (default `[]`)
- `--message`: Notification message
- `--severity`: Notification severity label
- `--title`: Notification title
- `--webhook-url`: HTTPS Microsoft Teams Workflows webhook URL; legacy incoming connector URLs are also supported; http is allowed only for loopback hosts
- `--webhook-url-env`: Environment variable containing the Microsoft Teams Workflows webhook URL (default `PLAYPUB_NOTIFY_WEBHOOK_URL`)
- `--webhook-url-file`: File containing the Microsoft Teams Workflows webhook URL

### playpub one-time-product-offers

Inspect Google Play one-time product offers

```sh
playpub one-time-product-offers
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub one-time-product-offers activate`: activate a one-time product offer
- `playpub one-time-product-offers batch-activate`: activate multiple one-time product offers
- `playpub one-time-product-offers batch-cancel`: Cancel multiple pre-order one-time product offers and pending orders
- `playpub one-time-product-offers batch-deactivate`: deactivate multiple one-time product offers
- `playpub one-time-product-offers batch-delete`: Delete multiple one-time product offers
- `playpub one-time-product-offers batch-get`: Get multiple one-time product offers
- `playpub one-time-product-offers batch-patch-absolute-discounts`: Batch patch one-time product offer absolute discounts
- `playpub one-time-product-offers batch-patch-availability`: Batch patch one-time product offer regional availability
- `playpub one-time-product-offers batch-patch-no-overrides`: Batch reset one-time product offer regional discounts to no override
- `playpub one-time-product-offers batch-patch-relative-discounts`: Batch patch one-time product offer relative discounts
- `playpub one-time-product-offers cancel`: cancel a one-time product offer
- `playpub one-time-product-offers create`: Create a one-time product offer
- `playpub one-time-product-offers deactivate`: deactivate a one-time product offer
- `playpub one-time-product-offers get`: Get a one-time product offer
- `playpub one-time-product-offers list`: List one-time product offers

#### playpub one-time-product-offers activate

activate a one-time product offer

```sh
playpub one-time-product-offers activate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer state update (default `false`)
- `--dry-run`: Print the planned one-time product offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### playpub one-time-product-offers batch-activate

activate multiple one-time product offers

```sh
playpub one-time-product-offers batch-activate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer batch state update (default `false`)
- `--dry-run`: Print the planned one-time product offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### playpub one-time-product-offers batch-cancel

Cancel multiple pre-order one-time product offers and pending orders

```sh
playpub one-time-product-offers batch-cancel [flags]
```

##### Flags

- `--confirm`: Cancel the pre-order offers and their pending orders (default `false`)
- `--dry-run`: Print the planned pre-order offer cancellation without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### playpub one-time-product-offers batch-deactivate

deactivate multiple one-time product offers

```sh
playpub one-time-product-offers batch-deactivate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer batch state update (default `false`)
- `--dry-run`: Print the planned one-time product offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### playpub one-time-product-offers batch-delete

Delete multiple one-time product offers

```sh
playpub one-time-product-offers batch-delete [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer batch deletion (default `false`)
- `--dry-run`: Print the planned one-time product offer batch deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to delete as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted

#### playpub one-time-product-offers batch-get

Get multiple one-time product offers

```sh
playpub one-time-product-offers batch-get [flags]
```

##### Flags

- `--offer`: Offer to fetch as productId/purchaseOptionId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options

#### playpub one-time-product-offers batch-patch-absolute-discounts

Batch patch one-time product offer absolute discounts

```sh
playpub one-time-product-offers batch-patch-absolute-discounts [flags]
```

##### Flags

- `--absolute-discount`: Absolute discount patch as productId/purchaseOptionId/offerId/REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--confirm`: Apply the one-time product offer absolute discount batch patch (default `false`)
- `--dry-run`: Print the planned one-time product offer absolute discount batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted
- `--regions-version`: Google Play regions version required by oneTimeProductOffers.batchUpdate

#### playpub one-time-product-offers batch-patch-availability

Batch patch one-time product offer regional availability

```sh
playpub one-time-product-offers batch-patch-availability [flags]
```

##### Flags

- `--availability`: Availability patch as productId/purchaseOptionId/offerId/REGION:available|noLongerAvailable; repeatable (default `[]`)
- `--confirm`: Apply the one-time product offer availability batch patch (default `false`)
- `--dry-run`: Print the planned one-time product offer availability batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted
- `--regions-version`: Google Play regions version required by oneTimeProductOffers.batchUpdate

#### playpub one-time-product-offers batch-patch-no-overrides

Batch reset one-time product offer regional discounts to no override

```sh
playpub one-time-product-offers batch-patch-no-overrides [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer no-override batch patch (default `false`)
- `--dry-run`: Print the planned one-time product offer no-override batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--no-override`: No-override patch as productId/purchaseOptionId/offerId/REGION; repeatable (default `[]`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted
- `--regions-version`: Google Play regions version required by oneTimeProductOffers.batchUpdate

#### playpub one-time-product-offers batch-patch-relative-discounts

Batch patch one-time product offer relative discounts

```sh
playpub one-time-product-offers batch-patch-relative-discounts [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer relative discount batch patch (default `false`)
- `--dry-run`: Print the planned one-time product offer relative discount batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent one-time product ID, or - for offers across products; inferred when omitted
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for offers across purchase options; inferred when omitted
- `--regions-version`: Google Play regions version required by oneTimeProductOffers.batchUpdate
- `--relative-discount`: Relative discount patch as productId/purchaseOptionId/offerId/REGION:0.75, where 0.75 means the user pays 75% of the purchase option price; repeatable (default `[]`)

#### playpub one-time-product-offers cancel

cancel a one-time product offer

```sh
playpub one-time-product-offers cancel [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer state update (default `false`)
- `--dry-run`: Print the planned one-time product offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### playpub one-time-product-offers create

Create a one-time product offer

```sh
playpub one-time-product-offers create [flags]
```

##### Flags

- `--absolute-discount`: Basic create regional absolute discount as REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--confirm`: Create the one-time product offer (default `false`)
- `--dry-run`: Print the planned one-time product offer creation without calling Google Play (default `false`)
- `--end-time`: Basic offer end time as RFC3339
- `--from-json`: Path to a Google Play API or playpub JSON one-time product offer body
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--no-override`: Basic create regional no-override price mode as REGION; repeatable (default `[]`)
- `--offer-id`: One-time product offer ID
- `--offer-tag`: Basic create offer tag; repeatable (default `[]`)
- `--pre-order`: Build a basic pre-order offer instead of a discounted offer (default `false`)
- `--price-change-behavior`: Basic pre-order price behavior: PRE_ORDER_PRICE_CHANGE_BEHAVIOR_TWO_POINT_LOWEST or PRE_ORDER_PRICE_CHANGE_BEHAVIOR_NEW_ORDERS_ONLY
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID
- `--redemption-limit`: Basic discounted offer redemption limit from 0 to 50 (default `0`)
- `--regions-version`: Google Play regions version required by oneTimeProductOffers.batchUpdate
- `--relative-discount`: Basic create regional relative discount as REGION:0.5, where 0.5 means the user pays 50% of the purchase option price; repeatable (default `[]`)
- `--release-time`: Basic pre-order offer release time as RFC3339
- `--start-time`: Basic offer start time as RFC3339

#### playpub one-time-product-offers deactivate

deactivate a one-time product offer

```sh
playpub one-time-product-offers deactivate [flags]
```

##### Flags

- `--confirm`: Apply the one-time product offer state update (default `false`)
- `--dry-run`: Print the planned one-time product offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### playpub one-time-product-offers get

Get a one-time product offer

```sh
playpub one-time-product-offers get [flags]
```

##### Flags

- `--offer-id`: One-time product offer ID
- `--product-id`: Parent one-time product ID
- `--purchase-option-id`: Parent one-time product purchase option ID

#### playpub one-time-product-offers list

List one-time product offers

```sh
playpub one-time-product-offers list [flags]
```

##### Flags

- `--page-size`: Maximum offers to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--product-id`: Parent one-time product ID, or - for all products
- `--purchase-option-id`: Parent one-time product purchase option ID, or - for all purchase options

### playpub one-time-products

Inspect Google Play one-time products

```sh
playpub one-time-products
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub one-time-products batch-delete`: Delete multiple one-time products
- `playpub one-time-products batch-get`: Get multiple one-time products
- `playpub one-time-products batch-patch-listings`: Batch patch localized one-time product listings
- `playpub one-time-products create`: Create a one-time product
- `playpub one-time-products delete`: Delete a one-time product
- `playpub one-time-products get`: Get a one-time product
- `playpub one-time-products list`: List one-time products
- `playpub one-time-products patch`: Patch a one-time product listing
- `playpub one-time-products purchase-option`: Manage one-time product purchase options

#### playpub one-time-products batch-delete

Delete multiple one-time products

```sh
playpub one-time-products batch-delete [flags]
```

##### Flags

- `--confirm`: Apply the one-time product batch deletion (default `false`)
- `--dry-run`: Print the planned one-time product batch deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID; repeatable, up to 100 (default `[]`)

#### playpub one-time-products batch-get

Get multiple one-time products

```sh
playpub one-time-products batch-get [flags]
```

##### Flags

- `--product-id`: One-time product ID; repeatable, up to 100 (default `[]`)

#### playpub one-time-products batch-patch-listings

Batch patch localized one-time product listings

```sh
playpub one-time-products batch-patch-listings [flags]
```

##### Flags

- `--confirm`: Apply the one-time product listing batch patch (default `false`)
- `--dry-run`: Print the planned one-time product listing batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--listing`: CSV listing patch productId,language,title,description; repeat for multiple localized listings (default `[]`)
- `--regions-version`: Google Play regions version required by oneTimeProducts.batchUpdate

#### playpub one-time-products create

Create a one-time product

```sh
playpub one-time-products create [flags]
```

##### Flags

- `--confirm`: Create the one-time product (default `false`)
- `--dry-run`: Print the planned one-time product creation without calling Google Play (default `false`)
- `--from-json`: Path to a Google Play API or playpub JSON one-time product body
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--legacy-compatible`: Mark the basic buy purchase option as legacy compatible (default `true`)
- `--listing`: Basic create listing as CSV language,title,description; repeatable (default `[]`)
- `--multi-quantity`: Enable multi-quantity purchases on the basic buy purchase option (default `false`)
- `--offer-tag`: Basic create offer tag on the product and purchase option; repeatable (default `[]`)
- `--price`: Basic create regional price as REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--product-id`: One-time product ID
- `--purchase-option-id`: Basic create purchase option ID (default `buy`)
- `--regions-version`: Google Play regions version required by oneTimeProducts.patch

#### playpub one-time-products delete

Delete a one-time product

```sh
playpub one-time-products delete [flags]
```

##### Flags

- `--confirm`: Apply the one-time product deletion (default `false`)
- `--dry-run`: Print the planned one-time product deletion without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID

#### playpub one-time-products get

Get a one-time product

```sh
playpub one-time-products get [flags]
```

##### Flags

- `--product-id`: One-time product ID

#### playpub one-time-products list

List one-time products

```sh
playpub one-time-products list [flags]
```

##### Flags

- `--page-size`: Maximum one-time products to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response

#### playpub one-time-products patch

Patch a one-time product listing

```sh
playpub one-time-products patch [flags]
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

#### playpub one-time-products purchase-option

Manage one-time product purchase options

```sh
playpub one-time-products purchase-option
```

##### Commands

- `playpub one-time-products purchase-option activate`: activate a one-time product purchase option
- `playpub one-time-products purchase-option batch-delete`: Delete one-time product purchase options
- `playpub one-time-products purchase-option batch-patch-availability`: Batch patch one-time product purchase option availability
- `playpub one-time-products purchase-option batch-patch-prices`: Batch patch one-time product purchase option regional prices
- `playpub one-time-products purchase-option deactivate`: deactivate a one-time product purchase option

##### playpub one-time-products purchase-option activate

activate a one-time product purchase option

```sh
playpub one-time-products purchase-option activate [flags]
```

###### Flags

- `--confirm`: Apply the purchase option state update (default `false`)
- `--dry-run`: Print the planned purchase option state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID
- `--purchase-option-id`: One-time product purchase option ID

##### playpub one-time-products purchase-option batch-delete

Delete one-time product purchase options

```sh
playpub one-time-products purchase-option batch-delete [flags]
```

###### Flags

- `--confirm`: Apply the purchase option batch deletion (default `false`)
- `--dry-run`: Print the planned purchase option batch deletion without calling Google Play (default `false`)
- `--force`: Also delete associated offers under each purchase option (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent one-time product ID, or - when deleting across products; inferred when omitted
- `--purchase-option`: Purchase option to delete as productId/purchaseOptionId; repeatable, up to 100 (default `[]`)

##### playpub one-time-products purchase-option batch-patch-availability

Batch patch one-time product purchase option availability

```sh
playpub one-time-products purchase-option batch-patch-availability [flags]
```

###### Flags

- `--availability`: Availability patch as productId/purchaseOptionId/REGION:available|noLongerAvailable|availableIfReleased|availableForOffersOnly; repeatable (default `[]`)
- `--confirm`: Apply the purchase option availability batch patch (default `false`)
- `--dry-run`: Print the planned purchase option availability batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--regions-version`: Google Play regions version required by oneTimeProducts.batchUpdate

##### playpub one-time-products purchase-option batch-patch-prices

Batch patch one-time product purchase option regional prices

```sh
playpub one-time-products purchase-option batch-patch-prices [flags]
```

###### Flags

- `--confirm`: Apply the purchase option price batch patch (default `false`)
- `--dry-run`: Print the planned purchase option price batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--price`: Regional price patch as productId/purchaseOptionId/REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--regions-version`: Google Play regions version required by oneTimeProducts.batchUpdate

##### playpub one-time-products purchase-option deactivate

deactivate a one-time product purchase option

```sh
playpub one-time-products purchase-option deactivate [flags]
```

###### Flags

- `--confirm`: Apply the purchase option state update (default `false`)
- `--dry-run`: Print the planned purchase option state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: One-time product ID
- `--purchase-option-id`: One-time product purchase option ID

### playpub orders

Inspect and refund Google Play orders

```sh
playpub orders
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub orders batch-get`: Get multiple Google Play orders
- `playpub orders get`: Get one Google Play order
- `playpub orders refund`: Refund one Google Play order

#### playpub orders batch-get

Get multiple Google Play orders

```sh
playpub orders batch-get [flags]
```

##### Flags

- `--order-id`: Google Play order ID, repeatable (default `[]`)

#### playpub orders get

Get one Google Play order

```sh
playpub orders get [flags]
```

##### Flags

- `--order-id`: Google Play order ID

#### playpub orders refund

Refund one Google Play order

```sh
playpub orders refund [flags]
```

##### Flags

- `--confirm`: Apply the refund (default `false`)
- `--dry-run`: Print the planned refund without calling Google Play (default `false`)
- `--order-id`: Google Play order ID
- `--revoke`: Revoke the purchased item after refunding (default `false`)

### playpub pricing

Build and inspect Google Play price conversions

```sh
playpub pricing
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub pricing build-price-patches`: Build regional price patch arguments from converted Play prices
- `playpub pricing convert-region-prices`: Convert one source price into Play region prices

#### playpub pricing build-price-patches

Build regional price patch arguments from converted Play prices

```sh
playpub pricing build-price-patches [flags]
```

##### Flags

- `--base-plan-id`: Base plan ID for subscription targets
- `--from-json`: Path to playpub pricing convert-region-prices JSON output
- `--offer-id`: Offer ID for --target subscription-offer-phase
- `--phase-index`: Zero-based offer phase index for --target subscription-offer-phase (default `-1`)
- `--product-id`: One-time product or subscription product ID
- `--purchase-option-id`: Purchase option ID for --target one-time-product
- `--sku`: In-app product SKU for --target in-app-product
- `--target`: Patch target: in-app-product, one-time-product, subscription-base-plan, or subscription-offer-phase

#### playpub pricing convert-region-prices

Convert one source price into Play region prices

```sh
playpub pricing convert-region-prices [flags]
```

##### Flags

- `--currency`: Source price currency code, for example USD
- `--nanos`: Fractional source price nanos, 0 to 999999999 (default `0`)
- `--units`: Whole source price units (default `0`)

### playpub publish

Run high-level Google Play publishing workflows

```sh
playpub publish
```

#### Commands

- `playpub publish internal`: Publish an Android App Bundle to the internal track

#### playpub publish internal

Publish an Android App Bundle to the internal track

```sh
playpub publish internal [flags]
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

### playpub purchases

Inspect and manage Google Play purchase tokens

```sh
playpub purchases
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub purchases product`: Get or mutate one in-app product purchase
- `playpub purchases subscription`: Get or mutate one subscription purchase
- `playpub purchases voided`: Inspect voided Google Play purchases

#### playpub purchases product

Get or mutate one in-app product purchase

```sh
playpub purchases product [flags]
```

##### Flags

- `--product-id`: Optional in-app product ID hint for stable output when Google omits line items
- `--token`: Purchase token

##### Commands

- `playpub purchases product acknowledge`: Acknowledge an in-app product purchase
- `playpub purchases product consume`: Consume an in-app product purchase

##### playpub purchases product acknowledge

Acknowledge an in-app product purchase

```sh
playpub purchases product acknowledge [flags]
```

###### Flags

- `--confirm`: Apply the product purchase mutation (default `false`)
- `--developer-payload`: Optional developer payload to attach to the acknowledgement
- `--dry-run`: Print the planned product purchase mutation without calling Google Play (default `false`)
- `--product-id`: In-app product ID
- `--token`: Purchase token

##### playpub purchases product consume

Consume an in-app product purchase

```sh
playpub purchases product consume [flags]
```

###### Flags

- `--confirm`: Apply the product purchase mutation (default `false`)
- `--dry-run`: Print the planned product purchase mutation without calling Google Play (default `false`)
- `--product-id`: In-app product ID
- `--token`: Purchase token

#### playpub purchases subscription

Get or mutate one subscription purchase

```sh
playpub purchases subscription [flags]
```

##### Flags

- `--token`: Purchase token

##### Commands

- `playpub purchases subscription acknowledge`: Acknowledge a subscription purchase through the legacy subscriptions API
- `playpub purchases subscription cancel`: Cancel a subscription purchase through the subscriptions v2 API
- `playpub purchases subscription revoke`: Revoke a subscription purchase

##### playpub purchases subscription acknowledge

Acknowledge a subscription purchase through the legacy subscriptions API

```sh
playpub purchases subscription acknowledge [flags]
```

###### Flags

- `--confirm`: Apply the subscription purchase mutation (default `false`)
- `--developer-payload`: Optional developer payload to attach to the acknowledgement
- `--dry-run`: Print the planned subscription purchase mutation without calling Google Play (default `false`)
- `--subscription-id`: Legacy subscription product ID
- `--token`: Purchase token

##### playpub purchases subscription cancel

Cancel a subscription purchase through the subscriptions v2 API

```sh
playpub purchases subscription cancel [flags]
```

###### Flags

- `--cancellation-type`: Cancellation type: userRequestedStopRenewals or developerRequestedStopPayments
- `--confirm`: Apply the subscription purchase mutation (default `false`)
- `--dry-run`: Print the planned subscription purchase mutation without calling Google Play (default `false`)
- `--token`: Purchase token

##### playpub purchases subscription revoke

Revoke a subscription purchase

```sh
playpub purchases subscription revoke [flags]
```

###### Flags

- `--confirm`: Apply the subscription revocation (default `false`)
- `--dry-run`: Print the planned subscription revocation without calling Google Play (default `false`)
- `--refund`: Refund type: full, prorated, or item
- `--refund-product-id`: Subscription product ID to refund when --refund item is used
- `--token`: Purchase token

#### playpub purchases voided

Inspect voided Google Play purchases

```sh
playpub purchases voided
```

##### Commands

- `playpub purchases voided list`: List voided purchases

##### playpub purchases voided list

List voided purchases

```sh
playpub purchases voided list [flags]
```

###### Flags

- `--end-time`: Newest seen-as-voided time in epoch milliseconds (default `0`)
- `--include-quantity-based-partial-refund`: Include quantity-based partial refunds (default `false`)
- `--max-results`: Maximum voided purchases to return (default `0`)
- `--start-index`: Zero-based voided purchase offset (default `0`)
- `--start-time`: Oldest seen-as-voided time in epoch milliseconds (default `0`)
- `--token`: Pagination token from a previous response
- `--type`: Voided purchase type: 0 for products, 1 for products and subscriptions (default `0`)

### playpub releases

Upload and manage Google Play releases

```sh
playpub releases
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub releases halt`: Halt a staged release
- `playpub releases list`: List releases for a track
- `playpub releases promote`: Promote a release from one track to another
- `playpub releases resume`: Resume a staged release
- `playpub releases upload`: Upload an APK or Android App Bundle to a track

#### playpub releases halt

Halt a staged release

```sh
playpub releases halt [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned halt workflow without calling Google Play (default `false`)
- `--track`: Track name (default `production`)
- `--version-code`: Version code to halt (default `0`)

#### playpub releases list

List releases for a track

```sh
playpub releases list [flags]
```

##### Flags

- `--track`: Track name, for example internal, alpha, beta, or production (default `internal`)

#### playpub releases promote

Promote a release from one track to another

```sh
playpub releases promote [flags]
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

#### playpub releases resume

Resume a staged release

```sh
playpub releases resume [flags]
```

##### Flags

- `--confirm`: Commit the edit after validation (default `false`)
- `--dry-run`: Print the planned resume workflow without calling Google Play (default `false`)
- `--status`: Target release status: completed or inProgress
- `--track`: Track name (default `production`)
- `--user-fraction`: Staged rollout fraction when status is inProgress (default `0`)
- `--version-code`: Version code to resume (default `0`)

#### playpub releases upload

Upload an APK or Android App Bundle to a track

```sh
playpub releases upload [flags]
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

### playpub reviews

Read and reply to Google Play reviews

```sh
playpub reviews
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub reviews get`: Get one Google Play review
- `playpub reviews list`: List Google Play reviews
- `playpub reviews reply`: Reply to a Google Play review

#### playpub reviews get

Get one Google Play review

```sh
playpub reviews get [flags]
```

##### Flags

- `--review-id`: Google Play review ID
- `--translation-language`: Language localization code for translated review text

#### playpub reviews list

List Google Play reviews

```sh
playpub reviews list [flags]
```

##### Flags

- `--max-results`: Maximum reviews to return (default `0`)
- `--start-index`: Zero-based review offset (default `0`)
- `--token`: Pagination token from a previous response
- `--translation-language`: Language localization code for translated reviews

#### playpub reviews reply

Reply to a Google Play review

```sh
playpub reviews reply [flags]
```

##### Flags

- `--confirm`: Apply the public review reply (default `false`)
- `--dry-run`: Print the planned review reply without calling Google Play (default `false`)
- `--review-id`: Google Play review ID
- `--text`: Public developer reply text

### playpub schema

Print the Google Play discovery schema

```sh
playpub schema [flags]
```

#### Flags

- `--method`: Filter by discovery method name or ID, for example list or androidpublisher.edits.tracks.list
- `--resource`: Filter by dotted discovery resource path, for example edits.tracks

### playpub search

Search playpub commands and flags

```sh
playpub search QUERY... [flags]
```

#### Flags

- `--limit`: Maximum number of matches; 0 returns all matches (default `20`)

### playpub snitch

Report playpub friction

```sh
playpub snitch
```

#### Commands

- `playpub snitch report`: Generate a GitHub issue URL for CLI friction

#### playpub snitch report

Generate a GitHub issue URL for CLI friction

```sh
playpub snitch report [flags]
```

##### Flags

- `--body`: Issue body
- `--command`: playpub command or workflow that caused friction
- `--label`: GitHub issue label; repeatable (default `[]`)
- `--repo`: GitHub repository as owner/name (default `aljrico/Google-Play-Console-CLI`)
- `--title`: Short issue title

### playpub status

Summarize Google Play release status

```sh
playpub status [flags]
```

#### Flags

- `--include-draft`: Include draft releases in the status summary (default `false`)
- `--package`: Android package name, for example com.example.app

### playpub subscription-offers

Inspect Google Play subscription offers

```sh
playpub subscription-offers
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub subscription-offers activate`: activate a subscription offer
- `playpub subscription-offers batch-activate`: activate multiple subscription offers
- `playpub subscription-offers batch-deactivate`: deactivate multiple subscription offers
- `playpub subscription-offers batch-get`: Get multiple subscription offers
- `playpub subscription-offers batch-patch-availability`: Batch patch subscription offer regional availability
- `playpub subscription-offers batch-patch-phase-absolute-discounts`: Batch patch subscription offer phase absolute discounts
- `playpub subscription-offers batch-patch-phase-free`: Batch patch subscription offer phases to free
- `playpub subscription-offers batch-patch-phase-prices`: Batch patch subscription offer phase prices
- `playpub subscription-offers batch-patch-phase-relative-discounts`: Batch patch subscription offer phase relative discounts
- `playpub subscription-offers create`: Create a draft subscription offer
- `playpub subscription-offers deactivate`: deactivate a subscription offer
- `playpub subscription-offers delete`: Delete a draft subscription offer
- `playpub subscription-offers get`: Get one subscription offer
- `playpub subscription-offers list`: List subscription offers

#### playpub subscription-offers activate

activate a subscription offer

```sh
playpub subscription-offers activate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--confirm`: Apply the subscription offer state update (default `false`)
- `--dry-run`: Print the planned subscription offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### playpub subscription-offers batch-activate

activate multiple subscription offers

```sh
playpub subscription-offers batch-activate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer batch state update (default `false`)
- `--dry-run`: Print the planned subscription offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/basePlanId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted

#### playpub subscription-offers batch-deactivate

deactivate multiple subscription offers

```sh
playpub subscription-offers batch-deactivate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer batch state update (default `false`)
- `--dry-run`: Print the planned subscription offer batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer`: Offer to update as productId/basePlanId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted

#### playpub subscription-offers batch-get

Get multiple subscription offers

```sh
playpub subscription-offers batch-get [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans
- `--offer`: Offer to fetch as productId/basePlanId/offerId; repeatable, up to 100 (default `[]`)
- `--product-id`: Parent subscription product ID, or - for offers across products

#### playpub subscription-offers batch-patch-availability

Batch patch subscription offer regional availability

```sh
playpub subscription-offers batch-patch-availability [flags]
```

##### Flags

- `--availability`: Availability patch as productId/basePlanId/offerId/REGION:true|false; repeatable (default `[]`)
- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer availability batch patch (default `false`)
- `--dry-run`: Print the planned subscription offer availability batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted
- `--regions-version`: Google Play regions version required by subscriptionOffers.batchUpdate

#### playpub subscription-offers batch-patch-phase-absolute-discounts

Batch patch subscription offer phase absolute discounts

```sh
playpub subscription-offers batch-patch-phase-absolute-discounts [flags]
```

##### Flags

- `--absolute-discount`: Phase absolute discount patch as productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]; phaseIndex is zero-based, so 0 is the first phase; repeatable (default `[]`)
- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer phase absolute discount batch patch (default `false`)
- `--dry-run`: Print the planned subscription offer phase absolute discount batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted
- `--regions-version`: Google Play regions version required by subscriptionOffers.batchUpdate

#### playpub subscription-offers batch-patch-phase-free

Batch patch subscription offer phases to free

```sh
playpub subscription-offers batch-patch-phase-free [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer phase free batch patch (default `false`)
- `--dry-run`: Print the planned subscription offer phase free batch patch without calling Google Play (default `false`)
- `--free`: Phase free patch as productId/basePlanId/offerId/phaseIndex/REGION; phaseIndex is zero-based, so 0 is the first phase; repeatable (default `[]`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted
- `--regions-version`: Google Play regions version required by subscriptionOffers.batchUpdate

#### playpub subscription-offers batch-patch-phase-prices

Batch patch subscription offer phase prices

```sh
playpub subscription-offers batch-patch-phase-prices [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer phase price batch patch (default `false`)
- `--dry-run`: Print the planned subscription offer phase price batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--price`: Phase price patch as productId/basePlanId/offerId/phaseIndex/REGION:CURRENCY:UNITS[:NANOS]; phaseIndex is zero-based, so 0 is the first phase; repeatable (default `[]`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted
- `--regions-version`: Google Play regions version required by subscriptionOffers.batchUpdate

#### playpub subscription-offers batch-patch-phase-relative-discounts

Batch patch subscription offer phase relative discounts

```sh
playpub subscription-offers batch-patch-phase-relative-discounts [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for offers across base plans; inferred when omitted
- `--confirm`: Apply the subscription offer phase relative discount batch patch (default `false`)
- `--dry-run`: Print the planned subscription offer phase relative discount batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Parent subscription product ID, or - for offers across products; inferred when omitted
- `--regions-version`: Google Play regions version required by subscriptionOffers.batchUpdate
- `--relative-discount`: Phase relative discount patch as productId/basePlanId/offerId/phaseIndex/REGION:0.75; phaseIndex is zero-based, so 0 is the first phase; 0.75 means the user pays 75% of the base plan price prorated over the phase duration; repeatable (default `[]`)

#### playpub subscription-offers create

Create a draft subscription offer

```sh
playpub subscription-offers create [flags]
```

##### Flags

- `--absolute-discount`: Basic create regional phase absolute discount as REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--base-plan-id`: Parent subscription base plan ID
- `--confirm`: Create the draft subscription offer (default `false`)
- `--dry-run`: Print the planned subscription offer creation without calling Google Play (default `false`)
- `--free-region`: Basic create region with new-subscriber availability and a free phase price mode; repeatable (default `[]`)
- `--from-json`: Path to a Google Play API or playpub JSON subscription offer body
- `--offer-id`: Subscription offer ID
- `--offer-tag`: Basic create offer tag; repeatable (default `[]`)
- `--other-regions-absolute-eur-discount`: Basic create first phase other-regions absolute EUR discount as EUR:UNITS[:NANOS]
- `--other-regions-absolute-usd-discount`: Basic create first phase other-regions absolute USD discount as USD:UNITS[:NANOS]
- `--other-regions-eur-price`: Basic create first phase other-regions EUR price as EUR:UNITS[:NANOS]
- `--other-regions-free`: Basic create free phase mode for other regions (default `false`)
- `--other-regions-relative-discount`: Basic create first phase other-regions relative discount as 0.5
- `--other-regions-usd-price`: Basic create first phase other-regions USD price as USD:UNITS[:NANOS]
- `--phase-2-absolute-discount`: Basic create second phase regional absolute discount as REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--phase-2-duration`: Basic create second phase duration as an ISO 8601 period, for example P1M
- `--phase-2-free-region`: Basic create second phase region with a free price mode; repeatable (default `[]`)
- `--phase-2-other-regions-absolute-eur-discount`: Basic create second phase other-regions absolute EUR discount as EUR:UNITS[:NANOS]
- `--phase-2-other-regions-absolute-usd-discount`: Basic create second phase other-regions absolute USD discount as USD:UNITS[:NANOS]
- `--phase-2-other-regions-eur-price`: Basic create second phase other-regions EUR price as EUR:UNITS[:NANOS]
- `--phase-2-other-regions-relative-discount`: Basic create second phase other-regions relative discount as 0.5
- `--phase-2-other-regions-usd-price`: Basic create second phase other-regions USD price as USD:UNITS[:NANOS]
- `--phase-2-price`: Basic create second phase regional price as REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--phase-2-recurrence`: Basic create second phase recurrence count (default `1`)
- `--phase-2-relative-discount`: Basic create second phase regional relative discount as REGION:0.5; repeatable (default `[]`)
- `--phase-duration`: Basic create phase duration as an ISO 8601 period, for example P7D or P1M
- `--phase-recurrence`: Basic create phase recurrence count (default `1`)
- `--price`: Basic create regional phase price as REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--product-id`: Parent subscription product ID
- `--regions-version`: Google Play regions version required by subscriptionOffers.create
- `--relative-discount`: Basic create regional phase relative discount as REGION:0.5, where 0.5 means the user pays 50% of the base plan price; repeatable (default `[]`)
- `--targeting-acquisition-scope`: Basic create acquisition targeting scope: any-subscription-in-app or this-subscription
- `--targeting-upgrade-billing-period`: Basic create upgrade targeting billing period duration as an ISO 8601 period
- `--targeting-upgrade-once-per-user`: Basic create upgrade targeting once-per-user rule (default `false`)
- `--targeting-upgrade-product-id`: Basic create upgrade targeting subscription product ID when scope is specific-subscription-in-app
- `--targeting-upgrade-scope`: Basic create upgrade targeting scope: this-subscription or specific-subscription-in-app

#### playpub subscription-offers deactivate

deactivate a subscription offer

```sh
playpub subscription-offers deactivate [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--confirm`: Apply the subscription offer state update (default `false`)
- `--dry-run`: Print the planned subscription offer state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### playpub subscription-offers delete

Delete a draft subscription offer

```sh
playpub subscription-offers delete [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--confirm`: Apply the subscription offer deletion (default `false`)
- `--dry-run`: Print the planned subscription offer deletion without calling Google Play (default `false`)
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### playpub subscription-offers get

Get one subscription offer

```sh
playpub subscription-offers get [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID
- `--offer-id`: Subscription offer ID
- `--product-id`: Parent subscription product ID

#### playpub subscription-offers list

List subscription offers

```sh
playpub subscription-offers list [flags]
```

##### Flags

- `--base-plan-id`: Parent subscription base plan ID, or - for all base plans
- `--page-size`: Maximum offers to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--product-id`: Parent subscription product ID, or - for all products

### playpub subscriptions

Inspect Google Play monetization subscriptions

```sh
playpub subscriptions
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub subscriptions base-plan`: Manage subscription base plans
- `playpub subscriptions batch-get`: Get multiple monetization subscriptions
- `playpub subscriptions batch-patch-listings`: Batch patch localized subscription listings
- `playpub subscriptions create`: Create a draft subscription
- `playpub subscriptions delete`: Delete a draft-only monetization subscription
- `playpub subscriptions get`: Get one monetization subscription
- `playpub subscriptions list`: List monetization subscriptions
- `playpub subscriptions patch`: Patch a subscription listing

#### playpub subscriptions base-plan

Manage subscription base plans

```sh
playpub subscriptions base-plan
```

##### Commands

- `playpub subscriptions base-plan activate`: activate a subscription base plan
- `playpub subscriptions base-plan batch-activate`: Batch activate subscription base plans
- `playpub subscriptions base-plan batch-deactivate`: Batch deactivate subscription base plans
- `playpub subscriptions base-plan batch-migrate-prices`: Batch migrate subscription base plan prices
- `playpub subscriptions base-plan batch-patch-prices`: Batch patch subscription base plan regional prices
- `playpub subscriptions base-plan deactivate`: deactivate a subscription base plan
- `playpub subscriptions base-plan delete`: Delete a draft-only subscription base plan

##### playpub subscriptions base-plan activate

activate a subscription base plan

```sh
playpub subscriptions base-plan activate [flags]
```

###### Flags

- `--base-plan-id`: Subscription base plan ID
- `--confirm`: Apply the base plan state update (default `false`)
- `--dry-run`: Print the planned base plan state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID

##### playpub subscriptions base-plan batch-activate

Batch activate subscription base plans

```sh
playpub subscriptions base-plan batch-activate [flags]
```

###### Flags

- `--base-plan`: Subscription base plan as productId/basePlanId; repeat for cross-subscription batches (default `[]`)
- `--base-plan-id`: Subscription base plan ID; repeat for multiple base plans (default `[]`)
- `--confirm`: Apply the base plan batch state update (default `false`)
- `--dry-run`: Print the planned base plan batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID, or - for base plans across subscriptions; inferred when --base-plan is used

##### playpub subscriptions base-plan batch-deactivate

Batch deactivate subscription base plans

```sh
playpub subscriptions base-plan batch-deactivate [flags]
```

###### Flags

- `--base-plan`: Subscription base plan as productId/basePlanId; repeat for cross-subscription batches (default `[]`)
- `--base-plan-id`: Subscription base plan ID; repeat for multiple base plans (default `[]`)
- `--confirm`: Apply the base plan batch state update (default `false`)
- `--dry-run`: Print the planned base plan batch state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID, or - for base plans across subscriptions; inferred when --base-plan is used

##### playpub subscriptions base-plan batch-migrate-prices

Batch migrate subscription base plan prices

```sh
playpub subscriptions base-plan batch-migrate-prices [flags]
```

###### Flags

- `--confirm`: Apply the base plan price migration (default `false`)
- `--dry-run`: Print the planned base plan price migration without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--migration`: Price migration as productId/basePlanId/REGION/RFC3339_TIME; repeat for multiple regions or base plans (default `[]`)
- `--price-increase-type`: Price increase type: optIn or optOut
- `--product-id`: Subscription product ID, or - for migrations across subscriptions; inferred from --migration values
- `--regions-version`: Google Play regions version required by batchMigratePrices

##### playpub subscriptions base-plan batch-patch-prices

Batch patch subscription base plan regional prices

```sh
playpub subscriptions base-plan batch-patch-prices [flags]
```

###### Flags

- `--confirm`: Apply the base plan price batch patch (default `false`)
- `--dry-run`: Print the planned base plan price batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--price`: Regional price patch as productId/basePlanId/REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--product-id`: Subscription product ID, or - for price patches across subscriptions; inferred from --price values
- `--regions-version`: Google Play regions version required by subscriptions.batchUpdate

##### playpub subscriptions base-plan deactivate

deactivate a subscription base plan

```sh
playpub subscriptions base-plan deactivate [flags]
```

###### Flags

- `--base-plan-id`: Subscription base plan ID
- `--confirm`: Apply the base plan state update (default `false`)
- `--dry-run`: Print the planned base plan state update without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--product-id`: Subscription product ID

##### playpub subscriptions base-plan delete

Delete a draft-only subscription base plan

```sh
playpub subscriptions base-plan delete [flags]
```

###### Flags

- `--base-plan-id`: Subscription base plan ID
- `--confirm`: Apply the base plan deletion (default `false`)
- `--dry-run`: Print the planned base plan deletion without calling Google Play (default `false`)
- `--product-id`: Subscription product ID

#### playpub subscriptions batch-get

Get multiple monetization subscriptions

```sh
playpub subscriptions batch-get [flags]
```

##### Flags

- `--product-id`: Subscription product ID; repeatable, up to 100 (default `[]`)

#### playpub subscriptions batch-patch-listings

Batch patch localized subscription listings

```sh
playpub subscriptions batch-patch-listings [flags]
```

##### Flags

- `--confirm`: Apply the subscription listing batch patch (default `false`)
- `--dry-run`: Print the planned subscription listing batch patch without calling Google Play (default `false`)
- `--latency-tolerance`: Propagation latency: latencySensitive or latencyTolerant (default `latencySensitive`)
- `--listing`: CSV listing patch productId,language,title,description; repeat for multiple localized listings (default `[]`)
- `--regions-version`: Google Play regions version required by subscriptions.batchUpdate

#### playpub subscriptions create

Create a draft subscription

```sh
playpub subscriptions create [flags]
```

##### Flags

- `--base-plan-id`: Basic create base plan ID
- `--billing-period`: Basic create billing period: P1W, P4W, P1M, P3M, P6M, or P1Y
- `--committed-payments`: Basic installments committed payments count (default `0`)
- `--confirm`: Create the draft subscription (default `false`)
- `--dry-run`: Print the planned subscription creation without calling Google Play (default `false`)
- `--eea-withdrawal-right-type`: Basic create EEA withdrawal right type: WITHDRAWAL_RIGHT_DIGITAL_CONTENT or WITHDRAWAL_RIGHT_SERVICE
- `--from-json`: Path to a Google Play API or playpub JSON subscription body
- `--installments`: Build a basic installments base plan instead of an auto-renewing base plan (default `false`)
- `--legacy-compatible`: Mark the basic auto-renewing base plan as legacy compatible (default `true`)
- `--listing`: Basic create listing as CSV language,title,description; repeatable (default `[]`)
- `--offer-tag`: Basic create base plan offer tag; repeatable (default `[]`)
- `--prepaid`: Build a basic prepaid base plan instead of an auto-renewing base plan (default `false`)
- `--price`: Basic create regional price as REGION:CURRENCY:UNITS[:NANOS]; repeatable (default `[]`)
- `--product-id`: Subscription product ID
- `--regional-streaming-tax`: Basic create US streaming tax type as US:STREAMING_TAX_TYPE; repeatable (default `[]`)
- `--regional-tax-tier`: Basic create regional reduced tax tier as REGION:TAX_TIER; repeatable (default `[]`)
- `--regions-version`: Google Play regions version required by subscriptions.create
- `--renewal-type`: Basic installments renewal type: RENEWAL_TYPE_RENEWS_WITHOUT_COMMITMENT or RENEWAL_TYPE_RENEWS_WITH_COMMITMENT
- `--restricted-country`: Basic create restricted payment country as REGION; repeatable (default `[]`)
- `--time-extension`: Basic prepaid time extension: TIME_EXTENSION_ACTIVE or TIME_EXTENSION_INACTIVE
- `--tokenized-digital-asset`: Basic create tokenized digital asset declaration: true or false

#### playpub subscriptions delete

Delete a draft-only monetization subscription

```sh
playpub subscriptions delete [flags]
```

##### Flags

- `--confirm`: Apply the subscription deletion (default `false`)
- `--dry-run`: Print the planned subscription deletion without calling Google Play (default `false`)
- `--product-id`: Subscription product ID

#### playpub subscriptions get

Get one monetization subscription

```sh
playpub subscriptions get [flags]
```

##### Flags

- `--product-id`: Subscription product ID

#### playpub subscriptions list

List monetization subscriptions

```sh
playpub subscriptions list [flags]
```

##### Flags

- `--page-size`: Maximum subscriptions to return, capped at 1000 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--show-archived`: Deprecated by Google; subscription archiving is no longer supported (default `false`)

#### playpub subscriptions patch

Patch a subscription listing

```sh
playpub subscriptions patch [flags]
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

### playpub system-apks

Inspect Google Play system APK variants

```sh
playpub system-apks
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub system-apks variants`: Inspect generated system APK variants

#### playpub system-apks variants

Inspect generated system APK variants

```sh
playpub system-apks variants
```

##### Commands

- `playpub system-apks variants list`: List system APK variants for an App Bundle version

##### playpub system-apks variants list

List system APK variants for an App Bundle version

```sh
playpub system-apks variants list [flags]
```

###### Flags

- `--version-code`: App Bundle version code (default `0`)

### playpub testers

Manage Google Play track tester groups

```sh
playpub testers
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub testers get`: Get Google Groups configured as testers for a track
- `playpub testers update`: Replace Google Groups configured as testers for a track

#### playpub testers get

Get Google Groups configured as testers for a track

```sh
playpub testers get [flags]
```

##### Flags

- `--track`: Track name, for example internal, alpha, beta, or production (default `internal`)

#### playpub testers update

Replace Google Groups configured as testers for a track

```sh
playpub testers update [flags]
```

##### Flags

- `--clear`: Remove all testing Google Groups from the track (default `false`)
- `--confirm`: Commit the tester update (default `false`)
- `--dry-run`: Print the planned tester update without calling Google Play (default `false`)
- `--google-group`: Testing Google Group email address, repeatable (default `[]`)
- `--track`: Track name, for example internal, alpha, beta, or production (default `internal`)

### playpub tracks

Manage Google Play release tracks

```sh
playpub tracks
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub tracks list`: List release tracks

### playpub users

Inspect and manage Google Play Console users

```sh
playpub users
```

#### Commands

- `playpub users create`: Grant developer-account access to a user
- `playpub users delete`: Remove all developer-account access for a user
- `playpub users list`: List users with access to a developer account
- `playpub users patch`: Replace developer-account access fields for a user

#### playpub users create

Grant developer-account access to a user

```sh
playpub users create [flags]
```

##### Flags

- `--confirm`: Apply the user creation (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned user creation without calling Google Play (default `false`)
- `--expiration-time`: Optional RFC3339 access expiration time
- `--permission`: Developer-account permission, repeatable (default `[]`)
- `--user-email`: Play Console user email

#### playpub users delete

Remove all developer-account access for a user

```sh
playpub users delete [flags]
```

##### Flags

- `--confirm`: Apply the user deletion (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned user deletion without calling Google Play (default `false`)
- `--name`: User resource name, developers/{developer}/users/{email}
- `--user-email`: Play Console user email

#### playpub users list

List users with access to a developer account

```sh
playpub users list [flags]
```

##### Flags

- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--page-size`: Maximum users to return; use -1 to disable pagination (default `0`)
- `--page-token`: Pagination token from a previous response

#### playpub users patch

Replace developer-account access fields for a user

```sh
playpub users patch [flags]
```

##### Flags

- `--confirm`: Apply the user patch (default `false`)
- `--developer`: Developer account ID or resource, for example 1234567890 or developers/1234567890
- `--dry-run`: Print the planned user patch without calling Google Play (default `false`)
- `--expiration-time`: Optional RFC3339 access expiration time
- `--name`: User resource name, developers/{developer}/users/{email}
- `--permission`: Developer-account permission, repeatable; replaces the account-level permission list when provided (default `[]`)
- `--user-email`: Play Console user email

### playpub validate

Validate a temporary Google Play edit

```sh
playpub validate [flags]
```

#### Flags

- `--package`: Android package name, for example com.example.app

### playpub vitals

Inspect Google Play Developer Reporting vitals

```sh
playpub vitals
```

#### Flags

- `--package`: Android package name, for example com.example.app

#### Commands

- `playpub vitals anomalies`: List Android vitals anomalies
- `playpub vitals errors`: Search Android vitals errors
- `playpub vitals metric-set`: Inspect Android vitals metric set metadata

#### playpub vitals anomalies

List Android vitals anomalies

```sh
playpub vitals anomalies
```

##### Commands

- `playpub vitals anomalies list`: List Android vitals anomalies

##### playpub vitals anomalies list

List Android vitals anomalies

```sh
playpub vitals anomalies list [flags]
```

###### Flags

- `--filter`: AIP-160 anomaly filter, for example activeBetween("2026-05-01T00:00:00Z", "2026-05-19T00:00:00Z")
- `--page-size`: Maximum anomalies to return, capped by Google at 100 (default `0`)
- `--page-token`: Pagination token from a previous response

#### playpub vitals errors

Search Android vitals errors

```sh
playpub vitals errors
```

##### Commands

- `playpub vitals errors issues`: Search grouped Android vitals error issues
- `playpub vitals errors reports`: Search Android vitals error reports

##### playpub vitals errors issues

Search grouped Android vitals error issues

```sh
playpub vitals errors issues
```

###### Commands

- `playpub vitals errors issues search`: Search grouped Android vitals error issues

###### playpub vitals errors issues search

Search grouped Android vitals error issues

```sh
playpub vitals errors issues search [flags]
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

##### playpub vitals errors reports

Search Android vitals error reports

```sh
playpub vitals errors reports
```

###### Commands

- `playpub vitals errors reports search`: Search individual Android vitals error reports

###### playpub vitals errors reports search

Search individual Android vitals error reports

```sh
playpub vitals errors reports search [flags]
```

###### Flags

- `--end-date`: End date, exclusive, in YYYY-MM-DD format
- `--filter`: AIP-160 filter expression for report fields
- `--page-size`: Maximum reports to return, capped by Google at 100 (default `0`)
- `--page-token`: Pagination token from a previous response
- `--start-date`: Start date, inclusive, in YYYY-MM-DD format
- `--time-zone`: Time zone for the interval; only UTC is supported when set

#### playpub vitals metric-set

Inspect Android vitals metric set metadata

```sh
playpub vitals metric-set
```

##### Commands

- `playpub vitals metric-set get`: Get Android vitals metric set freshness
- `playpub vitals metric-set query`: Query Android vitals metric rows

##### playpub vitals metric-set get

Get Android vitals metric set freshness

```sh
playpub vitals metric-set get [flags]
```

###### Flags

- `--metric-set`: Vitals metric set: anr-rate, crash-rate, error-count, excessive-wakeup-rate, lmk-rate, slow-rendering-rate, slow-start-rate, stuck-background-wakelock-rate

##### playpub vitals metric-set query

Query Android vitals metric rows

```sh
playpub vitals metric-set query [flags]
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

### playpub web

Inspect Play Console browser automation support

```sh
playpub web
```

#### Commands

- `playpub web status`: Explain the Play Console browser automation boundary

### playpub workflow

Run repo-local playpub workflows

```sh
playpub workflow
```

#### Flags

- `--file`: Workflow JSON file (default `.playpub/workflow.json`)

#### Commands

- `playpub workflow list`: List configured workflows
- `playpub workflow run`: Run one configured workflow

#### playpub workflow run

Run one configured workflow

```sh
playpub workflow run NAME [flags]
```

##### Flags

- `--confirm`: Execute the workflow shell steps (default `false`)
- `--dry-run`: Print the planned workflow steps without executing them (default `false`)
- `--workdir`: Working directory for shell steps; defaults to the workflow root

