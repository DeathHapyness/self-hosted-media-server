# 🛡️ AdGuard Home

O AdGuard Home atua como **servidor DNS local**, bloqueando anúncios e rastreadores para todos os dispositivos da rede que o utilizam como resolvedor.

## Criar diretórios

```bash
sudo mkdir -p /opt/adguard
cd /opt/adguard

mkdir -p work
mkdir -p conf
```

## Docker Compose

```yaml
services:
  adguardhome:
    image: adguard/adguardhome
    container_name: adguardhome
    restart: unless-stopped
    ports:
      - "53:53/tcp"
      - "53:53/udp"
      - "3000:3000/tcp"
      - "8081:80/tcp"
    volumes:
      - ./work:/opt/adguardhome/work
      - ./conf:/opt/adguardhome/conf
```

Inicie:

```bash
docker compose up -d
```

## Primeiro acesso

Acesse a interface de configuração inicial:

```text
http://IP_DO_SERVIDOR:3000
```

No primeiro acesso, o AdGuard Home guia a criação do usuário administrador e a configuração inicial (interface de escuta, porta do painel, upstream DNS). Como o repositório não inclui a config real (`conf/AdGuardHome.yaml` é ignorado pelo Git), esse arquivo é gerado automaticamente na primeira execução.

## Usando como DNS da rede

Depois de configurado, aponte o DNS do seu roteador (ou dos dispositivos individualmente) para o IP do servidor na porta `53`, para que o bloqueio de anúncios valha para toda a rede.

> ⚠️ A porta `80` do container está mapeada para `8081` no host neste setup, para não conflitar com outros serviços que já usem a porta 80.

## O que não é versionado

* `work/` — logs, estatísticas e histórico de consultas DNS
* `conf/AdGuardHome.yaml` — contém o hash da senha do administrador e a lista de clientes/dispositivos da rede

Apenas o `docker-compose.yml` (sem dados sensíveis) é versionado no repositório.