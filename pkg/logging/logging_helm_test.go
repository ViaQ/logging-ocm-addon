package logging

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakekube "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
	addonapiv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	fakeaddon "open-cluster-management.io/api/client/addon/clientset/versioned/fake"
	clusterv1 "open-cluster-management.io/api/cluster/v1"

	lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
	loggingapis "github.com/openshift/cluster-logging-operator/apis"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/scheme"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager/addontesting"
	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"
)

func TestManifestAddonAgent(t *testing.T) {
	cases := []struct {
		name                      string
		managedCluster            *clusterv1.ManagedCluster
		managedClusterAddOn       *addonapiv1alpha1.ManagedClusterAddOn
		configMaps                []runtime.Object
		secrets                   []runtime.Object
		lokistacks                []runtime.Object
		addOnDeploymentConfigs    []runtime.Object
		verifyClusterLogForwarder func(t *testing.T, objs []runtime.Object)
		verifyMTLSSecret          func(t *testing.T, objs []runtime.Object)
	}{
		{
			name:           "no configs",
			managedCluster: addontesting.NewManagedCluster("cluster1"),
			managedClusterAddOn: func() *addonapiv1alpha1.ManagedClusterAddOn {
				addon := addontesting.NewAddon("test", "cluster1")
				addon.Status.ConfigReferences = []addonapiv1alpha1.ConfigReference{
					{
						ConfigGroupResource: addonapiv1alpha1.ConfigGroupResource{
							Group:    "loki.grafana.com",
							Resource: "lokistacks",
						},
						ConfigReferent: addonapiv1alpha1.ConfigReferent{
							Namespace: "openshift-monitoring",
							Name:      "loki",
						},
					},
					{
						ConfigGroupResource: addonapiv1alpha1.ConfigGroupResource{
							Group:    "",
							Resource: "secrets",
						},
						ConfigReferent: addonapiv1alpha1.ConfigReferent{
							Namespace: "cluster1",
							Name:      "mTLS",
						},
					},
					{
						ConfigGroupResource: addonapiv1alpha1.ConfigGroupResource{
							Group:    "addon.open-cluster-management.io",
							Resource: "addondeploymentconfigs",
						},
						ConfigReferent: addonapiv1alpha1.ConfigReferent{
							Namespace: "cluster1",
							Name:      "loki-addon-config",
						},
					},
				}
				return addon
			}(),
			configMaps: []runtime.Object{&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loki-gateway-ca-bundle",
					Namespace: "openshift-monitoring",
				},
				Data: map[string]string{
					"service-ca.crt": "hubCA",
				},
			}},
			secrets: []runtime.Object{&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mTLS",
					Namespace: "cluster1",
				},
				Data: map[string][]byte{
					"tls.key":        []byte("spokeKey"),
					"tls.crt":        []byte("spokeCRT"),
					"service-ca.crt": []byte("spokeCA"),
				},
			}},
			lokistacks: []runtime.Object{&lokiv1.LokiStack{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loki",
					Namespace: "openshift-logging",
				},
			}},
			addOnDeploymentConfigs: []runtime.Object{&addonapiv1alpha1.AddOnDeploymentConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loki-addon-config",
					Namespace: "cluster1",
				},
				Spec: addonapiv1alpha1.AddOnDeploymentConfigSpec{
					CustomizedVariables: []addonapiv1alpha1.CustomizedVariable{
						{
							Name:  "lokiURL",
							Value: "myTenantURL",
						},
					},
				},
			}},
			verifyClusterLogForwarder: func(t *testing.T, objs []runtime.Object) {
				clusterLogForwarder := findClusterLogForwarder(objs)
				assert.NotNil(t, clusterLogForwarder)
				assert.Equal(t, "instance", clusterLogForwarder.Name)
				assert.Equal(t, "openshift-logging", clusterLogForwarder.Namespace)
				assert.Equal(t, "mtls-spoke-hub", clusterLogForwarder.Spec.Outputs[0].Secret.Name)
				assert.Equal(t, "myTenantURL", clusterLogForwarder.Spec.Outputs[0].URL)
			},
			verifyMTLSSecret: func(t *testing.T, objs []runtime.Object) {
				mTLSSecret := findSecret(objs)
				assert.NotNil(t, mTLSSecret)
				assert.Equal(t, "mtls-spoke-hub", mTLSSecret.Name)
				assert.Equal(t, "spokeKey", string(mTLSSecret.Data["tls.key"]))
				assert.Equal(t, "spokeCRT", string(mTLSSecret.Data["tls.crt"]))
				assert.Equal(t, "hubCA", string(mTLSSecret.Data["ca-bundle.crt"]))
			},
		},
	}

	for _, c := range cases {
		fakeKubeClient := fakekube.NewSimpleClientset(append(c.configMaps, c.secrets...)...)
		fakeAddonClient := fakeaddon.NewSimpleClientset(c.addOnDeploymentConfigs...)

		err := loggingapis.AddToScheme(scheme.Scheme)
		assert.NoError(t, err)

		err = operatorsv1.AddToScheme(scheme.Scheme)
		assert.NoError(t, err)

		err = operatorsv1alpha1.AddToScheme(scheme.Scheme)
		assert.NoError(t, err)

		agentAddon, err := addonfactory.NewAgentAddonFactory(AddonName, FS, "manifests/charts/logging-ocm-addon").
			WithConfigGVRs(
				schema.GroupVersionResource{Version: "v1", Resource: "configmaps"},
				schema.GroupVersionResource{Version: "v1", Group: "loki.grafana.com", Resource: "lokistacks"},
				utils.AddOnDeploymentConfigGVR,
			).
			WithGetValuesFuncs(
				GetValues(
					addonfactory.GetAddOnDeploymentConfigValues(
						addonfactory.NewAddOnDeploymentConfigGetter(fakeAddonClient),
						addonfactory.ToAddOnCustomizedVariableValues,
					),
					GetMTLSSecretValues(fakeKubeClient),
					GetCABundleValues(fakeKubeClient),
				),
			).WithAgentRegistrationOption(&agent.RegistrationOption{}).
			WithScheme(scheme.Scheme).
			BuildHelmAgentAddon()
		if err != nil {
			klog.Fatalf("failed to build agent %v", err)
		}

		objects, err := agentAddon.Manifests(c.managedCluster, c.managedClusterAddOn)
		require.NoError(t, err)
		require.Equal(t, 8, len(objects))

		c.verifyClusterLogForwarder(t, objects)
		c.verifyMTLSSecret(t, objects)
	}
}

func findClusterLogForwarder(objs []runtime.Object) *loggingv1.ClusterLogForwarder {
	for _, obj := range objs {
		switch obj := obj.(type) {
		case *loggingv1.ClusterLogForwarder:
			return obj
		}
	}

	return nil
}
func findSecret(objs []runtime.Object) *corev1.Secret {
	for _, obj := range objs {
		switch obj := obj.(type) {
		case *corev1.Secret:
			return obj
		}
	}

	return nil
}
