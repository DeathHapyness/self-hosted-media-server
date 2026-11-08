# 🐳 DockerBuddy

## O que é

O DockerBuddy é uma **colinha de comandos Docker no terminal**. Você digita
`dockerbuddy`, escolhe um número no menu e ele mostra os comandos mais usados
daquele assunto, cada um com uma explicação curta do que faz.

Projeto: [DeathHapyness/DockerBuddy](https://github.com/DeathHapyness/DockerBuddy)

```text
       DOCKER BUDDY

[1] Containers
[2] Images
[3] Volumes
[4] Networks
[5] Docker Compose
[6] Filtros (docker ps)

Selecione uma opção:
```

## Para que serve aqui

Depois de instalar o servidor, quase tudo que você vai fazer no dia a dia
passa por comandos Docker: ver os logs de um serviço que parou, reiniciar um
container, ver quanto espaço as imagens estão ocupando, limpar imagens velhas,
subir de novo um `docker-compose.yml` que você editou. O DockerBuddy deixa
esses comandos a um `dockerbuddy` de distância, sem precisar procurar na
internet.

Ele **só mostra** os comandos; não executa nada. Você copia o que precisar e
roda você mesmo.

## Não é um serviço

Diferente dos outros itens deste repositório, o DockerBuddy **não é container
nem serviço**: não abre porta, não fica rodando em segundo plano e não aparece
na Homepage. São só um script bash e alguns arquivos de texto.

## Ativado por padrão

O DockerBuddy não aparece na lista numerada de serviços. Logo depois dela (e
da pergunta do [Loger](loger.md)), o instalador explica o que ele é e
pergunta:

```text
Instalar o DockerBuddy? [S/n]:
```

- `Enter` ou `s`: instala o DockerBuddy depois de subir os serviços.
- `n`: não instala nesta execução.

No modo web, a mesma pergunta aparece com os botões **Sim** / **Não**.

## O que o instalador faz

1. Baixa o código do repositório do DockerBuddy (branch `main`, do GitHub) e
   extrai **só** o `dockerbuddy.sh`, os `data/*.txt`, o `README.md` e a
   `LICENSE` para `/opt/dockerbuddy`.
2. Cria o link `/usr/local/bin/dockerbuddy -> /opt/dockerbuddy/dockerbuddy.sh`,
   para o comando funcionar para qualquer usuário.

Rodar o instalador de novo **não sobrescreve** um `/opt/dockerbuddy` que já
exista (você pode ter editado os textos), e não mexe num
`/usr/local/bin/dockerbuddy` que não seja o link criado por ele.

Se o download falhar (sem internet, por exemplo), o instalador só avisa: os
serviços continuam instalados normalmente.

## Usar

```bash
dockerbuddy
```

Escolha um número e pronto.

> **Podman:** a colinha mostra os comandos com `docker`. Se você instalou com
> Podman, a maioria funciona trocando `docker` por `podman`.

## Editar os comandos

Os textos ficam em `/opt/dockerbuddy/data/*.txt`, um arquivo por opção do
menu. Para acrescentar um comando, edite o arquivo como root:

```bash
sudo nano /opt/dockerbuddy/data/containers.txt
```

O script colore sozinho:

- linhas como `==== TÍTULO ====` viram títulos;
- linhas como `-- seção --` viram seções;
- o que vem depois de `#` numa linha de comando vira explicação.

## Atualizar

```bash
sudo rm -rf /opt/dockerbuddy
```

Depois rode o instalador de novo. Ele baixa a versão atual. Edições que você
fez nos `.txt` se perdem; copie-as antes se quiser mantê-las.

## Desinstalar

Na opção "Desinstalar" do instalador, se o DockerBuddy estiver instalado, ele
pergunta `Também desinstalar o DockerBuddy? [s/N]`. Respondendo `s`, ele
remove o link `/usr/local/bin/dockerbuddy` e a pasta `/opt/dockerbuddy`.

Para remover à mão:

```bash
sudo rm /usr/local/bin/dockerbuddy
sudo rm -rf /opt/dockerbuddy
```
