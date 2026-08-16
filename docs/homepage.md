# 🏠 Homepage

O Homepage é um **dashboard inicial para o servidor**, permitindo centralizar e organizar o acesso aos serviços hospedados em um único painel web.

## Criar diretórios

```bash
sudo mkdir -p /opt/homepage
cd /opt/homepage

mkdir -p config
```

## Docker Compose

```yaml
services:
  homepage:
    image: ghcr.io/gethomepage/homepage:latest
    container_name: homepage
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - ./config:/app/config
    environment:
      HOMEPAGE_ALLOWED_HOSTS: IP_DO_SERVIDOR:3000
```

Inicie:

```bash
docker compose up -d
```

## Primeiro acesso

Acesse o Homepage pelo navegador:

```text
http://IP_DO_SERVIDOR:3000
```

A configuração do dashboard é feita através dos arquivos dentro do diretório `config/`.

Entre os principais arquivos estão:

```text
config/
├── bookmarks.yaml
├── docker.yaml
├── services.yaml
├── settings.yaml
└── widgets.yaml
```

O arquivo `services.yaml` é utilizado para adicionar e organizar os serviços exibidos no dashboard.

Exemplo:

```yaml
- Media:
    - Jellyfin:
        href: http://IP_DO_SERVIDOR:8096
        description: Servidor de filmes e séries

    - Navidrome:
        href: http://IP_DO_SERVIDOR:4533
        description: Servidor de músicas
```

Após alterar os arquivos de configuração, o Homepage atualiza o dashboard automaticamente.

## Organização dos serviços

O Homepage permite organizar os serviços em grupos, facilitando o acesso aos diferentes sistemas hospedados no servidor.

Exemplo:

```text
Homepage
│
├── Media
│   ├── Jellyfin
│   └── Navidrome
│
├── Downloads
│   └── qBittorrent
│
└── Server
    └── AdGuard Home
```

## Integração com Docker

O Homepage também pode consultar informações dos containers Docker e exibir seu estado diretamente no dashboard.

Para isso, pode ser necessário disponibilizar o socket do Docker:

```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock:ro
```

> ⚠️ O acesso ao `docker.sock` concede ao container acesso significativo ao Docker do host. Utilize somente se essa integração for realmente necessária.

## O que não é versionado

Dependendo da configuração utilizada, arquivos dentro de `config/` podem conter informações específicas do ambiente, como URLs internas, nomes de serviços, tokens de integração ou outras configurações.

Recomenda-se revisar os arquivos antes de versioná-los no Git.

Uma estrutura possível:

```text
/opt/homepage/
├── docker-compose.yml
└── config/
    ├── bookmarks.yaml
    ├── docker.yaml
    ├── services.yaml
    ├── settings.yaml
    └── widgets.yaml
```

Apenas os arquivos de configuração que não contenham informações sensíveis devem ser versionados no repositório.
