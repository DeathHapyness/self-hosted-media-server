# 📺 Sonarr

O Sonarr é utilizado para gerenciamento e organização automática de **séries de TV**.

Ele trabalha em conjunto com indexadores, como os gerenciados pelo **Prowlarr**, e com um cliente de download, como o **qBittorrent**. Dessa forma, é possível adicionar séries, acompanhar novos episódios, realizar downloads automaticamente e organizar os arquivos na biblioteca.

## 📁 Estrutura

```text
sonarr/

├── config/
├── docker-compose.yml
├── .gitignore
└── README.md
```

A pasta `config/` contém os dados internos do Sonarr e não deve ser enviada para o GitHub.

## Criar diretórios

```bash
mkdir -p ~/media-server/sonarr

cd ~/media-server/sonarr

mkdir -p config
```

## Docker Compose

```yaml
services:

  sonarr:

    image: lscr.io/linuxserver/sonarr:latest

    container_name: sonarr

    restart: unless-stopped

    environment:

      - PUID=1000

      - PGID=1000

      - TZ=America/Sao_Paulo

    volumes:

      - ./config:/config

      - /mnt/media:/media

    ports:

      - "8989:8989"
```

## Iniciar

```bash
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

```text
http://IP_DO_SERVIDOR:8989
```

> Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor. Evite colocar IPs pessoais ou públicos reais no repositório do GitHub.

## 📺 Adicionar séries

Depois de acessar o Sonarr:

```text
Series
→ Add New Series
```

Procure pela série desejada e configure:

```text
Series Type
Quality Profile
Root Folder
Season Monitoring
```

Depois confirme em:

```text
Add Series
```

O Sonarr passa a monitorar a série e seus episódios.

## 📂 Diretórios de mídia

O Sonarr utiliza `/mnt/media` do servidor como `/media` dentro do container.

### Séries

```text
Host:

/mnt/media/series

Container:

/media/series
```

### Downloads

```text
Host:

/mnt/media/downloads

Container:

/media/downloads
```

## 🔗 Integração com qBittorrent

No Sonarr, acesse:

```text
Settings
→ Download Clients
→ Add
→ qBittorrent
```

Exemplo:

```text
Host: qBittorrent

Porta: 8080

Username: SEU_USUARIO

Password: SUA_SENHA
```

Se os containers estiverem na mesma rede Docker, prefira utilizar o nome do container:

```text
Host: qbittorrent
Port: 8080
```

Depois utilize:

```text
Test
→ Save
```

O Sonarr poderá enviar os downloads para o qBittorrent e acompanhar o progresso automaticamente.

## 🔎 Integração com Prowlarr

O Prowlarr pode fornecer ao Sonarr os indexadores utilizados para encontrar episódios.

No Prowlarr:

```text
Settings
→ Apps
→ Add Application
→ Sonarr
```

Exemplo:

```text
Name: Sonarr

Sonarr Server:
http://sonarr:8989

API Key:
SUA_API_KEY
```

Depois:

```text
Test
→ Save
```

O Prowlarr poderá sincronizar os indexadores configurados com o Sonarr.

## 📥 Fluxo de download

O funcionamento típico é:

```text
Usuário
   ↓
Sonarr
   ↓
Prowlarr
   ↓
Indexadores
   ↓
Sonarr encontra o resultado
   ↓
qBittorrent
   ↓
Download
   ↓
Sonarr importa e organiza
   ↓
/mnt/media/series
   ↓
Jellyfin
```

Depois que um episódio termina de ser baixado, o Sonarr pode identificá-lo e organizá-lo automaticamente na biblioteca.

## 🎞️ Organização

Estrutura recomendada:

```text
/mnt/media/

├── downloads/

├── filmes/

└── series/

    ├── Breaking Bad (2008)/

    │   ├── Season 01/

    │   │   ├── Breaking Bad - S01E01.mkv

    │   │   └── Breaking Bad - S01E02.mkv

    │   │
    │   └── Season 02/
    │
    └── Stranger Things (2016)/

        └── Season 01/

            └── Stranger Things - S01E01.mkv
