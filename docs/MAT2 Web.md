# 🕵️ Instalação do MAT2 Web

O MAT2 Web é uma interface web para remover metadados de arquivos, ajudando a eliminar informações potencialmente identificáveis presentes em imagens, documentos, áudios e outros arquivos compatíveis.

Doc oficial:https://0xacab.org/jvoisin/mat2-web#container

Nesta configuração, o MAT2 Web é executado em um container Docker isolado.

## 1. Criar o diretório

```bash
mkdir -p ~/media-server/mat2-web
cd ~/media-server/mat2-web
```

## 2. Criar o Dockerfile

O MAT2 Web possui um limite padrão de upload de 16 MB. Nesta configuração, o limite da aplicação será aumentado para 150 MB.

```bash
nano Dockerfile
```

Utilize:

```dockerfile
FROM registry.0xacab.org/jvoisin/mat2-web:latest

RUN sed -i "s/16 \* 1024 \* 1024/150 * 1024 * 1024/" /var/www/mat2-web/main.py
```

## 3. Criar a configuração do Nginx

Crie o diretório:

```bash
mkdir -p nginx
```

Crie o arquivo:

```bash
nano nginx/default
```

Utilize:

```nginx
server {
    server_name _;
    listen 8080 default_server;
    listen [::]:8080 default_server;
    client_max_body_size 150M;

    root /var/www/mat2-web;

    location / {
        try_files $uri @yourapplication;
    }

    location @yourapplication {
        include uwsgi_params;
        uwsgi_pass unix:/run/uwsgi/uwsgi.sock;
    }
}
```

## 4. Criar a configuração adicional de upload

```bash
nano nginx/upload.conf
```

Utilize:

```nginx
client_max_body_size 150M;
```

> O limite de 150 MB precisa estar configurado tanto na aplicação quanto no Nginx.

## 5. Criar o `docker-compose.yml`

```bash
nano docker-compose.yml
```

Utilize:

```yaml
services:
  mat2-web:
    build:
      context: .
      dockerfile: Dockerfile

    container_name: mat2-web

    restart: unless-stopped

    ports:
      - "8282:8080"

    read_only: true

    tmpfs:
      - /tmp
      - /run/uwsgi
      - /app/upload

    security_opt:
      - no-new-privileges:true

    volumes:
      - ./nginx/upload.conf:/etc/nginx/conf.d/upload.conf:ro
      - ./nginx/default:/etc/nginx/sites-enabled/default:ro

    labels:
      - homepage.group=Privacidade
      - homepage.name=MAT2 Web
      - homepage.icon=mat2.png
      - homepage.href=http://IP_DO_SERVIDOR:8282/
      - homepage.description=Remocao de metadados de arquivos
```

> Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor.

> A porta `8282` é a porta utilizada no host. A aplicação dentro do container continua utilizando a porta `8080`.

## 6. Construir a imagem

Como existe um `Dockerfile` personalizado, construa a imagem:

```bash
docker compose build --no-cache
```

## 7. Iniciar o MAT2 Web

```bash
docker compose up -d
```

## 8. Verificar o container

```bash
docker compose ps
```

Ou:

```bash
docker ps
```

O container `mat2-web` deve aparecer com status `Up`.

## 9. Acessar a interface

Abra no navegador:

```text
http://IP_DO_SERVIDOR:8282
```

Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor.

Após acessar a interface, envie um arquivo para que o MAT2 Web possa processá-lo e remover seus metadados.

## 10. Verificar os logs

```bash
docker logs mat2-web
```

Para acompanhar os logs em tempo real:

```bash
docker logs -f mat2-web
```

## 11. Verificar o limite de upload

Verifique o limite da aplicação:

```bash
docker exec mat2-web grep MAX_CONTENT_LENGTH /var/www/mat2-web/main.py
```

O resultado esperado é:

```text
app.config['MAX_CONTENT_LENGTH'] = 150 * 1024 * 1024
```

Verifique também o limite do Nginx:

```bash
docker exec mat2-web nginx -T 2>&1 | grep -A2 -B2 client_max_body_size
```

O resultado deve indicar:

```text
client_max_body_size 150M;
```

## 12. Reiniciar o MAT2 Web

```bash
docker restart mat2-web
```

Ou:

```bash
docker compose restart
```

## 13. Parar o MAT2 Web

```bash
docker stop mat2-web
```

Ou:

```bash
docker compose stop
```

## 14. Iniciar o MAT2 Web

```bash
docker start mat2-web
```

Ou:

```bash
docker compose start
```

## 15. Remover o MAT2 Web

Para remover o container:

```bash
docker compose down
```

Para remover também a imagem construída localmente:

```bash
docker compose down --rmi local
```

> ⚠️ O MAT2 Web não deve ser tratado como sistema de armazenamento. Ele é utilizado para processar e sanitizar arquivos enviados através da interface.

## 16. Atualizar o MAT2 Web

Para atualizar a imagem base e reconstruir a imagem personalizada:

```bash
cd ~/media-server/mat2-web
```

```bash
docker compose down
```

```bash
docker compose build --no-cache
```

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
~/media-server/mat2-web/
├── docker-compose.yml
├── Dockerfile
└── nginx/
    ├── default
    └── upload.conf
```

## 🔧 Comandos rápidos

```bash
cd ~/media-server/mat2-web

docker compose up -d
docker compose stop
docker compose start
docker compose restart
docker compose logs -f
```

Para reconstruir:

```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```
