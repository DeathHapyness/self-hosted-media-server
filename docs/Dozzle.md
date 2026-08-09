# 🐳 Instalação do Dozzle

O Dozzle é uma interface web leve para visualizar os logs dos containers Docker em tempo real.

## 1. Criar o container

```bash
docker run -d \
  --name dozzle \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -p 9999:8080 \
  --restart unless-stopped \
  amir20/dozzle:latest
```

## 2. Verificar o container

```bash
docker ps
```

Verifique se o container `dozzle` está em execução.

## 3. Acessar a interface

Abra no navegador:

```text
http://IP_DO_SERVIDOR:9999
```

Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor.

## 4. Verificar os logs do Dozzle

```bash
docker logs dozzle
```

Para acompanhar os logs em tempo real:

```bash
docker logs -f dozzle
```

## 5. Reiniciar o Dozzle

```bash
docker restart dozzle
```

## 6. Parar o Dozzle

```bash
docker stop dozzle
```

## 7. Iniciar o Dozzle

```bash
docker start dozzle
```

## 8. Remover o Dozzle

Para remover o container:

```bash
docker rm -f dozzle
```

Para remover também a imagem:

```bash
docker rmi amir20/dozzle:latest
```

---

## 🔍 Acessar logs dos containers

Após acessar a interface web, o Dozzle exibirá os containers Docker disponíveis.

Selecione um container para visualizar seus logs em tempo real.

Exemplo:

```text
Navidrome
Jellyfin
qBittorrent
Dozzle
```

O Dozzle é especialmente útil para identificar erros, acompanhar inicializações e monitorar o funcionamento dos serviços sem precisar executar `docker logs` manualmente.
