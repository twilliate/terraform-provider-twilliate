package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/twilliate/terraform-provider-twilliate/internal"
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

	err := providerserver.Serve(context.Background(), internal.New, opts)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
}
