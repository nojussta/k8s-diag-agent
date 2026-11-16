# Experiment environment: `make up` -> `make check` -> `make scenario S=s1-oomkilled` -> `make down`.
# All tools are written in Go (cmd/labctl, cmd/diag-agent, cmd/agent, ...); Go >= 1.24 is required.
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
        alerts grafana prometheus alertmanager status down build test vet probe agent agent-local agent-gemini eval eval-all eval-gemini watch \
        image agent-deploy agent-secret agent-on agent-off agent-logs incluster storm netpol-check

help: ## Show the targets
	@grep -hE '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-14s %s\n", $$1, $$2}'

up: cluster repos monitoring logging agent-deploy rules app ## Create the whole environment (~6-10 min)
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

SCENARIOS := s1-oomkilled s2-bad-image s3-missing-config s4-latency-bad-config s5-traffic-drop-upstream \
             s6-latency-slow-dependency s7-log-injection

eval-all: build ## Whole evaluation at no cost: all scenarios x N runs (default 3), rules + pipeline + agent on the local model (~75 min)
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

receiver: agent-deploy ## Same as agent-deploy (the agent pod is also the alert receiver)

image: ## Build the agent image (diag-agent + mcp-server, FROM scratch) and load it into kind
	@mkdir -p build/linux
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath -ldflags='-s -w' -o build/linux/diag-agent ./cmd/diag-agent
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath -ldflags='-s -w' -o build/linux/mcp-server ./cmd/mcp-server
	docker build -q -f Dockerfile.agent -t $(AGENT_IMAGE) .
	kind load docker-image $(AGENT_IMAGE) --name $(KIND_CLUSTER)

# The NetworkPolicy needs the node network (the API server and kubelet reach the pod from it).
NODE_CIDR = $(shell docker network inspect kind --format '{{range .IPAM.Config}}{{.Subnet}} {{end}}' | tr ' ' '\n' | grep -m1 '\.')

agent-deploy: image ## Deploy the agent into the cluster (records alerts; make agent-on starts investigations)
	sed 's|NODE_CIDR|$(NODE_CIDR)|g' env/manifests/diag-agent.yaml | $(KUBECTL) apply -f -
	-$(KUBECTL) -n diag-agent delete deployment alert-receiver --ignore-not-found
	$(KUBECTL) -n diag-agent rollout restart deployment/diag-agent
	$(KUBECTL) -n diag-agent rollout status deployment/diag-agent --timeout=180s

agent-secret: ## Put GEMINI_API_KEY (and NOTIFY_WEBHOOK_URL, if set) from .env into the Secret diag-agent-llm
	@set -a; . ./.env; set +a; \
	$(KUBECTL) -n diag-agent create secret generic diag-agent-llm \
	  --from-literal=GEMINI_API_KEY="$$GEMINI_API_KEY" \
	  $${NOTIFY_WEBHOOK_URL:+--from-literal=NOTIFY_WEBHOOK_URL="$$NOTIFY_WEBHOOK_URL"} \
	  --dry-run=client -o yaml | $(KUBECTL) apply -f - >/dev/null && echo "secret diag-agent-llm updated"

agent-on: agent-secret ## In-cluster agent investigates every firing alert on its own
	$(KUBECTL) -n diag-agent set env deployment/diag-agent AUTO_INVESTIGATE=true $(if $(BUDGET),QUEUE_BUDGET=$(BUDGET),)
	$(KUBECTL) -n diag-agent rollout status deployment/diag-agent --timeout=180s

agent-off: ## In-cluster agent only records alerts (needed for make eval*, which start the methods themselves)
	$(KUBECTL) -n diag-agent set env deployment/diag-agent AUTO_INVESTIGATE=false
	$(KUBECTL) -n diag-agent rollout status deployment/diag-agent --timeout=180s

agent-logs: ## Logs of the in-cluster agent
	$(KUBECTL) -n diag-agent logs deployment/diag-agent --tail=100

netpol-check: ## Prove the NetworkPolicy is enforced: a probe pod with the agent's label tries allowed and forbidden destinations
	@$(KUBECTL) -n diag-agent delete pod netpol-probe --ignore-not-found >/dev/null
	@$(KUBECTL) -n diag-agent run netpol-probe --image=$(LOADGEN_IMAGE) --restart=Never --labels=app=diag-agent,component=probe \
	  --command -- sh -c 'try() { if nc -z -w 4 $$2 $$3 2>/dev/null; then r=open; else r=blocked; fi; \
	    if [ $$r = $$1 ]; then echo "PASS $$2:$$3 is $$r"; else echo "FAIL $$2:$$3 is $$r, expected $$1"; fi; }; \
	  try open kps-prometheus.monitoring.svc.cluster.local 9090; try open loki.logging.svc.cluster.local 3100; \
	  try open generativelanguage.googleapis.com 443; try blocked frontend.demo.svc.cluster.local 9898; \
	  try blocked backend.demo.svc.cluster.local 9898; try blocked kps-alertmanager.monitoring.svc.cluster.local 9093; \
	  try blocked example.com 80' >/dev/null
	@$(KUBECTL) -n diag-agent wait --for=jsonpath='{.status.phase}'=Succeeded pod/netpol-probe --timeout=120s >/dev/null
	@$(KUBECTL) -n diag-agent logs netpol-probe; $(KUBECTL) -n diag-agent delete pod netpol-probe >/dev/null

incluster: build ## Scenario diagnosed by the in-cluster agent alone (after make agent-on): make incluster S=s4-latency-bad-config [N=3]
	build/labctl scenario $(S) --repeat $(or $(N),1) --incluster

storm: build ## Send N synthetic alerts at once to the in-cluster agent and show what the queue did: make storm [N=20]
	build/labctl storm $(or $(N),20)

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
