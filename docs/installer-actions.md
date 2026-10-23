# O que o instalador faz — e o que ele deliberadamente NÃO faz

## O que ele faz

1. Verifica se está rodando como root (necessário para `/opt` e permissões).
2. Verifica se o SO é Linux.
3. Verifica se o Docker ou o Podman está instalado (`docker --version`).
4. Verifica se o Compose está instalado (`docker compose version`, ou
   `podman-compose` no caso do Podman).
5. Verifica se o daemon do Docker está ativo (`docker info`).
6. Verifica se `/mnt/media` existe **e** está de fato montado como um
   filesystem separado. Se não estiver, pergunta se você quer continuar.
7. Verifica espaço livre na pasta de dados (`/opt/media-server`) e em `/mnt/media`.
8. Verifica se as portas usadas pelos serviços estão livres, tentando
   primeiro extrair as portas do `docker-compose.yml` do repositório e
   usando uma lista padrão como fallback.
9. Cria os diretórios de cada serviço em `/opt/media-server/`, só criando o
   que ainda não existe.
10. Cria as subpastas de `/mnt/media` (`filmes`, `series`, `musicas`,
    `fotos`, `downloads`, `inbox`).
11. Ajusta permissões (modo `770`) apenas nos diretórios que os containers
    precisam gravar, mais `775` em `/mnt/media` (para o File Browser) e em
    `/mnt/media/downloads` (para o qBittorrent).
12. Lê `PUID`/`PGID` de um `.env`, se existir, e tenta aplicar como dono das
    pastas gravável — sem exigir isso para continuar.
13. Cria `.env` a partir de `.env.example`, se o `.env` ainda não existir.
14. Valida o `docker-compose.yml` com `docker compose config` antes de
    subir qualquer coisa.
15. Sobe os serviços com `docker compose up -d`.
16. Imprime um resumo claro de cada etapa, com `[✓]`/`[✗]`/`[!]`, e retorna
    código de saída `0` (sucesso), `1` (erro crítico) ou `130` (cancelado
    pelo usuário com Ctrl+C).
17. Se a validação ou a subida de um serviço falhar, analisa o erro e, para
    causas conhecidas (como `.env` do serviço ausente), explica o problema e
    **pergunta** se pode corrigir. Só corrige com sua confirmação e, depois,
    tenta de novo. Detalhes em [`readme-instalador.md`](readme-instalador.md).
18. Opcionalmente, roda em modo web: abre um servidor HTTP local, mostra um
    link de uso único (expira em 15 minutos) e encerra o servidor ao sair.
    Detalhes em [`modo-web.md`](modo-web.md).
19. Grava um histórico em JSON dentro de [`installer/json`](../installer/json),
    para auditoria do instalador, e um log em `/var/log/media-server-installer.log`.

## O que ele deliberadamente NÃO faz

- **Não edita o `docker-compose.yml` por conta própria** — no modo terminal,
  apenas valida o que já existe no repositório. No modo web, o arquivo só é
  alterado quando você mesmo o edita e clica em Salvar; antes de gravar, o
  conteúdo é validado e o original é copiado para `docker-compose.yml.bak`.
- **Não apaga nem sobrescreve** diretórios, configurações ou bancos de
  dados existentes durante a instalação. Só cria o que está faltando.
  A opção de desinstalar só apaga pastas de dados depois de uma confirmação
  explícita.
- **Não altera permissões em todo `/mnt/media`** — só toca em `/mnt/media`
  (raiz, para o File Browser) e `/mnt/media/downloads` (para o
  qBittorrent); as demais subpastas mantêm as permissões atuais.
- **Não usa `/mnt/media` sem montagem sem perguntar antes** — se o disco
  não estiver montado, a instalação só continua se você confirmar.
- **Não armazena senhas, tokens, IPs pessoais ou chaves de API no código**
  — qualquer credencial fica no `.env`, fora do controle de versão.
- **Não força a parada de processos** que estejam usando as portas
  necessárias — apenas informa quais portas estão ocupadas e interrompe a
  instalação para você decidir o que fazer.
- **Não instala, reinicia ou reinstala o Docker** — apenas verifica se ele está
  presente e funcionando; se não estiver, informa como resolver.
- **Não faz downgrade/upgrade de versões de containers** — quem controla
  isso é o `docker-compose.yml` do repositório.
- **Não aplica correções sem perguntar** — quando reconhece a causa de uma
  falha do Compose, o instalador explica o problema e oferece uma correção,
  mas só a aplica se você responder `s`. Se a correção for recusada ou não
  resolver, a instalação é interrompida.
- **Não continua a execução após um erro crítico** — qualquer falha que não
  foi corrigida interrompe o script, sem tentar "seguir em frente".
- **Não usa HTTPS no modo web** — a interface web é HTTP puro, sem
  criptografia. Use-a apenas em uma rede confiável e não exponha a porta para
  a internet. Veja [`modo-web.md`](modo-web.md).