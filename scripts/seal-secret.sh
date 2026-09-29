#!/usr/bin/env bash
# Usage: printf '%s' "$VALUE" | scripts/seal-secret.sh <dev|prod> <KEY>
# Encrypts one value against the public sealing cert and writes it into the overlay.
# The encrypted file is safe to commit. Only the cluster's key can decrypt it.
set -euo pipefail
env=${1:?usage: seal-secret.sh <dev|prod> <KEY>, value on stdin}
key=${2:?usage: seal-secret.sh <dev|prod> <KEY>, value on stdin}
root=$(cd "$(dirname "$0")/.." && pwd)
overlay="$root/k8s/overlays/$env"

cert=$(mktemp)
trap 'rm -f "$cert"' EXIT
(cd "$root/infra" && pulumi config get sealingCert) > "$cert"

kubectl create secret generic items-api-secrets --namespace "demo-$env" \
  --from-file="$key=/dev/stdin" --dry-run=client -o yaml \
  | kubeseal --cert "$cert" --format yaml > "$overlay/sealed-secret.yaml"

if ! grep -q sealed-secret.yaml "$overlay/kustomization.yaml"; then
  awk '{print} /- \.\.\/\.\.\/base/ {print "  - sealed-secret.yaml"}' \
    "$overlay/kustomization.yaml" > "$overlay/kustomization.tmp"
  mv "$overlay/kustomization.tmp" "$overlay/kustomization.yaml"
fi
echo "wrote $overlay/sealed-secret.yaml, commit it and Argo CD applies it"
