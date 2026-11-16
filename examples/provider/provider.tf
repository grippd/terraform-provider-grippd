terraform {
  required_providers {
    grippd = {
      source  = "grippd/grippd"
      version = "~> 0.1"
    }
  }
}

provider "grippd" {}
