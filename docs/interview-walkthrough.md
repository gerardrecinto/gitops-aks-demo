# Interview walkthrough

Short answers first. Each one ties back to something you can point at in this repo.

## 1. Why GitOps?

Git becomes the record of what should be running. Every change has an author, a diff, a review and a revert. The cluster is checked against Git continuously, so drift is visible and gets corrected. Here, `git log -- k8s/overlays/prod` is the full production deploy history.

## 2. Why Argo CD?

It has a clear model (Application, sync status, health status), a good UI for showing a team what is deployed, and it runs inside the cluster so nothing outside needs cluster credentials. Flux is a fine alternative and lighter on UI. I picked Argo because the diff view and app-of-apps pattern make the demo easy to follow.

## 3. Why GitHub Actions?

The code already lives in GitHub, public repos get free minutes, and OIDC gives short-lived cloud credentials without stored secrets. In a larger company I would look at the same pipeline in Jenkins or another CI. The design does not depend on the CI tool, only that it can build, scan, push, and commit.

## 4. Why AKS?

It is the managed Kubernetes on Azure, the control plane is free on the Free tier, and the OIDC issuer is available if I later want Workload Identity. For this project the platform matters less than the delivery model, so I chose the cheapest way to have a real cluster.

## 5. Why Pulumi?

The infra is written in Go, the same language as the app, so it gets types, tests and ordinary functions instead of a config language. `settings.go`, `cluster.go` and `gitops.go` each do one job. Terraform would work equally well; the important part is that the cluster comes from code and can be destroyed and rebuilt.

## 6. Why separate CI and CD?

CI proves a commit is good and produces an artifact. CD makes the environment match a declared state. Separating them means CI never holds cluster credentials, deploys are pull based, and a broken pipeline cannot half-apply a change. It also means the same deploy path handles a revert.

## 7. Why immutable image tags?

`latest` moves, so the same manifest can mean different things at different times and a rollback is guesswork. `sha-<commit>` maps to one build forever. `/version` returns the same value, so I can prove which commit is serving.

## 8. How does Argo CD detect changes?

It polls the repo, every 60 seconds here (the default is 3 minutes), renders the Kustomize overlay, and compares the result to live objects. A webhook from GitHub makes it near instant. Live changes are seen through Kubernetes watches, which is why drift shows up without waiting for the poll.

## 9. How does Argo CD self-heal?

With `selfHeal: true`, when the live state differs from Git, Argo re-applies Git. Scaling a Deployment by hand is undone in seconds. With `prune: true` it also deletes objects that were removed from Git.

## 10. How would this scale to multiple teams?

One AppProject per team to limit source repos, destination namespaces and allowed kinds. An ApplicationSet with a Git generator so a new folder becomes a new app. Separate clusters for prod and non-prod. A shared platform repo for add-ons and a config repo per team, or per service. CODEOWNERS on prod paths. Policy checks (Kyverno or Gatekeeper) in the cluster and in CI.

## 11. How would secrets be handled in a real enterprise?

| Option | Good | Cost |
|---|---|---|
| Plain Kubernetes Secret | Nothing to run | Base64 in Git is not encryption. Do not commit them |
| Sealed Secrets (used here) | Free, encrypted value in Git, simple | One key per cluster to back up, values are not shared across clusters unless you share the key |
| External Secrets with Azure Key Vault | Central rotation, audit, one source of truth | Key Vault, Workload Identity and an operator to run |

For a company I would use Key Vault with External Secrets and Workload Identity, so nothing secret lives in Git at all. I stayed with Sealed Secrets here because it shows the pattern for free.

## 12. How would I implement progressive delivery?

Argo Rollouts replaces the Deployment with a Rollout that shifts traffic in steps (canary or blue-green) and runs an analysis between steps against metrics such as error rate and latency. A failed analysis aborts and rolls back automatically. It needs an ingress or mesh that can split traffic and a metrics source.

## 13. How would I add observability?

Start with what is here: probes, `kubectl top`, Argo health, and the app's `/metrics`. Next step is Azure Managed Prometheus and Managed Grafana, or kube-prometheus-stack, scraping `/metrics`, with alerts on availability, error rate, pod restarts and Argo sync failures. Logs to Log Analytics or Loki. Argo CD notifications to chat for failed syncs. I left them out for cost, not because they are optional.

## 14. How would I secure the supply chain?

Already here: pinned action SHAs, Dependabot, Trivy on the image and the rendered manifests, provenance and SBOM attached to the image, a non-root distroless image with a read-only filesystem, and OIDC instead of stored cloud keys. To add: sign images with cosign and verify at admission (Kyverno or Ratify), pin the base image by digest, require signed commits, and use a private registry with a promotion step.

## 15. How would I reduce cloud costs further?

Destroy when not in use, which is the biggest saving. Use a smaller node if Argo CD is trimmed further. Skip the load balancer with a private cluster only if the demo never needs outbound access, which it does. Use spot nodes for user pools in a larger setup. Set a budget alert. In a real environment: right-size requests, autoscale to zero for non-prod, and schedule non-prod clusters to stop out of hours.

## Things to say about the tradeoffs

- Dev and prod share a cluster here. In production I would separate them at least by node pool, and more likely by cluster.
- The Free AKS tier has no uptime SLA.
- The CI job commits straight to main for dev. For prod, the promote workflow with a required reviewer is the approval gate. A stricter setup would open a PR instead.
- Sealed Secrets ties decryption to one key. Losing it means resealing everything, which is why the key is kept in Pulumi config.
