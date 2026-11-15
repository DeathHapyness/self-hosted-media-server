# 🐳 Dockge

O Dockge é um painel web para **gerenciar stacks do Docker Compose**. Pelo
navegador você cria, edita, sobe, para e atualiza stacks (`compose.yaml`), além
de ver os logs e abrir um terminal dentro dos containers.

Projeto: [louislam/dockge](https://github.com/louislam/dockge)

## Para que serve aqui

Os serviços deste repositório continuam sendo instalados pelo instalador, mas
o Dockge é útil para:

* subir serviços novos que não fazem parte do repositório, sem precisar de SSH;
* ver rapidamente quais containers estão rodando, parados ou com erro;
* acompanhar logs e reiniciar containers pelo navegador;
* atualizar imagens (`pull` + `up`) de uma stack com um clique.

> O Dockge só gerencia como **stack** o que está na pasta de stacks dele
> (veja abaixo). Os outros containers do servidor aparecem na lista, mas não
> dá para editar o compose deles pelo Dockge.

## 📁 Estrutura

```text
dockge/
└── docker-compose.yaml
```

Os dados ficam fora do repositório, dentro do `DATA_DIR` (padrão
`/opt/media-server`):

```text
/opt/media-server/dockge/
├── data/      # banco e configurações do Dockge (usuário, senha, etc.)
└── stacks/    # uma pasta por stack criada pelo Dockge
```

## Docker Compose

```yaml
services:
  dockge:
    image: louislam/dockge:1
    container_name: dockge
    restart: unless-stopped

    ports:
      - "5557:5001"

    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ${DATA_DIR:-/opt/media-server}/dockge/data:/app/data
      - ${DATA_DIR:-/opt/media-server}/dockge/stacks:${DATA_DIR:-/opt/media-server}/dockge/stacks

    environment:
      - DOCKGE_STACKS_DIR=${DATA_DIR:-/opt/media-server}/dockge/stacks
```

### ⚠️ A pasta de stacks

O volume das stacks precisa ter o **mesmo caminho absoluto dos dois lados**
(host e container), e `DOCKGE_STACKS_DIR` precisa apontar para esse mesmo
caminho. O Dockge roda `docker compose` através do socket do host; se os
caminhos forem diferentes, os volumes relativos das stacks (`./data`, etc.)
são criados no lugar errado do host.

Se quiser usar outra pasta (por exemplo o padrão do Dockge, `/opt/stacks`),
troque **as três ocorrências** juntas:

```yaml
    volumes:
      - /opt/stacks:/opt/stacks
    environment:
      - DOCKGE_STACKS_DIR=/opt/stacks
```

### Registries privados

Para baixar imagens de registries privados, compartilhe o arquivo de
autenticação do Docker com o container:

```yaml
    volumes:
      - /root/.docker/:/root/.docker
```

## Instalação pelo instalador

O Dockge é descoberto automaticamente pelo instalador, como os outros
serviços (pasta `dockge/` com `docker-compose.yaml`). Ao rodar o instalador
ele:

* aparece na lista de **"Quais serviços instalar?"**;
* tem as pastas `dockge/data` e `dockge/stacks` criadas dentro do `DATA_DIR`,
  com as permissões dos outros serviços;
* tem a porta `5557` verificada no diagnóstico de portas;
* é validado (`docker compose config`) e iniciado com
  `docker compose -p dockge up -d`.

Para instalar só o Dockge, rode o instalador e escolha apenas o número dele na
lista.

## Instalação manual

```bash
cd dockge
sudo mkdir -p /opt/media-server/dockge/{data,stacks}
docker compose up -d
```

Verificar:

```bash
docker compose ps
```

## Acesso

O Dockge utiliza a porta `5557`.

```text
http://IP_DO_SERVIDOR:5557
```

Exemplo:

```text
http://192.168.15.118:5557
```

No primeiro acesso ele pede para **criar o usuário administrador**. Faça isso
logo depois de subir o container: até lá, qualquer pessoa na rede que abrir a
página pode criar esse usuário.

Ele também aparece na Homepage, no card **Dockge**.

## 🧱 Criando uma stack

1. Clique em **+ Compose**.
2. Dê um nome para a stack (vira o nome da pasta em `dockge/stacks/`).
3. Cole ou escreva o `compose.yaml`.
4. Clique em **Deploy**.

A stack fica salva em:

```text
/opt/media-server/dockge/stacks/<nome-da-stack>/compose.yaml
```

Para trazer para o Dockge uma stack que já existe, copie a pasta dela (com o
`compose.yaml`) para dentro de `dockge/stacks/` e clique em **Scan Stacks
Folder** no menu do usuário.

## 🐳 Gerenciamento

Ver logs:

```bash
docker logs -f dockge
```

Reiniciar:

```bash
docker restart dockge
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

O Dockge tem acesso ao `/var/run/docker.sock`, ou seja, **controle total do
Docker e, na prática, do servidor**. Quem tiver o login do Dockge consegue
subir qualquer container com qualquer permissão.

* Não exponha a porta `5557` para a internet; acesse pela rede local ou pelo
  [Tailscale](tailscale.md).
* Use uma senha forte no usuário administrador.

## ⚠️ Observações

* As stacks criadas pelo Dockge ficam dentro do `DATA_DIR`. Se você remover os
  dados pelo instalador, **as stacks também são apagadas**. Faça backup de
  `dockge/stacks/` antes (o [Kopia](kopia.md) serve para isso).
* Esqueceu a senha? Com o container rodando, use o reset de senha do Dockge:

  ```bash
  docker exec -it dockge npm run reset-password
  ```

## Referências

* [Repositório oficial do Dockge](https://github.com/louislam/dockge)
