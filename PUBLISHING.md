# Publishing grippd/grippd to the OpenTofu registry

The first release, done once in this order. Each step is Austin's to run. Later releases are
step 4 alone.

## 1. Make your grippd org membership public

The signing-key form checks it. github.com/orgs/grippd/people → your row → Public.

## 2. Create the public repository and push

From this folder:

```sh
gh repo create grippd/terraform-provider-grippd --public \
  --description "OpenTofu provider for grippd" --source=. --remote=origin --push
```

## 3. Make the signing key and give the workflow its private half

The private key never leaves your machine except into the repository secret; nothing below
prints it.

```sh
gpg --quick-generate-key "grippd releases <releases@grippd.sh>" rsa4096 sign 3y
FPR=$(gpg --list-secret-keys --with-colons "releases@grippd.sh" | awk -F: '/^fpr/{print $10; exit}')
gpg --armor --export-secret-keys "$FPR" | gh secret set GPG_PRIVATE_KEY -R grippd/terraform-provider-grippd
gh secret set PASSPHRASE -R grippd/terraform-provider-grippd   # paste the passphrase at the prompt
gpg --armor --export "$FPR" > grippd-signing-key.asc            # public half, for step 5
```

Keep a backup of the key (for example `gpg --export-secret-keys --armor "$FPR"` into your
password manager), or a lost key means a new key submission.

## 4. Tag the release

```sh
git tag v0.1.0 && git push origin v0.1.0
```

The `release` workflow builds, signs and publishes it. Check the release has 13 zips,
`terraform-provider-grippd_0.1.0_SHA256SUMS`, `..._SHA256SUMS.sig` and `..._manifest.json`.

## 5. Register the key, then the provider

Both are GitHub issue forms on opentofu/registry; they have to be filled in the web UI (the
registry's bot reads the form, so `gh issue create` doesn't work).

**Key** — https://github.com/opentofu/registry/issues/new?template=provider_key.yml

- Provider Namespace: `grippd`
- Provider Name: leave empty (the key covers every provider in the namespace)
- Public Membership: tick it
- Provider GPG Key: the contents of `grippd-signing-key.asc`

**Provider** — https://github.com/opentofu/registry/issues/new?template=provider.yml

- Provider Repository: `grippd/terraform-provider-grippd`

The bot validates each issue and opens the pull request itself. Once both are merged,
`curl https://registry.opentofu.org/v1/providers/grippd/grippd/versions` lists 0.1.0.

## 6. Pin it in the runner

In the grippd monorepo: add `grippd/grippd 0.1.0` to `apps/runner/container/providers.lock`,
remove `grippd/grippd` from `PENDING` in `apps/runner/test/mirror.test.ts`, run `bun test` and
`bunx wrangler deploy --dry-run` in `apps/runner`.
