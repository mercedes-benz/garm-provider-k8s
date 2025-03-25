// SPDX-License-Identifier: MIT

package spec

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cloudbase/garm-provider-common/params"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/mercedes-benz/garm-provider-k8s/pkg/config"
)

var (
	runnerVolumeName      = "runner"
	runnerVolumeMountPath = "/runner"
	runnerVolumeEmptyDir  = &corev1.EmptyDirVolumeSource{}
)

const (
	GarmInstanceNameLabel = "garm/instance-name"
	GarmControllerIDLabel = "garm/controllerID"
	GarmFlavorLabel       = "garm/flavor"
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

type (
	OSType    string
	OSName    string
	OSArch    string
	OSVersion string
)

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
	instanceName, ok := pod.ObjectMeta.Labels[GarmInstanceNameLabel]
	if !ok {
		instanceName = pod.Name
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

	labels[GarmInstanceNameLabel] = ToValidLabel(bootstrapParams.Name)
	labels[GarmControllerIDLabel] = ToValidLabel(controllerID)
	labels[GarmPoolIDLabel] = ToValidLabel(bootstrapParams.PoolID)
	labels[GarmFlavorLabel] = ToValidLabel(bootstrapParams.Flavor)
	labels[GarmOSTypeLabel] = ToValidLabel(string(bootstrapParams.OSType))
	labels[GarmOSArchLabel] = ToValidLabel(string(bootstrapParams.OSArch))
	labels[GarmRunnerGroupLabel] = ToValidLabel(bootstrapParams.GitHubRunnerGroup)

	return labels
}

func FlavorToResourceRequirements(flavor string) corev1.ResourceRequirements {
	if _, ok := config.Config.Flavors[flavor]; !ok {
		return corev1.ResourceRequirements{}
	}

	return config.Config.Flavors[flavor]
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
	if len(pod.Spec.Containers) < 1 {
		return fmt.Errorf("pod %s has no runner container spec", pod.Name)
	}

	// Skip volume creation if a volume with the default name already exists in podTemplate
	for _, vol := range config.Config.PodTemplate.Spec.Volumes {
		if vol.Name == runnerVolumeName {
			return nil
		}
	}

	volume := corev1.Volume{
		Name: runnerVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: runnerVolumeEmptyDir,
		},
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, volume)

	return nil
}

func CreateRunnerVolumeMount(pod *corev1.Pod, runnerContainerName string) error {
	if len(pod.Spec.Containers) < 1 {
		return fmt.Errorf("pod %s has no runner container spec", pod.Name)
	}

	// Skip volumemount creation if a volumemount with the same path already exists
	// in podTemplate for the default container
	for _, container := range config.Config.PodTemplate.Spec.Containers {
		if container.Name == runnerContainerName {
			for _, volMounts := range container.VolumeMounts {
				// Volumemount paths e.g. /runner and /runner/ are threated equal
				// The last one in the pod spec will take precedence, which can lead to unexpected behavior
				if filepath.Clean(volMounts.MountPath) == runnerVolumeMountPath {
					return nil
				}
			}
		}
	}

	runnerContainer := &pod.Spec.Containers[0]

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
			Name:  "RUNNER_NO_DEFAULT_LABELS",
			Value: "true",
		},
		{
			Name:  "DISABLE_RUNNER_UPDATE",
			Value: "true",
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
		{
			Name:  "JIT_CONFIG_ENABLED",
			Value: fmt.Sprintf("%t", bootstrapParams.JitConfigEnabled),
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
