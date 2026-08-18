# 💾 Scrutiny

O Scrutiny é utilizado para monitoramento da saúde dos discos rígidos e SSDs através dos dados S.M.A.R.T.

Ele permite acompanhar informações como temperatura, saúde do disco, erros e outros indicadores fornecidos pelo S.M.A.R.T.

## 📁 Estrutura

```text
scrutiny/
├── config/
├── influxdb/
└── docker-compose.yml
```

As pastas `config/` e `influxdb/` contêm dados gerados pelo Scrutiny e não devem ser enviadas para o GitHub.

## Docker Compose

```yaml
services:
  scrutiny:
    image: ghcr.io/analogj/scrutiny:master-omnibus
    container_name: scrutiny
    restart: unless-stopped

    ports:
      - "8081:8080"

    cap_add:
      - SYS_RAWIO
      - SYS_ADMIN

    volumes:
      - ./config:/opt/scrutiny/config
      - ./influxdb:/opt/scrutiny/influxdb
      - /run/udev:/run/udev:ro
      - /dev:/dev
```

## Iniciar

```bash
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

O Scrutiny utiliza a porta `8081`.

```text
http://IP_DO_SERVIDOR:8081
```

Exemplo:

```text
http://192.168.15.118:8081
```

## 💾 Monitoramento dos discos

O Scrutiny precisa acessar os dispositivos `/dev` do servidor para conseguir consultar os dados S.M.A.R.T.

Exemplo de discos:

```text
/dev/sda
/dev/sdb
/dev/sdc
/dev/sdd
```

Verificar os discos do servidor:

```bash
lsblk -o NAME,SIZE,TYPE,MODEL,SERIAL
```

## 🔍 Verificar dispositivos dentro do container

```bash
docker exec scrutiny lsblk
```

Também é possível verificar os dispositivos:

```bash
docker exec scrutiny ls -l /dev/sd*
```

## 📊 S.M.A.R.T.

Para verificar os dados S.M.A.R.T. diretamente no servidor:

```bash
sudo smartctl -a /dev/sda
```

Exemplo para outro disco:

```bash
sudo smartctl -a /dev/sdb
```

## 🐳 Gerenciamento

Verificar o container:

```bash
docker ps --filter name=scrutiny
```

Ver logs:

```bash
docker logs scrutiny
```

Acompanhar os logs:

```bash
docker logs -f scrutiny
```

Reiniciar:

```bash
docker restart scrutiny
```

Parar:

```bash
docker compose down
```

Iniciar:

```bash
docker compose up -d
```

Atualizar:

```bash
docker compose pull
docker compose up -d
```

## 🔒 Configuração

Os dados do Scrutiny ficam armazenados localmente:

```text
scrutiny/config/
scrutiny/influxdb/
```

Essas pastas não devem ser enviadas para o GitHub.

O `.gitignore` pode conter:

```gitignore
config/
influxdb/
*.db
*.db-shm
*.db-wal
*.log
```

## ⚠️ Observação

O Scrutiny precisa de acesso aos dispositivos físicos do servidor para conseguir coletar informações S.M.A.R.T.

Em ambientes Docker, o acesso aos discos pode exigir permissões adicionais, como:

```yaml
cap_add:
  - SYS_RAWIO
  - SYS_ADMIN
```

Além do acesso:

```yaml
volumes:
  - /dev:/dev
  - /run/udev:/run/udev:ro
```

Sem esse acesso, o Scrutiny pode iniciar normalmente, mas não conseguir detectar ou coletar informações dos discos.