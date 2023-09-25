// SPDX-License-Identifier: MIT

package provider_test

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudbase/garm-provider-common/params"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/mercedes-benz/garm-provider-k8s/internal/provider"
	"github.com/mercedes-benz/garm-provider-k8s/internal/spec"
	"github.com/mercedes-benz/garm-provider-k8s/pkg/config"
)

var (
	poolID       = "ddce45e7-1bbb-4ecd-92cb-c733372b5cde"
	instanceName = "garm-HvjEdcLmnVrY"
	providerID   = strings.ToLower(instanceName)
	controllerID = uuid.New().String()
)

func TestCreateInstance(t *testing.T) {
	testCases := []struct {
		name                     string
		config                   *config.ProviderConfig
		bootstrapParams          params.BootstrapInstance
		expectedProviderInstance params.ProviderInstance
		expectedPodInstance      *corev1.Pod
		runtimeObjects           []runtime.Object
		err                      error
	}{
		{
			name: "Valid bootstrapParams and merge pod template spec",
			config: &config.ProviderConfig{
				KubeConfigPath:    "",
				ContainerRegistry: "localhost:5000",
				RunnerNamespace:   "runner",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "my-additional-volume",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{
										Medium:    "",
										SizeLimit: nil,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name: "runner",
								Env: []corev1.EnvVar{
									{
										Name:  "MY_ADDITIONAL_ENV",
										Value: "test",
									},
								},
							},
							{
								Name: "sidecar",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:              resource.MustParse("500m"),
										corev1.ResourceMemory:           resource.MustParse("500Mi"),
										corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
			bootstrapParams: params.BootstrapInstance{
				Name:          instanceName,
				PoolID:        poolID,
				Flavor:        "small",
				RepoURL:       "https://github.com/testorg",
				InstanceToken: "test-token",
				MetadataURL:   "https://metadata.test",
				CallbackURL:   "https://callback.test/status",
				Image:         "runner:ubuntu-22.04",
				OSType:        "linux",
				OSArch:        "arm64",
				Labels:        []string{"road-runner", "linux", "arm64", "kubernetes"},
			},
			expectedProviderInstance: params.ProviderInstance{
				ProviderID: providerID,
				Name:       instanceName,
				OSType:     "linux",
				OSName:     "",
				OSVersion:  "",
				OSArch:     "arm64",
				Status:     "running",
			},
			expectedPodInstance: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      providerID,
					Namespace: "runner",
					Labels: map[string]string{
						spec.GarmInstanceNameLabel: instanceName,
						spec.GarmFlavourLabel:      "small",
						spec.GarmOSArchLabel:       "arm64",
						spec.GarmOSTypeLabel:       "linux",
						spec.GarmPoolIDLabel:       "ddce45e7-1bbb-4ecd-92cb-c733372b5cde",
						spec.GarmControllerIDLabel: controllerID,
						spec.GarmRunnerGroupLabel:  "",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "my-additional-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "",
									SizeLimit: nil,
								},
							},
						},
						{
							Name: "runner",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "",
									SizeLimit: nil,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "runner",
							Image: "localhost:5000/runner:ubuntu-22.04",
							Env: []corev1.EnvVar{
								{
									Name:  "MY_ADDITIONAL_ENV",
									Value: "test",
								},
								{
									Name:  "RUNNER_ORG",
									Value: "testorg",
								},
								{
									Name:  "RUNNER_REPO",
									Value: "",
								},
								{
									Name:  "RUNNER_ENTERPRISE",
									Value: "",
								},
								{
									Name:  "RUNNER_GROUP",
									Value: "",
								},
								{
									Name:  "RUNNER_NAME",
									Value: instanceName,
								},
								{
									Name:  "RUNNER_LABELS",
									Value: "road-runner,linux,arm64,kubernetes",
								},
								{
									Name:  "RUNNER_WORKDIR",
									Value: "/runner/_work/",
								},
								{
									Name:  "GITHUB_URL",
									Value: "https://github.com",
								},
								{
									Name:  "RUNNER_EPHEMERAL",
									Value: "true",
								},
								{
									Name:  "RUNNER_TOKEN",
									Value: "dummy",
								},
								{
									Name:  "METADATA_URL",
									Value: "https://metadata.test",
								},
								{
									Name:  "BEARER_TOKEN",
									Value: "test-token",
								},
								{
									Name:  "CALLBACK_URL",
									Value: "https://callback.test/status",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "runner",
									ReadOnly:  false,
									MountPath: "/runner",
								},
							},
							ImagePullPolicy: "Always",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
								},
							},
						},
						{
							Name: "sidecar",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:              resource.MustParse("500m"),
									corev1.ResourceMemory:           resource.MustParse("500Mi"),
									corev1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
			runtimeObjects: []runtime.Object{},
			err:            nil,
		},
		{
			name: "Valid bootstrapParams and merge pod template spec with added sidecar",
			config: &config.ProviderConfig{
				KubeConfigPath:    "",
				ContainerRegistry: "localhost:5000",
				RunnerNamespace:   "runner",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "runner",
								Env: []corev1.EnvVar{
									{
										Name:  "MY_ADDITIONAL_ENV",
										Value: "test",
									},
								},
							},
							{
								Name:  "sidecar",
								Image: "localhost:5000/sidecar:latest",
								Env: []corev1.EnvVar{
									{
										Name:  "MY_SIDECAR_ENV",
										Value: "test",
									},
								},
							},
						},
					},
				},
			},
			bootstrapParams: params.BootstrapInstance{
				Name:          instanceName,
				PoolID:        poolID,
				Flavor:        "small",
				RepoURL:       "https://github.com/testorg",
				InstanceToken: "test-token",
				MetadataURL:   "https://metadata.test",
				CallbackURL:   "https://callback.test/status",
				Image:         "runner:ubuntu-22.04",
				OSType:        "linux",
				OSArch:        "arm64",
				Labels:        []string{"road-runner", "linux", "arm64", "kubernetes"},
			},
			expectedProviderInstance: params.ProviderInstance{
				ProviderID: providerID,
				Name:       instanceName,
				OSType:     "linux",
				OSName:     "",
				OSVersion:  "",
				OSArch:     "arm64",
				Status:     "running",
			},
			expectedPodInstance: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      providerID,
					Namespace: "runner",
					Labels: map[string]string{
						spec.GarmInstanceNameLabel: instanceName,
						spec.GarmFlavourLabel:      "small",
						spec.GarmOSArchLabel:       "arm64",
						spec.GarmOSTypeLabel:       "linux",
						spec.GarmPoolIDLabel:       "ddce45e7-1bbb-4ecd-92cb-c733372b5cde",
						spec.GarmControllerIDLabel: controllerID,
						spec.GarmRunnerGroupLabel:  "",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "runner",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "",
									SizeLimit: nil,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "runner",
							Image: "localhost:5000/runner:ubuntu-22.04",
							Env: []corev1.EnvVar{
								{
									Name:  "MY_ADDITIONAL_ENV",
									Value: "test",
								},
								{
									Name:  "RUNNER_ORG",
									Value: "testorg",
								},
								{
									Name:  "RUNNER_REPO",
									Value: "",
								},
								{
									Name:  "RUNNER_ENTERPRISE",
									Value: "",
								},
								{
									Name:  "RUNNER_GROUP",
									Value: "",
								},
								{
									Name:  "RUNNER_NAME",
									Value: instanceName,
								},
								{
									Name:  "RUNNER_LABELS",
									Value: "road-runner,linux,arm64,kubernetes",
								},
								{
									Name:  "RUNNER_WORKDIR",
									Value: "/runner/_work/",
								},
								{
									Name:  "GITHUB_URL",
									Value: "https://github.com",
								},
								{
									Name:  "RUNNER_EPHEMERAL",
									Value: "true",
								},
								{
									Name:  "RUNNER_TOKEN",
									Value: "dummy",
								},
								{
									Name:  "METADATA_URL",
									Value: "https://metadata.test",
								},
								{
									Name:  "BEARER_TOKEN",
									Value: "test-token",
								},
								{
									Name:  "CALLBACK_URL",
									Value: "https://callback.test/status",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "runner",
									ReadOnly:  false,
									MountPath: "/runner",
								},
							},
							ImagePullPolicy: "Always",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
								},
							},
						},
						{
							Name:  "sidecar",
							Image: "localhost:5000/sidecar:latest",
							Env: []corev1.EnvVar{
								{
									Name:  "MY_SIDECAR_ENV",
									Value: "test",
								},
							},
						},
					},
				},
			},
			runtimeObjects: []runtime.Object{},
			err:            nil,
		},
		{
			name: "Valid bootstrapParams no pod template spec",
			config: &config.ProviderConfig{
				KubeConfigPath:    "",
				ContainerRegistry: "localhost:5000",
				RunnerNamespace:   "runner",
				PodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			bootstrapParams: params.BootstrapInstance{
				Name:          instanceName,
				PoolID:        poolID,
				Flavor:        "small",
				RepoURL:       "https://github.com/testorg",
				InstanceToken: "test-token",
				MetadataURL:   "https://metadata.test",
				CallbackURL:   "https://callback.test/status",
				Image:         "runner:ubuntu-22.04",
				OSType:        "linux",
				OSArch:        "arm64",
				Labels:        []string{"road-runner", "linux", "arm64", "kubernetes"},
			},
			expectedProviderInstance: params.ProviderInstance{
				ProviderID: providerID,
				Name:       instanceName,
				OSType:     "linux",
				OSName:     "",
				OSVersion:  "",
				OSArch:     "arm64",
				Status:     "running",
			},
			expectedPodInstance: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      providerID,
					Namespace: "runner",
					Labels: map[string]string{
						spec.GarmInstanceNameLabel: instanceName,
						spec.GarmFlavourLabel:      "small",
						spec.GarmOSArchLabel:       "arm64",
						spec.GarmOSTypeLabel:       "linux",
						spec.GarmPoolIDLabel:       "ddce45e7-1bbb-4ecd-92cb-c733372b5cde",
						spec.GarmControllerIDLabel: controllerID,
						spec.GarmRunnerGroupLabel:  "",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "runner",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "",
									SizeLimit: nil,
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "runner",
							Image: "localhost:5000/runner:ubuntu-22.04",
							Env: []corev1.EnvVar{
								{
									Name:  "RUNNER_ORG",
									Value: "testorg",
								},
								{
									Name:  "RUNNER_REPO",
									Value: "",
								},
								{
									Name:  "RUNNER_ENTERPRISE",
									Value: "",
								},
								{
									Name:  "RUNNER_GROUP",
									Value: "",
								},
								{
									Name:  "RUNNER_NAME",
									Value: instanceName,
								},
								{
									Name:  "RUNNER_LABELS",
									Value: "road-runner,linux,arm64,kubernetes",
								},
								{
									Name:  "RUNNER_WORKDIR",
									Value: "/runner/_work/",
								},
								{
									Name:  "GITHUB_URL",
									Value: "https://github.com",
								},
								{
									Name:  "RUNNER_EPHEMERAL",
									Value: "true",
								},
								{
									Name:  "RUNNER_TOKEN",
									Value: "dummy",
								},
								{
									Name:  "METADATA_URL",
									Value: "https://metadata.test",
								},
								{
									Name:  "BEARER_TOKEN",
									Value: "test-token",
								},
								{
									Name:  "CALLBACK_URL",
									Value: "https://callback.test/status",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "runner",
									ReadOnly:  false,
									MountPath: "/runner",
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
								},
							},
						},
					},
				},
			},
			runtimeObjects: []runtime.Object{},
			err:            nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create the provider config
			config.Config = *tc.config

			// create a fake kubernetes client
			client := fake.NewSimpleClientset(tc.runtimeObjects...)

			// initialize the provider
			p, _ := provider.NewKubernetesProvider(client, controllerID, poolID)

			// trigger the instance creation
			actual, err := p.CreateInstance(context.Background(), tc.bootstrapParams)
			assert.Equal(t, tc.err, err)

			// get the created pod
			createdPod, err := client.CoreV1().Pods(config.Config.RunnerNamespace).Get(context.Background(), actual.ProviderID, metav1.GetOptions{})
			assert.Equal(t, tc.err, err)

			// compare created instance with expected instance
			assert.Equal(t, tc.expectedProviderInstance, actual)

			// compare created pod with expected pod
			assert.Equal(t, tc.expectedPodInstance, createdPod)
		})
	}
}

