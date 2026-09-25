## Instalação automática

Este repositório inclui um instalador (`install.py`) que prepara o servidor
do zero: verifica pré-requisitos, cria os diretórios necessários em
`/opt/media-server`, ajusta permissões, cria logs, valida o `docker-compose.yml` e sobe
os containers.

Instalador atualizado conforme novos programas são adicionados ao repo.

## <span style="color: red; font-weight: bold;">ADVERTÊNCIA:</span> 
<span style="color: darkorange;">Este script é fornecido "como está" e não se responsabiliza por quaisquer danos ou perda de dados. Use por sua própria conta e risco.</span>

<span style="color: #00fffb;">Esse script foi testado antes de estar nesse repo</span>

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

- Verifica sistema operacional, Docker,Podman, Docker Compose, daemon do Docker,
  montagem de `/mnt/media`, espaço em disco e portas livres(com permissão do user).
- Cria a estrutura de diretórios em `/opt/media-server/` para cada serviço
  com dados persistentes (lista completa e sempre atualizada em
  `SERVICE_DIRS` no `install.py`), sem apagar nada que já exista.
- Ajusta permissões das pastas que os containers precisam gravar.
- Cria um `.env` a partir de `.env.example`, caso ainda não exista.
- Valida o `docker-compose.yml` com `docker compose config`.
- Sobe os serviços com `docker compose up -d`.
- Cria um arquivo em json para log do instalador. 
- Pode instalar docker ou podman.
- Apaga e atualiza serviços.

### Executar novamente

O instalador é idempotente: rodar `sudo python3 install.py` de novo não
apaga configurações, bancos de dados ou arquivos de mídia — ele apenas cria
o que estiver faltando e informa o que já está configurado.

### Se `/mnt/media` não estiver montado

O instalador interrompe a execução e pergunta ao usuario se ele gostaria de montar isso e se sim cria um arquivo
dentro de `/mnt/media`.
O usuario tambem pode escolher outra pasta ou pode deixar na raiz dos arquivos.