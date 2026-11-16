# Made once for the environment and kept by grippd.
resource "grippd_env" "session_secret" {
  project  = "saas"
  env      = "production"
  name     = "AUTH_SECRET"
  apps     = ["web"]
  generate = "32-bytes-base64url"
}

# Asked of the team on the deploy form or with `grippd env set`.
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

# Minted by grippd's keys engine for each app, scoped and expiring.
resource "grippd_env" "media_key" {
  project = "saas"
  env     = "production"
  name    = "MEDIA_ACCESS_KEY"
  apps    = ["api"]
  from    = "files.media.access-key"
}
