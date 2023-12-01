// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"

	koanfYaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/validation"
	"sigs.k8s.io/yaml"
)

type ProviderConfig struct {
	KubeConfigPath    string                 `koanf:"kubeConfigPath"`
	ContainerRegistry string                 `koanf:"containerRegistry"`
	RunnerNamespace   string                 `koanf:"runnerNamespace"`
	PodTemplate       corev1.PodTemplateSpec `koanf:"podTemplate"`
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

	// unmarshal all koanf config keys into ProviderConfig struct
	if err := k.Unmarshal("", &Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// set the default namespace for runners
	if Config.RunnerNamespace == "" {
		Config.RunnerNamespace = "runner"
	}

	// containers is a required field in a PodSpec, so if it's not part of the patch
	// the conversion to unstructured will have a nil spec.containers field, which
	// will clear out the containers field in the merge.  We don't want that.
	if Config.PodTemplate.Spec.Containers == nil {
		Config.PodTemplate.Spec.Containers = []corev1.Container{}
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
