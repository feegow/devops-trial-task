# Alerta baseado em Logs (Loki) - Guia Prático

> **O que é isso?** Um alerta que monitora os logs das suas aplicações e dispara quando tem muita coisa dando errado. Simples assim!

## 🤔 Por que precisamos disso?

Você já deve ter passado por isso: a aplicação começa a dar erro, mas você só descobre quando alguém reclama. Com este alerta, o Loki fica de olho nos logs e te avisa **antes** que vire um problemão.

**Exemplo real:**
- Sua API começou a retornar erro 500
- Em 2 minutos, o Loki detecta a explosão de erros
- Você recebe o alerta e já pode agir
- Cliente nem percebe o problema (ou percebe bem menos!)

## 🎯 O Alerta: LogsErrorBurst

### O que ele faz?

Simples: conta quantos logs de erro aparecem por segundo. Se passar de **5 erros/segundo**, ele dispara um alerta.

**Por que 5?** É tipo o limite entre "tá tudo bem" e "opa, tem algo errado aqui". 

**Fazendo as contas:**
- **SLO**: 99.5% de disponibilidade (error budget de 0.5%)
- **Taxa normal**: ~50 req/s × 0.01% erro = **~0.005 erros/s** (super saudável!)
- **Taxa no limite**: ~50 req/s × 0.5% erro = **~0.25 erros/s** (no limite do SLO)
- **Threshold do alerta**: **5 erros/s** = 1000× acima do normal, ou ~100× acima do limite do SLO
- **Conclusão**: Se chegou em 5 erros/s, algo está **MUITO** errado!

### Como funciona?

```
Aplicação Python/Go 
    ↓
Emite logs JSON: {"level":"error", "route":"/api/...", ...}
    ↓
Promtail coleta os logs
    ↓
Loki armazena e processa
    ↓
Loki Ruler executa a query a cada 1 minuto
    ↓
Se taxa > 5 erros/s por 2 minutos consecutivos
    ↓
🚨 Alerta dispara no AlertManager!
```

## 📝 A Consulta LogQL

### Query do Alerta (no Loki Ruler)

Esta é a query que vai no arquivo `loki-ruler-config.yaml`:

```logql
sum(rate({namespace="apps"} |~ "error" [5m])) > 5
```

### Query para Visualizar no Grafana

```logql
sum(rate({namespace="apps"} |~ "error" [5m]))
```

---
### Traduzindo para o português:

1. **`{namespace="apps"}`** 
   → Olha só os logs das aplicações (não os do sistema)

2. **`|~ "error"`**
   → Busca por logs que contêm "error"
   → O `|~` é tipo um "contains" com regex

3. **`[5m]`**
   → Nos últimos 5 minutos

4. **`rate(...)`**
   → Calcula: quantos logs de erro por segundo

5. **`sum(...)`**
   → Soma tudo (de todas as aplicações)

6. **`> 5`**
   → Se passar de 5 erros/segundo → 🚨 ALERTA!

### Exemplos de logs que são capturados:

```json
// ✅ Estes disparam o alerta:
{"level":"error", "status":500, ...}           // Tem "error" e "500"
{"level":"ERROR", "route":"/api", ...}         // Tem "ERROR"
{"status":503, "message":"timeout"}            // Status 5xx
{"info":"Database error occurred"}             // Tem "error"

// ❌ Estes NÃO disparam:
{"level":"info", "status":200, ...}            // Tudo OK
{"level":"warn", "status":404, ...}            // Warning, mas não error
```

## 🏷️ Os Rótulos (Labels)

Quando você olha o alerta ou investiga no Grafana, você tem acesso a vários **rótulos** que te ajudam a entender o que tá acontecendo:

### Labels Automáticos (Kubernetes)

| Label | O que é | Exemplo | Pra que serve |
|-------|---------|---------|---------------|
| **namespace** | Onde o pod está rodando | `apps` | Separar logs de apps vs sistema |
| **container** | Nome do container | `available-schedules-python` | Filtrar por aplicação |
| **app** | Label do Kubernetes | `available-schedules-python` | Filtrar por aplicação |
| **pod** | Nome do pod específico | `available-schedules-python-abc123` | Investigar um pod específico |

### Labels Extraídos dos Logs JSON

Estes vêm do conteúdo do log em si (após fazer `| json`):

| Label | O que é | Exemplo | Pra que serve |
|-------|---------|---------|---------------|
| **app** | Nome da aplicação (do log) | `available-schedules` | Filtrar por app |
| **route** | Qual endpoint deu erro | `/v1/appoints/available-schedule` | Ver qual rota tá falhando |
| **env** | Ambiente | `production`, `staging` | Separar prod de staging |
| **version** | Versão do código | `v1.0.0`, `v2.0.0` | Ver se erro veio após deploy |
| **level** | Severidade | `error`, `info`, `warning` | Filtrar só erros |
| **status** | Status HTTP | `200`, `500`, `503` | Ver qual tipo de erro |