func TestGetInstance(t *testing.T) {
	testCases := []struct {
		name                     string
		config                   *config.ProviderConfig
		expectedProviderInstance params.ProviderInstance
		runtimeObjects           []runtime.Object
		wantErr                  error
	}{
		{
			name: "Get Instance",
			config: &config.ProviderConfig{
				KubeConfigPath:    "",
				ContainerRegistry: "localhost:5000",
				RunnerNamespace:   "runner",
			},
			expectedProviderInstance: params.ProviderInstance{
				ProviderID: providerID,
				Name:       instanceName,
				OSType:     "linux",
				OSName:     "",
				OSVersion:  "",
				OSArch:     "arm64",
				Status:     "running",
			},
			runtimeObjects: []runtime.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      providerID,
						Namespace: "runner",
						Labels: map[string]string{
							spec.GarmInstanceNameLabel: instanceName,
							spec.GarmFlavourLabel:      "small",
							spec.GarmOSArchLabel:       "arm64",
							spec.GarmOSTypeLabel:       "linux",
							spec.GarmPoolIDLabel:       "ddce45e7-1bbb-4ecd-92cb-c733372b5cde",
							spec.GarmControllerIDLabel: controllerID,
							spec.GarmRunnerGroupLabel:  "",
						},
					},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "runner",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{
										Medium:    "",
										SizeLimit: nil,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "runner",
								Image: "localhost:5000/runner:ubuntu-22.04",
								Env: []corev1.EnvVar{
									{
										Name:  "RUNNER_ORG",
										Value: "testorg",
									},
									{
										Name:  "RUNNER_REPO",
										Value: "",
									},
									{
										Name:  "RUNNER_ENTERPRISE",
										Value: "",
									},
									{
										Name:  "RUNNER_GROUP",
										Value: "",
									},
									{
										Name:  "RUNNER_NAME",
										Value: instanceName,
									},
									{
										Name:  "RUNNER_LABELS",
										Value: "road-runner,linux,arm64,kubernetes",
									},
									{
										Name:  "RUNNER_WORKDIR",
										Value: "/runner/_work/",
									},
									{
										Name:  "GITHUB_URL",
										Value: "https://github.com",
									},
									{
										Name:  "RUNNER_EPHEMERAL",
										Value: "true",
									},
									{
										Name:  "RUNNER_TOKEN",
										Value: "dummy",
									},
									{
										Name:  "METADATA_URL",
										Value: "https://metadata.test",
									},
									{
										Name:  "BEARER_TOKEN",
										Value: "test-token",
									},
									{
										Name:  "CALLBACK_URL",
										Value: "https://callback.test/status",
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "runner",
										ReadOnly:  false,
										MountPath: "/runner",
									},
								},
								ImagePullPolicy: "Always",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: "Running",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config.Config = *tc.config

			client := fake.NewSimpleClientset(tc.runtimeObjects...)

			p, _ := provider.NewKubernetesProvider(client, controllerID, poolID)

			actual, err := p.GetInstance(context.Background(), instanceName)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.expectedProviderInstance, actual)
		})
	}
}

