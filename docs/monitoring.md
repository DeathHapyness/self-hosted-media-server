# 📊 Monitoring

A pasta `monitoring/` é utilizada para centralizar os serviços de **monitoramento e observabilidade** do servidor.

Atualmente, ela contém:

* **Grafana** — utilizado para visualizar métricas e criar dashboards.
* **Prometheus** — utilizado para coletar e armazenar métricas.

## 📁 Estrutura

```text
monitoring/

├── grafana/
│   └── config/
│
├── prometheus/
│   ├── config/
│   │   └── prometheus.yml
│   └── data/
│
├── docker-compose.yml
├── .gitignore
└── README.md
```

As pastas `config/` e `data/` contêm dados internos dos serviços e não devem ser enviadas para o GitHub.

## Criar diretórios

```bash
mkdir -p ~/media-server/monitoring

cd ~/media-server/monitoring

mkdir -p grafana/config
mkdir -p prometheus/config
mkdir -p prometheus/data
```

## Docker Compose

Crie o arquivo:

```text
docker-compose.yml
```

com:

```yaml
services:

  prometheus:

    image: prom/prometheus:latest

    container_name: prometheus

    restart: unless-stopped

    volumes:
      - ./prometheus/config/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./prometheus/data:/prometheus

    ports:
      - "9090:9090"

  grafana:

    image: grafana/grafana:latest

    container_name: grafana

    restart: unless-stopped

    depends_on:
      - prometheus

    volumes:
      - ./grafana/config:/var/lib/grafana

    ports:
      - "3000:3000"
```

## Configuração do Prometheus

Crie:

```text
prometheus/config/prometheus.yml
```

Exemplo básico:

```yaml
global:
  scrape_interval: 15s

scrape_configs:

  - job_name: "prometheus"

    static_configs:
      - targets:
          - "prometheus:9090"
```

Essa configuração faz o Prometheus coletar métricas dele próprio.

Outros serviços podem ser adicionados posteriormente através de novos `scrape_configs`.

## Iniciar

Na pasta `monitoring/`:

```bash
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

Os containers esperados são:

```text
prometheus
grafana
```

## Acesso ao Prometheus

O Prometheus utiliza a porta `9090`.

```text
http://IP_DO_SERVIDOR:9090
```

Exemplo:

```text
http://IP_DO_SERVIDOR:9090
```

> Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor. Evite colocar IPs pessoais ou públicos reais no GitHub.

## Acesso ao Grafana

O Grafana utiliza a porta `3000`.

```text
http://IP_DO_SERVIDOR:3000
```

Exemplo:

```text
http://IP_DO_SERVIDOR:3000
```

## 🔗 Integração Grafana + Prometheus

O Grafana utiliza o Prometheus como fonte de dados para criar dashboards.

No Grafana:

```text
Connections
→ Data sources
→ Add new data source
→ Prometheus
```

Como os dois serviços estão na mesma rede Docker Compose, utilize:

```text
http://prometheus:9090
```

Depois:

```text
Save & test
```

Se a conexão estiver funcionando, o Grafana poderá consultar as métricas armazenadas pelo Prometheus.

## 📊 Fluxo de monitoramento

A arquitetura básica é:

```text
┌──────────────────────┐
│      Serviços        │
│                      │
│ Jellyfin             │
│ Radarr               │
│ Sonarr               │
│ Prowlarr             │
│ qBittorrent           │
│ Docker               │
└──────────┬───────────┘
           │
           │ Métricas
           ▼
┌──────────────────────┐
│     Prometheus       │
│                      │
│ Coleta e armazena    │
│ métricas             │
└──────────┬───────────┘
           │
           │ PromQL
           ▼
┌──────────────────────┐
│       Grafana        │
│                      │
│ Dashboards e gráficos│
└──────────────────────┘
```

## 📈 Monitoramento do servidor

Para monitorar informações do próprio servidor, como:

```text
CPU
RAM
Disco
Rede
Load
Temperatura
```

é recomendado utilizar um exporter, como o **Node Exporter**.

Nesse caso, o fluxo fica:

```text
Servidor
   │
   ▼
Node Exporter
   │
   ▼
Prometheus
   │
   ▼
Grafana
```

O Node Exporter expõe métricas do sistema em um endpoint que pode ser coletado pelo Prometheus.

## 🐳 Monitoramento dos containers

Para monitorar containers Docker, pode ser utilizado um exporter específico, como o **cAdvisor**.

Exemplo de arquitetura:

```text
Docker
   │
   ▼
cAdvisor
   │
   ▼
Prometheus
   │
   ▼
Grafana
```

Isso permite acompanhar informações relacionadas aos containers, como utilização de CPU, memória e rede.

## 📂 Estrutura de dados

### Grafana

No host:

```text
~/media-server/monitoring/grafana/config
```

Dentro do container:

```text
/var/lib/grafana
```

Essa pasta contém dados persistentes do Grafana.

### Prometheus

Configuração no host:

```text
~/media-server/monitoring/prometheus/config/prometheus.yml
```

Configuração dentro do container:

```text
/etc/prometheus/prometheus.yml
```

Dados do Prometheus no host:

```text
~/media-server/monitoring/prometheus/data
```

Dados dentro do container:

```text
/prometheus
```

## 🔒 .gitignore

O `.gitignore` deve conter:

```gitignore
grafana/config/
prometheus/data/
```

A configuração do Prometheus pode ser mantida no GitHub:

```text
prometheus/config/prometheus.yml
```

desde que ela não contenha informações sensíveis.

## 🐳 Gerenciamento

### Verificar containers

```bash
docker ps --filter name=prometheus
docker ps --filter name=grafana
```

### Ver logs do Prometheus

```bash
docker logs prometheus
```

### Ver logs do Grafana

```bash
docker logs grafana
```

### Acompanhar logs

```bash
docker logs -f prometheus
```

```bash
docker logs -f grafana
```

### Reiniciar

```bash
docker restart prometheus
docker restart grafana
```

### Parar

```bash
docker compose down
```

### Iniciar

```bash
docker compose up -d
```

### Atualizar

```bash
docker compose pull

docker compose up -d
```

## 🌐 Portas

### Prometheus

```text
Host:      9090
Container: 9090
```

### Grafana

```text
Host:      3000
Container: 3000
```

## 🔐 Configuração

Os dados persistentes ficam em:

```text
~/media-server/monitoring/
```

Estrutura:

```text
monitoring/

├── grafana/
│   └── config/

├── prometheus/
│   ├── config/
│   │   └── prometheus.yml
│   └── data/

├── docker-compose.yml
├── .gitignore
└── README.md
```

Não envie para o GitHub:

```text
grafana/config/
prometheus/data/
```

Esses diretórios podem conter bancos de dados, dashboards, credenciais, métricas e outras informações persistentes.
