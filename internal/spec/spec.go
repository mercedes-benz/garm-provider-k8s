// SPDX-License-Identifier: MIT

package spec

import (
	"fmt"
	"github.com/cloudbase/garm-provider-common/params"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/json"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	GarmRunnerNameLabel   = "garm/runner-name"
	GarmControllerIDLabel = "garm/controllerID"
	GarmImageLabel        = "garm/image"
	GarmFlavourLabel      = "garm/flavour"
	GarmOSTypeLabel       = "garm/os_type"
	GarmOSArchLabel       = "garm/os_arch"
	GarmOSNameLabel       = "garm/os_name"
	GarmOSVersionLabel    = "garm/os_version"
	GarmRunnerGroupLabel  = "garm/runner-group"
	GarmPoolIDLabel       = "garm/poolID"
)

type GitHubScopeDetails struct {
	BaseURL    string
	Repo       string
	Org        string
	Enterprise string
}

type Flavour string

const (
	Small  Flavour = "small"
	Medium Flavour = "medium"
	Large  Flavour = "large"
)

type OSType string
type OSName string
type OSArch string
type OSVersion string

type ExtraSpecs struct {
	OSName    OSName
	OSVersion OSVersion
}

type ImageDetails struct {
	OSType    OSType
	OSName    OSName
	OSArch    OSArch
	OSVersion OSVersion
}

var statusMap = map[string]string{
	"Running":   "running",
	"Succeeded": "stopped",
	"Pending":   "pending_create",
	"Failed":    "error",
	"Unknown":   "unknown",
}

func PodToInstance(pod *corev1.Pod, overwriteInstanceStatus params.InstanceStatus) (*params.ProviderInstance, error) {
	instanceName, ok := pod.ObjectMeta.Labels[GarmRunnerNameLabel]
	if !ok {
		return &params.ProviderInstance{}, fmt.Errorf("Error converting pod to params.Instance: pod  %s has no label: %s", pod.Name, GarmRunnerNameLabel)
	}

	// for garm to work properly, during creation of instance status needs to be set manually to "running", other than the status is derived from pod.Status.Phase
	if overwriteInstanceStatus == "" {
		overwriteInstanceStatus = params.InstanceStatus(statusMap[string(pod.Status.Phase)])
	}

	imageDetails := ExtractImageDetails(pod)

	return &params.ProviderInstance{
		ProviderID: pod.Name,
		Name:       instanceName,
		Status:     overwriteInstanceStatus,
		OSArch:     params.OSArch(imageDetails.OSArch),
		OSType:     params.OSType(imageDetails.OSType),
		OSName:     string(imageDetails.OSName),
		OSVersion:  string(imageDetails.OSVersion),
	}, nil
}

func ParamsToPodLabels(controllerID string, bootstrapParams params.BootstrapInstance) map[string]string {
	labels := make(map[string]string)
	extraSpecs := ExtraSpecs{}

	err := json.Unmarshal(bootstrapParams.ExtraSpecs, &extraSpecs)
	if err == nil {
		labels[GarmOSNameLabel] = string(extraSpecs.OSName)
		labels[GarmOSVersionLabel] = string(extraSpecs.OSVersion)
	}

	labels[GarmRunnerNameLabel] = ToValidLabel(bootstrapParams.Name)
	labels[GarmControllerIDLabel] = ToValidLabel(controllerID)
	labels[GarmPoolIDLabel] = ToValidLabel(bootstrapParams.PoolID)
	labels[GarmImageLabel] = ToValidLabel(bootstrapParams.Image)
	labels[GarmFlavourLabel] = ToValidLabel(bootstrapParams.Flavor)
	labels[GarmOSTypeLabel] = ToValidLabel(string(bootstrapParams.OSType))
	labels[GarmOSArchLabel] = ToValidLabel(string(bootstrapParams.OSArch))
	labels[GarmRunnerGroupLabel] = ToValidLabel(bootstrapParams.GitHubRunnerGroup)

	return labels
}

func FlavourToResourceRequirements(flavour Flavour) corev1.ResourceRequirements {
	resourceCPU := "500m"
	resourceMemory := "500Mi"

	switch flavour {
	case Medium:
		resourceCPU = "1000m"
		resourceMemory = "1Gi"
	case Large:
		resourceCPU = "2000m"
		resourceMemory = "2Gi"
	}

	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(resourceCPU),
			corev1.ResourceMemory: resource.MustParse(resourceMemory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(resourceCPU),
			corev1.ResourceMemory: resource.MustParse(resourceMemory),
		},
	}
}

