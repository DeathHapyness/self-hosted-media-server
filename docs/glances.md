# 📊 Glances

O **Glances** é uma ferramenta de monitoramento de sistema em tempo real (CPU, memória, disco, rede, sensores e containers Docker), disponível como painel web.

## 📁 Estrutura

```text
glances/
└── docker-compose.yml
```

O Glances não persiste dados em disco — ele apenas lê informações do host e do Docker em tempo real.

## Docker Compose

```yaml
services:
  glances:
    image: nicolargo/glances:latest-full
    container_name: glances
    restart: unless-stopped
    pid: host
    ports:
      - "61208:61208"
    environment:
      GLANCES_OPT: "-w"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
```

* `pid: host` — permite que o Glances enxergue os processos do host, e não apenas os do próprio container.
* `GLANCES_OPT: "-w"` — inicia o Glances em modo servidor web, expondo o painel na porta `61208`.
* `/var/run/docker.sock:ro` — acesso somente leitura ao socket do Docker, usado para exibir estatísticas dos containers em execução.

## Iniciar

```bash
cd glances
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

O Glances utiliza a porta `61208`.

```text
http://IP_DO_SERVIDOR:61208
```

Exemplo:

```text
http://192.168.15.118:61208
```

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=glances
```

Ver logs:

```bash
docker logs glances
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

* Não exponha a porta `61208` diretamente à internet — prefira Tailscale/VPN ou rede local.
* O acesso somente leitura ao socket do Docker (`:ro`) permite que o Glances leia estatísticas dos containers, mas evita que ele os controle ou modifique.

## 📚 Referências

* [Repositório oficial do Glances](https://github.com/nicolargo/glances)
* [Documentação oficial](https://glances.readthedocs.io/)
