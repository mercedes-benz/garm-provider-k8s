// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"

	"github.com/mercedes-benz/garm-provider-k8s/config"
)

type IKubeClientWrapper interface {
	CreateNamespaceIfNotExists(name string) (*corev1.Namespace, error)
	DeleteNamespace(name string) error
	CreatePod(pod *corev1.Pod, namespace string) (*corev1.Pod, error)
	GetPod(name, namespace string) (*corev1.Pod, error)
	DeletePod(name, namespace string) error
	ListPods(namespace string) (*corev1.PodList, error)
	ListPodsByLabels(labels map[string]string, namespace string) (*corev1.PodList, error)
}

type KubeClientWrapper struct {
	Client kubernetes.Interface
}

func NewKubeClient(config *config.Config) (IKubeClientWrapper, error) {
	var client kubernetes.Interface
	var err error

	if config.KubeConfigPath == "" {
		client, err = inClusterConfig()
	} else {
		client, err = outOfClusterConfig(config.KubeConfigPath)
	}
	if err != nil {
		return nil, err
	}

	return &KubeClientWrapper{
		Client: client,
	}, nil
}

func (w *KubeClientWrapper) CreateNamespaceIfNotExists(name string) (*corev1.Namespace, error) {
	if name == "" {
		return &corev1.Namespace{}, errors.New("please provide a namespace name to create")
	}

	existingNamespace, err := w.Client.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil && existingNamespace != nil {
		return existingNamespace, err
	}
	newNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	result, err := w.Client.CoreV1().Namespaces().Create(context.TODO(), newNamespace, metav1.CreateOptions{})
	if err != nil {
		return &corev1.Namespace{}, err
	}
	return result, nil
}

func (w *KubeClientWrapper) DeleteNamespace(name string) error {
	if name == "" {
		return errors.New("can not delete namespace: no namespace provided")
	}
	seconds := int64(0)
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &seconds,
	}

	err := w.Client.CoreV1().Namespaces().Delete(context.TODO(), name, deleteOptions)
	return err
}

func (w *KubeClientWrapper) CreatePod(pod *corev1.Pod, namespace string) (*corev1.Pod, error) {
	createdPod, err := w.Client.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return &corev1.Pod{}, err
	}
	return createdPod, nil
}

func (w *KubeClientWrapper) GetPod(name, namespace string) (*corev1.Pod, error) {
	var err error

	if name == "" {
		return &corev1.Pod{}, errors.New("can not get pod: no name provided")
	}

	if namespace != "" {
		pod, err := w.Client.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return &corev1.Pod{}, err
		}
		return pod, nil
	}

	pods, err := w.Client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return &corev1.Pod{}, err
	}

	for _, pod := range pods.Items {
		if pod.Name == name {
			return &pod, nil
		}
	}

	return &corev1.Pod{}, fmt.Errorf("no pod found with name %v", name)
}

func (w *KubeClientWrapper) DeletePod(name, namespace string) error {
	var err error

	// namespace exists
	if namespace != "" {
		err = w.Client.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	}

	listOption := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	}

	pods, err := w.Client.CoreV1().Pods("").List(context.Background(), listOption)
	if err != nil {
		return err
	}

	if len(pods.Items) > 1 {
		return fmt.Errorf("can not delete pod %v as there are multiple instances across namespaces", name)
	}

	for _, pod := range pods.Items {
		if pod.Name == name {
			err = w.Client.CoreV1().Pods(pod.Namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (w *KubeClientWrapper) ListPods(namespace string) (*corev1.PodList, error) {
	podList, err := w.Client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return podList, nil
}

func (w *KubeClientWrapper) ListPodsByLabels(labels map[string]string, namespace string) (*corev1.PodList, error) {
	labelSelector := metav1.LabelSelector{MatchLabels: labels}
	labelSelectorStr, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return nil, err
	}

	// supply empty string for ns to search across namespaces
	podList, err := w.Client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelectorStr.String(),
	})
	if err != nil {
		return nil, err
	}

	return podList, nil
}

func outOfClusterConfig(kubeConfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func inClusterConfig() (*kubernetes.Clientset, error) {
	config, err := initRestClient()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func initRestClient() (*rest.Config, error) {
	const (
		/* #nosec */
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	tlsClientConfig := rest.TLSClientConfig{}

	if _, err := certutil.NewPool(rootCAFile); err != nil {
		klog.Errorf("Expected to load root CA config from %s, but got err: %v", rootCAFile, err)
	} else {
		tlsClientConfig.CAFile = rootCAFile
	}

	return &rest.Config{
		Host:            "https://kubernetes.default.svc",
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(token),
		BearerTokenFile: tokenFile,
	}, nil
}
