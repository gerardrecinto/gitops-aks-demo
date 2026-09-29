package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// rootApp is the only Argo CD object Pulumi applies. Argo discovers the rest from Git.
const rootApp = "../argocd/root.yaml"

var argoValues = pulumi.Map{
	"dex":           pulumi.Map{"enabled": pulumi.Bool(false)},
	"notifications": pulumi.Map{"enabled": pulumi.Bool(false)},
	"configs": pulumi.Map{
		// Poll Git every minute so the demo reacts fast. The default is 3 minutes.
		"cm": pulumi.Map{"timeout.reconciliation": pulumi.String("60s")},
	},
}

// bootstrapGitOps installs Argo CD, restores the sealing key, and applies the root app.
// After this returns, Git is the source of truth and Pulumi is out of the deploy path.
func bootstrapGitOps(ctx *pulumi.Context, s settings, kubeconfig pulumi.StringOutput) error {
	provider, err := kubernetes.NewProvider(ctx, "aks", &kubernetes.ProviderArgs{
		Kubeconfig: kubeconfig,
	})
	if err != nil {
		return err
	}
	withProvider := pulumi.Provider(provider)

	// Restoring the same key on every rebuild keeps committed SealedSecrets decryptable.
	sealingKey, err := corev1.NewSecret(ctx, "sealing-key", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("sealed-secrets-key-bootstrap"),
			Namespace: pulumi.String("kube-system"),
			Labels: pulumi.StringMap{
				"sealedsecrets.bitnami.com/sealed-secrets-key": pulumi.String("active"),
			},
		},
		Type: pulumi.String("kubernetes.io/tls"),
		StringData: pulumi.StringMap{
			"tls.crt": s.sealingCert,
			"tls.key": s.sealingKey,
		},
	}, withProvider)
	if err != nil {
		return err
	}

	argocd, err := helmv3.NewRelease(ctx, "argocd", &helmv3.ReleaseArgs{
		Chart:           pulumi.String("argo-cd"),
		Version:         pulumi.String(s.argocdChartVersion),
		Namespace:       pulumi.String("argocd"),
		CreateNamespace: pulumi.Bool(true),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String("https://argoproj.github.io/argo-helm"),
		},
		Values: argoValues,
	}, withProvider)
	if err != nil {
		return err
	}

	_, err = yamlv2.NewConfigFile(ctx, "root-app", &yamlv2.ConfigFileArgs{
		File: pulumi.String(rootApp),
	}, withProvider, pulumi.DependsOn([]pulumi.Resource{argocd, sealingKey}))
	return err
}
