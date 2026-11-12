# Experiment environment: `make up` -> `make check` -> `make scenario S=s1-oomkilled` -> `make down`.
# All tools are written in Go (cmd/labctl, cmd/alert-receiver, cmd/agent, ...); Go >= 1.24 is required.
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
        alerts grafana prometheus alertmanager status down build test vet probe agent agent-local agent-gemini eval eval-all eval-gemini watch

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
	go build -o build/mcp-server ./cmd/mcp-server
	go build -o build/mcp-probe ./cmd/mcp-probe
	go build -o build/agent ./cmd/agent

# Every scenario run also applies the rule-based baseline (method "rules") to the fixed sequence's evidence.
AGENT := build/agent --alert {alert} --out {out} --notify {notify}

agent: build ## Scenario with the agent loop; local apfel when installed, else Gemini: make agent S=s1-oomkilled [N=3]
	build/labctl scenario $(S) --repeat $(or $(N),1) --method agent="$(AGENT)"

agent-local: build ## Scenario with the agent loop on the on-device model (apfel, free, offline)
	build/labctl scenario $(S) --repeat $(or $(N),1) --method agent="$(AGENT) --provider apfel"

agent-gemini: build ## Scenario with the agent loop on Gemini (needs GEMINI_API_KEY in .env; free tier is ~20 requests/day/model)
	build/labctl scenario $(S) --repeat $(or $(N),1) --method agent="$(AGENT) --provider gemini"

eval: build ## Evaluation run: rules + pipeline + agent loop on the local model, same fault: make eval S=s4-latency-bad-config N=3
	build/labctl scenario $(S) --repeat $(or $(N),1) \
	  --method pipeline="$(AGENT) --provider apfel --mode pipeline --evidence {evidence}" \
	  --method agent="$(AGENT) --provider apfel"

SCENARIOS := s1-oomkilled s2-bad-image s3-missing-config s4-latency-bad-config s5-traffic-drop-upstream

eval-all: build ## Whole evaluation at no cost: all 5 scenarios x N runs (default 3), rules + pipeline + agent on the local model (~75 min)
	@for s in $(SCENARIOS); do \
	  build/labctl scenario $$s --repeat $(or $(N),3) \
	    --method pipeline="$(AGENT) --provider apfel --mode pipeline --evidence {evidence}" \
	    --method agent="$(AGENT) --provider apfel" || exit 1; \
	done
	@build/labctl summary --since $$(date +%Y%m%d)

eval-gemini: build ## Same as eval with Gemini for both model methods (free tier: paced to 15 requests/min; GEMINI_MODEL picks the model)
	GEMINI_MIN_INTERVAL=$(or $(GEMINI_MIN_INTERVAL),4.2s) build/labctl scenario $(S) --repeat $(or $(N),1) \
	  --method pipeline-gemini="$(AGENT) --provider gemini --mode pipeline --evidence {evidence}" \
	  --method agent-gemini="$(AGENT) --provider gemini"

watch: build ## Automatic path: investigate every new alert at the receiver and post the report (NOTIFY_WEBHOOK_URL or a local stand-in)
	build/labctl watch --method "$(AGENT) $(if $(P),--provider $(P),)"

probe: build ## MCP server: list the 10 tools and run an investigation using MCP calls only (after make scenario S=... --keep)
	build/mcp-probe -ns demo

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
