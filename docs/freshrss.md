# FreshRSS

O **FreshRSS** é um agregador de feeds RSS/Atom self-hosted. Ele centraliza notícias, blogs, canais do YouTube e outras fontes em um único lugar, com atualização automática, leitura pelo navegador e sincronização com aplicativos de celular e desktop.

## Estrutura

```text
freshrss/
└── docker-compose.yml
```

Os dados do FreshRSS ficam em volumes nomeados do Docker, e não em pastas dentro do projeto:

| Volume                | Caminho no container            | Conteúdo                                        |
| --------------------- | ------------------------------- | ----------------------------------------------- |
| `freshrss_data`       | `/var/www/FreshRSS/data`        | Configuração, usuários, banco SQLite e feeds    |
| `freshrss_extensions` | `/var/www/FreshRSS/extensions`  | Extensões de terceiros                          |

## Docker Compose

```yaml
services:
  freshrss:
    image: docker.io/freshrss/freshrss:latest
    container_name: freshrss
    restart: unless-stopped
    ports:
      - "8083:80"
    environment:
      TZ: America/Sao_Paulo
      CRON_MIN: "3,33"
    volumes:
      - freshrss_data:/var/www/FreshRSS/data
      - freshrss_extensions:/var/www/FreshRSS/extensions

volumes:
  freshrss_data:
  freshrss_extensions:
```

* `8083:80` — a interface web fica disponível na porta `8083` do servidor.
* `TZ` — fuso horário usado nas datas dos artigos e nos logs.
* `CRON_MIN` — minutos de cada hora em que os feeds são atualizados automaticamente. Com `"3,33"`, a atualização ocorre duas vezes por hora (`hh:03` e `hh:33`). Para atualizar a cada 15 minutos, use `"*/15"`.

## Instalação

Entre na pasta do serviço:

```bash
cd freshrss
```

Inicie o container:

```bash
docker compose up -d
```

Verifique se está rodando:

```bash
docker ps --filter name=freshrss
```

Acesse:

```text
http://IP_DO_SERVIDOR:8083
```

## Configuração inicial

No primeiro acesso, o FreshRSS abre um assistente de instalação:

1. **Idioma** — selecione `Português (Brasil)`.
2. **Verificações** — o assistente confere as dependências do PHP. Na imagem oficial todas devem passar.
3. **Banco de dados** — selecione `SQLite`. Para uso pessoal ou de poucos usuários não é necessário um banco externo.
4. **Usuário administrador** — defina o nome de usuário e uma senha forte.
5. Conclua a instalação e faça login.

## Adicionando feeds

Pela interface web:

1. Clique em **Gerenciamento de inscrições** (ícone de engrenagem ao lado da lista de feeds).
2. Clique em **Adicionar um feed RSS**.
3. Cole a URL do feed ou do site. O FreshRSS tenta detectar o feed automaticamente.

### Importando de outro leitor

Se você já usa outro leitor de RSS, exporte suas inscrições em formato **OPML** e importe em:

```text
Gerenciamento de inscrições → Importar / exportar
```

## Acesso por aplicativos

O FreshRSS é compatível com a API do Google Reader, o que permite usar aplicativos externos.

### Habilitar a API

1. Acesse **Administração → Autenticação** e marque **Permitir acesso à API**.
2. Acesse **Perfil** e defina uma **Senha da API**. Ela é separada da senha de login.

Endereço da API para configurar nos aplicativos:

```text
http://IP_DO_SERVIDOR:8083/api/greader.php
```

Use o seu nome de usuário e a senha da API definida no perfil.

### Clientes recomendados

| Plataforma    | Aplicativo                                                  |
| ------------- | ----------------------------------------------------------- |
| Android       | [Readrops](https://github.com/readrops/Readrops), [FeedMe](https://play.google.com/store/apps/details?id=com.seazon.feedme) |
| iPhone / iOS  | [NetNewsWire](https://netnewswire.com/), Reeder             |
| PC / macOS    | [NetNewsWire](https://netnewswire.com/) (macOS), navegador  |

## Comandos úteis

Forçar a atualização de todos os feeds:

```bash
docker exec --user www-data freshrss php ./app/actualize_script.php
```

Ver os logs:

```bash
docker logs -f freshrss
```

Listar os usuários cadastrados:

```bash
docker exec --user www-data freshrss ./cli/list-users.php
```

## Atualização

```bash
cd freshrss
docker compose pull
docker compose up -d
```

Os feeds, usuários e configurações são mantidos, pois ficam nos volumes.

## Backup

Os dados ficam no volume `freshrss_data`. Para descobrir o nome completo do volume (o Docker adiciona o nome do projeto como prefixo):

```bash
docker volume ls | grep freshrss
```

Exemplo de backup do volume para um arquivo `.tar.gz`:

```bash
docker run --rm \
  -v freshrss_freshrss_data:/data:ro \
  -v "$(pwd)":/backup \
  alpine tar czf /backup/freshrss-data.tar.gz -C /data .
```

Também é recomendado exportar periodicamente as inscrições em OPML pela interface web.

## Remoção

Parar e remover o container, mantendo os dados:

```bash
docker compose down
```

Remover o container e apagar também os volumes (feeds, usuários e configurações):

```bash
docker compose down -v
```

> Atenção: o comando com `-v` apaga todos os dados do FreshRSS de forma permanente.

## Segurança

* Não exponha a porta `8083` diretamente para a internet. Para acesso remoto, prefira o [Tailscale](tailscale.md).
* Use senhas diferentes para o login e para a API.
* Mantenha a imagem atualizada.

## Referências

* [Documentação oficial do FreshRSS](https://freshrss.github.io/FreshRSS/)
* [FreshRSS no Docker Hub](https://hub.docker.com/r/freshrss/freshrss)
