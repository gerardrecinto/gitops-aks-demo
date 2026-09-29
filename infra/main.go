package main

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		s := loadSettings(ctx)

		c, err := newCluster(ctx, s)
		if err != nil {
			return err
		}
		if err := bootstrapGitOps(ctx, s, c.kubeconfig); err != nil {
			return err
		}

		ctx.Export("resourceGroup", c.resourceGroup.Name)
		ctx.Export("clusterName", c.aks.Name)
		ctx.Export("kubeconfig", c.kubeconfig)
		return nil
	})
}
