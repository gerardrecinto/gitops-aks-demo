import { PolicyPack, ReportViolation, ResourceValidationArgs } from "@pulumi/policy";

const aksType = "azure-native:containerservice:ManagedCluster";

// The URN is urn:pulumi:<stack>::<project>::<type>::<name>, so the stack name
// is the second colon-separated field.
const stackOf = (urn: string): string => urn.split("::")[0].split(":")[2];

new PolicyPack("aks-guardrails", {
    policies: [
        {
            name: "prod-aks-needs-standard-tier",
            description: "An AKS cluster in the prod stack must use the Standard tier, which carries the uptime SLA.",
            enforcementLevel: "mandatory",
            validateResource: (args: ResourceValidationArgs, reportViolation: ReportViolation) => {
                if (args.type !== aksType || stackOf(args.urn) !== "prod") {
                    return;
                }
                const tier = args.props.sku?.tier;
                if (tier !== "Standard" && tier !== "Premium") {
                    reportViolation(`prod cluster uses tier "${tier}", it must be Standard or Premium`);
                }
            },
        },
    ],
});
