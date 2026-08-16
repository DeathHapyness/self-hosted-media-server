# 🛡️ Gluetun + qBittorrent

O **Gluetun** é um cliente VPN executado em Docker que cria um túnel VPN e aplica um firewall para impedir que o tráfego dos containers escape pela conexão normal.

Neste setup, o **qBittorrent compartilha a rede do container Gluetun**. Dessa forma, o tráfego de rede do qBittorrent passa pelo túnel VPN.

Este projeto pode ser utilizado com:

* **OpenVPN** — para provedores que ainda oferecem OpenVPN
* **WireGuard** — recomendado para provedores modernos, como o Mullvad
* **Mullvad** — atualmente utilizando WireGuard

> ⚠️ **Importante:** o Mullvad encerrou completamente o suporte a OpenVPN em **15 de janeiro de 2026**. Portanto, configurações atuais do Mullvad devem utilizar WireGuard.

---

## 📁 Criar diretórios

Crie o diretório onde o stack ficará armazenado:

```bash
sudo mkdir -p /opt/gluetun
cd /opt/gluetun

mkdir -p gluetun
mkdir -p qbittorrent/config
mkdir -p qbittorrent/downloads
```

Opcionalmente, ajuste o proprietário dos diretórios para o usuário que executará o qBittorrent:

```bash
sudo chown -R $USER:$USER /opt/gluetun
```

---

# 🐳 Docker Compose

Crie o arquivo:

```bash
nano docker-compose.yml
```

Utilize:

```yaml
services:

  gluetun:
    image: qmcgaw/gluetun:latest
    container_name: gluetun
    restart: unless-stopped

    cap_add:
      - NET_ADMIN

    devices:
      - /dev/net/tun:/dev/net/tun

    ports:
      # qBittorrent WebUI
      - "8080:8080"

    volumes:
      - ./gluetun:/gluetun

    environment:
      # ==========================================
      # VPN
      # ==========================================

      VPN_SERVICE_PROVIDER=${VPN_SERVICE_PROVIDER}
      VPN_TYPE=${VPN_TYPE}

      # ==========================================
      # Servidor VPN
      # ==========================================

      SERVER_COUNTRIES=${SERVER_COUNTRIES}

      # ==========================================
      # OpenVPN
      # Utilizado somente quando VPN_TYPE=openvpn
      # ==========================================

      OPENVPN_USER=${OPENVPN_USER}
      OPENVPN_PASSWORD=${OPENVPN_PASSWORD}

      # ==========================================
      # Mullvad / WireGuard
      # Utilizado quando VPN_SERVICE_PROVIDER=mullvad
      # ==========================================

      WIREGUARD_PRIVATE_KEY=${WIREGUARD_PRIVATE_KEY}
      WIREGUARD_ADDRESSES=${WIREGUARD_ADDRESSES}

    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "https://www.google.com"]
      interval: 30s
      timeout: 10s
      retries: 3

  qbittorrent:
    image: lscr.io/linuxserver/qbittorrent:latest
    container_name: qbittorrent
    restart: unless-stopped

    environment:
      - PUID=${PUID}
      - PGID=${PGID}
      - TZ=${TZ}
      - WEBUI_PORT=8080

    volumes:
      - ./qbittorrent/config:/config
      - ./qbittorrent/downloads:/downloads

    network_mode: "service:gluetun"

    depends_on:
      gluetun:
        condition: service_healthy
```

### ⚠️ Sobre `network_mode`

O ponto mais importante deste Compose é:

```yaml
network_mode: "service:gluetun"
```

Isso faz o qBittorrent utilizar a mesma rede do container `gluetun`.

Consequentemente:

```text
Internet
   │
   ▼
┌───────────────┐
│    Gluetun    │
│     VPN       │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│  qBittorrent  │
└───────────────┘
```

O qBittorrent não possui uma interface de rede Docker independente. Todo o tráfego passa pela rede do Gluetun.

O projeto original utilizado como referência também emprega essa arquitetura.

---

# 🔐 Variáveis de ambiente

Para não colocar credenciais diretamente no `docker-compose.yml`, crie um arquivo `.env`:

```bash
nano .env
```

Não versione esse arquivo no Git.

Adicione:

```env
PUID=1000
PGID=1000
TZ=America/Sao_Paulo

VPN_SERVICE_PROVIDER=
VPN_TYPE=

SERVER_COUNTRIES=

OPENVPN_USER=
OPENVPN_PASSWORD=

WIREGUARD_PRIVATE_KEY=
WIREGUARD_ADDRESSES=
```

O `.env` deve ser adicionado ao `.gitignore`:

```gitignore
.env
gluetun/
qbittorrent/config/
qbittorrent/downloads/
```

---

# 🔵 Opção 1 — OpenVPN

