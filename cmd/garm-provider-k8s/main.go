// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudbase/garm-provider-common/execution"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/mercedes-benz/garm-provider-k8s/internal/provider"
	"github.com/mercedes-benz/garm-provider-k8s/pkg/config"
)

var signals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}

func main() {
	if err := kubernetesProvider(); err != nil {
		log.Fatal(err)
	}
}

func kubernetesProvider() error {
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	executionEnv, err := execution.GetEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get execution environment: %w", err)
	}

	configPath := flag.String("configpath", "", "absolute path to the config.yaml file")
	flag.Parse()

	if *configPath == "" {
		*configPath = executionEnv.ProviderConfigFile
	}

	err = config.NewConfig(*configPath)
	if err != nil {
		return fmt.Errorf("could not initialize config: %w", err)
	}

	// generate a k8s client config
	var restConfig *rest.Config
	if config.Config.KubeConfigPath == "" {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("could not initialize in-cluster config client: %w", err)
		}
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", config.Config.KubeConfigPath)
		if err != nil {
			return fmt.Errorf("could not initialize kubernetes config client: %w", err)
		}
	}

	// create a new kubernetes clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("could not initialize kube client: %w", err)
	}

	// create a new kubernetes provider
	prov := provider.NewKubernetesProvider(clientset, executionEnv.ControllerID)

	result, err := execution.Run(ctx, prov, executionEnv)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	if len(result) > 0 {
		fmt.Fprint(os.Stdout, result)
	}
	return nil
}
