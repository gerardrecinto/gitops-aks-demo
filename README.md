# gitops-aks-demo

A small GitOps setup on Azure Kubernetes Service. GitHub Actions does CI and writes the desired state to Git. Argo CD reads Git and makes the cluster match. Nothing in CI ever runs `kubectl`.

The app is a tiny Go API (`/health`, `/version`, `/api/items`). The point of the repo is everything around it.

```
 developer
    |  git push
    v
 GitHub repo ------------------------------------------.
    |                                                    |
    |  push to main                                      |  Argo CD polls
    v                                                    |  every 60s
 GitHub Actions (ci.yml)                                 |
    1. gofmt, vet, go test -race                         |
    2. build image, Trivy scan (HIGH, CRITICAL fail)     |
    3. push ghcr.io/<repo>:sha-<commit>                  |
    4. commit new tag to k8s/overlays/dev  -------------'
                                                         |
 promote.yml (manual, needs approval)                    |
    commits same tag to k8s/overlays/prod -------------->|
                                                         v
                                              Argo CD (in AKS)
                                               |            |
                                       items-api-dev   items-api-prod
                                       ns demo-dev     ns demo-prod
```

CI and CD are separate on purpose. CI has write access to Git and the registry. It has no credentials for the cluster. Argo CD has read access to Git and runs inside the cluster. A stolen CI token cannot touch the cluster directly.

## Demo

A full rehearsal on a live cluster: deploy, failed CI, bad tag and rollback, promotion, drift, self-heal, prune, sealed secret.

![demo](docs/assets/demo.gif)

Teardown: `make down` removes everything ([recording](docs/assets/teardown.gif)).

## Layout

```
app/              Go service, tests, Dockerfile
k8s/base/         Deployment, Service, ServiceAccount
k8s/overlays/     dev and prod: namespace, replicas, image tag, config
argocd/           Project, ApplicationSet (dev + prod), Sealed Secrets, root app
infra/            Pulumi program in Go: resource group, AKS, Argo CD bootstrap
scripts/          sealing key, seal a secret, GitHub to Azure OIDC setup
.github/          ci, promote, manifests, infra-preview, dependabot
docs/             cost, demo script, troubleshooting, interview walkthrough
```

Pulumi installs Argo CD and one `root` Application. From then on Git owns the cluster: `argocd/` describes the project, both environments and Sealed Secrets, and `k8s/overlays/*` describes the workload.

## Decisions worth knowing

| Choice | Why |
|---|---|
| One AKS cluster, two namespaces | A second cluster roughly doubles the bill. The cost is a shared node and shared blast radius, which I would not accept for real prod. |
| Free AKS tier, one D2as_v6 node | No control plane charge. The node is the main cost. |
| GHCR, not ACR | Free, and images survive a `make down`. A public package needs no pull secret. |
| Image tag is `sha-<full commit>` | Immutable and traceable. `/version` reports the same value. |
| dev deploys on every merge, prod on approval | `promote.yml` waits on the `production` environment reviewers, then commits the tag. Prod only changes when someone approves. |
| Sealed Secrets | Free, and encrypted values live in Git. The key pair is kept in Pulumi config so a rebuilt cluster can still decrypt. |
| Pulumi in Go | Same language as the app, real types, ordinary functions for reuse. |
| No ingress, no Log Analytics | Each adds an hourly or per-GB charge. Port-forward and `kubectl logs` are enough for a demo. |

## Prerequisites

`az`, `pulumi`, `kubectl`, `kubeseal`, `gh`, `go` 1.25+, and an Azure subscription. A public GitHub repo named `gerardrecinto/gitops-aks-demo` (the repo URL is written into `argocd/*.yaml` and `k8s/base/deployment.yaml`, so search and replace it if you use another name).

## Deploy

1. Push this repo to GitHub. The first CI run builds the image and points `dev` at it.
2. In the GitHub package settings for the image, make it public once.
3. In repo settings, create an environment called `production` with yourself as required reviewer.
4. Provision:

```
az login
cd infra && pulumi login && pulumi stack init demo && cd ..
scripts/gen-sealing-key.sh
make preview
make up
```

`make up` creates the cluster, installs Argo CD, restores the sealing key, and applies the root app. Argo then creates both namespaces and syncs dev.

5. Look at it:

```
make status            # applications, sync and health
make argocd-password
make argocd-ui         # https://localhost:8080, user admin
make app-ui            # dev on :8081, prod on :8082
curl localhost:8081/version
```

6. Promote to prod: Actions, `promote`, Run workflow, approve the deployment.

Optional: `scripts/setup-oidc.sh` and a `PULUMI_ACCESS_TOKEN` secret enable the `infra-preview` workflow, which runs `pulumi preview` on PRs using a GitHub OIDC token. No Azure secret is stored.

## Tear down

```
make down
```

That runs `pulumi destroy`, which removes the resource group and everything in it. Check with `az group list -o table`. The Pulumi state, the GHCR images and this repo remain, so `make up` rebuilds the same environment.

Details: [cost](docs/cost.md), [demo script](docs/demo-script.md), [troubleshooting](docs/troubleshooting.md), [interview walkthrough](docs/interview-walkthrough.md).

## Related

[argocd-gitops](https://github.com/gerardrecinto/argocd-gitops) holds the wider Argo CD patterns: multi-cluster ApplicationSets, RBAC and preview environments, with placeholder values. This repo is the small runnable version on one AKS cluster.
