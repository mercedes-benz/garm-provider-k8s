// SPDX-License-Identifier: MIT

package provider_test

import (
	"context"
	"os"
	"testing"

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"

	"github.com/mercedes-benz/garm-provider-k8s/internal/provider"
	"github.com/mercedes-benz/garm-provider-k8s/pkg/config"
)

func TestCreateInstanceFlavorPodTemplate(t *testing.T) {
	providerConfig := config.ProviderConfig{
		RunnerNamespace: "runner",
		PodTemplate: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				PriorityClassName: "runner",
				Containers: []corev1.Container{
					{
						Name: "runner",
						Env:  []corev1.EnvVar{{Name: "COMMON_ENV", Value: "common"}},
					},
				},
			},
		},
		Flavors: map[string]corev1.ResourceRequirements{
			"nano": {
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("512Mi")},
			},
			"small": {
				Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
			},
		},
		FlavorPodTemplates: map[string]corev1.PodTemplateSpec{
			"small": {
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:          "dockerd",
							Image:         "localhost:5000/dind:latest",
							RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways),
							VolumeMounts: []corev1.VolumeMount{
								{Name: "runner", MountPath: "/runner"},
								{Name: "docker-sock", MountPath: "/docker"},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:         "runner",
							Env:          []corev1.EnvVar{{Name: "DOCKER_HOST", Value: "unix:///docker/docker.sock"}},
							VolumeMounts: []corev1.VolumeMount{{Name: "docker-sock", MountPath: "/docker"}},
						},
					},
					Volumes: []corev1.Volume{
						{Name: "docker-sock", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					},
				},
			},
		},
	}

	createPod := func(t *testing.T, cfg config.ProviderConfig, flavor string) *corev1.Pod {
		t.Helper()
		config.Config = cfg
		t.Cleanup(func() { config.Config = config.ProviderConfig{} })

		client := fake.NewSimpleClientset()
		p, err := provider.NewKubernetesProvider(client, controllerID, poolID)
		require.NoError(t, err)

		instance, err := p.CreateInstance(context.Background(), params.BootstrapInstance{
			Name:          instanceName,
			PoolID:        poolID,
			Flavor:        flavor,
			RepoURL:       "https://github.com/testorg",
			InstanceToken: "test-token",
			MetadataURL:   "https://metadata.test",
			CallbackURL:   "https://callback.test/status",
			Image:         "localhost:5000/runner:ubuntu-22.04",
			OSType:        "linux",
			OSArch:        "amd64",
		})
		require.NoError(t, err)

		pod, err := client.CoreV1().Pods("runner").Get(context.Background(), instance.ProviderID, metav1.GetOptions{})
		require.NoError(t, err)
		return pod
	}

	envNames := func(c corev1.Container) []string {
		names := []string{}
		for _, e := range c.Env {
			names = append(names, e.Name)
		}
		return names
	}
	mountNames := func(c corev1.Container) []string {
		names := []string{}
		for _, m := range c.VolumeMounts {
			names = append(names, m.Name)
		}
		return names
	}
	volumeNames := func(pod *corev1.Pod) []string {
		names := []string{}
		for _, v := range pod.Spec.Volumes {
			names = append(names, v.Name)
		}
		return names
	}

	t.Run("flavor with an overlay gets it on top of the common template", func(t *testing.T) {
		pod := createPod(t, providerConfig, "small")

		require.Len(t, pod.Spec.InitContainers, 1)
		sidecar := pod.Spec.InitContainers[0]
		assert.Equal(t, "dockerd", sidecar.Name)
		assert.Equal(t, ptr.To(corev1.ContainerRestartPolicyAlways), sidecar.RestartPolicy)
		assert.ElementsMatch(t, []string{"runner", "docker-sock"}, mountNames(sidecar))

		// this overlay adds no regular container, so the runner is still the
		// only one, and it keeps what the provider, the flavor and the common
		// template gave it
		require.Len(t, pod.Spec.Containers, 1)
		runner := pod.Spec.Containers[0]
		assert.Equal(t, "runner", runner.Name)
		assert.Equal(t, "localhost:5000/runner:ubuntu-22.04", runner.Image)
		assert.Equal(t, providerConfig.Flavors["small"], runner.Resources)
		assert.Contains(t, envNames(runner), "RUNNER_NAME")
		assert.Contains(t, envNames(runner), "COMMON_ENV")
		assert.Contains(t, envNames(runner), "DOCKER_HOST")
		assert.ElementsMatch(t, []string{"runner", "docker-sock"}, mountNames(runner))

		assert.ElementsMatch(t, []string{"runner", "docker-sock"}, volumeNames(pod))
		assert.Equal(t, "runner", pod.Spec.PriorityClassName)
		assert.Equal(t, corev1.RestartPolicyNever, pod.Spec.RestartPolicy)
	})

	t.Run("flavor without an overlay gets the common template only", func(t *testing.T) {
		pod := createPod(t, providerConfig, "nano")

		assert.Empty(t, pod.Spec.InitContainers)
		require.Len(t, pod.Spec.Containers, 1)
		runner := pod.Spec.Containers[0]
		assert.Equal(t, providerConfig.Flavors["nano"], runner.Resources)
		assert.Contains(t, envNames(runner), "COMMON_ENV")
		assert.NotContains(t, envNames(runner), "DOCKER_HOST")
		assert.Equal(t, []string{"runner"}, mountNames(runner))
		assert.Equal(t, []string{"runner"}, volumeNames(pod))

		// identical to what the same config without any overlay produces
		withoutOverlays := providerConfig
		withoutOverlays.FlavorPodTemplates = nil
		assert.Equal(t, createPod(t, withoutOverlays, "nano"), pod)
	})

	t.Run("overlay can replace the runner volume and its mount", func(t *testing.T) {
		cfg := providerConfig
		cfg.FlavorPodTemplates = map[string]corev1.PodTemplateSpec{
			"small": {
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:         "runner",
							VolumeMounts: []corev1.VolumeMount{{Name: "runner", MountPath: "/runner/"}},
						},
					},
					Volumes: []corev1.Volume{
						{Name: "runner", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/mnt/work"}}},
					},
				},
			},
		}
		pod := createPod(t, cfg, "small")

		// exactly one source: no default emptyDir next to the overlay's volume
		require.Len(t, pod.Spec.Volumes, 1)
		assert.Nil(t, pod.Spec.Volumes[0].EmptyDir)
		assert.NotNil(t, pod.Spec.Volumes[0].HostPath)
		require.Len(t, pod.Spec.Containers, 1)
		assert.Equal(t, []corev1.VolumeMount{{Name: "runner", MountPath: "/runner/"}}, pod.Spec.Containers[0].VolumeMounts)

		// other flavors still get the default volume
		nano := createPod(t, cfg, "nano")
		require.Len(t, nano.Spec.Volumes, 1)
		assert.NotNil(t, nano.Spec.Volumes[0].EmptyDir)
	})

	t.Run("overlay without containers keeps the runner container", func(t *testing.T) {
		cfg := providerConfig
		cfg.FlavorPodTemplates = map[string]corev1.PodTemplateSpec{
			"small": {Spec: corev1.PodSpec{Hostname: "h"}},
		}
		pod := createPod(t, cfg, "small")

		require.Len(t, pod.Spec.Containers, 1)
		assert.Equal(t, "runner", pod.Spec.Containers[0].Name)
		assert.Equal(t, "h", pod.Spec.Hostname)
	})

	t.Run("overlay loaded from a config file", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "config.*.yaml")
		require.NoError(t, err)
		_, err = f.WriteString(`
runnerNamespace: runner
podTemplate:
  spec:
    tolerations:
      - key: common
        operator: Exists
    containers:
      - name: runner
        env:
          - name: COMMON_ENV
            value: common
flavors:
  small:
    limits:
      memory: 4Gi
flavorPodTemplates:
  small:
    spec:
      initContainers:
        - name: dockerd
          image: localhost:5000/dind:latest
          restartPolicy: Always
      containers:
        - name: runner
          env:
            - name: COMMON_ENV
              value: overridden
      volumes:
        - name: docker-sock
          emptyDir: {}
`)
		require.NoError(t, err)
		require.NoError(t, f.Close())
		require.NoError(t, config.NewConfig(f.Name()))

		small := createPod(t, config.Config, "small")
		require.Len(t, small.Spec.InitContainers, 1)
		require.Len(t, small.Spec.Containers, 1)
		assert.Contains(t, small.Spec.Containers[0].Env, corev1.EnvVar{Name: "COMMON_ENV", Value: "overridden"})
		assert.Equal(t, "4Gi", small.Spec.Containers[0].Resources.Limits.Memory().String())
		assert.ElementsMatch(t, []string{"runner", "docker-sock"}, volumeNames(small))
		assert.Len(t, small.Spec.Tolerations, 1)
	})
}
