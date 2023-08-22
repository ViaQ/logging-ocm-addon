package logging_helm

import (
	"context"
	"embed"
	"fmt"

	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"
	addonapiv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

//go:embed manifests
//go:embed manifests/charts/logging-omc-addon
//go:embed manifests/charts/logging-omc-addon/templates/_helpers.tpl
var FS embed.FS

const (
	AddonName             = "logginghelm"
	InstallationNamespace = "default"
)

type userValues struct {
	MTLS mTLS `json:"mTLS"`
}

type mTLS struct {
	Key      string `json:"key"`
	Cert     string `json:"cert"`
	CABundle string `json:"caBundle"`
}

func NewRegistrationOption(kubeConfig *rest.Config, addonName, agentName string) *agent.RegistrationOption {
	return &agent.RegistrationOption{
		CSRConfigurations: agent.KubeClientSignerConfigurations(addonName, agentName),
		CSRApproveCheck:   utils.DefaultCSRApprover(agentName),
		PermissionConfig:  AddonRBAC(kubeConfig),
		Namespace:         InstallationNamespace,
	}
}

func GetCABundleValues(kubeClient kubernetes.Interface) addonfactory.GetValuesFunc {
	return func(
		cluster *clusterv1.ManagedCluster,
		addon *addonapiv1alpha1.ManagedClusterAddOn,
	) (addonfactory.Values, error) {
		overrideValues := addonfactory.Values{}
		for _, config := range addon.Status.ConfigReferences {
			if config.ConfigGroupResource.Group != "loki.grafana.com" ||
				config.ConfigGroupResource.Resource != "lokistacks" {
				continue
			}

			caBundleName := config.Name + "-ca-bundle"
			configMap, err := kubeClient.CoreV1().ConfigMaps(config.Namespace).Get(context.Background(), caBundleName, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			caBundle, ok := configMap.Data["service-ca.crt"]
			if !ok {
				return nil, fmt.Errorf("no service-ca.crt in configmap %s/%s", config.Namespace, config.Name)
			}

			userJsonValues := userValues{
				MTLS: mTLS{
					CABundle: caBundle,
				},
			}
			values, err := addonfactory.JsonStructToValues(userJsonValues)
			if err != nil {
				return nil, err
			}
			overrideValues = addonfactory.MergeValues(overrideValues, values)
		}

		return overrideValues, nil
	}
}

func GetMTLSSecretValues(kubeClient kubernetes.Interface) addonfactory.GetValuesFunc {
	return func(
		cluster *clusterv1.ManagedCluster,
		addon *addonapiv1alpha1.ManagedClusterAddOn,
	) (addonfactory.Values, error) {
		overrideValues := addonfactory.Values{}
		for _, config := range addon.Status.ConfigReferences {
			if config.ConfigGroupResource.Group != "" ||
				config.ConfigGroupResource.Resource != "secrets" {
				continue
			}

			secret, err := kubeClient.CoreV1().Secrets(config.Namespace).Get(context.Background(), config.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}

			key, ok := secret.Data["tls.key"]
			if !ok {
				return nil, fmt.Errorf("no tls.key in secret %s/%s", config.Namespace, config.Name)
			}

			cert, ok := secret.Data["tls.crt"]
			if !ok {
				return nil, fmt.Errorf("no tls.crt in secret %s/%s", config.Namespace, config.Name)
			}

			userJsonValues := userValues{
				MTLS: mTLS{
					Key:  string(key),
					Cert: string(cert),
				},
			}
			values, err := addonfactory.JsonStructToValues(userJsonValues)
			if err != nil {
				return nil, err
			}
			overrideValues = addonfactory.MergeValues(overrideValues, values)
		}

		return overrideValues, nil
	}
}

func GetValues(getValuesFuncs ...addonfactory.GetValuesFunc) addonfactory.GetValuesFunc {
	return func(
		cluster *clusterv1.ManagedCluster,
		addon *addonapiv1alpha1.ManagedClusterAddOn,
	) (addonfactory.Values, error) {
		overrideValues := addonfactory.Values{}
		for _, function := range getValuesFuncs {
			values, err := function(cluster, addon)
			if err != nil {
				return addonfactory.Values{}, err
			}
			err = mergo.Merge(&overrideValues, values)
			if err != nil {
				return addonfactory.Values{}, err
			}
		}

		return overrideValues, nil
	}
}
