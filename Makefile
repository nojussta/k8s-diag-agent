# Experiment environment: `make up` -> `make check` -> `make scenario S=s1-oomkilled` -> `make down`.
# All tools are written in Go (cmd/labctl, cmd/alert-receiver, ...); Go >= 1.24 is required.
# Every command uses a separate kubeconfig file (.kube/config), so work cluster
# contexts (~/.kube/config) are never used or modified.

SHELL := /bin/bash
include env/versions.env

export KUBECONFIG := $(CURDIR)/.kube/config
KUBECTL := kubectl --context kind-$(KIND_CLUSTER)
HELM := helm --kube-context kind-$(KIND_CLUSTER)
S ?= s1-oomkilled
LABCTL := build/labctl
GOSRC := go.mod go.sum $(shell find cmd internal -name '*.go')
GOARCH := $(shell go env GOARCH)

.PHONY: help up cluster repos monitoring logging app receiver rules check scenario summary evidence \
        alerts grafana prometheus alertmanager status down build test vet

help: ## Show the targets
	@grep -hE '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

up: cluster repos monitoring logging receiver rules app ## Create the whole environment (~6-10 min)
	@echo "Environment ready. Next: make check"

cluster: ## Create the kind cluster
	@mkdir -p .kube
	@if kind get clusters 2>/dev/null | grep -qx $(KIND_CLUSTER); then \
	  echo "Cluster $(KIND_CLUSTER) already exists"; \
	  kind export kubeconfig --name $(KIND_CLUSTER) --kubeconfig $(KUBECONFIG); \
	else \
	  kind create cluster --config env/kind-config.yaml --kubeconfig $(KUBECONFIG) --wait 120s; \
	fi

repos:
	@helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
	@helm repo add grafana https://grafana.github.io/helm-charts >/dev/null 2>&1 || true
	@helm repo update prometheus-community grafana >/dev/null

monitoring: ## Prometheus + Alertmanager + Grafana (kube-prometheus-stack)
	$(HELM) upgrade --install kps prometheus-community/kube-prometheus-stack \
	  --version $(KUBE_PROMETHEUS_STACK_VERSION) -n monitoring --create-namespace \
	  -f env/helm/kube-prometheus-stack.values.yaml --wait --timeout 15m

logging: ## Loki + Alloy (logs and Kubernetes events)
	$(HELM) upgrade --install loki grafana/loki --version $(LOKI_CHART_VERSION) \
	  -n logging --create-namespace -f env/helm/loki.values.yaml --wait --timeout 10m
	$(HELM) upgrade --install alloy grafana/alloy --version $(ALLOY_CHART_VERSION) \
	  -n logging -f env/helm/alloy.values.yaml --wait --timeout 5m

build: $(LABCTL) ## Build all binaries into build/

$(LABCTL): $(GOSRC)
	go build -o $(LABCTL) ./cmd/labctl

test: ## Go vet and unit tests
	go vet ./...
	go test ./...

receiver: ## Alert receiver (Go image -> kind) and agent RBAC
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath -ldflags='-s -w' -o build/alert-receiver ./cmd/alert-receiver
	docker build -q -f Dockerfile.receiver -t $(RECEIVER_IMAGE) .
	kind load docker-image $(RECEIVER_IMAGE) --name $(KIND_CLUSTER)
	$(KUBECTL) apply -f env/manifests/diag-agent.yaml
	$(KUBECTL) -n diag-agent rollout restart deployment/alert-receiver
	$(KUBECTL) -n diag-agent rollout status deployment/alert-receiver --timeout=180s

rules: ## Lab alert rules
	$(KUBECTL) apply -f env/manifests/alert-rules.yaml

app: ## Demo application (healthy baseline)
	$(KUBECTL) apply -f env/manifests/demo-app.yaml
	$(KUBECTL) -n demo rollout status deployment/backend --timeout=180s
	$(KUBECTL) -n demo rollout status deployment/frontend --timeout=180s
	$(KUBECTL) -n demo rollout status deployment/loadgen --timeout=180s

check: $(LABCTL) ## Check that the environment works (including the alert path to the receiver)
	$(LABCTL) check

scenario: $(LABCTL) ## Run a fault scenario: make scenario S=s1-oomkilled
	$(LABCTL) scenario $(S)

summary: $(LABCTL) ## Results table (Markdown) from results/
	@$(LABCTL) summary

evidence: $(LABCTL) ## Collect diagnostic data for the latest alert (equivalent of manual diagnosis)
	$(LABCTL) evidence --latest

alerts: $(LABCTL) ## Show the alerts the receiver got
	@$(LABCTL) alerts

grafana: ## Grafana http://localhost:3000 (admin / kursinis)
	$(KUBECTL) -n monitoring port-forward svc/kps-grafana 3000:80

prometheus: ## Prometheus http://localhost:9090
	$(KUBECTL) -n monitoring port-forward svc/kps-prometheus 9090:9090

alertmanager: ## Alertmanager http://localhost:9093
	$(KUBECTL) -n monitoring port-forward svc/kps-alertmanager 9093:9093

status: ## Pod status in all lab namespaces
	$(KUBECTL) get pods -A -o wide | grep -E 'NAMESPACE|demo|monitoring|logging|diag-agent'

down: ## Delete the cluster
	kind delete cluster --name $(KIND_CLUSTER) --kubeconfig $(KUBECONFIG)
