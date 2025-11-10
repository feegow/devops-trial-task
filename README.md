# DevOps Senior — Trial Task

Este é um desafio para a posição de DevOps Senior. As seções abaixo guiarão você pela configuração e pelos objetivos do projeto.

## Requisitos

Antes de começar, certifique-se de que você tem as seguintes ferramentas instaladas:

- `docker`
- `kubectl`
- `helm`
- `kind`
- `make`
- `k6` (opcional, para testes de carga)

## Passos

1. **Configuração do Host**: Adicione a seguinte entrada ao seu arquivo `/etc/hosts` para acessar a aplicação localmente:

    ```sh
    127.0.0.1 dev.local
    ```

2. **Subir o Ambiente**: Crie o cluster Kubernetes local, o ingress controller e a stack de observabilidade.

    ```sh
    make up
    ```

3. **Deploy da Aplicação**: Faça o build, carregue a imagem no cluster e realize o deploy da `orders-api`, incluindo o HPA e o ServiceMonitor.

    ```sh
    make deploy
    ```

4. **Verificar a Aplicação**: Acesse o endpoint de health check para confirmar que a aplicação está no ar.
    - `http://dev.local/healthz`
5. **Acessar o Grafana**: Explore os dashboards de monitoramento.
    - **URL**: `http://localhost/`
    - **Credenciais**: admin/admin
    - *Nota: Os datasources para Loki e Tempo já estão pré-configurados.*
6. **Gerar Carga (Opcional)**: Use o `k6` para gerar carga na aplicação e observar seu comportamento.

    ```sh
    make load
    ```

7. **Testar Alertas**: Force a ativação dos alertas para validar a configuração.

    ```sh
    make fire-alerts
    ```

    Para reverter e silenciar os alertas, use `make calm`.

## Intencionalmente "ruim"

Para este desafio, alguns componentes foram configurados de forma subótima intencionalmente. Seu objetivo é melhorá-los.

- Alertas ruidosos
- HPA subótimo
- Painel incompleto

## Onde mexer

Os arquivos a seguir são os pontos de partida para suas alterações:

- `infra/observability/prometheus-rules.yaml` e `alertmanager.yaml`
- `dashboards/grafana/app-latency.json`
- `infra/apps/orders-api/hpa.yaml`
- `apps/orders-api` (instrumentação/flags)

## Desligar

Para parar e limpar todos os recursos do cluster, execute:

```sh
make down
```
