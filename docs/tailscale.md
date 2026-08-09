# 🔗 Acesso Remoto com Tailscale

Para acessar o servidor fora da rede doméstica, o projeto utiliza o **Tailscale** como uma rede privada entre os dispositivos.

Com o Tailscale, não é necessário expor diretamente as portas do Jellyfin, Navidrome, qBittorrent, AdGuard Home ou SSH na internet.

A arquitetura fica:

```text
                         INTERNET
                             │
                             │
                         Tailscale
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
           📱 iPhone       💻 PC        🖥️ Servidor
                                        100.x.x.x
                                             │
                    ┌──────────────┬─────────┼──────────────┐
                    │              │         │              │
                    ▼              ▼         ▼              ▼
                 Jellyfin      Navidrome  qBittorrent   AdGuard Home
                  :8096          :4533       :8080          :3000
```

## Instalação

No servidor Linux:

```bash
curl -fsSL https://tailscale.com/install.sh | sh
```

Depois autentique o servidor:

```bash
sudo tailscale up
```

Verifique os dispositivos conectados:

```bash
tailscale status
```

Para descobrir o IP Tailscale do servidor:

```bash
tailscale ip
```

O servidor receberá um endereço da rede `100.x.x.x`.

Exemplo:

```text
100.95.231.76
```

## Acessando os serviços remotamente

Depois de instalar o Tailscale nos dispositivos e entrar na mesma conta, os serviços podem ser acessados através do IP Tailscale do servidor.

### Jellyfin

```text
http://IP DO SEU TAILSCALE:8096
```

### Navidrome

```text
http://IP DO SEU TAILSCALE:4533
```

### qBittorrent

```text
http://IP DO SEU TAILSCALE:8080
```

### AdGuard Home

```text
http://IP DO SEU TAILSCALE:3000
```

### SSH

```bash
ssh rique@IP DO SEU TAILSCALE
```

## Dispositivos

O Tailscale pode conectar diferentes dispositivos à mesma rede privada:

```text
Servidor
IP DO SEU TAILSCALE
      │
      ├── 📱 iPhone
      ├── 💻 PC / Linux
      └── 🤖 Android
```

Basta instalar o Tailscale em cada dispositivo e entrar na mesma conta.

## Vantagens

* 🔒 Não é necessário abrir portas no roteador
* 🌐 Acesso aos serviços fora de casa
* 📱 Acesso pelo celular
* 💻 Acesso pelo computador
* 🔑 Autenticação através da conta Tailscale
* 🛡️ Rede privada entre os dispositivos
* 🖥️ Possibilidade de acessar o servidor via SSH remotamente

> O Tailscale é utilizado neste projeto como camada de acesso remoto à infraestrutura, mantendo os serviços internos sem exposição direta à internet.