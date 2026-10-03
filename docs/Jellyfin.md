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

## Referências

* [Repositório oficial do Jellyfin](https://github.com/jellyfin/jellyfin)
* [Documentação oficial](https://jellyfin.org/docs/)
* [Instalação via Docker](https://jellyfin.org/docs/general/installation/container)
