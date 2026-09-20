# 💾 Kopia

O **Kopia** é uma ferramenta de backup rápida e segura, com deduplicação, compressão e criptografia dos dados. Neste projeto ele roda em modo **servidor**, disponibilizando uma interface web para criar, restaurar e agendar snapshots.

## 📁 Estrutura

```text
kopia/
├── config/
├── cache/
├── logs/
├── .env
└── docker-compose.yml
```

* `config/` — configuração do repositório (`repository.config`).
* `cache/` — cache local do Kopia.
* `logs/` — logs do servidor.

Essas pastas contêm dados/estado gerados pelo serviço e não devem ser enviadas para o GitHub.

## 🔑 Configurar o `.env`

O acesso à interface web do Kopia é protegido por usuário e senha, definidos via variáveis de ambiente lidas de um arquivo `.env` local ao serviço (nunca hardcoded no `docker-compose.yml`).

```bash
cd kopia
```

Crie o arquivo `.env` com o seguinte conteúdo, substituindo pelos valores desejados:

```env
KOPIA_SERVER_USERNAME=usuario
KOPIA_SERVER_PASSWORD=senha-forte-aqui
```

> ⚠️ Nunca versione o arquivo `.env` real — ele contém a credencial de acesso ao repositório de backup.

## Docker Compose

```yaml
services:
  kopia:
    image: kopia/kopia:latest
    container_name: kopia
    hostname: kopia
    restart: unless-stopped

    env_file:
      - .env

    environment:
      KOPIA_CONFIG_PATH: /app/config/repository.config
      KOPIA_CACHE_DIRECTORY: /app/cache
      KOPIA_LOG_DIR: /app/logs

    ports:
      - "51515:51515"

    command:
      - server
      - start
      - --insecure
      - --address=0.0.0.0:51515
      - --server-username=${KOPIA_SERVER_USERNAME}
      - --server-password=${KOPIA_SERVER_PASSWORD}

    volumes:
      - ./config:/app/config
      - ./cache:/app/cache
      - ./logs:/app/logs

      # Destino do backup
      - /mnt/backup/kopia-repository:/repository

      # Origens disponíveis somente para leitura
      - /home/rique:/data/home:ro
      - /var/lib/docker/volumes/n8n_n8n_data/_data:/data/n8n:ro
      - /mnt/media:/data/media:ro
```

* `/repository` — destino onde o repositório de backup do Kopia é armazenado.
* `/data/home`, `/data/n8n`, `/data/media` — origens montadas somente leitura (`ro`), disponíveis para seleção como fontes de backup dentro da interface do Kopia.
* `--insecure` — desabilita TLS na interface web; recomendado acessar apenas via rede local ou Tailscale.

## Iniciar

```bash
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

No primeiro acesso, é necessário criar (ou conectar a) um repositório de backup pela própria interface web, apontando para `/repository`.

## Acesso

O Kopia utiliza a porta `51515`.

```text
http://IP_DO_SERVIDOR:51515
```

Exemplo:

```text
http://192.168.15.118:51515
```

Use as credenciais definidas em `KOPIA_SERVER_USERNAME` / `KOPIA_SERVER_PASSWORD`.

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=kopia
```

Ver logs:

```bash
docker logs kopia
```

Acompanhar os logs:

```bash
docker logs -f kopia
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

* Não exponha a porta `51515` diretamente à internet — o servidor roda com `--insecure` (sem TLS), então prefira acesso via Tailscale/VPN ou rede local.
* Trate o arquivo `.env` (usuário/senha) como uma credencial: nunca o versione nem o compartilhe.
* As origens montadas como `:ro` evitam que o Kopia grave nos dados originais — apenas o destino `/repository` recebe escrita.

## 📚 Referências

* [Repositório oficial do Kopia](https://github.com/kopia/kopia)
* [Documentação oficial](https://kopia.io/docs/)
