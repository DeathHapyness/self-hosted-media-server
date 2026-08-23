# 🤖 Instalação do n8n

O n8n é uma plataforma de automação que permite criar workflows conectando APIs, serviços e aplicações.

Doc oficial n8n:https://n8n-brasil.github.io/n8n-Doc-PT-BR/primeiros-passos/instalacao-self-hosted

## 1. Criar o diretório

```bash
mkdir -p ~/media-server/n8n
cd ~/media-server/n8n
```

## 2. Criar o `docker-compose.yml`

```bash
nano docker-compose.yml
```

Utilize:

```yaml
services:
  n8n:
    image: docker.n8n.io/n8nio/n8n:latest
    container_name: n8n

    restart: unless-stopped

    ports:
      - "5678:5678"

    volumes:
      - n8n_data:/home/node/.n8n

    environment:
      - TZ=America/Sao_Paulo
      - GENERIC_TIMEZONE=America/Sao_Paulo
      - N8N_SECURE_COOKIE=false

    labels:
      - homepage.group=Automação
      - homepage.name=N8n
      - homepage.icon=n8n.png
      - homepage.href=http://IP_DO_SERVIDOR:5678/
      - homepage.description=Automacoes

volumes:
  n8n_data:
```

> Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor.

> `N8N_SECURE_COOKIE=false` é adequado para acesso HTTP local. Para acesso externo utilizando HTTPS, essa configuração deve ser revista.

## 3. Iniciar o n8n

```bash
docker compose up -d
```

## 4. Verificar o container

```bash
docker compose ps
```

Ou:

```bash
docker ps
```

Verifique se o container `n8n` está com status `Up`.

## 5. Acessar a interface

Abra no navegador:

```text
http://IP_DO_SERVIDOR:5678
```

Na primeira execução, será necessário configurar a conta do proprietário da instância.

## 6. Verificar os logs

```bash
docker logs n8n
```

Para acompanhar os logs em tempo real:

```bash
docker logs -f n8n
```

## 7. Reiniciar o n8n

```bash
docker restart n8n
```

Ou:

```bash
docker compose restart n8n
```

## 8. Parar o n8n

```bash
docker stop n8n
```

Ou:

```bash
docker compose stop n8n
```

## 9. Iniciar o n8n

```bash
docker start n8n
```

Ou:

```bash
docker compose start n8n
```

## 10. Remover o n8n

Para remover somente o container:

```bash
docker rm -f n8n
```

Para remover também a imagem:

```bash
docker rmi docker.n8n.io/n8nio/n8n:latest
```

> ⚠️ O volume `n8n_data` contém os dados do n8n, incluindo workflows, credenciais e configurações. Não remova o volume caso queira preservar esses dados.

## 11. Atualizar o n8n

Entre no diretório do Compose:

```bash
cd ~/media-server/n8n
```

Baixe a versão mais recente:

```bash
docker compose pull
```

Recrie o container:

```bash
docker compose up -d
```

Verifique:

```bash
docker compose ps
```

## 📁 Estrutura

O diretório ficará organizado desta forma:

```text
~/media-server/n8n/
└── docker-compose.yml
```

Os dados persistentes ficam no volume Docker:

```text
n8n_data
```

## 🔧 Comandos rápidos

```bash
cd ~/media-server/n8n

docker compose up -d
docker compose stop
docker compose start
docker compose restart
docker compose logs -f
```
