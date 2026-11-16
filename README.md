# terraform-provider-grippd

The OpenTofu provider for [grippd](https://grippd.sh), published as
[`grippd/grippd`](https://search.opentofu.org/provider/grippd/grippd). grippd's compiler writes
its resources; customers' OpenTofu runs install it like any other provider.

It has one resource, [`grippd_env`](docs/resources/env.md): a declaration of a value grippd
keeps (generated, asked for or minted). Only the declaration is in state. The provider takes no
configuration and makes no network calls.

## Develop

```sh
mise install
mise run test          # unit and acceptance tests, run against OpenTofu
go build -o bin/ .     # then point a dev_overrides block at bin/ for registry.opentofu.org/grippd/grippd
```

## Release

Push a tag `vX.Y.Z`. The `release` workflow runs the tests, imports the signing key from the
`GPG_PRIVATE_KEY` and `PASSPHRASE` secrets with plain `gpg` and runs GoReleaser (pinned in `mise.toml`), which publishes the zips,
`SHA256SUMS`, its signature and the registry manifest. The OpenTofu registry picks up new tags
of a registered provider on its own.

The signing key's public half is registered with the OpenTofu registry for the `grippd`
namespace. Rotating it means a new key submission there before the next release.