## 🔍 Como Usar na Prática

### 1. Ver logs de erro no Grafana

**Acesse:** http://dev.local/grafana/explore

**Datasource:** Selecione **Loki** no dropdown

**Query básica:**
```logql
{namespace="apps"} |~ "error"
```

**O que você vê:**
- Lista de todos os logs que contém "error"
- Você pode clicar em cada um para ver detalhes
- Dá pra ver o timestamp, mensagem completa, etc

### 2. Filtrar por aplicação

**Python (v1):**
```logql
{namespace="apps", container="available-schedules-python"} |~ "error"
```

**Go (v2):**
```logql
{namespace="apps", container="available-schedules-go"} |~ "error"
```

**Ou usando o label `app`:**
```logql
{namespace="apps", app="available-schedules-python"} |~ "error"
```

### 3. Ver a taxa de erro (query do alerta, mas sem a condição!)

**Query simplificada:**
```logql
sum(rate({namespace="apps"} |~ "error" [5m]))
```

**O que aparece:**
- Um número tipo `8.5` = 8.5 erros por segundo
- Um gráfico mostrando como a taxa varia ao longo do tempo
- Se tá > 5, você já sabe: o alerta vai disparar!

**Por aplicação (com `by (container)`):**
```logql
sum by (container) (rate({namespace="apps"} |~ "error" [5m]))
```

## 🚨 Quando o Alerta Dispara

### O que você recebe

```
🔴 LogsErrorBurst - FIRING

Taxa de logs de erro: 8.52 logs/segundo
Namespace: apps
Job: available-schedules-python

Janela: 5 minutos
Threshold: >5 logs de erro/segundo

[Ver no Grafana] [Ver Runbook]
```

### O que fazer?

1. **Calma, respira** 🧘‍♂️
   - Você tem tempo, o alerta já te avisou cedo

2. **Abra o Grafana** 
   - Clique no link "Ver no Grafana" do alerta
   - Ou acesse: http://dev.local/grafana/explore

3. **Veja os logs de erro**
   ```logql
   {namespace="apps", job="<o-job-que-alertou>"} | json | level="error"
   ```

4. **Identifique o padrão**
   - É uma rota específica?
   - É um tipo de erro específico?
   - Começou depois de um deploy?

5. **Correlacione com métricas**
   - Veja se a latência também subiu
   - Veja se tem trace no Tempo mostrando o problema
   - Veja as métricas HTTP do Prometheus

6. **Aja!**
   - Rollback se foi após deploy
   - Escale se é falta de recursos
   - Fix o bug se é código
   - Reinicie o serviço se tá travado

## 📊 Configuração do Alerta

### Parâmetros (você pode ajustar!)

| Parâmetro | Valor Atual | O que faz | Quando ajustar |
|-----------|-------------|-----------|----------------|
| **Threshold** | `> 5` | Quantos erros/s ativa o alerta | Se tá alertando demais, aumenta. Se tá quieto demais, diminui |
| **Janela** | `[5m]` | Período que olha no passado | 5min é bom equilíbrio |
| **For** | `2m` | Quanto tempo acima do threshold para disparar | Evita alertas de spikes momentâneos |
| **Severity** | `warning` | Gravidade do alerta | `warning` = investiga, `critical` = acorda às 3h |

## 📁 Arquivos Importantes

### Onde está cada coisa:

```
📦 infra/observability/
├── 📄 loki-ruler-config.yaml           ← O alerta tá aqui!
│   └── Rules do Loki com query LogQL
│
├── 📄 values/loki-values.yaml          ← Config do Loki
│   ├── Loki Ruler habilitado
│   └── Pipeline Promtail (extrai labels)
│
└── 📄 prometheus-rules.yaml            ← Alertas de métricas
    └── (este é diferente, usa PromQL)

📦 apps/
├── 📁 available-schedules-python/
│   └── 📄 main.py                      ← Logs JSON estruturados
│
└── 📁 available-schedules-go/
    └── 📄 main.go                      ← Logs JSON estruturados
```

### Para editar o alerta:

```bash
# 1. Edite a regra
vim infra/observability/loki-ruler-config.yaml

# 2. Aplique
kubectl apply -f infra/observability/loki-ruler-config.yaml

# 3. Reinicie o Loki (para carregar nova regra)
kubectl rollout restart statefulset/loki -n observability

# 4. Aguarde ~1 minuto e a nova regra já está ativa!
```

---

**Dúvidas?** Abra um issue ou fala com o time de SRE! 🚀
