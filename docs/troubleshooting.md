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

AdGuard Home:

```bash
docker inspect adguardhome \
  --format '{{range .Mounts}}{{println .Source "->" .Destination}}{{end}}'
```