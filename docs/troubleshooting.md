# Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Pods in `ImagePullBackOff` right after `make up` | No image exists for the tag in the overlay, or the GHCR package is private | Let `ci` finish on main, then make the package public in GitHub package settings |
| `items-api-prod` degraded on first install | Prod still points at `sha-0000000` | Run the `promote` workflow once |
| `ci` cannot push the tag commit | Branch protection on main blocks `GITHUB_TOKEN` | Allow GitHub Actions to bypass the rule, or keep main unprotected in this repo and rely on the `production` environment for prod |
| `promote` sits at "Waiting" | The `production` environment needs an approval | Approve it in the Actions run |
| Application stays `Unknown` or `ComparisonError` | Repo URL in `argocd/*.yaml` does not match the real repo, or the repo is private | `kubectl -n argocd get app root -o yaml` and read `status.conditions`. Fix the URL |
| `SealedSecret` says it cannot decrypt | Sealed against a different cert than the cluster holds | Confirm `sealingCert` in Pulumi config is the one used with `seal-secret.sh`, then reseal |
| `items-api-dev` fails on a `SealedSecret` kind not found | CRD not installed yet | It retries. `SkipDryRunOnMissingResource` and the retry policy cover the first sync |
| `pulumi up` fails with a quota or SKU error | B2ms not available in the region or quota is zero | `pulumi config set nodeSize <2 vCPU, 8 GB size that your region offers>` or change `location` |
| `pulumi destroy` hangs on Kubernetes resources | The cluster is unreachable | `az group delete -n rg-gitops-aks-demo --yes`, then `pulumi refresh --yes` |
| `infra-preview` fails to log in | Missing OIDC variables | Run `scripts/setup-oidc.sh` and add `PULUMI_ACCESS_TOKEN` |
| Trivy fails the `image` job | A HIGH or CRITICAL finding with a fix | Bump the base image or the dependency. Only add a `.trivyignore` entry with a written reason |
| UI change not visible for a minute | Argo polls every 60 seconds | Hard refresh with the annotation shown in the demo script |

Useful commands:

```
kubectl -n argocd get applications
kubectl -n argocd describe application items-api-dev
kubectl -n demo-dev get events --sort-by=.lastTimestamp
kubectl -n demo-dev rollout status deploy/items-api
kubectl -n demo-dev logs deploy/items-api
kubectl top pods -A
```
