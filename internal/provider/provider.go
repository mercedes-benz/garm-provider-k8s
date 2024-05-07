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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/mercedes-benz/garm-provider-k8s/internal/spec"
	"github.com/mercedes-benz/garm-provider-k8s/pkg/config"
	"github.com/mercedes-benz/garm-provider-k8s/pkg/diff"
)

type Provider struct {
	ControllerID  string
	ClientSet     kubernetes.Interface
	LabelSelector labels.Selector
}

func (p Provider) CreateInstance(_ context.Context, bootstrapParams params.BootstrapInstance) (params.ProviderInstance, error) {
	podName := strings.ToLower(bootstrapParams.Name)
	labels := spec.ParamsToPodLabels(p.ControllerID, bootstrapParams)
	resourceRequirements := spec.FlavorToResourceRequirements(bootstrapParams.Flavor)

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
			Namespace: config.Config.RunnerNamespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:            "runner",
					Image:           bootstrapParams.Image,
					Resources:       resourceRequirements,
					Env:             envs,
					ImagePullPolicy: corev1.PullAlways,
				},
			},
		},
	}

	err = p.ensureNamespace(config.Config.RunnerNamespace)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("ensuring runner namespace %s failed: %w", config.Config.RunnerNamespace, err)
	}

	err = spec.CreateRunnerVolume(pod)
	if err != nil {
		return params.ProviderInstance{}, err
	}

	mergedPod, err := mergePodSpecs(pod, config.Config.PodTemplate)
	if err != nil {
		return params.ProviderInstance{}, err
	}

	pod, err = p.ClientSet.CoreV1().
		Pods(config.Config.RunnerNamespace).
		Create(context.Background(), mergedPod, metav1.CreateOptions{})
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error calling CreateInstance: can not create pod %v: %w", pod.Name, err)
	}

	result, err := spec.PodToInstance(pod, params.InstanceRunning)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error calling CreateInstance: can not map pod %v to params.Instance: %w", pod.Name, err)
	}

	return *result, nil
}

func (p Provider) ensureNamespace(runnerNamespace string) error {
	_, err := p.ClientSet.CoreV1().
		Namespaces().
		Get(context.Background(), runnerNamespace, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	// if namespace doesn't exist
	// there is no need for creating again
	if apierrors.IsNotFound(err) {
		_, err = p.ClientSet.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: runnerNamespace,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
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
	podName := strings.ToLower(instance)
	err := p.ClientSet.CoreV1().
		Pods(config.Config.RunnerNamespace).
		Delete(context.Background(), podName, metav1.DeleteOptions{})
	if err != nil {
		// if pod is not found, return nil so garm can delete the instance
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error calling DeleteInstance: can not delete instance %s: %w", instance, err)
	}
	return nil
}

func (p Provider) GetInstance(_ context.Context, instance string) (params.ProviderInstance, error) {
	podName := strings.ToLower(instance)

	pod, err := p.ClientSet.CoreV1().
		Pods(config.Config.RunnerNamespace).
		Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("error calling GetInstance: can not get instance %s: %s", instance, err)
	}

	result, err := spec.PodToInstance(pod, "")
	if err != nil {
		return params.ProviderInstance{}, err
	}

	return *result, nil
}

func (p Provider) ListInstances(_ context.Context, _ string) ([]params.ProviderInstance, error) {
	pods, err := p.ClientSet.
		CoreV1().
		Pods(config.Config.RunnerNamespace).
		List(context.Background(), metav1.ListOptions{
			LabelSelector: p.LabelSelector.String(),
		})
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

	pods, err := p.ClientSet.
		CoreV1().
		Pods(config.Config.RunnerNamespace).
		List(context.Background(), metav1.ListOptions{
			LabelSelector: p.LabelSelector.String(),
		})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		err := p.ClientSet.CoreV1().
			Pods(config.Config.RunnerNamespace).
			Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
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

func NewKubernetesProvider(clientSet kubernetes.Interface, controllerID, poolID string) (*Provider, error) {
	labelSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{
			spec.GarmControllerIDLabel: controllerID,
			spec.GarmPoolIDLabel:       poolID,
		},
	})
	if err != nil {
		return nil, err
	}
	return &Provider{
		ControllerID:  controllerID,
		ClientSet:     clientSet,
		LabelSelector: labelSelector,
	}, nil
}
