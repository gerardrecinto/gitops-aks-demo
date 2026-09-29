package main

import (
	"encoding/base64"

	containerservice "github.com/pulumi/pulumi-azure-native-sdk/containerservice/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type cluster struct {
	resourceGroup *resources.ResourceGroup
	aks           *containerservice.ManagedCluster
	kubeconfig    pulumi.StringOutput
}

func newCluster(ctx *pulumi.Context, s settings) (*cluster, error) {
	rg, err := resources.NewResourceGroup(ctx, "rg", &resources.ResourceGroupArgs{
		ResourceGroupName: pulumi.String("rg-" + name),
		Location:          pulumi.String(s.location),
		Tags:              s.tags,
	})
	if err != nil {
		return nil, err
	}

	aks, err := containerservice.NewManagedCluster(ctx, "aks", &containerservice.ManagedClusterArgs{
		ResourceName:      pulumi.String("aks-" + name),
		ResourceGroupName: rg.Name,
		Location:          rg.Location,
		DnsPrefix:         pulumi.String(name),
		Tags:              s.tags,
		// Free tier: no control plane charge and no uptime SLA. Fine for a demo.
		Sku: &containerservice.ManagedClusterSKUArgs{
			Name: pulumi.String("Base"),
			Tier: pulumi.String("Free"),
		},
		Identity: &containerservice.ManagedClusterIdentityArgs{
			Type: containerservice.ResourceIdentityTypeSystemAssigned,
		},
		OidcIssuerProfile: &containerservice.ManagedClusterOIDCIssuerProfileArgs{
			Enabled: pulumi.Bool(true),
		},
		NetworkProfile: &containerservice.ContainerServiceNetworkProfileArgs{
			NetworkPlugin:     pulumi.String("azure"),
			NetworkPluginMode: pulumi.String("overlay"),
			LoadBalancerSku:   pulumi.String("standard"),
		},
		AgentPoolProfiles: containerservice.ManagedClusterAgentPoolProfileArray{
			&containerservice.ManagedClusterAgentPoolProfileArgs{
				Name:         pulumi.String("system"),
				Mode:         pulumi.String("System"),
				Count:        pulumi.Int(s.nodeCount),
				VmSize:       pulumi.String(s.nodeSize),
				OsType:       pulumi.String("Linux"),
				OsDiskSizeGB: pulumi.Int(32),
				OsDiskType:   pulumi.String("Managed"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	creds := containerservice.ListManagedClusterUserCredentialsOutput(ctx,
		containerservice.ListManagedClusterUserCredentialsOutputArgs{
			ResourceGroupName: rg.Name,
			ResourceName:      aks.Name,
		})
	kubeconfig := creds.Kubeconfigs().Index(pulumi.Int(0)).Value().ApplyT(
		func(b string) (string, error) {
			raw, err := base64.StdEncoding.DecodeString(b)
			return string(raw), err
		}).(pulumi.StringOutput)

	return &cluster{rg, aks, pulumi.ToSecret(kubeconfig).(pulumi.StringOutput)}, nil
}
