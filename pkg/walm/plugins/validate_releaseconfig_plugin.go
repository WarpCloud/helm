package plugins

import (
	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/walm"
	"bytes"
	"k8s.io/helm/pkg/tiller/environment"
	"k8s.io/helm/pkg/kube"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"fmt"
	"transwarp/release-config/pkg/apis/transwarp/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"github.com/sirupsen/logrus"
	"github.com/ghodss/yaml"
)

const (
	AutoGenLabelKey = "auto-gen"
)

// ValidateReleaseConfig plugin is used to make sure:
// 1. release have and only have one ReleaseConfig
// 2. ReleaseConfig has the same namespace and name with the release

func init() {
	walm.Register("ValidateReleaseConfig", &walm.WalmPluginRunner{
		Run:  ValidateReleaseConfig,
		Type: walm.Pre_Install,
	})
}

func ValidateReleaseConfig(context *walm.WalmPluginManagerContext, args string) (err error) {
	resources, err := buildResources(context.KubeClient, context.R)
	if err != nil {
		return err
	}
	releaseConfigResources := []*resource.Info{}
	newResources := []runtime.Object{}
	for _, resource := range resources {
		if resource.Object.GetObjectKind().GroupVersionKind().Kind == "ReleaseConfig" {
			if resource.Name != context.R.Name {
				continue
			}
			releaseConfigResources = append(releaseConfigResources, resource)
		} else {
			newResources = append(newResources, resource.Object)
		}
	}

	releaseConfigNum := len(releaseConfigResources)
	if releaseConfigNum == 0 {
		return fmt.Errorf("release must have one ReleaseConfig resource")
	} else if releaseConfigNum == 1 {
		return nil
	} else if releaseConfigNum == 2 {
		var autoGenReleaseConfig, releaseConfig *v1beta1.ReleaseConfig
		for _, releaseConfigResource := range releaseConfigResources {
			rc := releaseConfigResource.Object.(*v1beta1.ReleaseConfig)
			if len(rc.Labels) > 0 && rc.Labels[AutoGenLabelKey] == "true" {
				autoGenReleaseConfig = rc
			} else {
				releaseConfig = rc
			}
		}
		if autoGenReleaseConfig == nil {
			return fmt.Errorf("release can not have more than one ReleaseConfig resource")
		}
		if releaseConfig == nil {
			return fmt.Errorf("release can not have more than one auto gen ReleaseConfig resource")
		}
		releaseConfig.Spec.Dependencies = autoGenReleaseConfig.Spec.Dependencies
		releaseConfig.Spec.DependenciesConfigValues = autoGenReleaseConfig.Spec.DependenciesConfigValues
		releaseConfig.Spec.ConfigValues = autoGenReleaseConfig.Spec.ConfigValues
		releaseConfig.Spec.ChartName = autoGenReleaseConfig.Spec.ChartName
		releaseConfig.Spec.ChartVersion = autoGenReleaseConfig.Spec.ChartVersion
		releaseConfig.Spec.ChartAppVersion = autoGenReleaseConfig.Spec.ChartAppVersion
		if releaseConfig.Labels == nil {
			releaseConfig.Labels = map[string]string{}
		}
		for k, v := range autoGenReleaseConfig.Labels {
			releaseConfig.Labels[k] = v
		}

		newResources = append(newResources, releaseConfig)
		context.R.Manifest, err = buildManifest(newResources)
		if err != nil {
			return err
		}
	} else if releaseConfigNum > 2 {
		return fmt.Errorf("release can not have more than one ReleaseConfig resource")
	}

	return
}

func buildResources(kubeClient environment.KubeClient, r *release.Release) (kube.Result, error) {
	return kubeClient.BuildUnstructured(r.Namespace, bytes.NewBufferString(r.Manifest))
}

func buildManifest(resources []runtime.Object) (string,  error) {
	var sb strings.Builder
	for _, resource := range resources {
		resourceBytes, err := yaml.Marshal(resource)
		if err != nil {
			logrus.Errorf("failed to marshal k8s resource : %s", err.Error())
			return "", err
		}
		sb.WriteString("\n---\n")
		sb.Write(resourceBytes)
	}
	return sb.String(), nil
}
