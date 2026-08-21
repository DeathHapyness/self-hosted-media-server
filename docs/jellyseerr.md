# 🎬 Jellyseerr

O Jellyseerr é utilizado para gerenciamento de **solicitações de filmes e séries**.

Ele fornece uma interface onde os usuários podem pesquisar um catálogo de filmes e séries e solicitar conteúdos. Essas solicitações podem ser encaminhadas automaticamente para o **Radarr** e **Sonarr**, que ficam responsáveis pelo gerenciamento do conteúdo.

O Jellyseerr pode trabalhar em conjunto com:

* Jellyfin
* Radarr
* Sonarr
* Prowlarr
* qBittorrent

## 📁 Estrutura

```text
jellyseerr/

├── config/
├── docker-compose.yml
├── .gitignore
└── README.md
```

A pasta `config/` contém os dados internos do Jellyseerr e não deve ser enviada para o GitHub.

## Criar diretórios

```bash
mkdir -p ~/media-server/jellyseerr

cd ~/media-server/jellyseerr

mkdir -p config
```

## Docker Compose

```yaml
services:

  jellyseerr:

    image: fallenbagel/jellyseerr:latest

    container_name: jellyseerr

    restart: unless-stopped

    environment:

      - LOG_LEVEL=info

      - TZ=America/Sao_Paulo

    volumes:

      - ./config:/app/config

    ports:

      - "5055:5055"
```

## Iniciar

```bash
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

O Jellyseerr utiliza a porta `5055`.

```text
http://IP_DO_SERVIDOR:5055
```

Exemplo:

```text
http://IP_DO_SERVIDOR:5055
```

> Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor. Evite colocar IPs pessoais ou públicos reais no GitHub.

## 🔐 Configuração inicial

Ao acessar o Jellyseerr pela primeira vez, será apresentado o assistente de configuração.

O primeiro passo é conectar o Jellyfin.

```text
Jellyfin
→ Connect
```

Informe o endereço do Jellyfin.

Se estiver utilizando Docker e os containers estiverem na mesma rede:

```text
http://jellyfin:8096
```

Caso o Jellyfin esteja fora da rede Docker:

```text
http://IP_DO_SERVIDOR:8096
```

Depois, faça login com uma conta do Jellyfin.

## 🔗 Integração com Radarr

O Jellyseerr utiliza o Radarr para processar solicitações de filmes.

No Jellyseerr:

```text
Settings
→ Services
→ Radarr
→ Add Radarr Server
```

Exemplo:

```text
Name:
Radarr

Hostname:
radarr

Port:
7878

API Key:
SUA_API_KEY
```

Se os containers estiverem na mesma rede Docker, prefira:

```text
http://radarr:7878
```

Depois:

```text
Test
→ Save
```

## 🔗 Integração com Sonarr

Para séries:

```text
Settings
→ Services
→ Sonarr
→ Add Sonarr Server
```

Exemplo:

```text
Name:
Sonarr

Hostname:
sonarr

Port:
8989

API Key:
SUA_API_KEY
```

Se estiver na mesma rede Docker:

```text
http://sonarr:8989
```

Depois:

```text
Test
→ Save
```

## 📚 Catálogo

O Jellyseerr permite que os usuários pesquisem:

```text
Filmes
Séries
Atores
Diretores
Gêneros
```

O usuário pode selecionar um título e solicitar:

```text
Request
```

Dependendo da configuração, o pedido pode ser processado automaticamente.

## 📥 Fluxo de solicitação

A arquitetura funciona aproximadamente assim:

```text
┌───────────────┐
│    Usuário    │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│  Jellyseerr   │
│    Catálogo   │
└───────┬───────┘
        │
        ├─────────────────┐
        ▼                 ▼
   ┌─────────┐       ┌─────────┐
   │ Radarr  │       │ Sonarr  │
   │ Filmes  │       │ Séries  │
   └────┬────┘       └────┬────┘
        │                 │
        └────────┬────────┘
                 ▼
          ┌──────────────┐
          │   Prowlarr   │
          │  Indexadores  │
          └───────┬──────┘
                  ▼
          ┌──────────────┐
          │ qBittorrent  │
          │   Download   │
          └───────┬──────┘
                  ▼
             /mnt/media
                  │
                  ▼
          ┌──────────────┐
          │   Jellyfin   │
          │    Biblioteca│
          └──────────────┘
```

O Jellyseerr **não realiza os downloads diretamente**. Ele funciona como uma interface de solicitação e encaminha os pedidos para Radarr ou Sonarr.

## 🎬 Solicitação de filme

Exemplo:

```text
Usuário
   ↓
Pesquisa "Interstellar"
   ↓
Seleciona o filme
   ↓
Request
   ↓
Jellyseerr
   ↓
Radarr
   ↓
Prowlarr
   ↓
Indexador
   ↓
Radarr
   ↓
qBittorrent
   ↓
Download
   ↓
Radarr organiza
   ↓
Jellyfin
```

## 📺 Solicitação de série

Para séries:

```text
Usuário
   ↓
Pesquisa a série
   ↓
Seleciona temporadas
   ↓
Request
   ↓
Jellyseerr
   ↓
Sonarr
   ↓
Prowlarr
   ↓
Indexador
   ↓
qBittorrent
   ↓
Download
   ↓
Sonarr organiza
   ↓
Jellyfin
```

## 👥 Usuários

O Jellyseerr pode utilizar as contas existentes no Jellyfin para autenticação.

Isso permite que diferentes usuários tenham acesso ao catálogo sem precisar criar uma conta completamente separada.

As permissões e solicitações podem ser gerenciadas através do próprio Jellyseerr.

## 📂 Configuração

A configuração fica em:

```text
~/media-server/jellyseerr/config
```

Dentro do container:

```text
/app/config
```

Essa pasta contém os dados persistentes do Jellyseerr.

## 🔒 .gitignore

O `.gitignore` deve conter:

```gitignore
config/
```

Não envie a pasta `config/` para o GitHub.

Ela pode conter:

```text
Banco de dados
Configurações
Sessões
Dados dos usuários
Configurações de integração
```

## 🐳 Gerenciamento

### Verificar container

```bash
docker ps --filter name=jellyseerr
```

### Ver logs

```bash
docker logs jellyseerr
```

### Acompanhar logs

```bash
docker logs -f jellyseerr
```

### Reiniciar

```bash
docker restart jellyseerr
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

## 🔗 Stack completa

A estrutura final do servidor pode ficar:

```text
                         ┌───────────────┐
                         │  Jellyseerr   │
                         │   Catálogo    │
                         └───────┬───────┘
                                 │
                    ┌────────────┴────────────┐
                    ▼                         ▼
              ┌──────────┐              ┌──────────┐
              │  Radarr  │              │  Sonarr  │
              │  Filmes  │              │  Séries  │
              └────┬─────┘              └────┬─────┘
                   │                         │
                   └───────────┬─────────────┘
                               ▼
                        ┌──────────────┐
                        │   Prowlarr   │
                        │  Indexadores │
                        └──────┬───────┘
                               ▼
                        ┌──────────────┐
                        │ qBittorrent  │
                        │   Download   │
                        └──────┬───────┘
                               ▼
                          /mnt/media
                               │
                               ▼
                        ┌──────────────┐
                        │   Jellyfin   │
                        │   Biblioteca │
                        └──────────────┘
```

## 🌐 Porta

O Jellyseerr utiliza a porta `5055`.

```text
Host:      5055
Container: 5055
```
