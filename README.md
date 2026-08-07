# Self-Hosted Media Server

Infraestrutura de servidor de mídia self-hosted utilizando Docker,
Jellyfin e qBittorrent.

## 🚀 Sobre o projeto

Projeto criado para estudar e demonstrar conceitos de Linux,
Docker, Docker Compose, gerenciamento de containers, volumes
persistentes e serviços self-hosted.

O servidor utiliza o Jellyfin para gerenciamento e reprodução
da biblioteca de mídia e o qBittorrent para gerenciamento de
downloads.

## 🛠️ Tecnologias

- Linux
- Docker
- Docker Compose
- Jellyfin
- qBittorrent
- Git / GitHub

## 🏗️ Arquitetura

```text
                 ┌─────────────────┐
                 │   qBittorrent   │
                 │                 │
                 │    Downloads    │
                 └────────┬────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │  Media Storage  │
                 │                 │
                 │  ├── filmes     │
                 │  └── series     │
                 └────────┬────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │     Jellyfin    │
                 │                 │
                 │ Media Server    │
                 └────────┬────────┘
                          │
             ┌────────────┼────────────┐
             ▼            ▼            ▼
           Desktop       TV          Mobile
📁 Estrutura
media-server/
├── jellyfin/
│   └── docker-compose.yml
├── qbittorrent/
│   └── docker-compose.yml
├── .gitignore
└── README.md

As configurações internas dos containers e os arquivos de mídia
não são versionados no Git.

🐳 Containers
Jellyfin

Responsável pelo gerenciamento e reprodução da biblioteca de mídia.

Porta:

8096

Acesso:

http://SERVER_IP:8096
qBittorrent

Cliente de download com interface Web.

Porta:

8080

Acesso:

http://SERVER_IP:8080
💾 Volumes

O Jellyfin utiliza volumes persistentes para separar configuração,
cache e mídia.

/config
/cache
/media

A biblioteca de mídia é organizada em:

/media
├── filmes
└── series
🔐 Segurança

Arquivos de configuração, credenciais, tokens e biblioteca de
mídia não são armazenados neste repositório.

O projeto utiliza .gitignore para evitar o versionamento
acidental desses arquivos.

📚 Objetivos de aprendizado
Administração de servidores Linux
Docker e containers
Docker Compose
Persistência de dados
Gerenciamento de volumes
Configuração de serviços self-hosted
Redes e portas
Git e GitHub
Documentação de infraestrutura
🔮 Próximos passos
 Adicionar Radarr
 Adicionar Sonarr
 Automatizar organização da biblioteca
 Implementar monitoramento
 Implementar backup
 Melhorar documentação da infraestrutura
👨‍💻 Autor

Projeto desenvolvido como parte do meu laboratório pessoal de
Linux, Docker e infraestrutura.
