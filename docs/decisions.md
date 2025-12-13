# Aqui algumas decisões para o desafio



## Problema com meu dev.local
(antes)
- Eu tive um problema para usar o dev.local, entao 

(depois)
- Alterei a stack para usar
dev.lab

## Http Echo Image

Aqui eu tive um problema de Access Denied da imagem:

(antes)
```yaml
containers:
  - name: echo
    image: ghcr.io/caarlos0/httpecho:latest
    args: ["-port", "8080", "-i"]
    ports:
      - containerPort: 8080
```

(depois)
Alterei a imagem e os args para:

```yaml
containers:
  - name: echo
    image: hashicorp/http-echo:1.0
    args:
      - "-text=alert-debugger ok"
      - "-listen=:8080"
    ports:
      - containerPort: 8080
```


### Ruidos no prometheus-rules.yaml

(antes)

- Notei que o arquivo estava configuardo com oscilacao/salto curto, no qual
disparva o alerta muito rapido (1m)

- Notei que os thresholds estavam sensiveis (0,5% 5xx e p95 > 150ms), ao menos
nos testes locais com k6 estavam disparando muito rapido.

- Sem protecao com baixa amostragem, pouco trafego e 1 erro distorcendo a %
e ativava o alerta sem impacto real.

- Faltava label team e runbook

(depois)

- Alterado:
  - HighErrorRate: > 0.0.5 (0.5%) para 0.02 (2%)
  - HighLatencyP95: > 0.15s para 0.30s
  Tomei essa decisao pois o teste de carga mostrou taxa de falha 1-1.5% e
  latencias bem abaixo de 150ms na maior parte do tempo 0.5%/150ms o que gerava
  alertas constantes sem acao real.
  
- Adicionei condicao de trafego minimo
```sum(rate(http_requests_total{...}[5m])) > 1```
a ideia foi tentar evitar quando o volume e baixo e as metricas ficam instaveis
(um erro virava varios %).

- Adicionado team: sre e annotation runbook_url
- Adicionado runbook/latency-p95.md seguindo o modelo do
erro-5xx.md
- Tentei focar o tempo de detectacao pois a ideia foi tentar reduzir o flapping
e focar em incidente de verdade.


### LogsErrorBurst

(antes)

- Aplicacoes nao possuiam nivel de log configurado
- O Loki nao estava configurado no ambiente corretamente para essa atividade
- Para criar o alerta LogErrorBurst, precisava de um componete que monitorasse
os logs continuamentes.
  

(depois)

- Configurado level logs nas Apps Go e Python
- Configurei para enviar alertas ao mesmo AlertManager do cluster dev
- Criei um ConfigMap com a regra LogsErrorBust separado para facilitar o versionamento


### Dashboard Grafana

(antes)

- Faltava completar com RPS, 4xx/5xx, p50,p90 e p95

(depois)

- Adicionado as metricas solicitadas
- Adicionado dashboard de monitoracao do Logs de Erros das Apps
