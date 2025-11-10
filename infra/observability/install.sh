#!/usr/bin/env bash
set -euo pipefail
kubectl create ns observability || true
kubectl create ns ingress || true

# ingress-nginx
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress \
  --set controller.hostNetwork=true \
  --set controller.kind=Deployment

# kube-prometheus-stack (release "kps")
helm upgrade --install kps prometheus-community/kube-prometheus-stack \
  --namespace observability \
  -f infra/observability/values/kps-values.yaml

# Loki
helm upgrade --install loki grafana/loki-stack \
  --namespace observability \
  -f infra/observability/values/loki-values.yaml

# Tempo
helm upgrade --install tempo grafana/tempo \
  --namespace observability \
  -f infra/observability/values/tempo-values.yaml

# OTel Collector (deployment simples)
kubectl apply -n observability -f infra/observability/values/otel-collector.yaml

# Datasources do Grafana
kubectl apply -n observability -f infra/observability/grafana-datasources.yaml

# Alertmanager route + alertas "ruidosos"
kubectl apply -n observability -f infra/observability/alertmanager.yaml
kubectl apply -n observability -f infra/observability/prometheus-rules.yaml

# Webhook de debug de alertas
kubectl apply -n observability -f infra/observability/alert-debugger/deploy.yaml