// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudbase/garm-provider-common/params"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/mercedes-benz/garm-provider-k8s/client"
	"github.com/mercedes-benz/garm-provider-k8s/config"
	"github.com/mercedes-benz/garm-provider-k8s/internal/spec"
	"github.com/mercedes-benz/garm-provider-k8s/pkg/diff"
)

type Provider struct {
	ControllerID string
	Config       *config.Config
	KubeClient   client.IKubeClientWrapper
}

func (p Provider) CreateInstance(_ context.Context, bootstrapParams params.BootstrapInstance) (params.ProviderInstance, error) {
	podName := strings.ToLower(bootstrapParams.Name)
	labels := spec.ParamsToPodLabels(p.ControllerID, bootstrapParams)
	fullImageName := spec.GetFullImagePath(p.Config.ContainerRegistry, bootstrapParams.Image)
	resourceRequirements := spec.FlavourToResourceRequirements(spec.Flavour(bootstrapParams.Flavor))

	gitHubScopeDetails, err := spec.ExtractGitHubScopeDetails(bootstrapParams.RepoURL)
	if err != nil {
		return params.ProviderInstance{}, err
	}

	envs := spec.GetRunnerEnvs(gitHubScopeDetails, bootstrapParams)

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: p.Config.RunnerNamespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "runner",
					Image:           fullImageName,
					Resources:       resourceRequirements,
					Env:             envs,
					ImagePullPolicy: "Always",
				},
			},
		},
	}

	err = spec.CreateRunnerVolume(pod)
	if err != nil {
		return params.ProviderInstance{}, err
	}

	mergedPod, err := mergePodSpecs(pod, p.Config.PodTemplate)
	if err != nil {
		return params.ProviderInstance{}, err
	}

	createdPod, err := p.KubeClient.CreatePod(mergedPod, p.Config.RunnerNamespace)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error calling CreateInstance: can not create pod %v in namespace %v: %w", pod.Name, p.Config.RunnerNamespace, err)
	}

	result, err := spec.PodToInstance(createdPod, params.InstanceRunning)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error calling CreateInstance: can not map pod %v to params.Instance: %w", pod.Name, err)
	}

	return *result, nil
}

func mergePodSpecs(pod *corev1.Pod, template corev1.PodTemplateSpec) (*corev1.Pod, error) {
	if reflect.ValueOf(template).IsZero() {
		return pod, nil
	}

	patch, _, err := diff.CreateTwoWayMergePatch(pod.Spec, template, corev1.PodTemplateSpec{})
	if err != nil {
		return nil, err
	}

	mergeBytes, err := diff.StrategicMergePatch(pod, patch, corev1.Pod{})
	if err != nil {
		return nil, err
	}

	mergedPod := &corev1.Pod{}
	json.Unmarshal(mergeBytes, mergedPod)
	return mergedPod, nil
}

func (p Provider) DeleteInstance(_ context.Context, instance string) error {
	err := p.KubeClient.DeletePod(instance, p.Config.RunnerNamespace)
	if err != nil {
		return fmt.Errorf("error calling DeleteInstance: can not delete instance %s: %w", instance, err)
	}
	return nil
}

func (p Provider) GetInstance(_ context.Context, instance string) (params.ProviderInstance, error) {
	pod, err := p.KubeClient.GetPod(instance, p.Config.RunnerNamespace)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error calling GetInstance: can not get instance %s: %s", instance, err)
	}

	result, err := spec.PodToInstance(pod, "")
	if err != nil {
		return params.ProviderInstance{}, err
	}

	return *result, nil
}

func (p Provider) ListInstances(_ context.Context, poolID string) ([]params.ProviderInstance, error) {
	labels := make(map[string]string)
	labels[spec.GarmPoolIDLabel] = spec.ToValidLabel(poolID)

	pods, err := p.KubeClient.ListPodsByLabels(labels, p.Config.RunnerNamespace)
	if err != nil {
		return []params.ProviderInstance{}, fmt.Errorf("could not list pods: %w", err)
	}
	result := make([]params.ProviderInstance, len(pods.Items))
	for i, item := range pods.Items {
		pod := item
		instance, err := spec.PodToInstance(&pod, "")
		if err != nil {
			return []params.ProviderInstance{}, err
		}
		result[i] = *instance
	}
	return result, nil
}

func (p Provider) RemoveAllInstances(ctx context.Context) error {
	log := log.FromContext(ctx)

	labels := make(map[string]string)
	labels[spec.GarmControllerIDLabel] = p.ControllerID

	pods, err := p.KubeClient.ListPodsByLabels(labels, "")
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		err = p.KubeClient.DeletePod(pod.Name, pod.Namespace)
		if err != nil {
			log.Error(err, fmt.Sprintf("Error deleting some pods: %v in namespace %v", pod.Name, pod.Namespace))
		}
	}
	return nil
}

func (p Provider) Stop(_ context.Context, instance string, force bool) error {
	panic(fmt.Sprintf("Stop() not implemented, called with instance: %s, force: %t", instance, force))
}

func (p Provider) Start(_ context.Context, instance string) error {
	panic(fmt.Sprintf("Start() not implemented, called with instance: %s", instance))
}

func NewKubernetesProvider(
	kubeClient client.IKubeClientWrapper,
	config *config.Config,
	controllerID string,
) (*Provider, error) {
	return &Provider{
		KubeClient:   kubeClient,
		Config:       config,
		ControllerID: controllerID,
	}, nil
}
