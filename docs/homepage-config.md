# ⚙️ Homepage — Referência da pasta `config/`

Este documento explica **para que serve cada arquivo** dentro de `homepage/config/`. Todo o conteúdo aqui é genérico/ilustrativo — nenhum IP, token ou dado real deve ser copiado destes exemplos para produção, e nada do seu ambiente real foi usado para escrevê-lo.

> Exemplo de config em [homepage/config](homepage/config/).

## 📁 Visão geral

```text
homepage/config/
├── settings.yaml      # aparência e comportamento geral do dashboard
├── bookmarks.yaml      # links rápidos (sem widget/status)
├── services.yaml       # serviços exibidos no dashboard, com widgets
├── docker.yaml          # conexões com o(s) Docker socket(s) para auto-descoberta
├── widgets.yaml         # widgets no topo do dashboard (clima, busca, recursos...)
├── kubernetes.yaml      # conexão com cluster(s) Kubernetes (opcional)
├── proxmox.yaml         # integração com Proxmox VE (opcional)
├── custom.css            # CSS customizado do dashboard
├── custom.js             # JavaScript customizado do dashboard
└── logs/                 # logs internos do Homepage (não versionar)
```

Nenhum desses arquivos é obrigatório — o Homepage sobe com uma configuração padrão mesmo sem eles. Crie apenas os que você for usar.

---

## `settings.yaml`

Configurações globais: título, tema, idioma, layout dos grupos, plano de fundo, provedores usados pelos widgets, etc.

```yaml
title: Meu Servidor
theme: dark          # dark | light
color: slate
layout:
  Media:
    style: row
    columns: 3
  Downloads:
    style: row
    columns: 2

providers:
  openweathermap: openweathermapapikey
```

> `providers` centraliza chaves de API reaproveitadas por múltiplos widgets — trate como segredo.

## `bookmarks.yaml`

Links rápidos, sem status/widget — útil para documentação, roteador, wiki interna, etc.

```yaml
- Desenvolvedor:
    - GitHub:
        - abbr: GH
          href: https://github.com/

- Rede:
    - Roteador:
        - abbr: RT
          href: http://IP_DO_SERVIDOR/
```

## `services.yaml`

Os serviços exibidos no dashboard, organizados em grupos. É o arquivo mais usado no dia a dia.

```yaml
- Mídia:
    - Jellyfin:
        icon: jellyfin.png
        href: http://IP_DO_SERVIDOR:8096
        description: Filmes e séries
        widget:
          type: jellyfin
          url: http://IP_DO_SERVIDOR:8096
          key: SUA_API_KEY_AQUI

- Downloads:
    - qBittorrent:
        icon: qbittorrent.png
        href: http://IP_DO_SERVIDOR:8080
        description: Gerenciador de downloads
```

Como alternativa a manter tudo aqui, muitos serviços podem ser **descobertos automaticamente** via labels no próprio `docker-compose.yml` (veja `docker.yaml` abaixo) — é o padrão já usado pelos `docker-compose.yml` deste repositório, por exemplo:

```yaml
labels:
  - homepage.group=Mídia
  - homepage.name=Jellyfin
  - homepage.icon=jellyfin.png
  - homepage.href=http://IP_DO_SERVIDOR:8096
  - homepage.description=Filmes e séries
```

## `docker.yaml`

Define a quais Docker sockets o Homepage deve se conectar para descobrir containers via `labels` (necessário montar `/var/run/docker.sock` no container, como descrito em [docs/homepage.md](homepage.md)).

```yaml
my-docker:
  socket: /var/run/docker.sock

# Exemplo de host remoto via TCP+TLS (opcional):
# outro-servidor:
#   host: 192.168.1.50
#   port: 2376
#   tls: true
```

> Se você expõe o Docker via TCP, use sempre TLS e restrinja o acesso por firewall/VPN — a API do Docker equivale a acesso root no host.

## `widgets.yaml`

Widgets exibidos no topo do dashboard (clima, barra de busca, uso de recursos do sistema, saudação, etc.).

```yaml
- resources:
    cpu: true
    memory: true
    disk: /

- search:
    provider: duckduckgo
    target: _blank

- openmeteo:
    label: Minha Cidade
    latitude: 0
    longitude: 0
    units: metric
```

Widgets que dependem de serviços de terceiros (clima, notícias, etc.) costumam pedir uma chave de API — prefira referenciar via `{{HOMEPAGE_VAR_...}}` (variáveis de ambiente do container) em vez de escrever a chave direto no arquivo.

## `kubernetes.yaml`

Só é relevante se você conectar o Homepage a um cluster **Kubernetes** — não é o caso deste repositório, que é 100% Docker Compose. Pode ficar ausente ou vazio.

```yaml
mode: cluster   # ou "kubeconfig"
```

Se usado com `mode: kubeconfig`, o Homepage lê o `kubeconfig` do cluster — trate esse arquivo como credencial (acesso administrativo ao cluster).

## `proxmox.yaml`

Integração com o widget do Proxmox VE, caso o servidor rode dentro (ou ao lado) de um host Proxmox.

```yaml
- cluster:
    url: https://IP_DO_PROXMOX:8006
    username: homepage@pve
    password: "{{HOMEPAGE_VAR_PROXMOX_TOKEN}}"
    realm: pam
    node: pve
```

Recomendações:

* Crie um usuário/token **dedicado e somente-leitura** no Proxmox para o Homepage, nunca use a conta root/admin.
* Nunca escreva o token diretamente no arquivo — use uma variável de ambiente (`{{HOMEPAGE_VAR_...}}`) definida no `docker-compose.yml`/`.env` do Homepage.

## `custom.css`

CSS customizado, injetado no dashboard. Útil para ajustes visuais pontuais.

```css
:root {
  --color-primary: #5c6bc0;
}

body {
  font-family: "Inter", sans-serif;
}
```

## `custom.js`

JavaScript customizado, injetado e executado no navegador junto com o dashboard.

```js
console.log("Homepage carregado");
```

> ⚠️ Este script roda no contexto do seu navegador com acesso à página. Só coloque aqui código de sua autoria ou de fontes em que você confia plenamente — nunca cole scripts de origem desconhecida.

## `logs/`

Diretório onde o Homepage grava seus próprios logs internos. Não deve ser versionado (já coberto pelo `.gitignore`). Útil para depuração:

```bash
docker logs homepage
```

ou, se persistido em disco:

```bash
tail -f homepage/config/logs/*.log
```

---

## 🔒 Antes de versionar qualquer coisa em `config/`

Revise manualmente cada arquivo e nunca faça commit de:

* IPs públicos, domínios reais ou nomes de host internos;
* Tokens de API, senhas ou variáveis `key:` preenchidas (Jellyfin, Proxmox, provedores de clima, etc.);
* `kubeconfig` ou qualquer credencial de cluster;
* Logs (`logs/`).

Se quiser compartilhar sua configuração publicamente (ex.: como referência para outras pessoas), prefira criar cópias `*.example.yaml` com os valores acima substituídos por placeholders, em vez de versionar os arquivos reais usados em produção.

## 📚 Referências

* [Documentação oficial do Homepage](https://gethomepage.dev/)
* [Configuração de serviços](https://gethomepage.dev/configs/services/)
* [Configuração de widgets](https://gethomepage.dev/widgets/)
* [Integração com Docker](https://gethomepage.dev/configs/docker/)
* [docs/homepage.md](homepage.md) — instalação e primeiro acesso
