# 🖥️ Self-Hosted Media Server

Projeto de infraestrutura **self-hosted** desenvolvido para estudo e demonstração de conhecimentos em **Linux, Docker, Docker Compose, gerenciamento de containers, armazenamento persistente e serviços de mídia**.

O servidor reúne diferentes serviços em containers, permitindo centralizar filmes, séries, músicas e outros conteúdos em uma infraestrutura própria.

---

## 🚀 Sobre o projeto

A proposta deste projeto é construir e documentar um servidor pessoal utilizando tecnologias open source.

Atualmente, o servidor possui:

* 🎬 **Jellyfin** — servidor de filmes e séries
* 🎵 **Navidrome** — servidor de música
* 📥 **qBittorrent** — gerenciamento de downloads
* 📷 **Immich** — gerenciamento de fotos e vídeos

O projeto está sendo desenvolvido de forma incremental, adicionando novos serviços conforme a infraestrutura evolui.

---

## 🏗️ Arquitetura

```text
                         Self-Hosted Server
                                │
                    ┌───────────┴───────────┐
                    │                       │
                 Docker                Storage
                    │                       │
       ┌────────────┼────────────┐          │
       │            │            │          │
       ▼            ▼            ▼          ▼
   Jellyfin    Navidrome    qBittorrent   Media
       │            │            │
       │            │            │
       ▼            ▼            ▼
    Filmes       Música       Downloads
    Séries
```

---

## 🐳 Stack

| Tecnologia     | Função                          |
| -------------- | ------------------------------- |
| Linux          | Sistema operacional do servidor |
| Docker         | Containerização                 |
| Docker Compose | Gerenciamento dos serviços      |
| Jellyfin       | Servidor de mídia               |
| Navidrome      | Servidor de música              |
| qBittorrent    | Gerenciamento de downloads      |
| Immich         | Gerenciamento de fotos e vídeos |
| Git            | Versionamento                   |
| GitHub         | Documentação e código           |

---

## 📦 Serviços

### 🎬 Jellyfin

O Jellyfin é responsável pelo gerenciamento e reprodução da biblioteca de filmes e séries.

Porta:

```text
8096
```

Acesso:

```text
http://SERVER_IP:8096
```

Estrutura de mídia:

```text
/media
├── filmes
└── series
```

---

### 🎵 Navidrome

O Navidrome é utilizado como servidor de música.

Porta:

```text
4533
```

Acesso:

```text
http://SERVER_IP:4533
```

Estrutura:

```text
music/
├── Artista/
│   ├── Álbum/
│   │   ├── 01 - Música.mp3
│   │   └── 02 - Música.mp3
│   └── Outro Álbum/
│       └── 01 - Música.mp3
```

---

### 📥 qBittorrent

O qBittorrent é utilizado para gerenciamento dos downloads.

O armazenamento é separado por categorias:

```text
/downloads
├── filmes
└── series
```

Os dados são armazenados em volumes persistentes para que os containers possam ser recriados sem perder os arquivos.

---

### 📷 Immich

O Immich é utilizado para gerenciamento e armazenamento de fotos e vídeos pessoais.

A instalação do Immich possui seu próprio Docker Compose e não compartilha as configurações internas com os demais serviços.

---

## 📁 Estrutura do projeto

```text
media-server/
│
├── jellyfin/
│   └── docker-compose.yml
│
├── navidrome/
│   └── docker-compose.yml
│
├── qbittorrent/
│   └── docker-compose.yml
│
├── .gitignore
│
└── README.md
```

As configurações internas dos containers, bancos de dados, cache e arquivos de mídia **não são versionados no Git**.

---

## 💾 Persistência

Os containers utilizam volumes persistentes para manter os dados mesmo quando um container é recriado.

Exemplo do Jellyfin:

```text
./config → /config
./cache  → /cache
./media  → /media
```

Exemplo do Navidrome:

```text
./data  → /data
./music → /music
```

---

## 🔐 Segurança

Informações sensíveis não fazem parte do repositório.

O `.gitignore` impede o versionamento de:

```text
config/
cache/
data/
media/
music/
.env
*.secret
*.key
*.pem
*.log
```

Credenciais, tokens e dados pessoais devem permanecer somente no servidor.

---

## 🧰 Gerenciamento

Os serviços são executados através do Docker Compose.

Exemplo:

```bash
docker compose up -d
```

Para visualizar os containers:

```bash
docker ps
```

Para acompanhar os logs:

```bash
docker logs -f CONTAINER_NAME
```

Para atualizar um serviço:

```bash
docker compose pull
docker compose up -d
```

---

## 📚 Conhecimentos praticados

Este projeto está sendo utilizado para desenvolver conhecimentos em:

* Linux
* Administração de servidores
* Docker
* Docker Compose
* Containers
* Volumes persistentes
* Redes
* Gerenciamento de serviços
* Armazenamento
* Git
* GitHub
* Documentação de infraestrutura
* Self-hosting

---

## 🔮 Roadmap

### Infraestrutura

* [x] Docker
* [x] Docker Compose
* [x] Jellyfin
* [x] qBittorrent
* [x] Navidrome
* [ ] Immich
* [ ] Reverse Proxy
* [ ] HTTPS
* [ ] DNS
* [ ] Monitoramento
* [ ] Backup automático

### Automação

* [ ] Radarr
* [ ] Sonarr
* [ ] Organização automática de mídia
* [ ] Atualização automática dos serviços

### Monitoramento

* [ ] Prometheus
* [ ] Grafana
* [ ] Uptime Kuma
* [ ] Alertas de indisponibilidade

### Backup

* [ ] Backup das configurações
* [ ] Backup dos bancos de dados
* [ ] Backup externo
* [ ] Teste de restauração

---

## 🎯 Objetivo

O objetivo principal é transformar este servidor em um laboratório prático de **Linux, Docker, infraestrutura e DevOps**, documentando a evolução da arquitetura e os conhecimentos adquiridos durante o desenvolvimento.

O projeto continuará sendo expandido conforme novos serviços e práticas de infraestrutura forem implementados.

---

## 👨‍💻 Autor

Projeto desenvolvido como laboratório pessoal de infraestrutura, Linux, Docker e serviços self-hosted.

⭐ Projeto em evolução.
