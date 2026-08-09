# O que o install.py faz — e o que ele deliberadamente NÃO faz

## ✅ O que ele faz

1. Verifica se está rodando como root (necessário para `/opt` e permissões).
2. Verifica se o SO é Linux.
3. Verifica se o Docker está instalado (`docker --version`).
4. Verifica se o plugin Docker Compose está instalado (`docker compose version`).
5. Verifica se o daemon do Docker está ativo (`docker info`).
6. Verifica se `/mnt/media` existe **e** está de fato montado (`os.path.ismount`).
7. Verifica espaço livre em `/opt` e em `/mnt/media`.
8. Verifica se as portas usadas pelos serviços estão livres, tentando
   primeiro extrair as portas do `docker-compose.yml` do repositório e
   usando uma lista padrão como fallback.
9. Cria os diretórios de cada serviço em `/opt/media-server/`, só criando o
   que ainda não existe.
10. Cria as subpastas de `/mnt/media` (`filmes`, `series`, `musicas`,
    `fotos`, `downloads`, `inbox`) — somente depois de confirmar que o
    mount está ativo.
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

## ❌ O que ele deliberadamente NÃO faz

- **Não cria ou edita o `docker-compose.yml`** — apenas valida o que já
  existe no repositório.
- **Não apaga nem sobrescreve** diretórios, configurações ou bancos de
  dados existentes. Só cria o que está faltando.
- **Não altera permissões em todo `/mnt/media`** — só toca em `/mnt/media`
  (raiz, para o File Browser) e `/mnt/media/downloads` (para o
  qBittorrent); as demais subpastas mantêm as permissões atuais.
- **Não grava nada em `/mnt/media` se o disco não estiver montado** — a
  instalação é interrompida antes de qualquer criação de diretório.
- **Não armazena senhas, tokens, IPs pessoais ou chaves de API no código**
  — qualquer credencial fica no `.env`, fora do controle de versão.
- **Não força a parada de processos** que estejam usando as portas
  necessárias — apenas informa quais portas estão ocupadas e interrompe a
  instalação para você decidir o que fazer.
- **Não reinicia ou reinstala o Docker** — apenas verifica se ele está
  presente e funcionando; se não estiver, informa como resolver.
- **Não faz downgrade/upgrade de versões de containers** — quem controla
  isso é o `docker-compose.yml` do repositório.
- **Não continua a execução após um erro crítico** — qualquer falha em uma
  verificação interrompe o script imediatamente, sem tentar "seguir em
  frente" ou aplicar workarounds automáticos.