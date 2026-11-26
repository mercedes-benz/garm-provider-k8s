// SPDX-License-Identifier: MIT

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"github.com/mercedes-benz/garm-provider-k8s/pkg/config"
)

func TestFlavorPodTemplates(t *testing.T) {
	load := func(t *testing.T, content string) error {
		t.Helper()
		f, err := setupTempFile(content)
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = os.Remove(f.Name())
			config.Config = config.ProviderConfig{}
		})
		return config.NewConfig(f.Name())
	}

	t.Run("absent", func(t *testing.T) {
		require.NoError(t, load(t, `
flavors:
  nano:
    limits:
      memory: 512Mi
`))
		assert.Empty(t, config.Config.FlavorPodTemplates)
	})

	t.Run("overlay with a native sidecar", func(t *testing.T) {
		require.NoError(t, load(t, `
flavors:
  small:
    limits:
      memory: 4Gi
flavorPodTemplates:
  small:
    spec:
      initContainers:
        - name: dockerd
          image: dind:latest
          restartPolicy: Always
          resources:
            limits:
              memory: 4Gi
      volumes:
        - name: docker-sock
          emptyDir: {}
`))
		require.Contains(t, config.Config.FlavorPodTemplates, "small")
		spec := config.Config.FlavorPodTemplates["small"].Spec
		require.Len(t, spec.InitContainers, 1)
		assert.Equal(t, "dockerd", spec.InitContainers[0].Name)
		assert.Equal(t, ptr.To(corev1.ContainerRestartPolicyAlways), spec.InitContainers[0].RestartPolicy)
		assert.Equal(t, "4Gi", spec.InitContainers[0].Resources.Limits.Memory().String())
		require.Len(t, spec.Volumes, 1)
		// must not be nil, or the merge would drop the runner container
		assert.NotNil(t, spec.Containers)
		assert.Contains(t, config.Config.Flavors, "small")
	})

	t.Run("empty overlay", func(t *testing.T) {
		require.NoError(t, load(t, `
flavorPodTemplates:
  small:
`))
		require.Contains(t, config.Config.FlavorPodTemplates, "small")
		assert.NotNil(t, config.Config.FlavorPodTemplates["small"].Spec.Containers)
	})

	t.Run("malformed overlay is an error", func(t *testing.T) {
		assert.ErrorContains(t, load(t, `
flavorPodTemplates:
  small:
    spec:
      initContainers: "not a list"
`), "flavorPodTemplates")
	})

	t.Run("unknown field is an error", func(t *testing.T) {
		assert.ErrorContains(t, load(t, `
flavorPodTemplates:
  small:
    spec:
      initContainer:
        - name: dockerd
`), "flavorPodTemplates")
	})
}