func TestDeleteInstance(t *testing.T) {
	testCases := []struct {
		name                     string
		config                   *config.ProviderConfig
		expectedProviderInstance params.ProviderInstance
		runtimeObjects           []runtime.Object
		wantErr                  error
	}{
		{
			name: "Delete Instance Success",
			config: &config.ProviderConfig{
				KubeConfigPath:    "",
				ContainerRegistry: "localhost:5000",
				RunnerNamespace:   "runner",
			},
			runtimeObjects: []runtime.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      providerID,
						Namespace: "runner",
						Labels: map[string]string{
							spec.GarmInstanceNameLabel: instanceName,
							spec.GarmFlavourLabel:      "small",
							spec.GarmOSArchLabel:       "arm64",
							spec.GarmOSTypeLabel:       "linux",
							spec.GarmPoolIDLabel:       "ddce45e7-1bbb-4ecd-92cb-c733372b5cde",
							spec.GarmControllerIDLabel: controllerID,
							spec.GarmRunnerGroupLabel:  "",
						},
					},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "runner",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{
										Medium:    "",
										SizeLimit: nil,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "runner",
								Image: "localhost:5000/runner:ubuntu-22.04",
								Env: []corev1.EnvVar{
									{
										Name:  "RUNNER_ORG",
										Value: "testorg",
									},
									{
										Name:  "RUNNER_REPO",
										Value: "",
									},
									{
										Name:  "RUNNER_ENTERPRISE",
										Value: "",
									},
									{
										Name:  "RUNNER_GROUP",
										Value: "",
									},
									{
										Name:  "RUNNER_NAME",
										Value: instanceName,
									},
									{
										Name:  "RUNNER_LABELS",
										Value: "road-runner,linux,arm64,kubernetes",
									},
									{
										Name:  "RUNNER_WORKDIR",
										Value: "/runner/_work/",
									},
									{
										Name:  "GITHUB_URL",
										Value: "https://github.com",
									},
									{
										Name:  "RUNNER_EPHEMERAL",
										Value: "true",
									},
									{
										Name:  "RUNNER_TOKEN",
										Value: "dummy",
									},
									{
										Name:  "METADATA_URL",
										Value: "https://metadata.test",
									},
									{
										Name:  "BEARER_TOKEN",
										Value: "test-token",
									},
									{
										Name:  "CALLBACK_URL",
										Value: "https://callback.test/status",
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "runner",
										ReadOnly:  false,
										MountPath: "/runner",
									},
								},
								ImagePullPolicy: "Always",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: "Running",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create the provider config
			config.Config = *tc.config

			// create a fake kubernetes client
			client := fake.NewSimpleClientset(tc.runtimeObjects...)

			// initialize the provider
			p, _ := provider.NewKubernetesProvider(client, controllerID, poolID)

			labels := make(map[string]string)
			labels[spec.GarmPoolIDLabel] = spec.ToValidLabel(poolID)

			labelSelector := metav1.LabelSelector{MatchLabels: labels}
			labelSelectorStr, _ := metav1.LabelSelectorAsSelector(&labelSelector)

			// get all pods in the configured namespace
			pods, err := p.ClientSet.
				CoreV1().
				Pods(config.Config.RunnerNamespace).
				List(context.Background(), metav1.ListOptions{
					LabelSelector: labelSelectorStr.String(),
				})
			assert.NoError(t, err)
			assert.Equal(t, len(pods.Items), len(tc.runtimeObjects))

			// trigger the instance deletion
			err = p.DeleteInstance(context.Background(), instanceName)
			assert.Equal(t, tc.wantErr, err)

			if tc.wantErr == nil && err == nil {
				pods, err := p.ClientSet.
					CoreV1().
					Pods(config.Config.RunnerNamespace).
					List(context.Background(), metav1.ListOptions{
						LabelSelector: labelSelectorStr.String(),
					})
				assert.NoError(t, err)
				assert.Equal(t, len(pods.Items), 0)
			}
		})
	}
}

