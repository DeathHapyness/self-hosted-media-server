# 🔗 Links Hub

Página única e minimalista para guardar os sites que você mais usa no dia a dia.
Só nome + URL (com ícone e categoria opcionais), sem login e sem busca.

Projeto: [DeathHapyness/Links-Hub](https://github.com/DeathHapyness/Links-Hub)

## Diferença para o Homepage

O [Homepage](homepage.md) é o painel dos serviços do servidor, configurado por
arquivos YAML. O Links Hub é para os seus sites do dia a dia, adicionados e
removidos direto pela página, sem editar arquivo nenhum.

## Docker Compose

```yaml
services:
  links-hub:
    image: ghcr.io/deathhapyness/links-hub:latest
    container_name: links-hub
    restart: unless-stopped

    ports:
      - "5000:5000"

    volumes:
      - ${DATA_DIR:-/opt/media-server}/links-hub/data:/app/data
```

Inicie:

```bash
docker compose up -d
```

Acesse `http://IP-DO-SERVIDOR:5000`.

## Dados

Os links ficam no SQLite em `${DATA_DIR}/links-hub/data/links.db`, então
sobrevivem a atualizações do container. O container roda como o UID 1000, que
precisa ter permissão de escrita nessa pasta (o instalador já cria e ajusta).
