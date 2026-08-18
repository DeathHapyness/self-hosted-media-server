<h1 align="center">🖥️ Self-Hosted Media Server</h1>

<div align="center">
  <img src="assets/self-hosted-media-server-banner.png" alt="Self-Hosted Media Server" width="900" height="350" style="border-radius: 20px;">
</div>

<div align="center">

![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![Jellyfin](https://img.shields.io/badge/Jellyfin-Media-00A4DC?logo=jellyfin&logoColor=white)
![Navidrome](https://img.shields.io/badge/Navidrome-Music-1DB954?logo=navidrome&logoColor=white)
![qBittorrent](https://img.shields.io/badge/qBittorrent-Downloads-2F67BA?logo=qbittorrent&logoColor=white)
![Radarr](https://img.shields.io/badge/Radarr-Filmes-FFC230?logo=radarr&logoColor=white)
![Scrutiny](https://img.shields.io/badge/Scrutiny-Disk%20Monitoring-5C6BC0?logo=linux&logoColor=white)
![AdGuard Home](https://img.shields.io/badge/AdGuard%20Home-DNS-68BC71?logo=adguard&logoColor=white)
![Tailscale](https://img.shields.io/badge/Tailscale-VPN-242424?logo=tailscale&logoColor=white)
![Dozzle](https://img.shields.io/badge/Dozzle-Logs-1E90FF?logo=docker&logoColor=white)
![Filebrowser](https://img.shields.io/badge/Filebrowser-Arquivos-4B4B4B?logo=files&logoColor=white)
![Homepage](https://img.shields.io/badge/Homepage-Dashboard-1C1C1C?logo=docker&logoColor=white)

</div>

Servidor de mídia **self-hosted** executado em Linux utilizando Docker e Docker Compose.

O projeto reúne serviços para gerenciamento, organização, monitoramento e reprodução de:

* 🎬 Filmes
* 📺 Séries
* 🎵 Música
* ⬇️ Downloads
* 🛡️ DNS / Bloqueio de anúncios
* 🔗 Acesso remoto
* 📋 Logs dos containers
* 📁 Gerenciamento de arquivos
* 💾 Monitoramento de discos
* 🎛️ Dashboard dos serviços

O objetivo é construir uma infraestrutura de mídia pessoal utilizando **containers, volumes persistentes, organização de arquivos, monitoramento, acesso remoto e serviços independentes**.

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
                  ┌────────────────────┼────────────────────┐
                  │                    │                    │
                  ▼                    ▼                    ▼
              Homepage              Tailscale           Filebrowser
              Dashboard            Acesso remoto        Arquivos
                  │
        ┌─────────┼─────────┬────────────┬────────────┐
        │         │         │            │            │
        ▼         ▼         ▼            ▼            ▼
    Jellyfin  Navidrome  Radarr     qBittorrent   Scrutiny
      :8096      :4533     :7878        :8080        :8081
        │         │         │            │            │
        │         │         │            │            │
        ▼         ▼         ▼            ▼            ▼
      Filmes    Música    Filmes      Downloads     S.M.A.R.T.
      Séries              │
                          │
                          ▼
                    /mnt/media
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
           filmes      series      downloads


                    ┌───────────────┐
                    │  AdGuard Home │
                    │ DNS / Bloqueio│
                    └───────┬───────┘
                            │
                            ▼
                         Rede LAN


                    ┌───────────────┐
                    │    Dozzle     │
                    │ Docker Logs   │
                    └───────────────┘
```

O acesso remoto entre dispositivos é realizado através do **Tailscale**, evitando a necessidade de expor diretamente os serviços administrativos na internet.

---

## 📦 Serviços

| Serviço | Função | Porta | Documentação |
|---|---|---:|---|
| Jellyfin | Filmes e séries | `8096` | [docs/jellyfin.md](docs/jellyfin.md) |
| Navidrome | Servidor de música | `4533` | [docs/navidrome.md](docs/navidrome.md) |
| qBittorrent | Gerenciamento de downloads | `8080` | [docs/qbittorrent.md](docs/qbittorrent.md) |
| Radarr | Gerenciamento automático de filmes | `7878` | [docs/radarr.md](docs/radarr.md) |
| Scrutiny | Monitoramento S.M.A.R.T. dos discos | `8081` | [docs/scrutiny.md](docs/scrutiny.md) |
| AdGuard Home | DNS e bloqueio de anúncios | `53 / 3000` | [docs/adguard.md](docs/adguard.md) |
| Dozzle | Visualização de logs dos containers | `9999` | [docs/dozzle.md](docs/dozzle.md) |
| Filebrowser | Gerenciamento de arquivos | `8081` | [docs/filebrowser.md](docs/filebrowser.md) |
| Homepage | Dashboard dos serviços | `3001` | [docs/homepage.md](docs/homepage.md) |
| Tailscale | VPN e acesso remoto privado | — | [docs/tailscale.md](docs/tailscale.md) |
| Gluetun | VPN para o qBittorrent | — | [docs/gluetun-qbittorrent.md](docs/gluetun-qbittorrent.md) |
| spotDL | Download manual de músicas | CLI | [docs/navidrome.md](docs/navidrome.md) |

> **Nota:** O Scrutiny utiliza a porta `8081` internamente. Caso exista outro serviço utilizando a mesma porta no host, configure uma porta externa diferente no `docker-compose.yml`.

---

## 💾 Armazenamento

A mídia do servidor está organizada em:

```text
/mnt/media/
├── filmes/
├── series/
├── musicas/
└── downloads/
```

Os containers acessam esses diretórios através de volumes Docker.

### Radarr

```text
Host:
  /mnt/media

Container:
  /media
```

Filmes:

```text
Host:
  /mnt/media/filmes

Container:
  /media/filmes
```

Downloads:

```text
Host:
  /mnt/media/downloads

Container:
  /media/downloads
```

### Jellyfin

O Jellyfin utiliza os diretórios de mídia para disponibilizar filmes, séries e outros conteúdos para reprodução.

---

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
* [🎵 Navidrome + spotDL](docs/navidrome.md)
* [📥 qBittorrent](docs/qbittorrent.md)

### 🖥️ Serviços

* [🏠 Homepage](docs/homepage.md)
* [🛡️ AdGuard Home](docs/adguard.md)
* [📋 Dozzle](docs/dozzle.md)
* [📁 Filebrowser](docs/filebrowser.md)
* [💾 Scrutiny](docs/scrutiny.md)
* [🔐 Gluetun + qBittorrent](docs/gluetun-qbittorrent.md)

---

## 🔐 Segurança

Evite expor diretamente serviços administrativos para a internet.

Serviços como:

```text
qBittorrent  → 8080
AdGuard Home → 3000
Dozzle       → 9999
Filebrowser  → porta configurada
Radarr       → 7878
Scrutiny     → porta configurada
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

## 🧠 O que este projeto demonstra

* Linux
* Docker
* Docker Compose
* Containers
* Volumes e persistência de dados
* Administração de servidores
* DNS
* Bloqueio de anúncios
* Gerenciamento de mídia
* Gerenciamento de downloads
* Monitoramento S.M.A.R.T.
* Git e GitHub
* Self-hosting
* Redes
* VPN
* Acesso remoto
* Troubleshooting
* Observabilidade
* Gerenciamento de arquivos

---

## 🗺️ Roadmap

* [ ] Integrar **Sonarr** para gerenciamento automático de séries
* [ ] Configurar **reverse proxy + HTTPS**
* [ ] Implementar **backup automático das configurações**
* [x] Regras de DNS Rewrite no AdGuard Home
* [x] Listas de bloqueio personalizadas no AdGuard Home
* [x] Dashboard centralizado com Homepage
* [x] Gerenciamento automático de filmes com Radarr
* [x] Monitoramento S.M.A.R.T. com Scrutiny

---

## ⚠️ Aviso

Este projeto é destinado a aprendizado, administração de servidores e gerenciamento de mídia que o usuário possui ou tem autorização para armazenar.

O responsável pela implantação deve verificar as leis, licenças e termos de serviço aplicáveis ao conteúdo e aos serviços utilizados.

---

## 📄 Licença

Este projeto é disponibilizado para fins educacionais e de uso pessoal.

Consulte o arquivo [`LICENSE`](LICENSE) para mais informações.