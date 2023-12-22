// SPDX-License-Identifier: MIT

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/mercedes-benz/garm-provider-k8s/pkg/config"
)

func TestGetConfig(t *testing.T) {
	testCases := []struct {
		name      string
		config    string
		expected  config.ProviderConfig
		wantError bool
	}{
		{
			name: "valid configuration withouth PodTemplateSpec",
			expected: config.ProviderConfig{
				KubeConfigPath:    "/path/to/kubeconfig",
				ContainerRegistry: "sample.registry.com",
				RunnerNamespace:   "test-namespace",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			config: `
kubeConfigPath: "/path/to/kubeconfig"
containerRegistry: "sample.registry.com"
runnerNamespace: "test-namespace"
`,
		},
		{
			name: "valid configuration - expect default namespace",
			expected: config.ProviderConfig{
				KubeConfigPath:    "/path/to/kubeconfig",
				ContainerRegistry: "sample.registry.com",
				RunnerNamespace:   "runner",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			config: `
kubeConfigPath: "/path/to/kubeconfig"
containerRegistry: "sample.registry.com"
`,
		},
		{
			name: "valid configuration without a defined registry is fine - using configured container runtime registries",
			expected: config.ProviderConfig{
				KubeConfigPath:    "/path/to/kubeconfig",
				ContainerRegistry: "",
				RunnerNamespace:   "runner",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			config: `
kubeConfigPath: "/path/to/kubeconfig"
containerRegistry: ""
runnerNamespace: "runner"
`,
		},
		{
			name: "invalid configuration with a invalid namespace name",
			expected: config.ProviderConfig{
				KubeConfigPath:    "/path/to/kubeconfig",
				ContainerRegistry: "",
				RunnerNamespace:   "runner",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			config: `
kubeConfigPath: "/path/to/kubeconfig"
containerRegistry: ""
runnerNamespace: "this_is_An_invalid_namespace_name"
`,
			wantError: true,
		},
		{
			name: "valid configuration with imagePullSecrets",
			expected: config.ProviderConfig{
				KubeConfigPath:    "/path/to/kubeconfig",
				ContainerRegistry: "sample.registry.com",
				RunnerNamespace:   "test-namespace",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
				ImagePullSecrets: []string{
					"my-imagepullsecret1",
					"my-imagepullsecret2",
				},
			},
			config: `
kubeConfigPath: "/path/to/kubeconfig"
containerRegistry: "sample.registry.com"
runnerNamespace: "test-namespace"
imagePullSecrets:
  - my-imagepullsecret1
  - my-imagepullsecret2
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempConfigFile, err := setupTempFile(tc.config)
			defer os.Remove(tempConfigFile.Name())

			require.NoError(t, err)

			// Call NewConfig
			err = config.NewConfig(tempConfigFile.Name())

			// check if we expected an error
			assert.Equal(t, tc.wantError, func() bool {
				return err != nil
			}())

			if tc.wantError == false && err == nil {
				assert.Equal(t, tc.expected.KubeConfigPath, config.Config.KubeConfigPath)
				assert.Equal(t, tc.expected.ContainerRegistry, config.Config.ContainerRegistry)
				assert.Equal(t, tc.expected.RunnerNamespace, config.Config.RunnerNamespace)
				assert.Equal(t, tc.expected.PodTemplate, config.Config.PodTemplate)
				assert.Equal(t, tc.expected.ImagePullSecrets, config.Config.ImagePullSecrets)
			}

			// empty the global config for the next run
			config.Config = config.ProviderConfig{}
		})
	}
}

func setupTempFile(content string) (*os.File, error) {
	tmpfile, err := os.CreateTemp("", "testconfig.*.yaml")
	if err != nil {
		return nil, err
	}

	_, err = tmpfile.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	err = tmpfile.Close()
	if err != nil {
		return nil, err
	}
	return tmpfile, nil
}
