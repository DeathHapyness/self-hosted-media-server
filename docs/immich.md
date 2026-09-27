# 📸 Immich

O **Immich** é uma alternativa self-hosted ao Google Fotos, com backup automático de fotos e vídeos pelo celular, linha do tempo, álbuns, compartilhamento, reconhecimento facial e busca inteligente por conteúdo (machine learning local).

## 📁 Estrutura

```text
immich-app/
├── library/
├── postgres/
├── .env
├── env-example.env
└── docker-compose.yml
```

* `library/` — fotos e vídeos enviados, miniaturas e vídeos transcodificados (`UPLOAD_LOCATION`).
* `postgres/` — banco de dados PostgreSQL do Immich (`DB_DATA_LOCATION`).

Essas pastas contêm mídia pessoal e dados gerados pelo serviço e não devem ser enviadas para o GitHub.

## Serviços

O Immich é composto por quatro containers:

| Container                  | Função                                                    |
| -------------------------- | --------------------------------------------------------- |
| `immich_server`            | API e interface web                                       |
| `immich_machine_learning`  | Reconhecimento facial e busca inteligente (CLIP)          |
| `immich_redis`             | Fila de tarefas (Valkey)                                  |
| `immich_postgres`          | Banco de dados (PostgreSQL com extensão de busca vetorial) |

## 🔑 Configurar o `.env`

As configurações do Immich (caminhos, versão e credenciais do banco) são lidas de um arquivo `.env` local ao serviço.

```bash
cd immich-app
```

Exemplo em: [`immich-app/env-example.env`](../immich-app/env-example.env)

Crie o arquivo `.env` a partir do exemplo e substitua pelos valores desejados:

```bash
cp env-example.env .env
```

```env
UPLOAD_LOCATION=./library
DB_DATA_LOCATION=./postgres
TZ=America/Sao_Paulo
IMMICH_VERSION=v3

DB_PASSWORD=senha-forte-aqui
DB_USERNAME=postgres
DB_DATABASE_NAME=immich
```

* `UPLOAD_LOCATION` — onde as fotos e vídeos serão armazenados. Pode apontar para o disco de mídia (ex.: `/mnt/media/fotos`). Deve ser um caminho local.
* `DB_DATA_LOCATION` — onde o banco de dados será armazenado. **Não pode** ser um compartilhamento de rede (NFS/SMB); prefira um SSD.
* `IMMICH_VERSION` — versão das imagens. Pode ser fixada em uma versão específica (ex.: `v2.1.0`).
* `DB_PASSWORD` — use apenas caracteres `A-Za-z0-9`, sem espaços ou caracteres especiais.

> ⚠️ Nunca versione o arquivo `.env` real — ele contém a senha do banco de dados.

## Docker Compose

O Immich recomenda sempre utilizar o `docker-compose.yml` da versão atual, disponível em:

```text
https://github.com/immich-app/immich/releases/latest/download/docker-compose.yml
```

Estrutura do arquivo:

```yaml
name: immich

services:
  immich-server:
    container_name: immich_server
    image: ghcr.io/immich-app/immich-server:${IMMICH_VERSION:-release}
    volumes:
      - ${UPLOAD_LOCATION}:/data
      - /etc/localtime:/etc/localtime:ro
    env_file:
      - .env
    ports:
      - '2283:2283'
    depends_on:
      - redis
      - database
    restart: always
    healthcheck:
      disable: false

  immich-machine-learning:
    container_name: immich_machine_learning
    image: ghcr.io/immich-app/immich-machine-learning:${IMMICH_VERSION:-release}
    volumes:
      - model-cache:/cache
    env_file:
      - .env
    restart: always
    healthcheck:
      disable: false

  redis:
    container_name: immich_redis
    image: docker.io/valkey/valkey:8-bookworm
    healthcheck:
      test: redis-cli ping || exit 1
    restart: always

  database:
    container_name: immich_postgres
    image: ghcr.io/immich-app/postgres:14-vectorchord0.4.3-pgvectors0.2.0
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_USER: ${DB_USERNAME}
      POSTGRES_DB: ${DB_DATABASE_NAME}
      POSTGRES_INITDB_ARGS: '--data-checksums'
    volumes:
      - ${DB_DATA_LOCATION}:/var/lib/postgresql/data
    shm_size: 128mb
    restart: always

volumes:
  model-cache:
```

