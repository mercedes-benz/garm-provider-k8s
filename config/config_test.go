// SPDX-License-Identifier: MIT

package config_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/mercedes-benz/garm-provider-k8s/config"
)

func TestCreateInstance(t *testing.T) {
	testCases := []struct {
		name      string
		config    string
		expected  config.Config
		wantError error
	}{
		{
			name: "valid config, no PodTemplateSpec",
			expected: config.Config{
				KubeConfigPath:    "/path/to/kubeconfig",
				ContainerRegistry: "sample.registry.com",
				RunnerNamespace:   "test-namespace",
				PodTemplate:       corev1.PodTemplateSpec{},
			},
			config: `
KubeConfigPath: "/path/to/kubeconfig"
ContainerRegistry: "sample.registry.com"
RunnerNamespace: "test-namespace"
`,
			wantError: nil,
		},
		{
			name: "valid config, set default namespace",
			expected: config.Config{
				KubeConfigPath:    "/path/to/kubeconfig",
				ContainerRegistry: "sample.registry.com",
				RunnerNamespace:   "runner",
				PodTemplate:       corev1.PodTemplateSpec{},
			},
			config: `
KubeConfigPath: "/path/to/kubeconfig"
ContainerRegistry: "sample.registry.com"
RunnerNamespace: ""
`,
			wantError: nil,
		},
		{
			name:     "invalid config, no ContainerRegistry set",
			expected: config.Config{},
			config: `
KubeConfigPath: "/path/to/kubeconfig"
RunnerNamespace: "test-namespace"
`,
			wantError: errors.New("property `ContainerRegistry` in provider config must be set"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempConfigFile, err := setupTempFile(tc.config)
			defer os.Remove(tempConfigFile.Name())

			require.NoError(t, err)

			// Call NewConfig
			config, err := config.NewConfig(tempConfigFile.Name())

			assert.Equal(t, tc.wantError, err)

			if tc.wantError == nil && err == nil {
				assert.Equal(t, tc.expected.KubeConfigPath, config.KubeConfigPath)
				assert.Equal(t, tc.expected.ContainerRegistry, config.ContainerRegistry)
				assert.Equal(t, tc.expected.RunnerNamespace, config.RunnerNamespace)
				assert.Equal(t, tc.expected.PodTemplate, config.PodTemplate)
			}
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
