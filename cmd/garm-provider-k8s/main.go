// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/cloudbase/garm-provider-common/execution"

	"github.com/mercedes-benz/garm-provider-k8s/client"
	"github.com/mercedes-benz/garm-provider-k8s/config"
	"github.com/mercedes-benz/garm-provider-k8s/internal/provider"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer func() {
		stop()
		os.Exit(1)
	}()

	// util.SetupLogging()

	executionEnv, err := execution.GetEnvironment()
	if err != nil {
		log.Println(err)
		return
	}

	var configPath *string
	configPath = flag.String("configpath", "", "absolute path to the config.yaml file")
	flag.Parse()

	if *configPath == "" {
		*configPath = executionEnv.ProviderConfigFile
	}

	config, err := config.NewConfig(*configPath)
	if err != nil {
		log.Printf("could not initialize config: %v", err)
		return
	}

	clientWrapper, err := client.NewKubeClient(config)
	if err != nil {
		log.Printf("could not initialize kube client: %s", err.Error())
		return
	}

	prov, err := provider.NewKubernetesProvider(clientWrapper, config, executionEnv.ControllerID)
	if err != nil {
		log.Print(err)
		return
	}

	result, err := execution.Run(ctx, prov, executionEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command: %+v\n", err)
		return
	}
	if len(result) > 0 {
		fmt.Fprint(os.Stdout, result)
	}
}
