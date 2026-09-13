# 📶 Speedtest Tracker

O Speedtest Tracker executa testes de velocidade de internet periodicamente e mantém um **histórico** (velocidade de download/upload, latência) disponível em um painel web.

## 📁 Estrutura

```text
speedtest-tracker/
├── config/
├── .env.example
├── .gitignore
└── docker-compose.yml
```

A pasta `config/` contém o banco de dados (SQLite) e demais dados gerados pelo serviço — não deve ser enviada para o GitHub.

## 🔑 Configurar a APP_KEY

O Speedtest Tracker exige uma chave de aplicação (`APP_KEY`) usada para criptografia interna. Ela é lida de um arquivo `.env` local ao serviço (nunca fica hardcoded no `docker-compose.yml`).

```bash
cd speedtest-tracker
cp .env.example .env
```

Gere uma chave aleatória:

```bash
openssl rand -base64 32
```

Edite o `.env` e preencha:

```env
APP_KEY=base64:COLE_AQUI_A_CHAVE_GERADA=
```

> ⚠️ Nunca reutilize uma `APP_KEY` de exemplo encontrada na internet e nunca versione o arquivo `.env` real.

## Docker Compose

```yaml
services:
  speedtest-tracker:
    image: lscr.io/linuxserver/speedtest-tracker:latest
    container_name: speedtest-tracker
    restart: unless-stopped

    env_file:
      - .env

    environment:
      - PUID=1000
      - PGID=1000
      - TZ=America/Sao_Paulo
      - DB_CONNECTION=sqlite
      - SPEEDTEST_SCHEDULE=*/30 * * * *

    volumes:
      - ./config:/config

    ports:
      - "8765:80"
```

* `PUID`/`PGID` — ajuste para o UID/GID do seu usuário (`id -u` / `id -g`), para evitar problemas de permissão na pasta `config/`.
* `SPEEDTEST_SCHEDULE` — expressão cron que define a frequência dos testes automáticos (exemplo: `*/30 * * * *` roda a cada 30 minutos).

## Iniciar

```bash
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

O Speedtest Tracker utiliza a porta `8765` (mapeada para a porta interna `80` do container).

```text
http://IP_DO_SERVIDOR:8765
```

Exemplo:

```text
http://192.168.15.118:8765
```

No primeiro acesso é necessário criar o usuário administrador pelo próprio painel.

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=speedtest-tracker
```

Ver logs:

```bash
docker logs speedtest-tracker
```

Acompanhar os logs:

```bash
docker logs -f speedtest-tracker
```

Rodar um teste manualmente (sem esperar o cron):

```bash
docker exec speedtest-tracker php artisan app:speedtest:run
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

* Não exponha a porta `8765` diretamente à internet — prefira Tailscale/VPN ou rede local.
* Trate o arquivo `.env` (com a `APP_KEY`) como uma credencial: nunca o versione nem o compartilhe.
* O banco de dados em `config/` guarda o histórico de testes; trate-o como dado do ambiente, não como código.

## 📚 Referências

* [Repositório oficial do Speedtest Tracker](https://github.com/alexjustesen/speedtest-tracker)
* [Documentação oficial](https://docs.speedtest-tracker.dev/)
* [Imagem Docker (LinuxServer.io)](https://docs.linuxserver.io/images/docker-speedtest-tracker/)
