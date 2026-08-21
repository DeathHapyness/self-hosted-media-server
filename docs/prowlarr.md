# 🔎 Prowlarr

O Prowlarr é utilizado para gerenciamento centralizado de **indexadores**.

Ele trabalha em conjunto com aplicações como **Radarr** e **Sonarr**, permitindo pesquisar e gerenciar indexadores em um único lugar. Os resultados encontrados podem ser enviados para os aplicativos de gerenciamento de mídia, que posteriormente utilizam um cliente de download, como o qBittorrent.

## 📁 Estrutura

```text
prowlarr/

├── config/
├── docker-compose.yml
├── .gitignore
└── README.md
```

A pasta `config/` contém os dados internos do Prowlarr e não deve ser enviada para o GitHub.

## Criar diretórios

```bash
mkdir -p ~/media-server/prowlarr

cd ~/media-server/prowlarr

mkdir -p config
```

## Docker Compose

```yaml
services:

  prowlarr:

    image: lscr.io/linuxserver/prowlarr:latest

    container_name: prowlarr

    restart: unless-stopped

    environment:

      - PUID=1000

      - PGID=1000

      - TZ=America/Sao_Paulo

    volumes:

      - ./config:/config

    ports:

      - "9696:9696"
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
http://IP_DO_SERVIDOR:9696
```

Exemplo:

```text
http://IP_DO_SERVIDOR:9696
```

> Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor. Evite colocar IPs pessoais ou públicos reais no repositório do GitHub.

## 🔎 Configuração de indexadores

Depois de acessar o Prowlarr:

```text
Indexers
→ Add Indexer
```

O Prowlarr exibirá os indexadores disponíveis e suas respectivas configurações.

Ao adicionar um indexador, utilize:

```text
Nome do indexador
URL
Credenciais, caso necessárias
API Key, caso necessária
```

Depois de configurar:

```text
Test
→ Save
```

O botão **Test** permite verificar se o Prowlarr consegue se comunicar corretamente com o indexador.

## 🔗 Integração com Radarr

O Prowlarr pode enviar automaticamente os indexadores configurados para o Radarr.

No Prowlarr:

```text
Settings
→ Apps
→ Add Application
→ Radarr
```

Exemplo:

```text
Name: Radarr

Prowlarr Server:
http://IP_DO_SERVIDOR:9696

Radarr Server:
http://radarr:7878

API Key:
SUA_API_KEY
```

> Se Prowlarr e Radarr estiverem na mesma rede Docker, normalmente é preferível utilizar o nome do container, por exemplo `http://radarr:7878`, em vez do IP do servidor.

Depois de configurar, utilize:

```text
Test
→ Save
```

## 🔗 Integração com Sonarr

Para séries, o Prowlarr pode ser integrado ao Sonarr:

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

## 🌐 Generic Torznab

Caso um indexador forneça uma **API Torznab compatível**, é possível adicioná-lo através do:

```text
Indexers
→ Add Indexer
→ Generic Torznab
```

Nesse caso, o Prowlarr precisa do endereço do endpoint Torznab e, quando necessário, de uma API Key.

Exemplo:

```text
URL:
https://EXEMPLO.COM/api/torznab

API Key:
SUA_API_KEY
```

> Uma URL comum de um site não é necessariamente uma API Torznab. O site precisa disponibilizar um endpoint compatível.

## 🧩 Indexadores personalizados

Se um site não possuir uma definição disponível no Prowlarr e também não fornecer uma API Torznab compatível, pode ser necessário criar uma **definição personalizada**.

O Prowlarr utiliza definições baseadas em **Cardigann/YAML** para determinados indexadores.

A definição descreve como o Prowlarr deve:

```text
Pesquisar
   ↓
Acessar a página de resultados
   ↓
Encontrar os resultados
   ↓
Extrair título, tamanho, categoria etc.
   ↓
Obter o link de download
```

Definições personalizadas devem ser utilizadas somente quando o site e sua forma de acesso forem compatíveis com o funcionamento do Prowlarr.

## 📂 Configuração

A configuração do Prowlarr fica em:

```text
~/media-server/prowlarr/config
```

Essa pasta contém:

```text
Banco de dados
Configurações
Logs
Definições
Estado do aplicativo
```

A pasta `config/` não deve ser enviada para o GitHub.

## 🔒 .gitignore

O `.gitignore` deve conter:

```gitignore
config/
```

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=prowlarr
```

Ver logs:

```bash
docker logs prowlarr
```

Acompanhar logs em tempo real:

```bash
docker logs -f prowlarr
```

Reiniciar:

```bash
docker restart prowlarr
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

## 🔗 Fluxo do servidor

Uma configuração típica pode utilizar:

```text
                  ┌─────────────┐
                  │   Prowlarr  │
                  │  Indexadores │
                  └──────┬──────┘
                         │
              ┌──────────┴──────────┐
              ▼                     ▼
        ┌──────────┐          ┌──────────┐
        │  Radarr  │          │  Sonarr  │
        │  Filmes  │          │  Séries  │
        └─────┬────┘          └─────┬────┘
              │                     │
              └──────────┬──────────┘
                         ▼
                  ┌─────────────┐
                  │ qBittorrent │
                  └──────┬──────┘
                         ▼
                    /mnt/media
                         │
              ┌──────────┴──────────┐
              ▼                     ▼
           Filmes                 Séries
              │                     │
              └──────────┬──────────┘
                         ▼
                     Jellyfin
```

O Prowlarr funciona como o **gerenciador central de indexadores**, enquanto Radarr e Sonarr utilizam esses indexadores para realizar suas pesquisas.

## 🌐 Porta

O Prowlarr utiliza a porta `9696`.

```text
Host:      9696
Container: 9696
```
