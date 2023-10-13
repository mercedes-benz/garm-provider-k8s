// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"

	koanfYaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type Config struct {
	KubeConfigPath    string                 `koanf:"kubeConfigPath"`
	ContainerRegistry string                 `koanf:"containerRegistry"`
	RunnerNamespace   string                 `koanf:"runnerNamespace"`
	PodTemplate       corev1.PodTemplateSpec `koanf:"podTemplate"`
}

func NewConfig(configPath string) (*Config, error) {
	k := koanf.New(".")

	if configPath == "" {
		return nil, errors.New("no config file path provided")
	}

	if err := k.Load(file.Provider(configPath), koanfYaml.Parser()); err != nil {
		return nil, err
	}

	config := Config{}
	config.KubeConfigPath = k.String("kubeConfigPath")
	config.ContainerRegistry = k.String("containerRegistry")
	config.RunnerNamespace = k.String("runnerNamespace")

	podTemplateBytes, err := yaml.Marshal(k.Get("podTemplate"))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PodTemplate into bytes: %v", err)
	}

	var podTemplateSpec corev1.PodTemplateSpec
	if err := yaml.Unmarshal(podTemplateBytes, &podTemplateSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PodTemplate bytes into corev1.PodTemplateSpec: %v", err)
	}

	config.PodTemplate = podTemplateSpec

	if config.RunnerNamespace == "" {
		config.RunnerNamespace = "runner"
	}

	return &config, nil
}
