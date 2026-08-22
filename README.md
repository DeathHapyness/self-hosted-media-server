<h1 align="center">🖥️ Self-Hosted Media Server</h1>

<div align="center">

  <img src="assets/self-hosted-media-server-banner.png" alt="Self-Hosted Media Server" width="900" height="350" style="border-radius: 20px;">

</div>

<div align="center">

![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker\&logoColor=white)
![Jellyfin](https://img.shields.io/badge/Jellyfin-Media-00A4DC?logo=jellyfin\&logoColor=white)
![Navidrome](https://img.shields.io/badge/Navidrome-Music-1DB954?logo=navidrome\&logoColor=white)
![qBittorrent](https://img.shields.io/badge/qBittorrent-Downloads-2F67BA?logo=qbittorrent\&logoColor=white)
![Radarr](https://img.shields.io/badge/Radarr-Filmes-FFC230?logo=radarr\&logoColor=white)
![Sonarr](https://img.shields.io/badge/Sonarr-Séries-35C5F0?logo=sonarr\&logoColor=white)
![Prowlarr](https://img.shields.io/badge/Prowlarr-Indexadores-5C5C5C?logo=prowlarr\&logoColor=white)
![Jellyseerr](https://img.shields.io/badge/Jellyseerr-Requests-8B5CF6?logo=jellyfin\&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-Monitoring-E6522C?logo=prometheus\&logoColor=white)
![Grafana](https://img.shields.io/badge/Grafana-Dashboards-F46800?logo=grafana\&logoColor=white)
![Scrutiny](https://img.shields.io/badge/Scrutiny-Disk%20Monitoring-5C6BC0?logo=linux\&logoColor=white)
![AdGuard Home](https://img.shields.io/badge/AdGuard%20Home-DNS-68BC71?logo=adguard\&logoColor=white)
![Tailscale](https://img.shields.io/badge/Tailscale-VPN-242424?logo=tailscale\&logoColor=white)
![Dozzle](https://img.shields.io/badge/Dozzle-Logs-1E90FF?logo=docker\&logoColor=white)
![Filebrowser](https://img.shields.io/badge/Filebrowser-Arquivos-4B4B4B?logo=files\&logoColor=white)
![Homepage](https://img.shields.io/badge/Homepage-Dashboard-1C1C1C?logo=docker\&logoColor=white)

</div>

Servidor de mídia **self-hosted** executado em Linux utilizando Docker e Docker Compose.

O projeto reúne serviços para gerenciamento, organização, monitoramento e reprodução de:

* 🎬 Filmes
* 📺 Séries
* 🎵 Música
* ⬇️ Downloads
* 🔎 Indexadores
* 📋 Solicitações de filmes e séries
* 🛡️ DNS / Bloqueio de anúncios
* 🔗 Acesso remoto
* 📊 Monitoramento e métricas
* 📈 Dashboards
* 📋 Logs dos containers
* 📁 Gerenciamento de arquivos
* 💾 Monitoramento de discos
* 🎛️ Dashboard dos serviços

O objetivo é construir uma infraestrutura de mídia pessoal utilizando **containers, volumes persistentes, organização de arquivos, automação, monitoramento, acesso remoto e serviços independentes**.

---

## 🚀 INSTALADOR

Antes de usar, leia:

[`docs/readme-instalador.md`](docs/readme-instalador.md)

Para entender exatamente o que o instalador faz e o que ele deliberadamente não faz:

[`docs/installer-actions.md`](docs/installer-actions.md)

---

## 🏗️ Arquitetura

```text
                              ┌─────────────────┐
                              │     Usuário     │
                              └────────┬────────┘
                                       │
              ┌────────────────────────┼────────────────────────┐
              │                        │                        │
              ▼                        ▼                        ▼
          Jellyseerr               Homepage                Filebrowser
          Solicitações             Dashboard                Arquivos
              │
              │
       ┌──────┴──────┐
       │             │
       ▼             ▼
    Radarr         Sonarr
    Filmes         Séries
       │             │
       └──────┬──────┘
              │
              ▼
         ┌─────────────┐
         │  Prowlarr   │
         │ Indexadores  │
         └──────┬──────┘
                │
                ▼
         ┌─────────────┐
         │ qBittorrent │
         │  Downloads  │
         └──────┬──────┘
                │
                ▼
            /mnt/media
                │
       ┌────────┼─────────┐
       │        │         │
       ▼        ▼         ▼
    filmes    series    downloads
       │        │
       └────┬───┘
            │
            ▼
        ┌─────────┐
        │ Jellyfin│
        │  Media  │
        └─────────┘


        ┌──────────────────────┐
        │      Monitoring      │
        │                      │
        │ Prometheus + Grafana │
        └──────────┬───────────┘
                   │
                   ▼
              Métricas
                   │
       ┌───────────┼───────────┐
       ▼           ▼           ▼
     Docker      Sistema     Serviços


        ┌──────────────────────┐
        │    AdGuard Home      │
        │   DNS / Bloqueio     │
        └──────────┬───────────┘
                   │
                   ▼
                Rede LAN


        ┌──────────────────────┐
        │       Tailscale      │
        │    Acesso remoto     │
        └──────────────────────┘


        ┌──────────────────────┐
        │        Dozzle        │
        │    Docker Logs       │
        └──────────────────────┘
```

O **Jellyseerr** funciona como interface de solicitação para filmes e séries.

O **Radarr** gerencia filmes e o **Sonarr** gerencia séries.

O **Prowlarr** centraliza os indexadores utilizados pelo Radarr e Sonarr.

O **qBittorrent** funciona como cliente de download.

O **Jellyfin** disponibiliza a mídia organizada para reprodução.

O monitoramento é realizado através do **Prometheus + Grafana**, enquanto o **Scrutiny** é utilizado para monitoramento S.M.A.R.T. dos discos.

O acesso remoto entre dispositivos é realizado através do **Tailscale**, evitando a necessidade de expor diretamente os serviços administrativos na internet.

---

## 📦 Serviços

| Serviço      | Função                                |       Porta | Documentação                                               |
| ------------ | ------------------------------------- | ----------: | ---------------------------------------------------------- |
| Jellyfin     | Filmes e séries                       |      `8096` | [docs/jellyfin.md](docs/jellyfin.md)                       |
| Navidrome    | Servidor de música                    |      `4533` | [docs/navidrome.md](docs/navidrome.md)                     |
| qBittorrent  | Gerenciamento de downloads            |      `8080` | [docs/qbittorrent.md](docs/qbittorrent.md)                 |
| Radarr       | Gerenciamento automático de filmes    |      `7878` | [docs/radarr.md](docs/radarr.md)                           |
| Sonarr       | Gerenciamento automático de séries    |      `8989` | [docs/sonarr.md](docs/sonarr.md)                           |
| Prowlarr     | Gerenciamento de indexadores          |      `9696` | [docs/prowlarr.md](docs/prowlarr.md)                       |
| Jellyseerr   | Catálogo e solicitações               |      `5055` | [docs/jellyseerr.md](docs/jellyseerr.md)                   |
| Prometheus   | Coleta e armazenamento de métricas    |      `9090` | [docs/monitoring.md](docs/monitoring.md)                   |
| Grafana      | Dashboards e visualização de métricas |      `3000` | [docs/monitoring.md](docs/monitoring.md)                   |
| Scrutiny     | Monitoramento S.M.A.R.T. dos discos   |      `8081` | [docs/scrutiny.md](docs/scrutiny.md)                       |
| AdGuard Home | DNS e bloqueio de anúncios            | `53 / 3000` | [docs/adguard.md](docs/adguard.md)                         |
| Dozzle       | Visualização de logs dos containers   |      `9999` | [docs/dozzle.md](docs/dozzle.md)                           |
| Filebrowser  | Gerenciamento de arquivos             |      `8081` | [docs/filebrowser.md](docs/filebrowser.md)                 |
| Homepage     | Dashboard dos serviços                |      `3001` | [docs/homepage.md](docs/homepage.md)                       |
| Tailscale    | VPN e acesso remoto privado           |           — | [docs/tailscale.md](docs/tailscale.md)                     |
| Gluetun      | VPN para o qBittorrent                |           — | [docs/gluetun-qbittorrent.md](docs/gluetun-qbittorrent.md) |
| spotDL       | Download manual de músicas            |         CLI | [docs/navidrome.md](docs/navidrome.md)                     |

> **Nota:** algumas portas acima podem entrar em conflito dependendo da configuração dos containers.

> O **Grafana** normalmente utiliza `3000`, enquanto o **AdGuard Home** também pode utilizar `3000` para sua interface web.

> O **Scrutiny** e o **Filebrowser** também não devem utilizar a mesma porta externa do host.

> Caso exista conflito, altere a porta externa no `docker-compose.yml`. A porta interna do container pode permanecer inalterada.


---

## 🔄 Fluxo de mídia

### 🎬 Filmes

```text
Jellyseerr
    │
    ▼
  Radarr
    │
    ▼
 Prowlarr
    │
    ▼
qBittorrent
    │
    ▼
/mnt/media/downloads
    │
    ▼
  Radarr
    │
    ▼
/mnt/media/filmes
    │
    ▼
 Jellyfin
```

## 📋 Requisitos

* CPU x86_64
* 4 GB de RAM ou mais
* SSD para o sistema e containers
* HD/SSD para armazenamento de mídia
* Conexão de rede local
* Linux
* Docker
* Docker Compose

Guia completo de instalação do Docker:

[docs/docker.md](docs/docker.md)

---

## 📚 Documentação

### 🐳 Infraestrutura

* [🐳 Instalação do Docker](docs/docker.md)
* [🔗 Tailscale](docs/tailscale.md)
* [🔍 Troubleshooting](docs/troubleshooting.md)

### 🎬 Mídia

* [🎬 Jellyfin](docs/jellyfin.md)
* [🎬 Radarr](docs/radarr.md)
* [📺 Sonarr](docs/sonarr.md)
* [🔎 Prowlarr](docs/prowlarr.md)
* [🎬 Jellyseerr](docs/jellyseerr.md)
* [🎵 Navidrome + spotDL](docs/navidrome.md)
* [📥 qBittorrent](docs/qbittorrent.md)

### 📊 Monitoramento

* [📊 Monitoring — Grafana + Prometheus](docs/monitoring.md)
* [💾 Scrutiny](docs/scrutiny.md)

### 🖥️ Serviços

* [🏠 Homepage](docs/homepage.md)
* [🛡️ AdGuard Home](docs/adguard.md)
* [📋 Dozzle](docs/dozzle.md)
* [📁 Filebrowser](docs/filebrowser.md)
* [🔐 Gluetun + qBittorrent](docs/gluetun-qbittorrent.md)

---

## 🔐 Segurança

Evite expor diretamente serviços administrativos para a internet.

Serviços como:

```text
qBittorrent   → 8080
AdGuard Home  → 3000
Dozzle        → 9999
Filebrowser   → porta configurada
Radarr        → 7878
Sonarr        → 8989
Prowlarr      → 9696
Jellyseerr    → 5055
Prometheus    → 9090
Grafana       → 3000
Scrutiny      → porta configurada
```

devem preferencialmente ser acessados através de:

* Tailscale
* VPN
* Reverse proxy com HTTPS
* Firewall
* Rede local

O Dozzle utiliza o socket Docker:

```text
/var/run/docker.sock
```

Portanto, o acesso ao Dozzle deve ser tratado com o mesmo cuidado dado ao acesso administrativo do Docker.

O Filebrowser possui acesso aos diretórios montados e pode realizar alterações nos arquivos. Proteja adequadamente o acesso ao serviço.

Radarr, Sonarr, Prowlarr e Jellyseerr também possuem acesso a APIs e/ou bibliotecas de mídia e devem ser protegidos adequadamente.

Nunca versionar:

```text
.env
*.env
*.key
*.pem
*.token
config/
data/
cache/
work/
*.db
*.sqlite
*.sqlite3
*.log
```

Também não devem ser versionados:

* Senhas
* Tokens
* API Keys
* Credenciais
* Bancos de dados
* Configurações privadas
* Mídia pessoal

---


## ⚠️ Aviso de responsabilidade

Este projeto é fornecido **"como está"**, exclusivamente para fins educacionais e de uso pessoal.

Os responsáveis por este projeto **não se responsabilizam por danos, perda de dados, falhas de hardware, problemas de software, indisponibilidade de serviços ou qualquer outro dano** que possa ocorrer durante a instalação, configuração ou utilização deste projeto.

Antes de executar qualquer script, comando ou alteração no sistema, recomenda-se:

* Realizar backup dos dados importantes.
* Verificar os comandos antes de executá-los.
* Confirmar as configurações de armazenamento e volumes.
* Utilizar o projeto em um ambiente de teste quando possível.
* Manter o sistema e os serviços atualizados.

A utilização deste projeto é de **inteira responsabilidade do usuário**.

Ao utilizar este projeto, o usuário reconhece que possui conhecimento suficiente para administrar seu próprio sistema e assume os riscos relacionados às alterações realizadas na máquina.


---

## 📄 Licença

Este projeto é disponibilizado para fins educacionais e de uso pessoal.

Consulte o arquivo [`LICENSE`](LICENSE) para mais informações.
