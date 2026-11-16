---
page_title: "grippd_env Resource - grippd"
description: |-
  One value grippd keeps for an environment. The value itself is never in state or the plan.
---

# grippd_env (Resource)

One value grippd keeps for an environment: generated once, asked of the team, or resolved from
something grippd mints. Set exactly one of `generate`, `ask` or `from`.

Changing `project`, `env`, `name` or `generate` replaces the declaration, which for `generate`
means a new value. Changing `apps`, `ask` or `from` updates it in place.

## Example Usage

```terraform
resource "grippd_env" "session_secret" {
  project  = "saas"
  env      = "production"
  name     = "AUTH_SECRET"
  apps     = ["web"]
  generate = "32-bytes-base64url"
}

resource "grippd_env" "stripe_key" {
  project = "saas"
  env     = "production"
  name    = "STRIPE_KEY"
  apps    = ["api"]
  ask = {
    question = "Stripe secret key"
    help     = "Stripe: Developers → API keys"
    pattern  = "^sk_live_"
  }
}

resource "grippd_env" "media_key" {
  project = "saas"
  env     = "production"
  name    = "MEDIA_ACCESS_KEY"
  apps    = ["api"]
  from    = "files.media.access-key"
}
```

## Schema

### Required

- `project` (String) The grippd project.
- `env` (String) The environment, such as `dev`, `staging` or `production`.
- `name` (String) The variable's name, UPPER_SNAKE_CASE.
- `apps` (List of String) The apps that get the value. At least one, no repeats.

### Optional

- `generate` (String) Made once for the environment: `32-bytes-base64url`, `uuid` or `password`.
- `ask` (Attributes) Given by the team on the deploy form or with `grippd env set`. See below.
- `from` (String) Resolved by grippd when the app deploys, such as `database.main.url` or `files.media.access-key`.

### Read-Only

- `id` (String) `<project>/<env>/<name>`.

### Nested Schema for `ask`

- `question` (String) What the deploy form asks.
- `help` (String) Where to find the value.
- `pattern` (String) A pattern the value must match.
- `required` (Boolean) Deploys wait until a required value is set. Defaults to `true`.
