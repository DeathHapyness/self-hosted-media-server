# 🎵 Navidrome

O Navidrome é responsável pela biblioteca musical.

Ele lê os arquivos presentes em:

```text
/opt/navidrome/music
```

e disponibiliza a biblioteca através de clientes compatíveis.

## Criar diretórios

```bash
sudo mkdir -p /opt/navidrome

cd /opt/navidrome

mkdir -p data
mkdir -p music
```

## Docker Compose

```yaml
services:
  navidrome:
    image: deluan/navidrome:latest
    container_name: navidrome
    restart: unless-stopped

    ports:
      - "4533:4533"

    environment:
      ND_SCANSCHEDULE: 1h
      ND_LOGLEVEL: info
      ND_SESSIONTIMEOUT: 24h
      TZ: America/Sao_Paulo

    volumes:
      - ./data:/data
      - ./music:/music:ro
```

Inicie:

```bash
docker compose up -d
```

Acesse:

```text
http://IP_DO_SERVIDOR:4533
```

---

## 🎧 Clientes Recomendados

O Navidrome funciona como **servidor**, enquanto os aplicativos abaixo funcionam como clientes.

### 💻 PC — Feishin

Interface moderna para servidores compatíveis com Subsonic, incluindo Navidrome.

[Feishin — GitHub](https://github.com/jeffvli/feishin)

### 📱 iPhone / iOS

**Amperfy** — [App Store](https://apps.apple.com/app/amperfy-music-player/id1530145038)

**play:Sub** — [App Store](https://apps.apple.com/app/play-sub/id955329386)

### 🤖 Android

**Symfonium** — suporta Navidrome e outros servidores de música. [Site oficial](https://symfonium.app/)

---

## 🎶 spotDL

O spotDL pode ser utilizado para downloads manuais de músicas.

### Instalar dependências

```bash
sudo apt install -y python3 python3-pip ffmpeg
```

Instale:

```bash
pip install spotdl
```

Verifique:

```bash
spotdl --version
```

### Baixando músicas

Para manter a organização da biblioteca:

```bash
spotdl \
  --output "/opt/navidrome/music/{artist}/{album}/{track-number} - {title}.{output-ext}" \
  "LINK_DO_SPOTIFY"
```

Exemplo:

```bash
spotdl \
  --output "/opt/navidrome/music/{artist}/{album}/{track-number} - {title}.{output-ext}" \
  "https://open.spotify.com/track/ID"
```

Sempre coloque a URL entre aspas.

Isso permite organizar automaticamente:

```text
/opt/navidrome/music/
│
├── Artista/
│   ├── Álbum/
│   │   ├── 01 - Música.mp3
│   │   ├── 02 - Música.mp3
│   │   └── 03 - Música.mp3
│   │
│   └── Outro Álbum/
│
└── Outro Artista/
```

---

## 🔄 Atualizando o Navidrome

Depois de adicionar novas músicas:

```bash
docker exec -it navidrome /app/navidrome scan
```

Também existe um scan automático configurado:

```yaml
ND_SCANSCHEDULE: 1h
```

---

## 🔄 Fluxo da Música

```text
              Spotify URL
                   │
                   ▼
                spotDL
                   │
                   ▼
        /opt/navidrome/music
                   │
          ┌────────┴────────┐
          │                 │
       Artista           Artista
          │                 │
        Álbum             Álbum
          │                 │
        Faixas            Faixas
          │                 │
          └────────┬────────┘
                   ▼
               Navidrome
                   │
          ┌────────┼────────┐
          ▼        ▼        ▼
        Feishin   iOS    Android
          PC
```