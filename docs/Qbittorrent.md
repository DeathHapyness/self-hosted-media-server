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