Use esta opção para provedores que ainda disponibilizam OpenVPN.

O Gluetun utiliza as seguintes variáveis para autenticação OpenVPN:

```env
VPN_SERVICE_PROVIDER=nome-do-provedor
VPN_TYPE=openvpn

OPENVPN_USER=seu_usuario
OPENVPN_PASSWORD=sua_senha

SERVER_COUNTRIES=Netherlands
```

O Gluetun possui suporte a OpenVPN e utiliza `OPENVPN_USER` e `OPENVPN_PASSWORD` como credenciais padrão para esse modo.

### Exemplo

Um exemplo utilizando um provedor compatível seria:

```env
VPN_SERVICE_PROVIDER=airvpn
VPN_TYPE=openvpn

OPENVPN_USER=SEU_USUARIO
OPENVPN_PASSWORD=SUA_SENHA

SERVER_COUNTRIES=Netherlands
```

Depois:

```bash
docker compose up -d
```

Verifique os logs:

```bash
docker compose logs -f gluetun
```

Procure por uma mensagem indicando que a VPN foi estabelecida com sucesso.

---

# 🟢 Opção 2 — Mullvad + WireGuard

## ⚠️ Mullvad não utiliza mais OpenVPN

Para instalações atuais do Mullvad, utilize:

```env
VPN_SERVICE_PROVIDER=mullvad
VPN_TYPE=wireguard
```

A documentação atual do Gluetun para Mullvad utiliza WireGuard e requer a chave privada e o endereço WireGuard.

---

## 1. Gerar uma configuração WireGuard

No Mullvad, gere uma configuração WireGuard para o dispositivo.

Você precisará obter:

```text
Private key
Address
```

Por exemplo:

```text
Private key:
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=

Address:
10.64.222.21/32
```

**Não compartilhe sua chave privada.**

---

## 2. Configurar o `.env`

Edite:

```bash
nano .env
```

E configure:

```env
PUID=1000
PGID=1000
TZ=America/Sao_Paulo

VPN_SERVICE_PROVIDER=mullvad
VPN_TYPE=wireguard

SERVER_COUNTRIES=Brazil

WIREGUARD_PRIVATE_KEY=SUA_CHAVE_PRIVADA
WIREGUARD_ADDRESSES=10.64.222.21/32
```

Você também pode escolher uma cidade específica:

```env
SERVER_CITIES=Sao Paulo
```

ou utilizar outro país:

```env
SERVER_COUNTRIES=Netherlands
```

O Gluetun permite filtrar servidores do Mullvad por país, cidade ou hostname.

---

# 🚀 Iniciar os containers

Depois de configurar o `.env`:

```bash
docker compose up -d
```

Confira:

```bash
docker compose ps
```

O resultado deverá indicar os containers em execução:

```text
NAME          STATUS
gluetun       Up
qbittorrent   Up
```

---

# 📋 Verificar os logs do Gluetun

Para acompanhar a conexão:

```bash
docker compose logs -f gluetun
```

Ou:

```bash
docker logs -f gluetun
```

Para verificar apenas as últimas linhas:

```bash
docker logs gluetun --tail 100
```

---

# 🌐 Acessar o qBittorrent

Como a porta `8080` foi publicada pelo Gluetun:

```text
http://IP_DO_SERVIDOR:8080
```

Exemplo:

```text
http://192.168.1.10:8080
```

O qBittorrent está compartilhando a rede do Gluetun, portanto a porta é publicada no serviço `gluetun`, e não no serviço `qbittorrent`.

---

# 🔑 Senha inicial do qBittorrent

Na primeira inicialização, o qBittorrent gera uma senha temporária para o usuário:

```text
admin
```

Consulte:

```bash
docker logs qbittorrent
```

Procure pela senha temporária gerada pelo qBittorrent.

Depois de entrar no WebUI, altere a senha em:

```text
Tools
  → Options
    → Web UI
```

---

# 🔒 Configurar o qBittorrent para utilizar somente a VPN

Mesmo utilizando:

```yaml
network_mode: "service:gluetun"
```

é recomendado configurar o qBittorrent para utilizar explicitamente a interface VPN.

No qBittorrent:

```text
Tools
→ Options
→ Advanced
→ Network Interface
```

Selecione:

```text
tun0
```

para OpenVPN.

Para WireGuard, dependendo da configuração utilizada pelo Gluetun, a interface pode ser:

```text
wg0
```

A interface deve corresponder ao túnel VPN criado pelo Gluetun.

Essa configuração evita que o qBittorrent tente utilizar outra interface caso a VPN seja interrompida. O projeto de referência também recomenda fazer o bind do qBittorrent à interface VPN.

---

# 🧪 Testar o IP público

É importante confirmar que o tráfego está realmente saindo pelo VPN.

