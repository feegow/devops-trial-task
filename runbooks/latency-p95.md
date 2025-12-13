# Runbook — Latência p95 alta nas APIs Available Schedules

> Objetivo: orientar o diagnóstico e a mitigação de latência p95 elevada nas APIs `available-schedules-python` (v1) e `available-schedules-go` (v2), e entender se a degradação é de backend, ingress ou UI.

## 1. Sinais de detecção

- **Alertas**: verifique o Alertmanager para alertas de `HighLatencyP95` (p95 acima do limiar por tempo sustentado).
- **Dashboards**: abra `Available Schedules — Latency & Errors` no Grafana (`http://dev.lab/grafana/`). Observe:
  - Painel "Latency p95 por rota" — quais rotas estão degradadas (v1 vs v2).
  - Painel de "RPS (req/s)" — houve pico de carga junto com a degradação?
  - Painel "Error rate 5xx (%)" — latência alta está antecedendo erros 5xx?
  - (Se existir) painel de latência do frontend/ingress — degradação é só backend ou caminho completo?
- **Logs**: Loki (`Explore > Loki datasource`) filtrando `app="available-schedules"` e buscando por mensagens de timeout, retries, erros upstream, ou aumentos de tempo de processamento.
- **Traces**: Tempo (`Explore > Tempo datasource`) para identificar spans longos/outliers; filtre por rota e compare duração entre serviços.

## 2. Checklist de isolamento rápido

1. **Qual rota?** Veja o painel "Latency p95 por rota" e identifique `route` afetada.
2. **Qual versão?** O problema é só na v1 (Python), só na v2 (Go) ou em ambas?
3. **Cresceu o tráfego?** Compare com "RPS (req/s)". Se subiu, pode ser saturação/limite de recursos.
4. **Tem erro junto?** Se 5xx também subiu, pode haver timeout/upstream falhando.
5. **Saturação**: `kubectl top pods -n apps` e reinícios/evictions em `kubectl get pods -n apps`.
6. **HPA**: o autoscaling reagiu? (`kubectl get hpa -n apps`).

## 3. Fluxo de diagnóstico detalhado

### 3.1 Backend Python (v1)
- Logs:
  - `kubectl logs deploy/available-schedules-python -n apps --tail=200`
  - Procure por exceções, timeouts e mensagens relacionadas a lentidão.
- Possível injeção de latência (cenário de teste do desafio):
  - Verifique `EXTRA_LATENCY_MS` e `ERROR_RATE` nos manifests (`infra/apps/available-schedules-python`).
- Métricas:
  - Compare latência p95 e p90/p50 por rota para ver se a degradação é generalizada ou só outliers.
  - Use o painel de p95 por rota para confirmar o endpoint mais afetado.

### 3.2 Backend Go (v2)
- Logs:
  - `kubectl logs deploy/available-schedules-go -n apps --tail=200`
  - Procure por slow requests, timeouts, GC spikes, ou erros upstream.
- Métricas:
  - Use `http_request_duration_seconds_bucket` para ver se a latência aumentou em buckets maiores.
  - Verifique se a degradação é por rota (`route`) ou global.

### 3.3 Frontend (UI Node.js)
- Logs:
  - `kubectl logs deploy/available-schedules-web -n apps --tail=200`
  - Procure por timeouts ao chamar as APIs, retries, ou falhas de fetch.
- Sintoma comum:
  - Backend ok, mas UI lenta → pode ser cache, bundle grande, ou latência no ingress.

### 3.4 Ingress / Rede (caminho do usuário)
- Ingress controller:
  - `kubectl logs -n ingress deploy/ingress-nginx-controller --tail=200`
  - Procure por `upstream_response_time`, `upstream_status`, `504`, `499`, `timeout`.
- Serviços/endpoints:
  - `kubectl get svc,ep -n apps`
  - Confirme endpoints saudáveis.
- Se existir painel de ingress:
  - Compare latência no ingress vs latência no backend para saber onde está o “tempo gasto”.

### 3.5 OTel / Traces (quando disponíveis)
- No Tempo, filtre por rota e observe:
  - quais spans concentram maior duração (handler, middleware, chamadas downstream)
  - presença de `status.error` ou spans longos com retries
- Se a API Go não estiver emitindo traces (ponto citado no README):
  - documente a limitação e, se fizer bônus, instrumente spans OTLP no Go.

## 4. Mitigações rápidas

- **Se for pico de carga/saturação**
  - Escalar réplicas temporariamente:
    - `kubectl scale deploy/available-schedules-go -n apps --replicas=3`
    - `kubectl scale deploy/available-schedules-python -n apps --replicas=3`
  - Verificar HPA e limites de CPU/memória (ajustar targets/stabilization se necessário).

- **Se for injeção de latência (cenário do desafio)**
  - Reduzir `EXTRA_LATENCY_MS` e aplicar os manifests.
  - Confirmar via dashboard que p95 voltou ao normal.

- **Se for problema transitório**
  - `kubectl rollout restart deploy/available-schedules-{python,go,web} -n apps`

- **Rollback (se houve mudança recente)**
  - `kubectl rollout undo deploy/available-schedules-python -n apps`
  - `kubectl rollout undo deploy/available-schedules-go -n apps`
  - `kubectl rollout undo deploy/available-schedules-web -n apps`

## 5. Pós-incidente

1. Registrar causa-raiz (RCA) e sinais observados (p95, RPS, 5xx, rota afetada).
2. Se necessário, criar/ajustar alertas:
   - latência por rota (se hoje só existe global)
   - correlação latência + saturação (CPU/memória) para reduzir falsos positivos
3. Melhorar o dashboard:
   - adicionar p50/p90/p95 por rota e separar v1 vs v2 se necessário
4. Atualizar este runbook com o aprendizado.

## 6. Artefatos úteis

- **Dashboards**:
  - `Available Schedules — Latency & Errors`
  - `Kubernetes / Compute Resources / Namespace`
- **Consultas Prometheus**:
  - Latência p95 global:
    - `histogram_quantile(0.95, sum by (le)(rate(http_request_duration_seconds_bucket{job=~"available-schedules-(python|go)"}[5m])))`
  - Latência p95 por rota:
    - `histogram_quantile(0.95, sum by (route, le)(rate(http_request_duration_seconds_bucket{job=~"available-schedules-(python|go)"}[5m])))`
  - RPS:
    - `sum(rate(http_requests_total{job=~"available-schedules-(python|go)"}[5m]))`
- **Comandos**:
  - `kubectl top pods -n apps`
  - `kubectl get hpa -n apps`
  - `kubectl get events -n apps --sort-by=.lastTimestamp | tail`

> Checklist rápido de encerramento: rota identificada ✔ causa provável definida ✔ mitigação aplicada ✔ p95 estabilizado ✔ (se aplicável) HPA/recursos ajustados ✔ runbook atualizado ✔
