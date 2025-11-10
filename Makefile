SHELL := /bin/bash
KIND_CLUSTER := devops-lab
APP_NAME := orders-api
APP_IMG := $(APP_NAME):local
NAMESPACE_APPS := apps
NAMESPACE_OBS := observability

.PHONY: up down deploy app.build app.load helm.repos obs.install load fire-alerts calm dashboards.export

up: ## Cria cluster kind e instala stack de observabilidade + ingress
	kind create cluster --name $(KIND_CLUSTER) --config infra/kind/cluster.yaml || true
	./infra/observability/repos.sh
	./infra/observability/install.sh
	kubectl create ns $(NAMESPACE_APPS) || true
	kubectl apply -f infra/ingress/ingress.yaml

app.build: ## Build da aplicação
	docker build -t $(APP_IMG) ./apps/$(APP_NAME)

app.load: app.build ## Carrega imagem no kind
	kind load docker-image $(APP_IMG) --name $(KIND_CLUSTER)

deploy: app.load ## Deploy da aplicação, HPA e ServiceMonitor
	kubectl apply -n $(NAMESPACE_APPS) -f infra/apps/$(APP_NAME)/deployment.yaml
	kubectl apply -n $(NAMESPACE_APPS) -f infra/apps/$(APP_NAME)/service.yaml
	kubectl apply -n $(NAMESPACE_APPS) -f infra/apps/$(APP_NAME)/hpa.yaml
	kubectl apply -n $(NAMESPACE_APPS) -f infra/apps/$(APP_NAME)/servicemonitor.yaml

load: ## Gera carga para acionar alertas
	k6 run tests/k6/orders_surge.js || echo "Instale k6 para rodar este alvo"

fire-alerts: ## Aumenta erro/latência por 15m
	kubectl -n $(NAMESPACE_APPS) set env deploy/$(APP_NAME) ERROR_RATE=0.10 EXTRA_LATENCY_MS=400 || true
	@echo "Mantendo flags ligadas por ~15m; depois rode 'make calm' para normalizar"

calm: ## Normaliza flags
	kubectl -n $(NAMESPACE_APPS) set env deploy/$(APP_NAME) ERROR_RATE=0.01 EXTRA_LATENCY_MS=0 || true

obs.install: ## (re)instala componentes de observabilidade
	./infra/observability/install.sh

dashboards.export:
	@echo "Exporte via UI do Grafana e salve em dashboards/grafana/*.json"

down: ## Destroi cluster
	kind delete cluster --name $(KIND_CLUSTER)