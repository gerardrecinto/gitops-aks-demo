# Demo script

Run everything from the repo root with `KUBECONFIG` set by the Makefile. Keep three panes open:

```
make argocd-ui                                  # pane 1, browser on https://localhost:8080
watch -n2 kubectl -n argocd get applications    # pane 2
make app-ui                                     # pane 3, dev :8081, prod :8082
```

Argo CD polls Git every 60 seconds. To skip the wait:

```
kubectl -n argocd annotate app items-api-dev argocd.argoproj.io/refresh=hard --overwrite
```

## 1. Infrastructure provisioning

```
make preview
make up
make status
```

Expect a plan with a resource group, a managed cluster, a Helm release and a few Kubernetes objects. After `up`, about ten minutes later, `make status` lists `root`, `sealed-secrets`, `items-api-dev` and `items-api-prod`. All should reach `Synced` and `Healthy` once an image exists for the tag they point at.

## 2. Normal deployment

Change a name in `app/internal/api/api.go` (the `items` slice), then:

```
git checkout -b feature/rename-item
git commit -am "rename an item"
git push -u origin feature/rename-item
```

Open a PR. `ci` runs test and image jobs but does not push or deploy. Merge it. On main, `ci` pushes `sha-<commit>` and its last job commits `deploy dev: sha-<commit>` to `k8s/overlays/dev`. Within a minute:

```
curl -s localhost:8081/version
curl -s localhost:8081/api/items
```

`version` shows the new commit. Prod still shows the old one. Argo CD shows `items-api-dev` going `OutOfSync`, then `Synced`, and the Deployment rolling with no downtime.

## 3. Failed CI

On a branch, break a test, for example change the expected count in `TestItems` from 3 to 4, and open a PR.

Expect the `test` job to fail, `image` and `deploy-dev` to be skipped, and no new commit in `k8s/`. Nothing reaches the cluster. To show the scan gate instead, add a dependency with a known HIGH CVE; the Trivy step fails the `image` job.

## 4. Failed Kubernetes deployment

Point dev at a tag that does not exist:

```
sed -i.bak 's/newTag: .*/newTag: "sha-doesnotexist"/' k8s/overlays/dev/kustomization.yaml && rm k8s/overlays/dev/kustomization.yaml.bak
git commit -am "bad tag" && git push
```

Expect Argo to sync, the new pod to show `ErrImagePull` and then `ImagePullBackOff`, and the app to show `Progressing`, then `Degraded` after about a minute. The old pod keeps serving because the rollout uses `maxUnavailable: 0`, so `curl localhost:8081/health` still answers. Check with `kubectl -n demo-dev get pods`.

Recover with Git, not the cluster: `git revert HEAD && git push`.

## 5. Rollback to a previous revision

```
git log --oneline -- k8s/overlays/dev
git revert <the deploy commit you want to undo>
git push
```

Expect Argo to sync the older tag and `/version` to show the previous commit. History is in Git and in the Argo CD app's History tab. The Rollback button is disabled while auto sync is on, which is deliberate: a rollback that lives in the cluster and not in Git is drift.

## 6. Promotion from dev to prod

Actions, `promote`, Run workflow, leave the tag empty. The run pauses at `production` waiting for a reviewer. Approve it.

Expect a commit `deploy prod: sha-...` to `k8s/overlays/prod`, `items-api-prod` syncing, and `curl localhost:8082/version` matching dev. If you pass a tag that is not in GHCR, the job fails before it writes anything.

## 7. Drift detected

Self healing hides drift in seconds, so turn it off through Git, which is itself the point. In `argocd/applicationset.yaml` set `selfHeal: false` and push. Wait for `root` to sync, then:

```
kubectl -n demo-dev scale deploy items-api --replicas=5
kubectl -n demo-dev get pods
```

Expect five pods and `items-api-dev` showing `OutOfSync`. The UI diff view shows `replicas: 1` in Git against `5` live. Argo reports it and leaves it alone.

## 8. Self healing

Set `selfHeal` back to `true`, push, and wait for `root` to sync. `items-api-dev` syncs and the Deployment returns to one replica. Then repeat the drift:

```
kubectl -n demo-dev scale deploy items-api --replicas=5
kubectl -n demo-dev get pods -w
```

Expect the extra pods to disappear within seconds, with no human action. `kubectl -n demo-dev delete deploy items-api` behaves the same way: Argo recreates it.

## 9. Pruning

Add `k8s/base/extra-configmap.yaml` with any ConfigMap, list it under `resources` in `k8s/base/kustomization.yaml`, push, and see it appear in both namespaces. Delete it from Git and push again. Expect the ConfigMap to be removed from the cluster. `prune: true` is why Git deletions become cluster deletions.

## 10. Secrets

```
printf '%s' 'demo-value' | scripts/seal-secret.sh dev API_KEY
git add k8s/overlays/dev && git commit -m "seal dev api key" && git push
curl -s localhost:8081/version
```

Pods read Secrets at start, so run `kubectl -n demo-dev delete pod -l app.kubernetes.io/name=items-api` and expect `"api_key_configured": true` from the new pod. The value never appears in Git, in the Argo UI or in `/version`. `kubectl -n demo-dev get secret items-api-secrets` shows the decrypted Secret, created by the controller from the `SealedSecret`.

## 11. Infrastructure destruction

```
make down
az group list -o table
```

Expect `rg-gitops-aks-demo` to be gone. Billing stops. `make up` brings the same environment back, and the committed SealedSecrets still decrypt because the key pair lives in Pulumi config.
