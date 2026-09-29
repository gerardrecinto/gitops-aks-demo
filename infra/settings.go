package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const name = "gitops-aks-demo"

// settings is read once in main and passed down, so no other file touches config.
type settings struct {
	location           string
	nodeSize           string
	nodeCount          int
	argocdChartVersion string
	sealingCert        pulumi.StringOutput
	sealingKey         pulumi.StringOutput
	tags               pulumi.StringMap
}

func loadSettings(ctx *pulumi.Context) settings {
	cfg := config.New(ctx, "")
	return settings{
		location:           cfg.Get("location"),
		nodeSize:           cfg.Get("nodeSize"),
		nodeCount:          cfg.GetInt("nodeCount"),
		argocdChartVersion: cfg.Get("argocdChartVersion"),
		sealingCert:        cfg.RequireSecret("sealingCert"),
		sealingKey:         cfg.RequireSecret("sealingKey"),
		tags: pulumi.StringMap{
			"project":    pulumi.String(name),
			"managed-by": pulumi.String("pulumi"),
			"stack":      pulumi.String(ctx.Stack()),
		},
	}
}
