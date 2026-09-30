STACK      ?= demo
KUBECONFIG_FILE ?= $(HOME)/.kube/gitops-aks-demo
export KUBECONFIG := $(KUBECONFIG_FILE)

# The Pulumi program is pure Go, so cgo is off. State is local, no Pulumi Cloud account needed.
export CGO_ENABLED := 0
export PULUMI_BACKEND_URL ?= file://$(HOME)/.pulumi-state
export PULUMI_CONFIG_PASSPHRASE_FILE ?= $(HOME)/.pulumi-demo-passphrase

.PHONY: help preview up down kubeconfig argocd-password argocd-ui app-ui status test

help:
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-18s %s\n", $$1, $$2}'

test: ## Run the app tests
	cd app && go vet ./... && go test -race ./...

preview: ## Show what Pulumi would change
	cd infra && pulumi preview --stack $(STACK)

up: ## Create AKS and bootstrap Argo CD
	cd infra && pulumi up --stack $(STACK) --yes
	$(MAKE) kubeconfig

down: ## Destroy everything in Azure so billing stops
	cd infra && pulumi destroy --stack $(STACK) --yes
	rm -f $(KUBECONFIG_FILE)

kubeconfig: ## Write the cluster kubeconfig to $(KUBECONFIG_FILE)
	@mkdir -p $(dir $(KUBECONFIG_FILE))
	cd infra && pulumi stack output kubeconfig --show-secrets --stack $(STACK) > $(KUBECONFIG_FILE)
	chmod 600 $(KUBECONFIG_FILE)

argocd-password: ## Print the initial Argo CD admin password
	@kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d; echo

argocd-ui: ## Argo CD UI on https://localhost:8080
	kubectl -n argocd port-forward svc/argocd-server 8080:443

app-ui: ## dev app on http://localhost:8081, prod on 8082
	kubectl -n demo-dev port-forward svc/items-api 8081:80 & \
	kubectl -n demo-prod port-forward svc/items-api 8082:80 & wait

status: ## Sync and health for every Argo CD application
	kubectl -n argocd get applications
