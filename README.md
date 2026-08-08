# 🖥️ Self-Hosted Media Server

![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge\&logo=linux\&logoColor=black)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge\&logo=docker\&logoColor=white)
![Docker Compose](https://img.shields.io/badge/Docker%20Compose-2496ED?style=for-the-badge\&logo=docker\&logoColor=white)
![Jellyfin](https://img.shields.io/badge/Jellyfin-00A4DC?style=for-the-badge\&logo=jellyfin\&logoColor=white)
![Navidrome](https://img.shields.io/badge/Navidrome-000000?style=for-the-badge)
![qBittorrent](https://img.shields.io/badge/qBittorrent-2F67BA?style=for-the-badge\&logo=qbittorrent\&logoColor=white)
![Tailscale](https://img.shields.io/badge/Tailscale-4A4DE6?style=for-the-badge\&logo=tailscale\&logoColor=white)
![GitHub](https://img.shields.io/badge/GitHub-181717?style=for-the-badge\&logo=github\&logoColor=white)

Servidor de mídia **self-hosted** executado em Linux utilizando Docker e Docker Compose.

O projeto reúne serviços para gerenciamento e reprodução de:

* 🎬 Filmes
* 📺 Séries
* 🎵 Música
* ⬇️ Downloads
* 🔗 Acesso remoto

O objetivo é construir uma infraestrutura de mídia pessoal utilizando **containers, volumes persistentes, organização de arquivos, acesso remoto e serviços independentes**.

---

## 📑 Sumário

* [🏗️ Arquitetura](#️-arquitetura)
* [📋 Requisitos](#-requisitos)
* [🐳 Instalação do Docker](#-instalação-do-docker)
* [👤 Usar Docker sem sudo](#-usar-docker-sem-sudo)
* [🔗 Acesso Remoto com Tailscale](#-acesso-remoto-com-tailscale)
* [🗺️ Roadmap](#️-roadmap)
* [⚠️ Aviso](#️-aviso)

---

# 🏗️ Arquitetura

```text
                         ┌─────────────────┐
                         │     Usuário     │
                         └────────┬────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │             │             │
                    ▼             ▼             ▼
                Jellyfin      Navidrome     qBittorrent
                  :8096          :4533          :8080
                    │             │             │
                    │             │             ▼
                    │             │          Downloads
                    │             │             │
                    │             │             ▼
                    │             │      /opt/jellyfin/media
                    │             │
                    │             ▼
                    │      /opt/navidrome/music
                    │
                    ▼
              Filmes / Séries


```

### Serviços

| Serviço     | Função                     |  Porta |
| ----------- | -------------------------- | -----: |
| Jellyfin    | Filmes e séries            | `8096` |
| Navidrome   | Servidor de música         | `4533` |
| qBittorrent | Gerenciamento de downloads | `8080` |
| spotDL      | Download manual de músicas |    CLI |
| Tailscale   | Acesso remoto privado      |      — |

---

# 📋 Requisitos

## Hardware

O projeto pode ser executado em um computador antigo, mini PC, notebook ou servidor dedicado.

Recomendação inicial:

* CPU x86_64
* 4 GB de RAM ou mais
* SSD para o sistema e containers
* HD/SSD para armazenamento de mídia
* Conexão de rede local

O armazenamento necessário depende principalmente da quantidade de filmes, séries e músicas.

---

# 🐧 Sistema Operacional

O projeto foi desenvolvido para Linux.

Exemplo de ambiente:

```text
Linux
Docker
Docker Compose
```

O Docker Engine possui suporte oficial para diversas distribuições Linux.

---

# 🐳 Instalação do Docker

## 1. Atualizar o sistema

Em Ubuntu/Debian:

```bash
sudo apt update
sudo apt upgrade -y
```

## 2. Instalar dependências

```bash
sudo apt install -y ca-certificates curl
```

## 3. Adicionar a chave oficial do Docker

```bash
sudo install -m 0755 -d /etc/apt/keyrings

sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  -o /etc/apt/keyrings/docker.asc

sudo chmod a+r /etc/apt/keyrings/docker.asc
```

## 4. Adicionar o repositório

```bash
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF
```

Atualize:

```bash
sudo apt update
```

## 5. Instalar Docker

```bash
sudo apt install -y \
  docker-ce \
  docker-ce-cli \
  containerd.io \
  docker-buildx-plugin \
  docker-compose-plugin
```

## 6. Testar

```bash
sudo systemctl status docker
```

```bash
sudo docker run hello-world
```

Verifique o Compose:

```bash
docker compose version
```

---

# 👤 Usar Docker sem sudo

Adicione seu usuário ao grupo Docker:

```bash
sudo usermod -aG docker $USER
```

Depois encerre a sessão e entre novamente.

Teste:

```bash
docker ps
```

---

# 📁 Estrutura do Projeto

```text
media-server/
│
├── jellyfin/
│   ├── docker-compose.yml
│   ├── config/
│   ├── cache/
│   └── media/
│       ├── filmes/
│       └── series/
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
├── .gitignore
└── README.md
```

> As pastas de configuração, banco de dados, cache e mídia não devem ser enviadas para o GitHub.

---

# 🎬 Jellyfin

O Jellyfin é responsável pela biblioteca de **filmes e séries**.

## Criar diretórios

```bash
sudo mkdir -p /opt/jellyfin

cd /opt/jellyfin

mkdir -p config
mkdir -p cache
mkdir -p media/filmes
mkdir -p media/series
```

## Docker Compose

Crie:

```bash
nano docker-compose.yml
```

```yaml
services:
  jellyfin:
    image: jellyfin/jellyfin:latest
    container_name: jellyfin
    restart: unless-stopped

    ports:
      - "8096:8096"

    volumes:
      - ./config:/config
      - ./cache:/cache
      - ./media:/media

    environment:
      - TZ=America/Sao_Paulo
```

Inicie:

```bash
docker compose up -d
```

Acesse:

```text
http://IP_DO_SERVIDOR:8096
```

---

# 📥 qBittorrent

O qBittorrent é utilizado para gerenciamento de downloads.

## Criar diretórios

```bash
sudo mkdir -p /opt/qbittorrent
cd /opt/qbittorrent

mkdir -p config
```

## Docker Compose

```yaml
services:
  qbittorrent:
    image: lscr.io/linuxserver/qbittorrent:latest
    container_name: qbittorrent
    restart: unless-stopped

    environment:
      - PUID=1000
      - PGID=1000
      - TZ=America/Sao_Paulo
      - WEBUI_PORT=8080

    ports:
      - "8080:8080"
      - "6881:6881"
      - "6881:6881/udp"

    volumes:
      - ./config:/config
      - /opt/jellyfin/media:/downloads
```

Inicie:

```bash
docker compose up -d
```

Acesse:

```text
http://IP_DO_SERVIDOR:8080
```

O diretório `/downloads` do qBittorrent corresponde ao diretório de mídia do Jellyfin:

```text
Host:
/opt/jellyfin/media

Container qBittorrent:
/downloads

Container Jellyfin:
/media
```

Isso permite que os arquivos baixados pelo qBittorrent sejam posteriormente encontrados pelo Jellyfin.

---

# 🎵 Navidrome

O Navidrome é responsável pela biblioteca musical.

Ele lê os arquivos presentes em:

```text
/opt/navidrome/music
```

e disponibiliza a biblioteca através de clientes compatíveis.

## Criar diretórios

```bash
sudo mkdir -p /opt/navidrome

cd /opt/navidrome

mkdir -p data
mkdir -p music
```

## Docker Compose

```yaml
services:
  navidrome:
    image: deluan/navidrome:latest
    container_name: navidrome
    restart: unless-stopped

    ports:
      - "4533:4533"

    environment:
      ND_SCANSCHEDULE: 1h
      ND_LOGLEVEL: info
      ND_SESSIONTIMEOUT: 24h
      TZ: America/Sao_Paulo

    volumes:
      - ./data:/data
      - ./music:/music:ro
```

Inicie:

```bash
docker compose up -d
```

Acesse:

```text
http://IP_DO_SERVIDOR:4533
```

---

# 🎧 Clientes Recomendados

O Navidrome funciona como **servidor**, enquanto os aplicativos abaixo funcionam como clientes.

## 💻 PC — Feishin

Para desktop, uma opção recomendada é o **Feishin**.

Ele oferece uma interface moderna para servidores compatíveis com Subsonic, incluindo Navidrome.

[Feishin — GitHub](https://github.com/jeffvli/feishin)

---

## 📱 iPhone / iOS

No iPhone, você pode utilizar um cliente compatível com Navidrome/Subsonic.

### Amperfy

[Amperfy — App Store](https://apps.apple.com/app/amperfy-music-player/id1530145038)

### play:Sub

[play:Sub — App Store](https://apps.apple.com/app/play-sub/id955329386)

---

## 🤖 Android

No Android, uma opção interessante é o **Symfonium**.

Ele suporta Navidrome e outros servidores de música.

[Symfonium — site oficial](https://symfonium.app/)

---

# 🎶 spotDL

O spotDL pode ser utilizado para downloads manuais de músicas.

## Instalar dependências

```bash
sudo apt install -y python3 python3-pip ffmpeg
```

Instale:

```bash
pip install spotdl
```

Verifique:

```bash
spotdl --version
```

---

# 📥 Baixando músicas

Para manter a organização da biblioteca:

```bash
spotdl \
  --output "/opt/navidrome/music/{artist}/{album}/{track-number} - {title}.{output-ext}" \
  "LINK_DO_SPOTIFY"
```

Exemplo:

```bash
spotdl \
  --output "/opt/navidrome/music/{artist}/{album}/{track-number} - {title}.{output-ext}" \
  "https://open.spotify.com/track/ID"
```

Sempre coloque a URL entre aspas.

Isso permite organizar automaticamente:

```text
/opt/navidrome/music/
│
├── Artista/
│   ├── Álbum/
│   │   ├── 01 - Música.mp3
│   │   ├── 02 - Música.mp3
│   │   └── 03 - Música.mp3
│   │
│   └── Outro Álbum/
│
└── Outro Artista/
```

---

# 🔄 Atualizando o Navidrome

Depois de adicionar novas músicas:

```bash
docker exec -it navidrome /app/navidrome scan
```

Também existe um scan automático configurado:

```yaml
ND_SCANSCHEDULE: 1h
```

---

# 🔄 Fluxo da Música

```text
              Spotify URL
                   │
                   ▼
                spotDL
                   │
                   ▼
        /opt/navidrome/music
                   │
          ┌────────┴────────┐
          │                 │
       Artista           Artista
          │                 │
        Álbum             Álbum
          │                 │
        Faixas            Faixas
          │                 │
          └────────┬────────┘
                   ▼
               Navidrome
                   │
          ┌────────┼────────┐
          ▼        ▼        ▼
        Feishin   iOS    Android
          PC
```

---

# 🔍 Verificação e Troubleshooting

## Verificar arquivos no host

```bash
find /opt/navidrome/music -type f
```

Somente MP3:

```bash
find /opt/navidrome/music -type f -name "*.mp3"
```

## Verificar arquivos dentro do Navidrome

```bash
docker exec -it navidrome find /music -type f
```

Se o arquivo estiver no host e dentro do container, o volume está funcionando corretamente.

## Verificar volumes

Jellyfin:

```bash
docker inspect jellyfin \
  --format '{{range .Mounts}}{{println .Source "->" .Destination}}{{end}}'
```

qBittorrent:

```bash
docker inspect qbittorrent \
  --format '{{range .Mounts}}{{println .Source "->" .Destination}}{{end}}'
```

Navidrome:

```bash
docker inspect navidrome \
  --format '{{range .Mounts}}{{println .Source "->" .Destination}}{{end}}'
```

---

# 🔐 Segurança

Evite expor diretamente serviços administrativos para a internet.

Principalmente:

```text
8080 → qBittorrent
```

Para acesso externo, considere utilizar:

* Tailscale
* VPN
* Reverse proxy
* HTTPS
* Autenticação adequada
* Firewall

Nunca coloque no GitHub:

```text
senhas
tokens
API keys
.env
config/
data/
cache/
bancos de dados
mídia
```

---

# 🔗 Acesso Remoto com Tailscale

Para acessar o servidor fora da rede doméstica, o projeto utiliza o **Tailscale** como uma rede privada entre os dispositivos.

Com o Tailscale, não é necessário expor diretamente as portas do Jellyfin, Navidrome, qBittorrent ou SSH na internet.

A arquitetura fica:

```text
                         INTERNET
                             │
                             │
                         Tailscale
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
           📱 iPhone       💻 PC        🖥️ Servidor
                                        100.x.x.x
                                             │
                              ┌──────────────┼──────────────┐
                              │              │              │
                              ▼              ▼              ▼
                           Jellyfin      Navidrome      qBittorrent
                            :8096          :4533           :8080
```

## Instalação

No servidor Linux:

```bash
curl -fsSL https://tailscale.com/install.sh | sh
```

Depois autentique o servidor:

```bash
sudo tailscale up
```

Verifique os dispositivos conectados:

```bash
tailscale status
```

Para descobrir o IP Tailscale do servidor:

```bash
tailscale ip
```

O servidor receberá um endereço da rede `100.x.x.x`.

Exemplo:

```text
100.95.231.76
```

## Acessando os serviços remotamente

Depois de instalar o Tailscale nos dispositivos e entrar na mesma conta, os serviços podem ser acessados através do IP Tailscale do servidor.

### Jellyfin

```text
http://100.95.231.76:8096
```

### Navidrome

```text
http://100.95.231.76:4533
```

### qBittorrent

```text
http://100.95.231.76:8080
```

### SSH

```bash
ssh rique@100.95.231.76
```

## Dispositivos

O Tailscale pode conectar diferentes dispositivos à mesma rede privada:

```text
Servidor
100.95.231.76
      │
      ├── 📱 iPhone
      ├── 💻 PC / Linux
      └── 🤖 Android
```

Basta instalar o Tailscale em cada dispositivo e entrar na mesma conta.

## Vantagens

* 🔒 Não é necessário abrir portas no roteador
* 🌐 Acesso aos serviços fora de casa
* 📱 Acesso pelo celular
* 💻 Acesso pelo computador
* 🔑 Autenticação através da conta Tailscale
* 🛡️ Rede privada entre os dispositivos
* 🖥️ Possibilidade de acessar o servidor via SSH remotamente

> O Tailscale é utilizado neste projeto como camada de acesso remoto à infraestrutura, mantendo os serviços internos sem exposição direta à internet.

---

# 🐙 GitHub

Inicialize o repositório:

```bash
git init
```

Utilize `main`:

```bash
git branch -M main
```

Adicione os arquivos:

```bash
git add .
```

Crie o commit:

```bash
git commit -m "docs: add media server setup"
```

Adicione o repositório:

```bash
git remote add origin URL_DO_SEU_REPOSITORIO
```

Envie:

```bash
git push -u origin main
```

---

# 🧹 .gitignore

Exemplo:

```gitignore
# Docker
*.env

# Jellyfin
jellyfin/config/
jellyfin/cache/
jellyfin/media/

# qBittorrent
qbittorrent/config/

# Navidrome
navidrome/data/
navidrome/music/

# Logs
*.log

# Databases
*.db
*.sqlite
*.sqlite3

# Sistema
.DS_Store
Thumbs.db
```

---

# 🧠 O que este projeto demonstra

Este projeto demonstra conhecimentos em:

* Linux
* Docker
* Docker Compose
* Containers
* Volumes
* Persistência de dados
* Permissões Linux
* Administração de servidores
* Gerenciamento de serviços
* Git/GitHub
* Self-hosting
* Organização de arquivos
* Troubleshooting
* Redes
* Acesso remoto
* Serviços multimídia

---

# 🗺️ Roadmap

* [ ] Integrar **Lidarr** para gerenciamento automático da biblioteca musical
* [ ] Configurar **reverse proxy + HTTPS**
* [ ] Implementar **backup das configurações**

---

# ⚠️ Aviso

Este projeto é destinado a aprendizado, administração de servidores e gerenciamento de mídia que o usuário possui ou tem autorização para armazenar.

O responsável pela implantação deve verificar as leis e os termos de serviço aplicáveis ao conteúdo utilizado.

---

# 👨‍💻 Projeto

Servidor de mídia **self-hosted** baseado em Linux e Docker.

```text
Linux
 │
 └── Docker
      │
      ├── Jellyfin
      │    └── Filmes / Séries
      │
      ├── qBittorrent
      │    └── Downloads
      │
      └── Navidrome
           └── Música
                │
                └── spotDL

Tailscale
    │
    ├── 📱 iPhone
    ├── 💻 PC
    └── 🌐 Acesso remoto
```

O objetivo é demonstrar, de forma prática, a construção e administração de uma infraestrutura de mídia utilizando tecnologias open source, containers, armazenamento persistente e acesso remoto seguro.
