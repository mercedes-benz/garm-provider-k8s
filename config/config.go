// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type Config struct {
	KubeConfigPath    string                 `mapstructure:"KubeConfigPath"`
	ContainerRegistry string                 `mapstructure:"ContainerRegistry"`
	RunnerNamespace   string                 `mapstructure:"RunnerNamespace"`
	PodTemplate       corev1.PodTemplateSpec `mapstructure:"PodTemplate"`
}

func NewConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, errors.New("no config file path provided")
	}

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Println("No configuration file found, using default values")
		}
	}

	config := Config{}
	config.KubeConfigPath = viper.GetString("KubeConfigPath")
	config.ContainerRegistry = viper.GetString("ContainerRegistry")
	config.RunnerNamespace = viper.GetString("RunnerNamespace")

	podTemplateBytes, err := yaml.Marshal(viper.Get("PodTemplate"))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PodTemplate into bytes: %v", err)
	}

	var podTemplateSpec corev1.PodTemplateSpec
	if err := yaml.Unmarshal(podTemplateBytes, &podTemplateSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PodTemplate bytes into corev1.PodTemplateSpec: %v", err)
	}

	config.PodTemplate = podTemplateSpec

	if config.ContainerRegistry == "" {
		return nil, errors.New("property `ContainerRegistry` in provider config must be set")
	}

	if config.RunnerNamespace == "" {
		config.RunnerNamespace = "runner"
	}

	return &config, nil
}