func TestRemoveAllInstances(t *testing.T) {
	testCases := []struct {
		name           string
		config         *config.ProviderConfig
		runtimeObjects []runtime.Object
		wantErr        error
	}{
		{
			name: "Remove All Instances Success",
			config: &config.ProviderConfig{
				KubeConfigPath:    "",
				ContainerRegistry: "localhost:5000",
				RunnerNamespace:   "runner",
			},
			runtimeObjects: []runtime.Object{
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      providerID,
						Namespace: "runner",
						Labels: map[string]string{
							spec.GarmInstanceNameLabel: instanceName,
							spec.GarmFlavourLabel:      "small",
							spec.GarmOSArchLabel:       "arm64",
							spec.GarmOSTypeLabel:       "linux",
							spec.GarmPoolIDLabel:       "ddce45e7-1bbb-4ecd-92cb-c733372b5cde",
							spec.GarmControllerIDLabel: controllerID,
							spec.GarmRunnerGroupLabel:  "",
						},
					},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "runner",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{
										Medium:    "",
										SizeLimit: nil,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "runner",
								Image: "localhost:5000/runner:ubuntu-22.04",
								Env: []corev1.EnvVar{
									{
										Name:  "RUNNER_ORG",
										Value: "testorg",
									},
									{
										Name:  "RUNNER_REPO",
										Value: "",
									},
									{
										Name:  "RUNNER_ENTERPRISE",
										Value: "",
									},
									{
										Name:  "RUNNER_GROUP",
										Value: "",
									},
									{
										Name:  "RUNNER_NAME",
										Value: instanceName,
									},
									{
										Name:  "RUNNER_LABELS",
										Value: "road-runner,linux,arm64,kubernetes",
									},
									{
										Name:  "RUNNER_WORKDIR",
										Value: "/runner/_work/",
									},
									{
										Name:  "GITHUB_URL",
										Value: "https://github.com",
									},
									{
										Name:  "RUNNER_EPHEMERAL",
										Value: "true",
									},
									{
										Name:  "RUNNER_TOKEN",
										Value: "dummy",
									},
									{
										Name:  "METADATA_URL",
										Value: "https://metadata.test",
									},
									{
										Name:  "BEARER_TOKEN",
										Value: "test-token",
									},
									{
										Name:  "CALLBACK_URL",
										Value: "https://callback.test/status",
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "runner",
										ReadOnly:  false,
										MountPath: "/runner",
									},
								},
								ImagePullPolicy: "Always",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: "Running",
					},
				},
				&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "garm-cdciedbcijh",
						Namespace: "runner",
						Labels: map[string]string{
							spec.GarmInstanceNameLabel: instanceName,
							spec.GarmFlavourLabel:      "small",
							spec.GarmOSArchLabel:       "arm64",
							spec.GarmOSTypeLabel:       "linux",
							spec.GarmPoolIDLabel:       "ddce45e7-1bbb-4ecd-92cb-c733372b5cde",
							spec.GarmControllerIDLabel: controllerID,
							spec.GarmRunnerGroupLabel:  "",
						},
					},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "runner",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{
										Medium:    "",
										SizeLimit: nil,
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "runner",
								Image: "localhost:5000/runner:ubuntu-22.04",
								Env: []corev1.EnvVar{
									{
										Name:  "RUNNER_ORG",
										Value: "testorg",
									},
									{
										Name:  "RUNNER_REPO",
										Value: "",
									},
									{
										Name:  "RUNNER_ENTERPRISE",
										Value: "",
									},
									{
										Name:  "RUNNER_GROUP",
										Value: "",
									},
									{
										Name:  "RUNNER_NAME",
										Value: instanceName,
									},
									{
										Name:  "RUNNER_LABELS",
										Value: "road-runner,linux,arm64,kubernetes",
									},
									{
										Name:  "RUNNER_WORKDIR",
										Value: "/runner/_work/",
									},
									{
										Name:  "GITHUB_URL",
										Value: "https://github.com",
									},
									{
										Name:  "RUNNER_EPHEMERAL",
										Value: "true",
									},
									{
										Name:  "RUNNER_TOKEN",
										Value: "dummy",
									},
									{
										Name:  "METADATA_URL",
										Value: "https://metadata.test",
									},
									{
										Name:  "BEARER_TOKEN",
										Value: "test-token",
									},
									{
										Name:  "CALLBACK_URL",
										Value: "https://callback.test/status",
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "runner",
										ReadOnly:  false,
										MountPath: "/runner",
									},
								},
								ImagePullPolicy: "Always",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("500m"),
										corev1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: "Running",
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config.Config = *tc.config

			client := fake.NewSimpleClientset(tc.runtimeObjects...)

			p, _ := provider.NewKubernetesProvider(client, controllerID, poolID)

			labels := make(map[string]string)
			labels[spec.GarmPoolIDLabel] = spec.ToValidLabel(poolID)

			labelSelector := metav1.LabelSelector{MatchLabels: labels}
			labelSelectorStr, _ := metav1.LabelSelectorAsSelector(&labelSelector)

			pods, err := p.ClientSet.
				CoreV1().
				Pods(config.Config.RunnerNamespace).
				List(context.Background(), metav1.ListOptions{
					LabelSelector: labelSelectorStr.String(),
				})
			assert.NoError(t, err)
			assert.Equal(t, len(pods.Items), len(tc.runtimeObjects))

			err = p.RemoveAllInstances(context.Background())
			assert.Equal(t, tc.wantErr, err)

			if tc.wantErr == nil && err == nil {
				pods, err := p.ClientSet.
					CoreV1().
					Pods(config.Config.RunnerNamespace).
					List(context.Background(), metav1.ListOptions{
						LabelSelector: labelSelectorStr.String(),
					})
				assert.NoError(t, err)
				assert.Equal(t, len(pods.Items), 0)
			}
		})
	}
}
