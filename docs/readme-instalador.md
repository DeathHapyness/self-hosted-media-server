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
| `--web-porta N` | Porta da interface web (padrão `8787`; `0` escolhe uma livre) |

Exemplo: `sudo ./instalador-linux-amd64 --dry-run`

### Modo terminal ou modo web

Ao iniciar, o instalador pergunta como você quer usá-lo:

- `1) Terminal (CLI)`: o padrão. Os menus e perguntas aparecem no próprio
  terminal. Apertar Enter sem digitar nada escolhe esta opção.
- `2) Navegador (Web)`: o instalador abre um pequeno servidor web e mostra um
  link de uso único. Você abre esse link no navegador de outro dispositivo da
  mesma rede e responde às mesmas perguntas por uma página, com botões e
  campos. O modo web também tem um editor para os `docker-compose.yml`.

**Atenção: o modo web não usa HTTPS.** A conexão é HTTP puro, sem
criptografia, então use-o apenas na sua rede doméstica e nunca exponha a porta
para a internet. Detalhes, recomendações de segurança e solução de problemas em
[`modo-web.md`](modo-web.md).

### O que o instalador faz

- Verifica sistema operacional, Docker ou Podman, Compose, daemon,
  montagem de `/mnt/media`, espaço em disco e portas livres.
- Cria a estrutura de diretórios em `/opt/media-server/` para cada serviço
  com dados persistentes (lista completa em `SERVICE_SUBDIRS`, no arquivo
  `installer/config.go`), sem apagar nada que já exista.
- Ajusta permissões das pastas que os containers precisam gravar.
- Cria um `.env` a partir de `.env.example`, caso ainda não exista.
- Valida cada `docker-compose.yml` com `docker compose config`.
- Quando a validação ou a subida de um serviço falha, analisa o erro do
  Compose e, se reconhecer a causa, explica o problema e oferece uma correção
  (veja "Diagnóstico e correção automática" abaixo).
- Sobe os serviços com `docker compose up -d`.
- Grava um log em `/var/log/media-server-installer.log` e um histórico em
  `installer/json/media-server-history.json`.
- Permite escolher entre Docker e Podman.
- Desinstala serviços e, se você confirmar, apaga as pastas de dados deles.
- Gerencia o disco de mídia: usa um disco existente ou cria um disco virtual.
- Opcionalmente, roda pelo navegador (modo web), com editor de
  `docker-compose.yml`.

### Diagnóstico e correção automática

Se `docker compose config` ou `docker compose up -d` falhar para algum serviço,
o instalador lê a mensagem de erro e procura causas conhecidas. Hoje ele
reconhece:

- **Arquivo `.env` do serviço ausente**: o `docker-compose.yml` usa `env_file`,
  mas o arquivo não existe. Se houver um arquivo de exemplo na pasta do serviço
  (`env-example.env`, `.env.example`, `env.example` ou `.env.sample`), o
  instalador oferece criar o `.env` a partir dele.
- **Variável não definida**: o compose usa uma variável que não está no
  `.env`. Se o `.env` não existir, ele cai no caso acima. Se existir, o
  instalador só informa qual variável está faltando, para você completar à mão.

Como a correção do `.env` funciona:

1. O instalador **sempre pergunta antes** (`Posso criar ... ? [s/N]`). Sem um
   `s`, nada é alterado.
2. Ele copia o arquivo de exemplo e preenche só os campos vazios ou com texto
   de exemplo (como `SUA SENHA AQUI`). Valores que já estão preenchidos no
   exemplo são mantidos.
3. Senhas e chaves (nomes com `PASSWORD`, `SECRET`, `TOKEN`, terminados em
   `_KEY`, e `APP_KEY`) são **geradas automaticamente** com valores
   aleatórios. Os demais campos são perguntados, com uma sugestão padrão.
4. O `.env` é criado com permissão `600` (só o dono lê), e o dono é o usuário
   que chamou o `sudo`, não o root.
5. Depois de corrigir, o instalador tenta validar ou subir o serviço de novo.

Os valores gerados ficam no próprio `.env` do serviço; consulte-os lá quando
precisar fazer login. Com `--dry-run`, o instalador só mostra o que criaria.

### Executar novamente

O instalador é idempotente: rodar de novo não apaga configurações, bancos de
dados ou arquivos de mídia. Ele apenas cria o que estiver faltando e informa o
que já está configurado.

### Se `/mnt/media` não estiver montado

O instalador avisa e pergunta se você quer continuar mesmo assim, usando
`/mnt/media` como uma pasta comum no disco raiz.

Pelo menu "Gerenciar disco de mídia" também é possível formatar e montar um
disco existente ou criar um disco virtual (um arquivo montado em `/mnt/media`).
