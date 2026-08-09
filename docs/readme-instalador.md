## Instalação automática

Este repositório inclui um instalador (`install.py`) que prepara o servidor
do zero: verifica pré-requisitos, cria os diretórios necessários em
`/opt/media-server`, ajusta permissões, valida o `docker-compose.yml` e sobe
os containers.

## <span style="color: red; font-weight: bold;">ADVERTÊNCIA:</span> 
<span style="color: darkorange;">Este script é fornecido "como está" e não se responsabiliza por quaisquer danos ou perda de dados. Use por sua própria conta e risco.</span>

### Pré-requisitos

- Linux
- Docker + plugin Docker Compose v2 (`docker compose version` deve funcionar)
- Disco de mídia montado em `/mnt/media`
- Python 3.8 ou superior (já vem instalado na maioria das distros)

### Como executar

```bash
git clone <url-do-repositorio>
cd <pasta-do-repositorio>
sudo python3 install.py
```

O script precisa rodar como root (`sudo`) porque cria diretórios em `/opt`
e ajusta permissões de pastas usadas pelos containers.

### O que o instalador faz

- Verifica sistema operacional, Docker, Docker Compose, daemon do Docker,
  montagem de `/mnt/media`, espaço em disco e portas livres.
- Cria a estrutura de diretórios em `/opt/media-server/` para cada serviço
  (AdGuard Home, Dozzle, File Browser, Jellyfin, Navidrome, qBittorrent,
  Homepage), sem apagar nada que já exista.
- Ajusta permissões das pastas que os containers precisam gravar.
- Cria um `.env` a partir de `.env.example`, caso ainda não exista.
- Valida o `docker-compose.yml` com `docker compose config`.
- Sobe os serviços com `docker compose up -d`.

### Executar novamente

O instalador é idempotente: rodar `sudo python3 install.py` de novo não
apaga configurações, bancos de dados ou arquivos de mídia — ele apenas cria
o que estiver faltando e informa o que já está configurado.

### Se `/mnt/media` não estiver montado

O instalador interrompe a execução imediatamente e não cria nenhum arquivo
dentro de `/mnt/media`, para evitar gravar dados no disco raiz por engano.
Monte o disco de mídia (por exemplo, via `/etc/fstab`) e rode o instalador
novamente.