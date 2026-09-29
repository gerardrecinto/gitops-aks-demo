#!/usr/bin/env bash
# One time. Lets GitHub Actions run `pulumi preview` against Azure with no stored secret.
# A federated credential trusts tokens GitHub issues for this repo's pull requests.
set -euo pipefail
repo=${1:-gerardrecinto/gitops-aks-demo}
app_name="gh-${repo//\//-}"

sub=$(az account show --query id -o tsv)
tenant=$(az account show --query tenantId -o tsv)
client=$(az ad app create --display-name "$app_name" --query appId -o tsv)
az ad sp create --id "$client" > /dev/null

az ad app federated-credential create --id "$client" --parameters "{
  \"name\": \"pull-request\",
  \"issuer\": \"https://token.actions.githubusercontent.com\",
  \"subject\": \"repo:${repo}:pull_request\",
  \"audiences\": [\"api://AzureADTokenExchange\"]
}" > /dev/null

for role in "Reader" "Azure Kubernetes Service Cluster User Role"; do
  az role assignment create --assignee "$client" --role "$role" --scope "/subscriptions/$sub" > /dev/null
done

gh variable set AZURE_CLIENT_ID --repo "$repo" --body "$client"
gh variable set AZURE_TENANT_ID --repo "$repo" --body "$tenant"
gh variable set AZURE_SUBSCRIPTION_ID --repo "$repo" --body "$sub"
echo "done. Also add the PULUMI_ACCESS_TOKEN repository secret."