func ExtractGitHubScopeDetails(gitRepoURL string) (GitHubScopeDetails, error) {
	if gitRepoURL == "" {
		return GitHubScopeDetails{}, fmt.Errorf("no gitRepoURL supplied")
	}
	u, err := url.Parse(gitRepoURL)
	if err != nil {
		return GitHubScopeDetails{}, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return GitHubScopeDetails{}, fmt.Errorf("invalid URL: %s", gitRepoURL)
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")

	scope := GitHubScopeDetails{
		BaseURL: u.Scheme + "://" + u.Host,
	}

	switch {
	case len(pathParts) == 1:
		scope.Org = pathParts[0]
	case len(pathParts) == 2 && pathParts[0] == "enterprises":
		scope.Enterprise = pathParts[1]
	case len(pathParts) == 2:
		scope.Org = pathParts[0]
		scope.Repo = pathParts[1]
	default:
		return GitHubScopeDetails{}, fmt.Errorf("URL does not match the expected patterns")
	}

	return scope, nil
}

func isValidLabelChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_' || r == '.'
}

func ToValidLabel(input string) string {
	var sb strings.Builder

	for _, r := range input {
		if isValidLabelChar(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}

	result := sb.String()

	// Ensure the resulting string starts and ends with an alphanumeric character
	if len(result) > 0 {
		if !unicode.IsLetter(rune(result[0])) && !unicode.IsNumber(rune(result[0])) {
			result = "a" + result[1:]
		}

		lastIndex := len(result) - 1
		if !unicode.IsLetter(rune(result[lastIndex])) && !unicode.IsNumber(rune(result[lastIndex])) {
			result = result[:lastIndex] + "0"
		}
	}

	return result
}

func CreateRunnerVolume(pod *corev1.Pod) error {
	runnerVolumeName := "runner"
	runnerVolumeMountPath := "/runner"
	runnerVolumeEmptyDir := &corev1.EmptyDirVolumeSource{}

	if len(pod.Spec.Containers) < 1 {
		return fmt.Errorf("pod %s has no runner container spec", pod.Name)
	}

	runnerContainer := &pod.Spec.Containers[0]

	volume := corev1.Volume{
		Name: runnerVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: runnerVolumeEmptyDir,
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)

	volumeMount := corev1.VolumeMount{
		Name:      runnerVolumeName,
		MountPath: runnerVolumeMountPath,
	}

	runnerContainer.VolumeMounts = append(runnerContainer.VolumeMounts, volumeMount)
	return nil
}

func GetRunnerEnvs(gitHubScope GitHubScopeDetails, bootstrapParams params.BootstrapInstance) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "RUNNER_ORG",
			Value: gitHubScope.Org,
		},
		{
			Name:  "RUNNER_REPO",
			Value: gitHubScope.Repo,
		},
		{
			Name:  "RUNNER_ENTERPRISE",
			Value: gitHubScope.Enterprise,
		},
		{
			Name:  "RUNNER_GROUP",
			Value: bootstrapParams.GitHubRunnerGroup,
		},
		{
			Name:  "RUNNER_NAME",
			Value: bootstrapParams.Name,
		},
		{
			Name:  "RUNNER_LABELS",
			Value: strings.Join(bootstrapParams.Labels, ","),
		},
		{
			Name:  "RUNNER_WORKDIR",
			Value: "/runner/_work/",
		},
		{
			Name:  "GITHUB_URL",
			Value: gitHubScope.BaseURL,
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
			Value: bootstrapParams.MetadataURL,
		},
		{
			Name:  "BEARER_TOKEN",
			Value: bootstrapParams.InstanceToken,
		},
		{
			Name:  "CALLBACK_URL",
			Value: bootstrapParams.CallbackURL,
		},
	}
}

func ExtractImageDetails(pod *corev1.Pod) *ImageDetails {
	return &ImageDetails{
		OSType:    OSType(pod.Labels[GarmOSTypeLabel]),
		OSName:    OSName(pod.Labels[GarmOSNameLabel]),
		OSVersion: OSVersion(pod.Labels[GarmOSVersionLabel]),
		OSArch:    OSArch(pod.Labels[GarmOSArchLabel]),
	}
}

func GetFullImagePath(containerRegistry, imageNameAndTag string) string {
	reg := filepath.Clean(containerRegistry)
	return reg + "/" + imageNameAndTag
}
