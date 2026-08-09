<h1 align="center">🖥️ Self-Hosted Media Server</h1>

<img src="assets/self-hosted-media-server-banner.png" alt="Self-Hosted Media Server" width="600" align="center">

<div align="center">

![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![Jellyfin](https://img.shields.io/badge/Jellyfin-Media-00A4DC?logo=jellyfin&logoColor=white)
![Navidrome](https://img.shields.io/badge/Navidrome-Music-1DB954?logo=navidrome&logoColor=white)
![qBittorrent](https://img.shields.io/badge/qBittorrent-Downloads-2F67BA?logo=qbittorrent&logoColor=white)
![AdGuard Home](https://img.shields.io/badge/AdGuard%20Home-DNS-68BC71?logo=adguard&logoColor=white)
![Tailscale](https://img.shields.io/badge/Tailscale-VPN-242424?logo=tailscale&logoColor=white)
![Dozzle](https://img.shields.io/badge/Dozzle-Logs-1E90FF?logo=docker&logoColor=white)
![Filebrowser](https://img.shields.io/badge/Filebrowser-Arquivos-4B4B4B?logo=files&logoColor=white)

Servidor de mídia **self-hosted** executado em Linux utilizando Docker e Docker Compose.

O projeto reúne serviços para gerenciamento e reprodução de:

* 🎬 Filmes
* 📺 Séries
* 🎵 Música
* ⬇️ Downloads
* 🛡️ DNS / Bloqueio de anúncios
* 🔗 Acesso remoto
* 📋 Logs dos containers
* 📁 Navegação de arquivos

O objetivo é construir uma infraestrutura de mídia pessoal utilizando **containers, volumes persistentes, organização de arquivos, acesso remoto e serviços independentes**.

---

## INSTALADOR 

Antes de usar leia [`docs/readme-instalador.md`](docs/readme-instalador.md)


 Quer saber exatamente o que o instalador faz (e o que ele deliberadamente não faz)? Veja [`docs/installer-actions.md`](docs/installer-actions.md).


---

## 🏗️ Arquitetura

```text
                         ┌─────────────────┐
                         │     Usuário     │
                         └────────┬────────┘
                                  │
              ┌─────────────┬─────┴────────┬─────────────┬─────────────┬─────────────┐
              │             │              │             │             │             │
              ▼             ▼              ▼             ▼             ▼             ▼
          Jellyfin      Navidrome     qBittorrent    AdGuard Home    Dozzle     Filebrowser
            :8096          :4533          :8080         :3000         :9999         :8081
              │             │              │           (DNS: 53)        │             │
              │             │              ▼                            │             │
              │             │           Downloads                      │             │
              │             │              │                            │             │
              │             │              ▼                            │             │
              │             │       /opt/jellyfin/media                 │             │
              │             │                                           │             │
              │             ▼                                           │             │
              │       /opt/navidrome/music                              │             │
              │                                                         │             │
              ▼                                                         ▼             ▼
        Filmes / Séries                                    logs de todos os    todos os diretórios
                                                                containers          montados
```

Acesso remoto entre dispositivos via **Tailscale**, sem expor portas diretamente na internet.

---

## 📦 Serviços

| Serviço      | Função                        |          Porta | Documentação                               |
| ------------ | ------------------------------ | --------------: | ------------------------------------------- |
| Jellyfin     | Filmes e séries                 |          `8096` | [docs/jellyfin.md](docs/jellyfin.md)         |
| Navidrome    | Servidor de música               |          `4533` | [docs/navidrome.md](docs/navidrome.md)       |
| qBittorrent  | Gerenciamento de downloads       |          `8080` | [docs/qbittorrent.md](docs/qbittorrent.md)   |
| AdGuard Home | DNS e bloqueio de anúncios       | `53` / `3000` | [docs/adguard.md](docs/adguard.md)           |
| Dozzle       | Visualização de logs em tempo real |          `9999` | [docs/dozzle.md](docs/dozzle.md)             |
| Filebrowser  | Navegação e gerenciamento de arquivos |          `8081` | [docs/filebrowser.md](docs/filebrowser.md)   |
| spotDL       | Download manual de músicas       |             CLI | [docs/navidrome.md](docs/navidrome.md)       |
| Tailscale    | Acesso remoto privado            |               — | [docs/tailscale.md](docs/tailscale.md)       |

---

## 📁 Estrutura do Projeto

```text
media-server/
│
├── jellyfin/
│   ├── docker-compose.yml
│   ├── config/
│   ├── cache/
│   └── media/
│
├── qbittorrent/
│   ├── docker-compose.yml
│   └── config/
│
├── navidrome/
│   ├── docker-compose.yml
│   ├── data/
│   └── music/
│
├── adguard/
│   ├── docker-compose.yml
│   ├── work/
│   └── conf/
│
├── dozzle/
│   └── docker-compose.yml
│
├── filebrowser/
│   ├── docker-compose.yml
│   ├── database.db
│   └── config/
│
├── docs/
│   ├── docker.md
│   ├── jellyfin.md
│   ├── navidrome.md
│   ├── qbittorrent.md
│   ├── adguard.md
│   ├── dozzle.md
│   ├── filebrowser.md
│   ├── tailscale.md
│   └── troubleshooting.md
│
├── .gitignore
└── README.md
```

> As pastas de configuração, banco de dados, cache e mídia não são versionadas — ver [.gitignore](.gitignore).

---

## 📋 Requisitos

* CPU x86_64
* 4 GB de RAM ou mais
* SSD para o sistema e containers
* HD/SSD para armazenamento de mídia
* Conexão de rede local
* Linux + Docker + Docker Compose

Guia completo de instalação do Docker: [docs/docker.md](docs/docker.md)

---

## 📚 Documentação detalhada

* [🐳 Instalação do Docker](docs/docker.md)
* [🎬 Jellyfin](docs/jellyfin.md)
* [🎵 Navidrome + spotDL](docs/navidrome.md)
* [📥 qBittorrent](docs/qbittorrent.md)
* [🛡️ AdGuard Home](docs/adguard.md)
* [📋 Dozzle](docs/dozzle.md)
* [📁 Filebrowser](docs/filebrowser.md)
* [🔗 Acesso remoto com Tailscale](docs/tailscale.md)
* [🔍 Verificação e Troubleshooting](docs/troubleshooting.md)

---

## 🔐 Segurança

Evite expor diretamente serviços administrativos para a internet (`8080` qBittorrent, `3000` AdGuard Home, `9999` Dozzle, `8081` Filebrowser). Prefira Tailscale, VPN, reverse proxy com HTTPS e firewall.

O Dozzle expõe logs de todos os containers via socket Docker (`/var/run/docker.sock`) — trate o acesso a ele com o mesmo cuidado dado ao próprio Docker. O Filebrowser tem acesso de leitura/escrita aos diretórios montados — altere a senha padrão (`admin/admin`) no primeiro acesso.

Nunca versionar: senhas, tokens, API keys, `.env`, `config/`, `data/`, `cache/`, `work/`, config real do AdGuard, bancos de dados, mídia.

---

## 🧠 O que este projeto demonstra

Linux · Docker · Docker Compose · Volumes e persistência de dados · Administração de servidores · DNS e bloqueio de anúncios · Git/GitHub · Self-hosting · Redes e acesso remoto · Troubleshooting · Observabilidade (logs) · Gerenciamento de arquivos via web

---

## 🗺️ Roadmap

* [ ] Integrar **Lidarr** para gerenciamento automático da biblioteca musical
* [ ] Configurar **reverse proxy + HTTPS**
* [ ] Implementar **backup das configurações**
* [x] Regras de DNS rewrite / listas de bloqueio personalizadas no AdGuard Home
* [ ] Substituir o Filebrowser por solução própria com monitor de discos integrado (forck do file browser)

---

## ⚠️ Aviso

Este projeto é destinado a aprendizado, administração de servidores e gerenciamento de mídia que o usuário possui ou tem autorização para armazenar. O responsável pela implantação deve verificar as leis e os termos de serviço aplicáveis ao conteúdo utilizado.