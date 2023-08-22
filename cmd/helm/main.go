package main

import (
	"context"
	goflag "flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	utilflag "k8s.io/component-base/cli/flag"
	logs "k8s.io/component-base/logs/api/v1"
	"k8s.io/klog/v2"
	addonv1alpha1client "open-cluster-management.io/api/client/addon/clientset/versioned"

	"github.com/ViaQ/logging-ocm-addon/pkg/logging_helm"

	"github.com/openshift/cluster-logging-operator/apis"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	cmdfactory "open-cluster-management.io/addon-framework/pkg/cmd/factory"
	"open-cluster-management.io/addon-framework/pkg/utils"
	"open-cluster-management.io/addon-framework/pkg/version"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	logs.AddFlags(logs.NewLoggingConfiguration(), pflag.CommandLine)

	command := newCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addon",
		Short: "logging omc addon - helm version",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			os.Exit(1)
		},
	}

	if v := version.Get().String(); len(v) == 0 {
		cmd.Version = "<unknown>"
	} else {
		cmd.Version = v
	}

	cmd.AddCommand(newControllerCommand())
	return cmd
}

func newControllerCommand() *cobra.Command {
	cmd := cmdfactory.
		NewControllerCommandConfig("logging-addon-helm-controller", version.Get(), runController).
		NewCommand()
	cmd.Use = "controller"
	cmd.Short = "Start the addon helm controller"

	return cmd
}

func runController(ctx context.Context, kubeConfig *rest.Config) error {
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	addonClient, err := addonv1alpha1client.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	mgr, err := addonmanager.New(kubeConfig)
	if err != nil {
		klog.Errorf("failed to new addon manager %v", err)
		return err
	}

	registrationOption := logging_helm.NewRegistrationOption(
		kubeConfig,
		logging_helm.AddonName,
		utilrand.String(5))

	err = apis.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	agentAddon, err := addonfactory.NewAgentAddonFactory(logging_helm.AddonName, logging_helm.FS, "manifests/charts/logging-omc-addon").
		WithConfigGVRs(
			schema.GroupVersionResource{Version: "v1", Resource: "configmaps"},
			schema.GroupVersionResource{Version: "v1", Group: "loki.grafana.com", Resource: "lokistacks"},
			utils.AddOnDeploymentConfigGVR,
		).
		WithGetValuesFuncs(
			logging_helm.GetValues(
				addonfactory.GetAddOnDeploymentConfigValues(
					addonfactory.NewAddOnDeploymentConfigGetter(addonClient),
					addonfactory.ToAddOnCustomizedVariableValues,
				),
				logging_helm.GetMTLSSecretValues(kubeClient),
				logging_helm.GetCABundleValues(kubeClient),
			),
		).WithAgentRegistrationOption(registrationOption).
		WithScheme(scheme.Scheme).
		BuildHelmAgentAddon()
	if err != nil {
		klog.Errorf("failed to build agent %v", err)
		return err
	}

	err = mgr.AddAgent(agentAddon)
	if err != nil {
		klog.Fatal(err)
	}

	err = mgr.Start(ctx)
	if err != nil {
		klog.Fatal(err)
	}
	<-ctx.Done()

	return nil
}