```

O Sonarr utiliza informações da série e dos episódios para manter os arquivos organizados.

## 📁 Root Folder

No Sonarr, configure como pasta principal:

```text
/media/series
```

Não utilize o caminho do host dentro do container.

Correto:

```text
/media/series
```

Incorreto:

```text
/mnt/media/series
```

O caminho `/mnt/media/series` existe no **host**, enquanto `/media/series` é o caminho correspondente dentro do container.

## 📥 Pasta de downloads

Uma configuração comum é:

```text
Host:

/mnt/media/downloads

Container:

/media/downloads
```

O qBittorrent pode utilizar essa pasta para downloads temporários/concluídos.

Depois que o download termina, o Sonarr realiza a importação para:

```text
/media/series
```

## 🏷️ Categorias do qBittorrent

É recomendado utilizar uma categoria específica para o Sonarr:

```text
sonarr
```

Por exemplo:

```text
Downloads/
└── sonarr/
```

No Sonarr:

```text
Settings
→ Download Clients
→ qBittorrent
→ Category
```

Configure:

```text
Category: sonarr
```

Isso ajuda a separar os downloads realizados pelo Sonarr de outros downloads do qBittorrent.

## 🔑 API Key

O Sonarr possui uma API Key utilizada para integrações com outros serviços.

Ela pode ser encontrada em:

```text
Settings
→ General
→ Security
→ API Key
```

Essa chave pode ser utilizada para integrar serviços como:

```text
Prowlarr
Seerr
Scripts
Outras aplicações
```

**Não publique a API Key no GitHub.**

## 🔒 Configuração

A configuração do Sonarr fica em:

```text
~/media-server/sonarr/config
```

Essa pasta contém:

```text
Banco de dados
Configurações
Logs
Metadados
Estado do aplicativo
```

A pasta `config/` não deve ser enviada para o GitHub.

## 📝 .gitignore

O `.gitignore` deve conter:

```gitignore
config/
```

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=sonarr
```

Ver logs:

```bash
docker logs sonarr
```

Acompanhar logs em tempo real:

```bash
docker logs -f sonarr
```

Reiniciar:

```bash
docker restart sonarr
```

Parar:

```bash
docker compose down
```

Iniciar:

```bash
docker compose up -d
```

Atualizar:

```bash
docker compose pull

docker compose up -d
```

## 🔗 Integração com Jellyfin

O Jellyfin deve ter acesso à mesma biblioteca de mídia.

No host:

```text
/mnt/media/series
```

O Sonarr utiliza:

```text
/media/series
```

O Jellyfin deve receber esse diretório como biblioteca de séries.

Exemplo:

```text
/mnt/media/
└── series/
    ├── Breaking Bad (2008)/
    └── Stranger Things (2016)/
```

Depois que o Sonarr organizar um episódio, o Jellyfin poderá detectá-lo durante sua atualização da biblioteca.

## 🔗 Stack completa

Uma configuração típica do servidor pode ser:

```text
                         ┌──────────────┐
                         │    Seerr     │
                         │  Solicitações│
                         └──────┬───────┘
                                │
                                ▼
                         ┌──────────────┐
                         │    Sonarr    │
                         │    Séries    │
                         └──────┬───────┘
                                │
                                ▼
                         ┌──────────────┐
                         │   Prowlarr   │
                         │  Indexadores │
                         └──────┬───────┘
                                │
                                ▼
                         ┌──────────────┐
                         │  qBittorrent │
                         │   Download   │
                         └──────┬───────┘
                                │
                                ▼
                         /mnt/media/series
                                │
                                ▼
                         ┌──────────────┐
                         │   Jellyfin   │
                         │    Séries    │
                         └──────────────┘
```

## 🌐 Porta

O Sonarr utiliza a porta `8989`.

```text
Host:      8989
Container: 8989
```
