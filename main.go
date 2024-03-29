package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/twilliate/terraform-provider-twilliate/internal"
	"log"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "enable debugger")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/twilliate/de",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), internal.New, opts)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
}
