package runtime

import (
	"fmt"
	"strings"

	"github.com/automationbroker/bundle-lib/clients"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BundleContainerName - name of the container in the pod defintion
	// Used for the default runtime.
	BundleContainerName = "apb"
	httpProxyEnvVar     = "HTTP_PROXY"
	httpsProxyEnvVar    = "HTTPS_PROXY"
	noProxyEnvVar       = "NO_PROXY"
)

// ProxyConfig - Contains a desired proxy configuration for the broker and
// the assets that it spawns
type ProxyConfig struct {
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string
}

// ExecutionContext - Contains the information necessary to track and clean up
// an APB run
type ExecutionContext struct {
	BundleName string
	// In k8s location is the namespace that the pod is running in
	Location string
	// Account/user that the bundle is running as
	Account     string
	Targets     []string
	Secrets     []string
	ExtraVars   string
	Image       string
	Action      string
	Policy      string
	ProxyConfig *ProxyConfig
	Metadata    map[string]string
	StateName   string
}

// RunBundleFunc - method that defines how to run a bundle
type RunBundleFunc func(ExecutionContext) (ExecutionContext, error)

// CopySecretsToNamespaceFunc - copy secrets to namespace
type CopySecretsToNamespaceFunc func(ec ExecutionContext, cn string, secrets []string) error

func defaultRunBundle(extContext ExecutionContext) (ExecutionContext, error) {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		return extContext, err
	}
	pullPolicy, err := checkPullPolicy(extContext.Policy)
	if err != nil {
		return extContext, err
	}
	volumes, volumeMounts := buildVolumeSpecs(extContext.Secrets, extContext.StateName)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   extContext.BundleName,
			Labels: extContext.Metadata,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  BundleContainerName,
					Image: extContext.Image,
					Args: []string{
						extContext.Action,
						"--extra-vars",
						extContext.ExtraVars,
					},
					Env:             createPodEnv(extContext),
					ImagePullPolicy: pullPolicy,
					VolumeMounts:    volumeMounts,
				},
			},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: extContext.Account,
			Volumes:            volumes,
		},
	}

	log.Infof(fmt.Sprintf("Creating pod %q in the %s namespace", pod.Name, extContext.Location))
	_, err = k8scli.Client.CoreV1().Pods(extContext.Location).Create(pod)

	return extContext, err
}

// Verify PullPolicy is acceptable
func checkPullPolicy(policy string) (v1.PullPolicy, error) {
	n := map[string]v1.PullPolicy{
		"always":       v1.PullAlways,
		"never":        v1.PullNever,
		"ifnotpresent": v1.PullIfNotPresent,
	}
	p := strings.ToLower(policy)
	value, _ := n[p]
	if value == "" {
		return "", fmt.Errorf("ImagePullPolicy: %s not found in [%s, %s, %s,]", policy, v1.PullAlways, v1.PullNever, v1.PullIfNotPresent)
	}

	return value, nil
}

func buildVolumeSpecs(secrets []string, stateName string) ([]v1.Volume, []v1.VolumeMount) {
	var optional bool
	var mountName string
	volumes := []v1.Volume{}
	volumeMounts := []v1.VolumeMount{}

	for _, secret := range secrets {
		mountName = "apb-" + secret
		volumes = append(volumes, v1.Volume{
			Name: mountName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
					Optional:   &optional,
					// Eventually, we can include: Items: []v1.KeyToPath here to specify specific keys in the secret
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      mountName,
			MountPath: "/etc/apb-secrets/" + mountName,
			ReadOnly:  true,
		})
	}
	if stateName != "" {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      stateName,
			MountPath: Provider.MountLocation(),
			ReadOnly:  true,
		})
		volumes = append(volumes, v1.Volume{
			Name: stateName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: stateName,
					},
				},
			},
		})
	}
	return volumes, volumeMounts
}

func createPodEnv(executionContext ExecutionContext) []v1.EnvVar {
	podEnv := []v1.EnvVar{
		v1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		v1.EnvVar{
			Name: "POD_NAMESPACE",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	if executionContext.ProxyConfig != nil {
		conf := executionContext.ProxyConfig

		log.Info("Proxy configuration present. Applying to APB before execution:")
		log.Infof("%s=\"%s\"", httpProxyEnvVar, conf.HTTPProxy)
		log.Infof("%s=\"%s\"", httpsProxyEnvVar, conf.HTTPSProxy)
		log.Infof("%s=\"%s\"", noProxyEnvVar, conf.NoProxy)

		podEnv = append(podEnv, []v1.EnvVar{
			v1.EnvVar{
				Name:  httpProxyEnvVar,
				Value: conf.HTTPProxy,
			},
			v1.EnvVar{
				Name:  httpsProxyEnvVar,
				Value: conf.HTTPSProxy,
			},
			v1.EnvVar{
				Name:  noProxyEnvVar,
				Value: conf.NoProxy,
			},
			v1.EnvVar{
				Name:  strings.ToLower(httpProxyEnvVar),
				Value: conf.HTTPProxy,
			},
			v1.EnvVar{
				Name:  strings.ToLower(httpsProxyEnvVar),
				Value: conf.HTTPSProxy,
			},
			v1.EnvVar{
				Name:  strings.ToLower(noProxyEnvVar),
				Value: conf.NoProxy,
			}}...)
	}

	return podEnv
}

// defaultCopySecretsToNamespace - copy secrets to namespace
func defaultCopySecretsToNamespace(ec ExecutionContext, cn string, secrets []string) error {
	k8scli, err := clients.Kubernetes()
	if err != nil {
		return err
	}
	for _, secrectName := range secrets {
		secretData, err := k8scli.Client.CoreV1().Secrets(cn).Get(secrectName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		oldMeta := secretData.ObjectMeta
		secretData.ObjectMeta = metav1.ObjectMeta{Name: oldMeta.Name, Namespace: ec.Location, Labels: oldMeta.Labels, Annotations: oldMeta.Annotations}
		_, err = k8scli.Client.CoreV1().Secrets(ec.Location).Create(secretData)
		if err != nil {
			return err
		}
	}
	return nil
}
