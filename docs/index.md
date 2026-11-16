---
page_title: "grippd Provider"
description: |-
  Declares the values grippd keeps for a project. Values never reach OpenTofu's state.
---

# grippd Provider

[grippd](https://grippd.sh) compiles a repository's `grippd.yaml` into OpenTofu. Most of it uses
each supplier's own provider; this one declares the things only grippd holds: environment values
that are generated, asked of the team, or minted by grippd's keys engine.

The provider records declarations only. grippd reads them from the plan, then makes, stores,
rotates and hands out the values itself, so a value is never in the plan, the state or the
repository. It takes no configuration and makes no network calls.

You don't usually write these resources by hand: `grippd` writes them when it compiles your project.

## Example Usage

```terraform
terraform {
  required_providers {
    grippd = {
      source  = "grippd/grippd"
      version = "~> 0.1"
    }
  }
}

provider "grippd" {}
```
