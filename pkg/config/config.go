// SPDX-License-Identifier: MIT

package config

import (
	"bytes"
	"errors"
	"fmt"

	koanfYaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/validation"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type ProviderConfig struct {
	KubeConfigPath    string                                 `koanf:"kubeConfigPath"`
	ContainerRegistry string                                 `koanf:"containerRegistry"`
	RunnerNamespace   string                                 `koanf:"runnerNamespace"`
	PodTemplate       corev1.PodTemplateSpec                 `koanf:"podTemplate"`
	Flavours          map[string]corev1.ResourceRequirements `koanf:"flavours"`
}

var Config ProviderConfig

func NewConfig(configPath string) error {
	k := koanf.New(".")

	if configPath == "" {
		return errors.New("no config file path provided")
	}

	// load the config file
	if err := k.Load(file.Provider(configPath), koanfYaml.Parser()); err != nil {
		return err
	}

	Config.Flavours = unmarshalFlavours(k)
	k.Delete("flavours")

	// unmarshal all koanf config keys into ProviderConfig struct
	if err := k.Unmarshal("", &Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	Config.PodTemplate = unmarshalPodTemplateSpec(k)

	// set the default namespace for runners
	if Config.RunnerNamespace == "" {
		Config.RunnerNamespace = "runner"
	}

	// validate the given runner namespace
	err := validateNamespace(Config.RunnerNamespace)
	if err != nil {
		return fmt.Errorf("failed to validate namespace: %v", err)
	}

	// validate the pod template spec
	err = validatePodTemplate()
	if err != nil {
		return fmt.Errorf("failed to marshal PodTemplate into bytes: %v", err)
	}
	return nil
}

// validatePodTemplate validates the pod template spec
// by unmarshalling it into a corev1.PodTemplateSpec
func validatePodTemplate() error {
	podTemplate, err := yaml.Marshal(Config.PodTemplate)
	if err != nil {
		return fmt.Errorf("failed to marshal PodTemplate into bytes: %v", err)
	}

	var podTemplateSpec corev1.PodTemplateSpec
	if err := yaml.Unmarshal(podTemplate, &podTemplateSpec); err != nil {
		return fmt.Errorf("failed to unmarshal PodTemplate bytes into corev1.PodTemplateSpec: %v", err)
	}
	return nil
}

// validateNamespace validates the namespace
// by checking if it is a valid DNS subdomain
func validateNamespace(namespace string) error {
	// response is a slice of strings
	// with multiple RFC1132 errors
	dnsValidationErrors := validation.NameIsDNSSubdomain(namespace, false)

	if len(dnsValidationErrors) > 0 {
		return fmt.Errorf("namespace %s is invalid: %v", namespace, dnsValidationErrors)
	}
	return nil
}

func unmarshalPodTemplateSpec(k *koanf.Koanf) corev1.PodTemplateSpec {
	defaultSpec := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{},
		},
	}

	// Extract the podTemplate as a raw map
	podTemplateMap := k.Get("podTemplate")
	if podTemplateMap == nil {
		return defaultSpec
	}

	// Convert the map to a YAML string
	podTemplateYAML, err := yaml.Marshal(podTemplateMap)
	if err != nil {
		return defaultSpec
	}

	var podTemplate corev1.PodTemplateSpec
	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(podTemplateYAML), len(podTemplateYAML))
	if err := decoder.Decode(&podTemplate); err != nil {
		return defaultSpec
	}
	return podTemplate
}

func unmarshalFlavours(k *koanf.Koanf) map[string]corev1.ResourceRequirements {
	var result map[string]corev1.ResourceRequirements

	// Extract the podTemplate as a raw map
	flavoursMap := k.Get("flavours")
	if flavoursMap == nil {
		return result
	}

	// Convert the map to a YAML string
	flavoursYAML, err := yaml.Marshal(flavoursMap)
	if err != nil {
		return result
	}

	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(flavoursYAML), len(flavoursYAML))
	if err := decoder.Decode(&result); err != nil {
		return result
	}
	return result
}
