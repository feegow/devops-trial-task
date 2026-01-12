## Entregáveis obrigatórios

### 1. Reduzir ruído de alertas

#### 1.1 HighErrorRate

```sh
**Antes:**
expr: ... > 0.005
for: 1m
labels:
  severity: critical
**Depois:**
expr: sum(rate(...[3m])) / sum(rate(...[3m])) > 0.01
for: 5m
labels:
  severity: warning
  service: available-schedules
  team: backend
annotations:
  runbook_url: "https://github.com/feegow/runbooks/blob/main/erro-5xx.md"

**Justificativa:** Aumentei a janela de `for` de 1m para 5m para evitar alertas em picos. Mudei severity de `critical` para `warning` pois 1% de erro não é crítico imediatamente. Adicionei labels `team` e `service` para roteamento correto no Alertmanager.
```

#### 1.2 HighLatencyP95

```sh

**Antes:**
expr: histogram_quantile(0.95, ...) > 0.15
for: 1m
labels:
  severity: critical
**Depois:**
expr: histogram_quantile(0.95, sum by (le)(rate(http_request_duration_seconds_bucket{job=~"available-schedules-(python|go)"}[5m]))) > 0.5
for: 5m
labels:
  severity: warning
  service: available-schedules
  team: backend
annotations:
  runbook_url: "https://github.com/feegow/devops-trial-task/blob/main/runbooks/erro-5xx.md"
**Justificativa:** Aumentei o threshold de 150ms para 500ms, mais realista para a aplicação. Janela `for` de 5m evita falsos positivos em picos momentâneos. Mudei severity para `warning` e adicionei labels para roteamento.
```

### 2. Definir 1 SLO + burn-rate
#### Definição do SLI/SLO

**SLI (Service Level Indicator)**:
- Métrica: Taxa de sucesso de requisições HTTP
- Fórmula: `requests_success / requests_total` onde sucesso = status < 500

**SLO (Service Level Objective)**:
- Target: 99,5% de disponibilidade
- Janela: 30 dias 
- Error Budget: 0,5% (~3,6 horas de indisponibilidade permitida por mês)


### 3. Alerta baseado em logs (Loki)
### Alerta LogsErrorBurst (Loki)

#### Consulta LogQL
sum by (app, route) (
  count_over_time(
    {namespace="apps"} 
    |~ `"status":5[0-9]{2}` 
    | regexp `"route":"(?P<route>[^"]+)"` 
    [5m]
  )
) > 10

### Rótulos utilizados
| Rótulo | Origem | Descrição |
|--------|--------|-----------|
| `app` | Pod label (Promtail) | Nome da aplicação |
| `route` | Extraído do log via regexp | Rota da requisição |
| `env` | Pod label (Promtail) | Ambiente (development/staging/production) |
| `version` | Pod label (Promtail) | Versão da aplicação |

### Threshold
- **Dispara quando**: Mais de 10 erros 5xx em 5 minutos
- **Severidade**: warning

### 4. Dashboard de aplicação

- Ver no Grafana
- Foi preciso alterar o código Python para ter a info na p50/p90/p95 por rota
- App Go pendente

### 5. CI/CD com testes unitários

#### Adicionado para Go:

- name: Run Go unit tests
  working-directory: apps/available-schedules-go
  run: go test -v ./...#### Execução local com act run: go test -v ./...
```

### Comando para rodar o act 
act -j build-and-test --container-daemon-socket - --pull=false

### Ganhos
- **Velocidade**: Feedback em ~1-2 minutos localmente vs ~3-5 minutos no GitHub Actions. Evita ciclos de "commit → esperar → corrigir → commit".
- **Segurança**: Identifica falhas de teste e vulnerabilidades (Trivy) antes do push, mantendo a branch principal sempre verde.
- **Economia**: Reduz consumo de minutos do GitHub Actions em execuções que falhariam.

Exemplo de execucao:
[ci/build-and-test] ✅ Success - Run Trivy security scan (Python)
[ci/build-and-test] ✅ Success - Run Trivy security scan (Go)
[ci/build-and-test] ✅ Success - Run Python unit tests
[ci/build-and-test] ✅ Success - Run Go unit tests
[ci/build-and-test] ⭐ Run Complete job
[ci/build-and-test] ✅  Success - Complete job


Adicionei o Trivy na pipeline tambem (ci.yaml)
---


## Entregáveis opcionais

**Comando:**
trivy image available-schedules-python:local

Antes da correcão 

Python (python-pkg)

Total: 3 (UNKNOWN: 0, LOW: 0, MEDIUM: 2, HIGH: 1, CRITICAL: 0)

┌──────────────────────┬────────────────┬──────────┬────────┬───────────────────┬───────────────┬─────────────────────────────────────────────────────┐
│       Library        │ Vulnerability  │ Severity │ Status │ Installed Version │ Fixed Version │                        Title                        │
├──────────────────────┼────────────────┼──────────┼────────┼───────────────────┼───────────────┼─────────────────────────────────────────────────────┤
│ pip (METADATA)       │ CVE-2025-8869  │ MEDIUM   │ fixed  │ 24.0              │ 25.3          │ pip: pip missing checks on symbolic link extraction │
│                      │                │          │        │                   │               │ https://avd.aquasec.com/nvd/cve-2025-8869           │
├──────────────────────┼────────────────┼──────────┤        ├───────────────────┼───────────────┼─────────────────────────────────────────────────────┤
│ starlette (METADATA) │ CVE-2025-62727 │ HIGH     │        │ 0.41.3            │ 0.49.1        │ starlette: Starlette DoS via Range header merging   │
│                      │                │          │        │                   │               │ https://avd.aquasec.com/nvd/cve-2025-62727          │
│                      ├────────────────┼──────────┤        │                   ├───────────────┼─────────────────────────────────────────────────────┤
│                      │ CVE-2025-54121 │ MEDIUM   │        │                   │ 0.47.2        │ starlette: Starlette denial-of-service              │
│                      │                │          │        │                   │               │ https://avd.aquasec.com/nvd/cve-2025-54121          │
└──────────────────────┴────────────────┴──────────┴────────┴───────────────────┴───────────────┴─────────────────────────────────────────────────────


Python (python-pkg)

Total: 1 (UNKNOWN: 0, LOW: 0, MEDIUM: 1, HIGH: 0, CRITICAL: 0)

┌────────────────┬───────────────┬──────────┬────────┬───────────────────┬───────────────┬─────────────────────────────────────────────────────┐
│    Library     │ Vulnerability │ Severity │ Status │ Installed Version │ Fixed Version │                        Title                        │
├────────────────┼───────────────┼──────────┼────────┼───────────────────┼───────────────┼─────────────────────────────────────────────────────┤
│ pip (METADATA) │ CVE-2025-8869 │ MEDIUM   │ fixed  │ 24.0              │ 25.3          │ pip: pip missing checks on symbolic link extraction │
│                │               │          │        │                   │               │ https://avd.aquasec.com/nvd/cve-2025-8869           │
└────────────────┴───────────────┴──────────┴────────┴───────────────────┴───────────────┴─────────────────────────────────────────────────────┘

### HPA Tuning

- Instalei o Metrics Server via Helm chart para monitorar o HPA corretamente
- Adicionei na instalação (`install.sh` e `repos.sh`)
- Ajustei os targets no HPA