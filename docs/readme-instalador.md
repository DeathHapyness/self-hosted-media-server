## Instalação automática

Este repositório inclui um instalador, escrito em Go, que prepara o servidor
do zero: verifica pré-requisitos, cria os diretórios necessários em
`/opt/media-server`, ajusta permissões, cria logs, valida os
`docker-compose.yml` e sobe os containers.

O instalador é atualizado conforme novos programas são adicionados ao repo.

## <span style="color: red; font-weight: bold;">ADVERTÊNCIA:</span> 
<span style="color: darkorange;">Este script é fornecido "como está" e não se responsabiliza por quaisquer danos ou perda de dados. Use por sua própria conta e risco.</span>

<span style="color: #00fffb;">Esse script foi testado antes de estar nesse repo</span>

### Pré-requisitos

- Linux (x86_64 ou ARM64, como Raspberry Pi 4/5)
- Docker com o plugin Docker Compose v2 (`docker compose version` deve funcionar),
  ou Podman com `podman compose` ou `podman-compose`
- Um disco dedicado para a mídia em `/mnt/media` é recomendado, mas opcional

Não é preciso ter Go instalado para usar o binário pronto.

### Como executar

O instalador precisa ficar dentro da pasta do repositório, porque é ali que ele
encontra os `docker-compose.yml` de cada serviço.

**Opção 1: binário pronto (recomendado)**

Baixe o arquivo da sua arquitetura na página de
[Releases](https://github.com/DeathHapyness/self-hosted-media-server/releases)
e coloque na raiz do repositório:

```bash
git clone https://github.com/DeathHapyness/self-hosted-media-server.git
cd self-hosted-media-server

# baixa sempre a versão mais recente; troque amd64 por arm64 se for ARM
wget https://github.com/DeathHapyness/self-hosted-media-server/releases/latest/download/instalador-linux-amd64
chmod +x instalador-linux-amd64
sudo ./instalador-linux-amd64
```

Para saber a arquitetura da máquina, use `uname -m`: `x86_64` usa `amd64`,
`aarch64` usa `arm64`.

Para conferir se o download não veio corrompido:

```bash
wget https://github.com/DeathHapyness/self-hosted-media-server/releases/latest/download/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing
```

**Opção 2: compilar a partir do código**

Requer Go instalado (veja a versão em `go.mod`):

```bash
git clone https://github.com/DeathHapyness/self-hosted-media-server.git
cd self-hosted-media-server
go build -o instalador .
sudo ./instalador
```

O instalador precisa rodar como root (`sudo`) porque cria diretórios em `/opt`
e ajusta permissões de pastas usadas pelos containers.

### Opções

| Opção | O que faz |
|---|---|
| `--dry-run` | Simula tudo sem criar arquivos nem rodar comandos destrutivos |
| `--version` | Mostra a versão do instalador |

Exemplo: `sudo ./instalador-linux-amd64 --dry-run`

### O que o instalador faz

- Verifica sistema operacional, Docker ou Podman, Compose, daemon,
  montagem de `/mnt/media`, espaço em disco e portas livres.
- Cria a estrutura de diretórios em `/opt/media-server/` para cada serviço
  com dados persistentes (lista completa em `SERVICE_SUBDIRS`, no arquivo
  `installer/config.go`), sem apagar nada que já exista.
- Ajusta permissões das pastas que os containers precisam gravar.
- Cria um `.env` a partir de `.env.example`, caso ainda não exista.
- Valida cada `docker-compose.yml` com `docker compose config`.
- Sobe os serviços com `docker compose up -d`.
- Grava um log em `/var/log/media-server-installer.log` e um histórico em
  `installer/json/media-server-history.json`.
- Permite escolher entre Docker e Podman.
- Desinstala serviços e, se você confirmar, apaga as pastas de dados deles.
- Gerencia o disco de mídia: usa um disco existente ou cria um disco virtual.

### Executar novamente

O instalador é idempotente: rodar de novo não apaga configurações, bancos de
dados ou arquivos de mídia. Ele apenas cria o que estiver faltando e informa o
que já está configurado.

### Se `/mnt/media` não estiver montado

O instalador avisa e pergunta se você quer continuar mesmo assim, usando
`/mnt/media` como uma pasta comum no disco raiz.

Pelo menu "Gerenciar disco de mídia" também é possível formatar e montar um
disco existente ou criar um disco virtual (um arquivo montado em `/mnt/media`).
