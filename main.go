package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/harvester/terraform-provider-harvester/internal/provider"
)

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	opts := &plugin.ServeOpts{ProviderFunc: provider.Provider}

	if debugMode {
		reattachConfig, closeCh, err := plugin.DebugServe(context.Background(), opts)
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Printf("Provider running in debug mode. To attach a debugger, use the following reattach config: %+v", reattachConfig)
		<-closeCh
		return
	}

	plugin.Serve(opts)
}
