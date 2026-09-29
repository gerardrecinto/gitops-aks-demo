#!/usr/bin/env bash
# One time. Creates the Sealed Secrets key pair and stores it as Pulumi secrets so
# every rebuilt cluster gets the same key back and committed SealedSecrets keep working.
set -euo pipefail
cd "$(dirname "$0")/../infra"

dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
openssl req -x509 -days 3650 -nodes -newkey rsa:4096 \
  -keyout "$dir/tls.key" -out "$dir/tls.crt" -subj "/CN=sealed-secret/O=sealed-secret" 2>/dev/null

pulumi config set --secret sealingCert "$(cat "$dir/tls.crt")"
pulumi config set --secret sealingKey "$(cat "$dir/tls.key")"
echo "sealing key stored in Pulumi config for stack $(pulumi stack --show-name)"
