# Copilot Instructions: DevOps Trial Task Repository

## Repository Overview
DevOps trial repository with Kubernetes observability stack for medical scheduling APIs. Contains: 2 backend APIs (Python FastAPI v1, Go net/http v2), Node.js frontend, observability stack (Prometheus, Grafana, Loki, Tempo, OTel).

**Stack**: Python 3.11+, Go 1.22+, Node.js 20+, Kubernetes (kind), Helm, YAML manifests.
**Size**: ~20 files, primarily YAML configs and 3 app directories.

## Critical Build and Validation Instructions

### Prerequisites (ALWAYS verify first)
`docker`, `kind`, `kubectl`, `helm`, `make` (required) | `k6` (optional for load testing)

### Validated Command Sequences (MUST run in order)

**1. Environment Validation** (run first):
```bash
./scripts/preflight.sh
```
Checks binaries, Docker access, kind, `/etc/hosts` for `dev.local`. Warns if missing but doesn't fail.

**2. Local Testing**:

Python (FastAPI):
```bash
cd apps/available-schedules-python
pip install -r requirements.txt && pip install pytest pytest-cov
pytest -q tests --junitxml=report.xml --cov=. --cov-report=term-missing
```
Time: ~30s deps, ~1s tests. NEVER commit `.coverage`, `report.xml`.

Go:
```bash
cd apps/available-schedules-go && go build ./...
```
Time: ~10s. No unit tests exist. NEVER commit binary.

Node.js:
```bash
cd apps/available-schedules-web && npm install --omit=dev
```
Time: ~5s. No tests exist. Package.json has zero dependencies.

**3. CI Pipeline** (`.github/workflows/ci.yml`):
Runs on all pushes/PRs: Python tests (pytest+coverage) → Go build → Node deps. Duration: ~2-3 min. ALWAYS ensure your changes pass.

**4. Local Cluster** (8-12 minutes):
```bash
make up      # Creates kind cluster, installs observability (idempotent)
make deploy  # Builds Docker images, loads to kind, deploys apps
```
Customize timeouts: `export HELM_TIMEOUT_INGRESS=8m HELM_TIMEOUT_KPS=15m`

**5. Validation**:
```bash
kubectl get pods -n apps -n observability
make load          # k6 load test (optional)
make fire-alerts   # Simulate errors (ERROR_RATE=0.10)
make calm          # Reset to normal
make down          # Tear down cluster
```

### Common Build Failures
1. **Docker not accessible**: Ensure daemon running (`docker info`)
2. **Helm timeout**: Increase via `export HELM_TIMEOUT_KPS=15m`
3. **Port conflicts**: Ports 80/443 must be free (ingress ports 80/443 are exposed on host ports 30080/30443)
4. **Image not in kind**: Re-run `make deploy`

## Project Layout

### Key Directories
```
.github/workflows/ci.yml     # CI: Python tests, Go build, Node deps
Makefile                     # Main automation (up/deploy/down/load/fire-alerts/calm)
scripts/preflight.sh         # Environment validation
apps/
  available-schedules-python/  # FastAPI v1 (port 8000, has tests)
    main.py, requirements.txt, Dockerfile, tests/test_smoke.py
  available-schedules-go/      # Go v2 (port 8080, no tests)
    main.go, go.mod, Dockerfile
  available-schedules-web/     # Node frontend (port 3000, no tests)
    server.mjs, package.json, Dockerfile
infra/
  kind/cluster.yaml          # Kind config (port mappings)
  apps/*/                    # K8s manifests per service
    deployment.yaml, service.yaml, hpa.yaml, servicemonitor.yaml
  ingress/                   # Ingress rules (/, /v1, /v2, /grafana)
  observability/             # Helm install scripts, Prometheus rules, Alertmanager
    install.sh, repos.sh, prometheus-rules.yaml, alertmanager.yaml
    values/                  # Helm values (kps, loki, tempo, otel)
dashboards/grafana/app-latency.json  # Dashboard with TODO panels
runbooks/erro-5xx.md         # 5xx incident runbook (Portuguese: filename "erro-5xx" is intentional; "erro" means "error")
tests/k6/available_schedules.js      # Load test script
```

### Critical Configuration Details

**Makefile**: `KIND_CLUSTER=devops-lab`, `NAMESPACE_APPS=apps`, `NAMESPACE_OBS=observability`

**Environment Variables** (in deployment.yaml):
- `ERROR_RATE`: Simulated error rate (0.01 default, 0.02 in Python deployment)
- `EXTRA_LATENCY_MS`: Simulated latency (0 default, 50 in Python deployment)
- `SERVICE_NAME`: "available-schedules"
- `OTEL_EXPORTER_OTLP_ENDPOINT`: Points to otel-collector.observability

**Ingress Routes** (via `http://dev.local`):
- `/` → web:3000, `/v1/appoints/available-schedule` → python:8000
- `/v2/appoints/available-schedule` → go:8080, `/healthz` → health checks
- `/grafana/` → Grafana UI (admin/admin)

**Observability**:
- Prometheus scrapes `/metrics` from Python/Go (ServiceMonitors)
- Alerts in `prometheus-rules.yaml`: HighErrorRate, HighLatencyP95 (intentionally noisy)
- Grafana datasources: Prometheus (default), Loki, Tempo (port 3200)
- OTel Collector receives traces on port 4318. **Go service has NO tracing** (intentional gap)
- Dashboard `app-latency.json` has incomplete TODO panels (p95 by route, errors by route)

**Known Intentional Gaps** (for trial task):
- Alerts have fragile thresholds
- HPAs lack stabilization, custom metrics
- Go service metrics only, no OTel traces
- No frontend tests or CI pipeline

## Best Practices

1. **ALWAYS** run `./scripts/preflight.sh` before `make up`
2. **NEVER commit**: `.coverage`, `report.xml` (test artifacts), Go binary (build artifact), `node_modules/` (dependencies)
3. **Test locally first**: Run `pytest` (Python), `go build` (Go) before pushing
4. **CI must pass**: Changes must pass `.github/workflows/ci.yml`
5. **Use Make targets**: `make deploy` over manual `kubectl apply` (ensures image loading)
6. **Check logs**: `kubectl logs -n apps deploy/<service>` after changes
7. **Document env changes**: Update Dockerfile defaults AND deployment.yaml

## README.md Summary
152-line Portuguese doc covering: requirements (docker/kind/kubectl/helm/make), setup steps (`make up`, `make deploy`), accessing apps (`http://dev.local`), Grafana dashboards, load testing (`make load`), alert testing (`make fire-alerts`/`make calm`), intentional issues for improvement, and observability challenge details (SRE track with SLOs, burn-rates, Loki alerts, dashboard completion, CI/CD unit tests).

---
**Trust these instructions**: All commands validated. Only search if errors occur or repo modified.
