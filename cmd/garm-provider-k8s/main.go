// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cloudbase/garm-provider-common/execution"
	"github.com/mercedes-benz/garm-provider-k8s/client"
	"github.com/mercedes-benz/garm-provider-k8s/config"
	"github.com/mercedes-benz/garm-provider-k8s/internal/provider"
	"log"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// util.SetupLogging()

	executionEnv, err := execution.GetEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	var configPath *string
	configPath = flag.String("configpath", "", "absolute path to the config.yaml file")
	flag.Parse()

	if *configPath == "" {
		*configPath = executionEnv.ProviderConfigFile
	}

	config, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("could not initialize config: %v", err)
	}

	clientWrapper, err := client.NewKubeClient(config)
	if err != nil {
		log.Fatalf("could not initialize kube client: %s", err.Error())
	}

	prov, err := provider.NewKubernetesProvider(clientWrapper, config, executionEnv.ControllerID)
	if err != nil {
		log.Fatal(err)
	}

	result, err := execution.Run(ctx, prov, executionEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command: %+v\n", err)
		os.Exit(1)
	}
	if len(result) > 0 {
		fmt.Fprint(os.Stdout, result)
	}
}