Execute:

```bash
docker exec -it qbittorrent sh
```

Dentro do container:

```bash
wget -qO- https://api.ipify.org
```

Ou:

```bash
curl -s https://ipinfo.io/ip
```

O IP retornado deve ser o IP público do servidor VPN, e **não o IP público da sua conexão residencial**.

Também é possível executar diretamente:

```bash
docker exec qbittorrent wget -qO- https://api.ipify.org
```

---

# 🛡️ Testar o Kill Switch

Uma das funções importantes do Gluetun é impedir que o tráfego escape pela conexão normal quando o túnel VPN cai.

Primeiro verifique o IP:

```bash
docker exec qbittorrent wget -qO- https://api.ipify.org
```

Depois pare o Gluetun:

```bash
docker stop gluetun
```

Tente novamente:

```bash
docker exec qbittorrent wget -qO- https://api.ipify.org
```

O tráfego deverá ficar bloqueado enquanto o Gluetun estiver parado.

Inicie novamente:

```bash
docker start gluetun
```

Depois aguarde a reconexão e teste novamente:

```bash
docker exec qbittorrent wget -qO- https://api.ipify.org
```

---

# 🔄 Reiniciar o stack

Para reiniciar:

```bash
docker compose restart
```

Para parar:

```bash
docker compose down
```

Para iniciar novamente:

```bash
docker compose up -d
```

---

# 📦 Atualizar as imagens

Baixe as imagens mais recentes:

```bash
docker compose pull
```

Recrie os containers:

```bash
docker compose up -d
```

Remova imagens antigas, se desejar:

```bash
docker image prune
```

---

# 🩺 Diagnóstico

## Gluetun não inicia

Verifique:

```bash
docker compose logs gluetun
```

Confira principalmente:

* credenciais da VPN;
* `VPN_SERVICE_PROVIDER`;
* `VPN_TYPE`;
* país/cidade selecionados;
* chave WireGuard;
* endereço WireGuard;
* existência de `/dev/net/tun`.

Verifique o dispositivo:

```bash
ls -l /dev/net/tun
```

Caso não exista, o host pode não ter o módulo TUN disponível.

---

## qBittorrent não abre

Confira os containers:

```bash
docker compose ps
```

Depois:

```bash
docker compose logs qbittorrent
```

Como o qBittorrent utiliza:

```yaml
network_mode: "service:gluetun"
```

ele depende da rede do Gluetun.

Também confirme que a porta está publicada no serviço `gluetun`:

```yaml
ports:
  - "8080:8080"
```

e não no serviço `qbittorrent`.

---

## VPN conecta, mas não há Internet

Verifique:

```bash
docker logs gluetun
```

Depois teste:

```bash
docker exec qbittorrent wget -qO- https://api.ipify.org
```

Também confira se o servidor selecionado pelo filtro está disponível.

Para reduzir problemas, inicialmente utilize apenas:

```env
SERVER_COUNTRIES=Brazil
```

e, depois que estiver funcionando, refine a configuração.

---

# 🔐 Segurança

Não coloque credenciais diretamente no `docker-compose.yml`.

Evite versionar:

```text
.env
gluetun/
qbittorrent/config/
```

O arquivo `.env` pode conter:

```text
OPENVPN_USER
OPENVPN_PASSWORD
WIREGUARD_PRIVATE_KEY
```

A chave privada do WireGuard deve ser tratada como uma credencial.

Se ela for exposta, gere uma nova configuração/chave no provedor VPN.

---

# 📚 Referências

* [Gluetun](https://github.com/qdm12/gluetun)
* [Gluetun — configuração OpenVPN](https://github.com/qdm12/gluetun-wiki/blob/main/setup/options/openvpn.md)
* [Gluetun — configuração Mullvad](https://github.com/qdm12/gluetun-wiki/blob/main/setup/providers/mullvad.md)
* [Gluetun + qBittorrent](https://github.com/tonyp7/gluetun-qbittorrent)
* [Mullvad — remoção do OpenVPN](https://mullvad.net/en/help/linux-openvpn-installation)

---

## ⚠️ Observação sobre Mullvad

Se você encontrar tutoriais antigos mostrando:

```env
VPN_SERVICE_PROVIDER=mullvad
VPN_TYPE=openvpn
OPENVPN_USER=...
```

**não siga essa parte para uma instalação atual.**

O OpenVPN foi removido dos servidores Mullvad em janeiro de 2026. Para Mullvad, utilize:

```env
VPN_SERVICE_PROVIDER=mullvad
VPN_TYPE=wireguard
WIREGUARD_PRIVATE_KEY=...
WIREGUARD_ADDRESSES=...
```

O próprio Gluetun removeu o exemplo OpenVPN da documentação do Mullvad justamente por essa mudança.
