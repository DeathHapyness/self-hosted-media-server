# 📁 Instalação do Filebrowser

O Filebrowser é uma interface web leve para navegar, visualizar e gerenciar arquivos armazenados no servidor.

## 1. Criar o container

```bash
docker run -d \
  --name filebrowser \
  -v /caminho/do/seu/hd:/srv \
  -v filebrowser_db:/database.db \
  -v filebrowser_config:/config \
  -p 8080:80 \
  --restart unless-stopped \
  filebrowser/filebrowser:latest
```

> Substitua `/caminho/do/seu/hd` pelo diretório real do servidor que você quer navegar/visualizar.

## 2. Verificar o container

```bash
docker ps
```

Verifique se o container `filebrowser` está em execução.

## 3. Acessar a interface

Abra no navegador:

```text
http://IP_DO_SERVIDOR:8080
```

Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor.

**Login padrão:**

```text
usuário: admin
senha: admin
```

> ⚠️ Altere a senha padrão no primeiro acesso, em **Settings → Profile**.

## 4. Verificar os logs do Filebrowser

```bash
docker logs filebrowser
```

Para acompanhar os logs em tempo real:

```bash
docker logs -f filebrowser
```

## 5. Reiniciar o Filebrowser

```bash
docker restart filebrowser
```

## 6. Parar o Filebrowser

```bash
docker stop filebrowser
```

## 7. Iniciar o Filebrowser

```bash
docker start filebrowser
```

## 8. Remover o Filebrowser

Para remover o container:

```bash
docker rm -f filebrowser
```

Para remover também a imagem:

```bash
docker rmi filebrowser/filebrowser:latest
```

---

## 🔍 Navegando pelos arquivos

Após acessar a interface web, o Filebrowser exibirá a estrutura de pastas do diretório montado.

Exemplo:

```text
/Documentos
/Fotos
/Downloads
/Backups
```

O Filebrowser é especialmente útil para visualizar, mover, renomear e fazer upload/download de arquivos direto pelo navegador, sem precisar de acesso SSH ou compartilhamento de rede para tarefas simples.