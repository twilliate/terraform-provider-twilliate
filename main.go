package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/twilliate/twilliate-provider-aws/twilliate"
	"log"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "enable debugger")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/twilliate/de",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), twilliate.New, opts)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
}
