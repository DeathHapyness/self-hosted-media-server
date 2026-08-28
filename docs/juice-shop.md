# 🧃 OWASP Juice Shop

O **OWASP Juice Shop** é uma aplicação web intencionalmente vulnerável, desenvolvida para o estudo e a prática de segurança em aplicações web.

O projeto permite praticar vulnerabilidades como:

- SQL Injection
- Cross-Site Scripting (XSS)
- Quebra de autenticação
- Controle de acesso inadequado
- Exposição de informações
- Falhas de validação
- Vulnerabilidades de API
- Problemas relacionados ao OWASP Top 10

> ⚠️ O Juice Shop deve ser utilizado somente para estudos e testes em ambientes autorizados. Não exponha a aplicação diretamente à internet.

## 📁 Estrutura

```text
juice-shop/
├── docker-compose.yml
├── .gitignore
└── README.md
```

O Juice Shop pode ser executado diretamente a partir da imagem oficial disponível no Docker Hub.

## 📂 Criar diretório

Crie uma pasta para o projeto:

```bash
mkdir -p ~/security-lab/juice-shop
cd ~/security-lab/juice-shop
```

## 🐳 Docker Compose

Crie o arquivo:

```text
docker-compose.yml
```

Adicione o seguinte conteúdo:

```yaml
services:
  juice-shop:
    image: bkimminich/juice-shop:latest
    container_name: juice-shop
    restart: unless-stopped
    ports:
      - "127.0.0.1:3000:3000"
```

O endereço `127.0.0.1` faz com que o Juice Shop fique disponível somente no próprio computador.

Isso ajuda a impedir que a aplicação vulnerável seja acessada por outros dispositivos da rede ou pela internet.

## ▶️ Iniciar

Dentro da pasta do projeto, execute:

```bash
docker compose up -d
```

Verifique o container:

```bash
docker compose ps
```

Também é possível verificar com:

```bash
docker ps --filter name=juice-shop
```

## 🌐 Acesso

O Juice Shop utiliza a porta `3000`.

Abra no navegador:

```text
http://localhost:3000
```

Caso esteja executando o Docker em uma máquina virtual ou servidor de laboratório:

```text
http://IP_DO_SERVIDOR:3000
```

> Substitua `IP_DO_SERVIDOR` pelo endereço do seu servidor. Nunca coloque IPs pessoais ou públicos reais no GitHub.

## 🐳 Executar sem Docker Compose

Também é possível executar o Juice Shop diretamente com o Docker.

Baixe a imagem:

```bash
docker pull bkimminich/juice-shop:latest
```

Crie e inicie o container:

```bash
docker run -d \
  --name juice-shop \
  --restart unless-stopped \
  -p 127.0.0.1:3000:3000 \
  bkimminich/juice-shop:latest
```

Verifique se o container está funcionando:

```bash
docker ps
```

## 🔄 Funcionamento

A estrutura do laboratório funciona aproximadamente assim:

```text
┌───────────────┐
│    Usuário    │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│   Navegador   │
│ localhost:3000│
└───────┬───────┘
        │
        ▼
┌───────────────┐
│    Docker     │
│  Porta 3000   │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│  Juice Shop   │
│ Aplicação Web │
└───────────────┘
```

O navegador acessa a porta `3000` do computador, que é encaminhada para a porta `3000` do container.

## 🎯 Objetivo do laboratório

O Juice Shop pode ser utilizado para aprender e praticar:

```text
Reconhecimento da aplicação
        ↓
Análise das funcionalidades
        ↓
Identificação de vulnerabilidades
        ↓
Exploração controlada
        ↓
Documentação das descobertas
        ↓
Estudo de formas de correção
```

O próprio Juice Shop possui desafios que são concluídos conforme vulnerabilidades são encontradas.

## 🏆 Painel de desafios

O Juice Shop possui um painel chamado **Score Board**, onde é possível acompanhar os desafios concluídos.

O painel normalmente pode ser encontrado durante a exploração da aplicação.

Ele apresenta desafios relacionados a:

- Autenticação
- Autorização
- Injeção
- XSS
- Criptografia
- Upload de arquivos
- Exposição de dados
- APIs
- Configuração insegura

## 🐳 Gerenciamento

### Verificar o container

```bash
docker ps --filter name=juice-shop
```

### Ver todos os containers

```bash
docker ps -a
```

### Ver logs

```bash
docker logs juice-shop
```

