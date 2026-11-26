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
	executioncommon "github.com/cloudbase/garm-provider-common/execution/common"
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

	// GARM passes the provider interface version it speaks via
	// GARM_INTERFACE_VERSION (empty means v0.1.0), and the execution package
	// exposes the pool ID in the matching version-specific environment, see
	// https://github.com/cloudbase/garm-provider-common/blob/main/execution/execution.go
	version := executionEnv.InterfaceVersion
	poolID := executionEnv.EnvironmentV010.PoolID
	switch version {
	case "":
		version = executioncommon.Version010
	case executioncommon.Version011:
		poolID = executionEnv.EnvironmentV011.PoolID
	}

	// create a new kubernetes provider
	prov, err := provider.NewKubernetesProvider(clientset, executionEnv.ControllerID, poolID, version)
	if err != nil {
		return fmt.Errorf("could not initialize provider: %w", err)
	}

	result, err := executionEnv.Run(ctx, prov)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	if len(result) > 0 {
		fmt.Fprint(os.Stdout, result)
	}
	return nil
}
