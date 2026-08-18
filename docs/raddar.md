# 🎬 Radarr

O Radarr é utilizado para gerenciamento e organização automática de filmes.

Ele trabalha em conjunto com um cliente de download, como o qBittorrent, permitindo adicionar filmes, acompanhar downloads e organizar automaticamente os arquivos na biblioteca.

## 📁 Estrutura

```text
radarr/
├── config/
├── docker-compose.yml
├── .gitignore
└── README.md
```

A pasta `config/` contém os dados internos do Radarr e não deve ser enviada para o GitHub.

## Criar diretórios

```bash
mkdir -p ~/media-server/radarr
cd ~/media-server/radarr
mkdir -p config
```

## Docker Compose

```yaml
services:
  radarr:
    image: lscr.io/linuxserver/radarr:latest
    container_name: radarr
    restart: unless-stopped

    environment:
      - PUID=1000
      - PGID=1000
      - TZ=America/Sao_Paulo

    volumes:
      - ./config:/config
      - /mnt/media:/media

    ports:
      - "7878:7878"
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

```text
http://IP_DO_SERVIDOR:7878
```

Exemplo:

```text
http://192.168.15.118:7878
```

## 📂 Diretórios de mídia

O Radarr utiliza `/mnt/media` do servidor como `/media` dentro do container.

### Filmes

```text
Host:
/mnt/media/filmes

Container:
/media/filmes
```

### Downloads

```text
Host:
/mnt/media/downloads

Container:
/media/downloads
```

## 🔗 Integração com qBittorrent

No Radarr, acesse:

```text
Settings
→ Download Clients
→ Add
→ qBittorrent
```

Exemplo:

```text
Host: 192.168.15.118
Porta: 8080
```

O Radarr utiliza o qBittorrent para realizar os downloads e depois organiza os filmes na biblioteca.

## 🎞️ Organização

Estrutura recomendada:

```text
/mnt/media/
├── downloads/
├── filmes/
│   ├── Filme 1 (2024)/
│   │   └── Filme 1 (2024).mkv
│   └── Filme 2 (2025)/
│       └── Filme 2 (2025).mkv
└── series/
```

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=radarr
```

Ver logs:

```bash
docker logs radarr
```

Reiniciar:

```bash
docker restart radarr
```

Parar:

```bash
docker compose down
```

Iniciar:

```bash
docker compose up -d
```

Atualizar:

```bash
docker compose pull
docker compose up -d
```

## 🔒 Configuração

A configuração fica em:

```text
~/media-server/radarr/config
```

Essa pasta contém banco de dados, configurações e logs e não deve ser enviada para o GitHub.

O `.gitignore` deve conter:

```gitignore
config/
```

## 🌐 Porta

O Radarr utiliza a porta `7878`.

```text
Host:      7878
Container: 7878
```