* `${UPLOAD_LOCATION}:/data` — biblioteca de fotos e vídeos.
* `model-cache` — volume Docker com os modelos de machine learning baixados no primeiro uso.
* `${DB_DATA_LOCATION}:/var/lib/postgresql/data` — dados do PostgreSQL.
* A imagem do banco é específica do Immich (`ghcr.io/immich-app/postgres`) — não substitua por uma imagem PostgreSQL comum.

> O arquivo oficial também fixa as imagens do Valkey e do PostgreSQL por digest (`@sha256:...`). Mantenha os digests do arquivo oficial ao atualizar.

### Aceleração por hardware (opcional)

Para usar GPU na transcodificação de vídeos ou no machine learning, baixe os arquivos `hwaccel.transcoding.yml` e `hwaccel.ml.yml` da mesma release e descomente os blocos `extends` no `docker-compose.yml`, escolhendo o backend (`nvenc`, `quicksync`, `vaapi`, `cuda`, `openvino`, etc.).

## Iniciar

```bash
cd immich-app
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

No primeiro acesso, é necessário criar o usuário administrador pela própria interface web.

## Acesso

O Immich utiliza a porta `2283`.

```text
http://IP_DO_SERVIDOR:2283
```

Exemplo:

```text
http://192.168.15.118:2283
```

### Aplicativo mobile

Instale o app **Immich** (Android / iOS) e informe a URL do servidor:

```text
http://IP_DO_SERVIDOR:2283
```

Fora de casa, use o IP ou nome do servidor no Tailscale. Em seguida, ative o **backup** no app e escolha os álbuns do celular que devem ser enviados.

## 🐳 Gerenciamento

Verificar os containers:

```bash
docker ps --filter name=immich
```

Ver logs do servidor:

```bash
docker logs immich_server
```

Acompanhar os logs:

```bash
docker logs -f immich_server
```

Logs do machine learning:

```bash
docker logs -f immich_machine_learning
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

> ⚠️ Antes de atualizar, leia as notas da release no GitHub. Algumas versões possuem *breaking changes* que exigem alterações no `docker-compose.yml` ou no `.env`.

## 💾 Backup

O Immich precisa de backup de duas partes:

* `library/` (`UPLOAD_LOCATION`) — fotos e vídeos originais.
* Banco de dados — álbuns, rostos, metadados e usuários.

O Immich gera dumps automáticos do banco em `UPLOAD_LOCATION/backups`. Portanto, fazer backup de `UPLOAD_LOCATION` com o [Kopia](kopia.md) já cobre as fotos e os dumps do banco.

Dump manual do banco:

```bash
docker exec -t immich_postgres pg_dumpall --clean --if-exists --username=postgres | gzip > immich-db.sql.gz
```

> Não copie a pasta `postgres/` com o container em execução — use os dumps.

## 🔒 Segurança

* Não exponha a porta `2283` diretamente à internet — prefira Tailscale/VPN ou rede local. Caso precise de acesso público, use reverse proxy com HTTPS.
* A pasta `library/` contém fotos e vídeos pessoais; nunca a versione e proteja o acesso ao disco.
* Use uma senha forte em `DB_PASSWORD` e crie senhas fortes para os usuários do Immich.
* O banco de dados não precisa de porta exposta — ele é acessado apenas pela rede interna do compose.

## 📚 Referências

* [Repositório oficial do Immich](https://github.com/immich-app/immich)
* [Documentação oficial](https://docs.immich.app/)
* [Instalação com Docker Compose](https://docs.immich.app/install/docker-compose)
* [Variáveis de ambiente](https://docs.immich.app/install/environment-variables)
* [Backup e restauração](https://docs.immich.app/administration/backup-and-restore)