### Acompanhar logs em tempo real

```bash
docker logs -f juice-shop
```

Para sair dos logs sem parar o container, pressione:

```text
Ctrl+C
```

### Reiniciar

```bash
docker restart juice-shop
```

Com Docker Compose:

```bash
docker compose restart
```

### Parar

```bash
docker compose stop
```

### Iniciar novamente

```bash
docker compose start
```

### Parar e remover o container

```bash
docker compose down
```

### Iniciar e recriar o container

```bash
docker compose up -d
```

## 🔄 Atualizar

Para baixar uma versão mais recente da imagem:

```bash
docker compose pull
```

Depois, recrie o container:

```bash
docker compose up -d
```

Também é possível executar os dois comandos juntos:

```bash
docker compose pull
docker compose up -d
```

Remova imagens antigas que não estão sendo utilizadas:

```bash
docker image prune
```

> Esse comando remove imagens Docker não utilizadas. Confira o que será removido antes de confirmar.

## 🧹 Reiniciar o laboratório

Para remover o container atual:

```bash
docker compose down
```

Depois, inicie uma nova instância:

```bash
docker compose up -d
```

Caso tenha executado sem Docker Compose:

```bash
docker rm -f juice-shop
```

Depois, crie novamente:

```bash
docker run -d \
  --name juice-shop \
  --restart unless-stopped \
  -p 127.0.0.1:3000:3000 \
  bkimminich/juice-shop:latest
```

## 🔒 .gitignore

Crie um arquivo chamado:

```text
.gitignore
```

Adicione:

```gitignore
.env
*.log
```

Não envie para o GitHub:

- Endereços IP pessoais
- Senhas
- Tokens
- Chaves de API
- Arquivos `.env`
- Logs contendo dados sensíveis

## ⚠️ Porta ocupada

Se a porta `3000` já estiver sendo utilizada, verifique os containers ativos:

```bash
docker ps
```

Você também pode verificar a porta no Linux:

```bash
sudo ss -lntp | grep :3000
```

Para utilizar outra porta, altere o arquivo `docker-compose.yml`:

```yaml
ports:
  - "127.0.0.1:3001:3000"
```

Depois, recrie o container:

```bash
docker compose up -d
```

Acesse:

```text
http://localhost:3001
```

## ❌ Nome do container já utilizado

Caso apareça uma mensagem informando que o nome `juice-shop` já está em uso, verifique o container:

```bash
docker ps -a --filter name=juice-shop
```

Remova o container antigo:

```bash
docker rm -f juice-shop
```

Depois, execute novamente:

```bash
docker compose up -d
```

## 🔍 Página não abre

Verifique se o container está em execução:

```bash
docker compose ps
```

Confira os logs:

```bash
docker compose logs --tail=100
```

Ou:

```bash
docker logs --tail 100 juice-shop
```

Verifique se a porta está respondendo:

```bash
curl http://localhost:3000
```

Se estiver utilizando outra porta, substitua `3000` pela porta configurada.

## 🔐 Segurança

O Juice Shop é intencionalmente vulnerável.

Por esse motivo:

- Não exponha a porta diretamente à internet.
- Não configure redirecionamento de porta no roteador.
- Não utilize senhas reais.
- Não coloque informações pessoais na aplicação.
- Não utilize o laboratório para atacar terceiros.
- Execute testes somente em ambientes autorizados.
- Mantenha o serviço ligado apenas durante os estudos.
- Prefira utilizar `127.0.0.1` na configuração da porta.
- Para acesso remoto, utilize uma VPN privada e regras de firewall.

## 🌐 Porta

O Juice Shop utiliza a porta `3000`.

```text
Host:      3000
Container: 3000
```

Mapeamento utilizado:

```yaml
ports:
  - "127.0.0.1:3000:3000"
```

## 📚 Referências

- [Repositório oficial do OWASP Juice Shop](https://github.com/juice-shop/juice-shop)
- [Documentação oficial do Juice Shop](https://pwning.owasp-juice.shop/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Documentação do Docker](https://docs.docker.com/)
- [Documentação do Docker Compose](https://docs.docker.com/compose/)

## ⚖️ Aviso legal

Este projeto deve ser utilizado exclusivamente para fins educacionais e em ambientes nos quais você tenha autorização explícita para realizar testes.

O autor não se responsabiliza pelo uso indevido das informações ou ferramentas apresentadas nesta documentação.