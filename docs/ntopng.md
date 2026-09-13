# 📡 ntopng

O ntopng é utilizado para **monitorar o tráfego de rede** em tempo real: hosts ativos, protocolos, volume de dados trafegado e comportamento da rede local.

Ele usa o Redis como backend para armazenar séries temporais e estado interno.

## 📁 Estrutura

```text
ntopng/
├── redis-data/
├── ntopng-data/
└── docker-compose.yml
```

As pastas `redis-data/` e `ntopng-data/` contêm dados gerados pelos containers e não devem ser enviadas para o GitHub (já cobertas pelo `.gitignore` da raiz do repositório).

## Docker Compose

```yaml
services:
  redis:
    image: redis:7-alpine
    container_name: ntopng-redis
    restart: unless-stopped
    volumes:
      - ./redis-data:/data
    command: redis-server --appendonly yes

  ntopng:
    image: ntop/ntopng:latest
    container_name: ntopng
    restart: unless-stopped
    depends_on:
      - redis
    command: >
      --redis=redis:6379
      --http-port=3000
    ports:
      - "3100:3000"
    volumes:
      - ./ntopng-data:/var/lib/ntopng
```

## Iniciar

```bash
cd ntopng
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

O ntopng utiliza a porta `3100` (mapeada para a porta interna `3000` do container).

```text
http://IP_DO_SERVIDOR:3100
```

Exemplo:

```text
http://192.168.15.118:3100
```

No primeiro acesso, o usuário/senha padrão é `admin` / `admin`. Troque a senha imediatamente pelo próprio painel (`Settings → Users`).

## 🌐 O que este compose monitora

Por padrão, este `docker-compose.yml` **não** usa `network_mode: host` nem privilégios de captura de pacotes — o ntopng enxerga apenas o tráfego da própria rede interna do Docker (bridge), não o tráfego real da sua LAN/roteador.

Isso é intencional: mantém o serviço simples e sem privilégios elevados. Ele já é útil para visualizar o tráfego entre os containers do próprio servidor.

Se você quiser que o ntopng monitore o tráfego real da interface de rede do host (todos os dispositivos da LAN), é necessário adaptar o compose, por exemplo:

```yaml
  ntopng:
    network_mode: "host"
    cap_add:
      - NET_ADMIN
      - NET_RAW
    command: >
      --redis=127.0.0.1:6379
      --interface=eth0
      --http-port=3100
```

> ⚠️ `network_mode: host` remove o isolamento de rede do container e expõe todas as portas do serviço diretamente na interface do host. Ajuste o `--interface` para o nome real da sua interface de rede (verifique com `ip a`). Use essa opção apenas se realmente precisar monitorar o tráfego físico da rede.

## 🐳 Gerenciamento

Verificar os containers:

```bash
docker ps --filter name=ntopng
```

Ver logs:

```bash
docker logs ntopng
docker logs ntopng-redis
```

Acompanhar os logs:

```bash
docker logs -f ntopng
```

Reiniciar:

```bash
docker compose restart
```

Parar:

```bash
docker compose down
```

Atualizar:

```bash
docker compose pull
docker compose up -d
```

## 🔒 Segurança

O ntopng expõe detalhes sobre os dispositivos e o tráfego da rede. Por isso:

* Não exponha a porta `3100` diretamente à internet.
* Troque a senha padrão (`admin`/`admin`) imediatamente após o primeiro acesso.
* Prefira acessar via Tailscale/VPN ou apenas pela rede local.
* Se usar `network_mode: host` para capturar tráfego real da LAN, redobre a atenção — o container passa a ter acesso direto à rede do host.

## 📚 Referências

* [Repositório oficial do ntopng](https://github.com/ntop/ntopng)
* [Documentação oficial do ntopng](https://www.ntop.org/guides/ntopng/)
* [Imagem Docker oficial](https://hub.docker.com/r/ntop/ntopng)
