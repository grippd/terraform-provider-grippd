// The grippd provider for OpenTofu: the things only grippd holds, declared in a project's
// compiled OpenTofu so they're planned and reviewed like everything else.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/grippd/terraform-provider-grippd/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Set by GoReleaser.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.opentofu.org/grippd/grippd",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
