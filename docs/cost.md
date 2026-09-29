# Cost

These are rough East US pay-as-you-go numbers from memory. Check them in the Azure pricing calculator before quoting them.

## Recurring Azure charges while the cluster exists

| Item | Approx. per hour | Notes |
|---|---|---|
| AKS control plane | $0 | Free tier, no uptime SLA |
| 1 x Standard_B2ms node | $0.08 | The main cost. Two vCPU, 8 GB, enough for Argo CD plus both namespaces |
| Standard load balancer | $0.03 | AKS creates it for outbound traffic |
| Public IP | $0.005 | Attached to that load balancer |
| 32 GB managed OS disk | under $0.01 | |

Total is about $0.12 an hour. A two hour demo is around $0.25. Left running all month it is about $85, which is the reason `make down` exists.

## What costs nothing

GitHub Actions on public repos, GHCR public images, Pulumi Cloud for one person, GitHub OIDC, Trivy, Argo CD, Sealed Secrets and Dependabot.

## What I left out, and the price of adding it

| Component | Why it is out |
|---|---|
| Azure Container Registry Basic | About $5 a month, and images vanish with the cluster unless it is kept outside the stack |
| Second AKS cluster | Doubles the node, load balancer and IP cost |
| Log Analytics, Container Insights, managed Prometheus | Billed per GB ingested and easy to leave on |
| Azure Key Vault | Cheap per operation, but needs Workload Identity and External Secrets. Complexity, not dollars |
| Ingress controller with its own public IP | One more hourly charge for a demo that works with port-forward |

## Keeping it low

- Destroy after every session.
- Set a budget alert on the subscription at $10.
- `az aks stop` pauses node compute but the disk, load balancer and IP keep billing. Destroy is cheaper and cleaner.
