// SPDX-License-Identifier: MIT

package provider_test

import (
	"context"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/google/uuid"
	"github.com/mercedes-benz/garm-provider-k8s/client"
	"github.com/mercedes-benz/garm-provider-k8s/config"
	"github.com/mercedes-benz/garm-provider-k8s/internal/provider"
	"github.com/stretchr/testify/assert"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
)

var (
	poolID                = "ddce45e7-1bbb-4ecd-92cb-c733372b5cde"
	instanceName          = "garm-hvjedclmnvry"
	controllerID          = uuid.New().String()
	mockKubeClient        = testclient.NewSimpleClientset()
	mockKubeClientWrapper = &client.KubeClientWrapper{Client: mockKubeClient}

	cfg = &config.Config{
		RunnerNamespace: "runner",
	}
)

func TestCreateInstance(t *testing.T) {
	testCases := []struct {
		name            string
		bootstrapParams params.BootstrapInstance
		expected        params.ProviderInstance
	}{
		{
			name: "Valid bootstrapParams",
			bootstrapParams: params.BootstrapInstance{
				Name:          instanceName,
				PoolID:        poolID,
				Flavor:        "small",
				RepoURL:       "https://github.com/testorg",
				InstanceToken: "test-token",
				MetadataURL:   "https://metadata.test",
				CallbackURL:   "https://callback.test",
				Image:         "ghcr.io/mercedes-benz/garm-provider-k8s/runner:ubuntu-22.04",
				OSType:        "linux",
				OSArch:        "arm64",
			},
			expected: params.ProviderInstance{
				ProviderID: instanceName,
				Name:       instanceName,
				OSType:     "linux",
				OSName:     "",
				OSVersion:  "",
				OSArch:     "arm64",
				Status:     "running",
			},
		},
	}

	p, _ := provider.NewKubernetesProvider(mockKubeClientWrapper, cfg, controllerID)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := p.CreateInstance(context.Background(), tc.bootstrapParams)
			assert.Nil(t, err)
			assert.Equal(t, actual, tc.expected)
		})
	}
}
