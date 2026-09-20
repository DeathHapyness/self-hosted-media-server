# 📶 LibreSpeed

O **LibreSpeed** é um servidor de teste de velocidade self-hosted (alternativa ao Speedtest/Ookla). Permite medir a velocidade de download, upload e latência entre um cliente e o próprio servidor, sem depender de serviços de terceiros.

## 📁 Estrutura

```text
librespeed/
├── config/
└── docker-compose.yml
```

A pasta `config/` guarda os dados persistentes do serviço (banco SQLite) — não deve ser enviada para o GitHub.

## Docker Compose

```yaml
services:
  librespeed:
    image: lscr.io/linuxserver/librespeed:latest
    container_name: librespeed
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=America/Sao_Paulo
      - DB_TYPE=sqlite
    volumes:
      - ./config:/config
    ports:
      - "8111:80"
    restart: unless-stopped
```

* `PUID`/`PGID` — ajuste para o UID/GID do seu usuário (`id -u` / `id -g`), evitando problemas de permissão na pasta `config/`.
* `DB_TYPE=sqlite` — armazena os resultados dos testes em um banco SQLite local, sem exigir um banco externo.

## Iniciar

```bash
cd librespeed
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

O LibreSpeed utiliza a porta `8111` (mapeada para a porta interna `80` do container).

```text
http://IP_DO_SERVIDOR:8111
```

Exemplo:

```text
http://192.168.15.118:8111
```

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=librespeed
```

Ver logs:

```bash
docker logs librespeed
```

Reiniciar:

```bash
docker compose restart
```

Parar:

```bash
docker compose down
```

Atualizar:

```bash
docker compose pull
docker compose up -d
```

## 🔒 Segurança

* Não exponha a porta `8111` diretamente à internet — prefira Tailscale/VPN ou rede local.
* Os resultados dos testes ficam salvos em `config/`; trate-os como dado do ambiente, não como código.

## 📚 Referências

* [Repositório oficial do LibreSpeed](https://github.com/librespeed/speedtest)
* [Imagem Docker (LinuxServer.io)](https://docs.linuxserver.io/images/docker-librespeed/)
