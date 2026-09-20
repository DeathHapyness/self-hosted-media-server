# 🦊 Forgejo

O **Forgejo** é uma plataforma Git self-hosted (fork do Gitea), com interface web para repositórios, issues, pull requests e um servidor SSH próprio para clone/push.

## 📁 Estrutura

```text
forgejo/
├── forgejo/
└── docker-compose.yml
```

A pasta `forgejo/` (montada em `/data` no container) guarda os repositórios, configuração e banco de dados do serviço — não deve ser enviada para o GitHub.

## Docker Compose

```yaml
networks:
  forgejo:

services:
  server:
    image: codeberg.org/forgejo/forgejo:16
    container_name: forgejo

    environment:
      - USER_UID=1000
      - USER_GID=1000
      - FORGEJO__server__SSH_PORT=222

    restart: unless-stopped

    networks:
      - forgejo

    volumes:
      - ./forgejo:/data
      - /etc/localtime:/etc/localtime:ro

    ports:
      - "3044:3000"
      - "222:22"
```

* `USER_UID`/`USER_GID` — ajuste para o UID/GID do seu usuário, evitando problemas de permissão na pasta `forgejo/`.
* `FORGEJO__server__SSH_PORT=222` — porta SSH interna usada pelo Forgejo para clone/push via `git@`; mapeada externamente para `222` para não conflitar com o SSH do host (`22`).

## Iniciar

```bash
cd forgejo
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

No primeiro acesso, é necessário concluir a instalação e criar o usuário administrador pela própria interface web.

## Acesso

O Forgejo utiliza a porta `3044` para a interface web (mapeada para a porta interna `3000`) e a porta `222` para SSH (mapeada para a porta interna `22`).

```text
http://IP_DO_SERVIDOR:3044
```

Exemplo:

```text
http://192.168.15.118:3044
```

Clone/push via SSH:

```bash
git clone ssh://git@IP_DO_SERVIDOR:222/usuario/repositorio.git
```

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=forgejo
```

Ver logs:

```bash
docker logs forgejo
```

Acompanhar os logs:

```bash
docker logs -f forgejo
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

* Não exponha as portas `3044`/`222` diretamente à internet — prefira Tailscale/VPN ou rede local.
* A pasta `forgejo/` contém o banco de dados e os repositórios; trate-a como dado do ambiente, não como código.
* Configure autenticação de dois fatores e chaves SSH para os usuários criados no Forgejo, evitando depender apenas de senha.

## 📚 Referências

* [Repositório oficial do Forgejo](https://codeberg.org/forgejo/forgejo)
* [Documentação oficial](https://forgejo.org/docs/latest/)
