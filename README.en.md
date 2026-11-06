<h1 align="center">Self-Hosted Media Server</h1>

<div align="center">

[🇧🇷 Português](README.md) | 🇺🇸 **English**

  <picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/banner-verde-agua.png">
  <source media="(prefers-color-scheme: light)" srcset="assets/banner-claro.png">
  <img src="assets/banner-verde-agua.png" alt="Self-Hosted Media Server" width="900">
</picture>

![Linux](https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black)
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Podman](https://img.shields.io/badge/Podman-892CA0?style=for-the-badge&logo=podman&logoColor=white)

<a href="https://buymeacoffee.com/rique.dev"><img src="https://img.shields.io/badge/Buy_Me_a_Coffee-FFDD00?style=for-the-badge&logo=buymeacoffee&logoColor=black" alt="Buy Me a Coffee"></a>
</div>

<div align="center">

A **self-hosted** media server running on Linux with Docker and Docker Compose.

The project brings together services for managing, organizing, monitoring and playing:

* Movies
* TV shows
* Music
* Downloads
* Indexers
* Movie and TV show requests
* DNS / Ad blocking
* Remote access
* Monitoring and metrics
* Dashboards
* Container logs
* File management
* Disk monitoring
* Service dashboard
* Image and video metadata removal
* Workflow automation
* Network traffic monitoring
* Internet speed history
* Photo and video backup and gallery
* Backup with deduplication and encryption
* Git server
* RSS feed aggregator
* Web security lab
* Server auditing (logins, commands, sudo)

The goal is to build a personal media infrastructure using **containers, persistent volumes, file organization, automation, monitoring, remote access and independent services**.

</div>

> **Note:** the detailed documentation in `docs/` is written in Portuguese. This README covers the overview, installation and security in English.

---

## INSTALLER

The installer is a single executable. Download the latest version from
[Releases](https://github.com/DeathHapyness/self-hosted-media-server/releases),
place it in the root of the cloned repository and run it with `sudo`:

```bash
sudo ./instalador-linux-amd64
```

Before using it, read:

[`docs/readme-instalador.md`](docs/readme-instalador.md) (Portuguese)

To understand exactly what the installer does, and what it deliberately does not do:

[`docs/installer-actions.md`](docs/installer-actions.md) (Portuguese)

The installer can also be used from a browser (web mode), which is handy on
headless servers. **Web mode does not use HTTPS**: only use it on your home network.
Details in [`docs/modo-web.md`](docs/modo-web.md) (Portuguese).

---


## Services

| Service      | Purpose                               |        Port | Documentation                                              |
| ------------ | ------------------------------------- | ----------: | ---------------------------------------------------------- |
| Jellyfin     | Movies and TV shows                   |      `8096` | [docs/Jellyfin.md](docs/Jellyfin.md)                       |
| Navidrome    | Music server                          |      `4533` | [docs/navidrome.md](docs/navidrome.md)                     |
| qBittorrent  | Download management                   |      `8080` | [docs/qbittorrent.md](docs/qbittorrent.md)                 |
| Radarr       | Automatic movie management            |      `7878` | [docs/raddar.md](docs/raddar.md)                           |
| Sonarr       | Automatic TV show management          |      `8989` | [docs/sonarr.md](docs/sonarr.md)                           |
| Prowlarr     | Indexer management                    |      `9696` | [docs/prowlarr.md](docs/prowlarr.md)                       |
| Jellyseerr   | Catalog and requests                  |      `5055` | [docs/jellyseerr.md](docs/jellyseerr.md)                   |
| Prometheus   | Metrics collection and storage        |      `9090` | [docs/monitoring.md](docs/monitoring.md)                   |
| Grafana      | Metrics dashboards and visualization  |      `3000` | [docs/monitoring.md](docs/monitoring.md)                   |
| Node Exporter | Host metrics for Prometheus          |      `9100` | [docs/monitoring.md](docs/monitoring.md)                   |
| cAdvisor     | Container metrics                     |      `8085` | [docs/monitoring.md](docs/monitoring.md)                   |
| Scrutiny     | Disk S.M.A.R.T. monitoring            |      `8090` | [docs/scrutiny.md](docs/scrutiny.md)                       |
| AdGuard Home | DNS and ad blocking                   | `53 / 8181` | [docs/adguard.md](docs/adguard.md)                         |
| Dozzle       | Container log viewer                  |      `9999` | [docs/Dozzle.md](docs/Dozzle.md)                           |
| Filebrowser  | File management                       |      `8082` | [docs/filebrowser.md](docs/filebrowser.md)                 |
| Homepage     | Service dashboard                     |      `3001` | [docs/homepage.md](docs/homepage.md)                       |
| Tailscale    | VPN and private remote access         |           — | [docs/tailscale.md](docs/tailscale.md)                     |
| Gluetun      | VPN for qBittorrent                   |           — | [docs/gluetun-qbittorrent.md](docs/gluetun-qbittorrent.md) |
| spotDL       | Manual music downloads                |         CLI | [docs/navidrome.md](docs/navidrome.md)                     |
| Mat2 Web     | Image and video metadata removal      |      `8282` | [docs/MAT2 Web.md](docs/MAT2%20Web.md)                     |
| n8n          | Workflow automation                   |      `5678` | [docs/n8n.md](docs/n8n.md)                                 |
| Juice Shop   | Web security lab (OWASP)              |      `3050` | [docs/juice-shop.md](docs/juice-shop.md)                   |
| ntopng       | Network traffic monitoring            |      `3100` | [docs/ntopng.md](docs/ntopng.md)                           |
| Speedtest Tracker | Speed test history               |      `8765` | [docs/speedtest-tracker.md](docs/speedtest-tracker.md)     |
| LibreSpeed   | Speed test server                     |      `8111` | [docs/librespeed.md](docs/librespeed.md)                   |
| Kopia        | Backup with deduplication and encryption |  `51515` | [docs/kopia.md](docs/kopia.md)                             |
| Forgejo      | Self-hosted Git server                | `3044 / 222` | [docs/forgejo.md](docs/forgejo.md)                        |
| Glances      | Real-time system monitoring           |     `61208` | [docs/glances.md](docs/glances.md)                         |
| Immich       | Photo and video backup and gallery    |      `2283` | [docs/immich.md](docs/immich.md)                           |
| Links Hub    | Minimal page with your favorite sites |      `5000` | [docs/links-hub.md](docs/links-hub.md)                     |
| FreshRSS     | RSS feed aggregator                   |      `8083` | [docs/freshrss.md](docs/freshrss.md)                       |
| Loger        | Server auditing (native, enabled by default) |   — | [docs/loger.md](docs/loger.md)                             |

> **Note:** the ports above are the external ports defined in this repository's `docker-compose.yml` files. None of them is shared between services.

> Several services use the same **internal** port (for example `3000` or `8080`), so each one is mapped to a different external port. When adding a new service, check that the external port you pick is not already in use.

> If there is a conflict, change the external port in `docker-compose.yml`. The container's internal port can stay the same.


---

## Media flow

### Movies

```text
Jellyseerr
    │
    ▼
  Radarr
    │
    ▼
 Prowlarr
    │
    ▼
qBittorrent
    │
    ▼
/mnt/media/downloads
    │
    ▼
  Radarr
    │
    ▼
/mnt/media/filmes
    │
    ▼
 Jellyfin
```

## Requirements

* x86_64 CPU
* 4 GB of RAM or more
* SSD for the system and containers
* HDD/SSD for media storage
* Local network connection
* Linux
* Docker
* Docker Compose

Full Docker installation guide:

[docs/docker.md](docs/docker.md) (Portuguese)

---

## Documentation

All documents below are in Portuguese.

### Infrastructure

* [Docker installation](docs/docker.md)
* [Tailscale](docs/tailscale.md)
* [Troubleshooting](docs/troubleshooting.md)

### Media

* [Jellyfin](docs/Jellyfin.md)
* [Radarr](docs/raddar.md)
* [Sonarr](docs/sonarr.md)
* [Prowlarr](docs/prowlarr.md)
* [Jellyseerr](docs/jellyseerr.md)
* [Navidrome + spotDL](docs/navidrome.md)
* [qBittorrent](docs/qbittorrent.md)

### Monitoring

* [Monitoring — Grafana + Prometheus](docs/monitoring.md)
* [Scrutiny](docs/scrutiny.md)
* [ntopng](docs/ntopng.md)
* [Speedtest Tracker](docs/speedtest-tracker.md)
* [LibreSpeed](docs/librespeed.md)
* [Glances](docs/glances.md)
* [Loger — server auditing](docs/loger.md)

### Services

* [Homepage](docs/homepage.md)
* [Homepage — config/ folder reference](docs/homepage-config.md)
* [AdGuard Home](docs/adguard.md)
* [Dozzle](docs/Dozzle.md)
* [Filebrowser](docs/filebrowser.md)
* [Gluetun + qBittorrent](docs/gluetun-qbittorrent.md)
* [Forgejo](docs/forgejo.md)
* [Immich](docs/immich.md)
* [FreshRSS](docs/freshrss.md)
* [Links Hub](docs/links-hub.md)

### Tools

* [Mat2 Web](docs/MAT2%20Web.md)
* [n8n](docs/n8n.md)
* [Juice Shop](docs/juice-shop.md)
* [Kopia](docs/kopia.md)

---

## Security

Avoid exposing administrative services directly to the internet.

Services such as:

```text
qBittorrent       → 8080
AdGuard Home      → 8181
Dozzle            → 9999
Filebrowser       → 8082
Homepage          → 3001
Radarr            → 7878
Sonarr            → 8989
Prowlarr          → 9696
Jellyseerr        → 5055
Prometheus        → 9090
Grafana           → 3000
Node Exporter     → 9100
cAdvisor          → 8085
Scrutiny          → 8090
Mat2 Web          → 8282
n8n               → 5678
Juice Shop        → 3050
ntopng            → 3100
Speedtest Tracker → 8765
LibreSpeed        → 8111
Kopia             → 51515
Forgejo           → 3044 / 222
Glances           → 61208
Immich            → 2283
FreshRSS          → 8083
Links Hub         → 5000
```

should preferably be accessed through:

* Tailscale
* VPN
* Reverse proxy with HTTPS
* Firewall
* Local network

Dozzle uses the Docker socket:

```text
/var/run/docker.sock
```

So access to Dozzle should be treated with the same care as administrative access to Docker itself.

Filebrowser has access to the mounted directories and can modify files. Protect access to it accordingly.

Radarr, Sonarr, Prowlarr and Jellyseerr also have access to APIs and/or media libraries and should be protected accordingly.

Never commit:

```text
.env
*.env
*.key
*.pem
*.token
config/
data/
cache/
work/
*.db
*.sqlite
*.sqlite3
*.log
```

Also never commit:

* Passwords
* Tokens
* API keys
* Credentials
* Databases
* Private configuration
* Personal media

---


## Disclaimer

This project is provided **"as is"**, solely for educational and personal use.

The maintainers of this project **are not responsible for any damage, data loss, hardware failure, software problems, service downtime or any other harm** that may occur while installing, configuring or using this project.

Before running any script, command or system change, it is recommended to:

* Back up your important data.
* Review commands before running them.
* Double-check storage and volume settings.
* Use the project in a test environment when possible.
* Keep the system and services up to date.

Use of this project is **entirely at the user's own risk**.

By using this project, the user acknowledges having enough knowledge to administer their own system and accepts the risks of the changes made to the machine.


---

## ☕ Support the project

This project is free and open source. If it helped you set up your server or saved you a few hours of configuration, consider buying me a coffee. Your support helps keep the project active and lets me spend more time on new services and improvements.

<a href="https://buymeacoffee.com/rique.dev"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" width="200"></a>

---

## License

This project is made available for educational and personal use.

See the [`LICENSE`](LICENSE) file for more information.

---

## Note on AI-assisted contributions

Claude appears in this repository's contributor list because some commits were recorded with its participation during development. This only reflects the history and technical authorship of those commits and does not mean Claude is a maintainer, owner or human collaborator of the project.

All decisions, reviews and responsibilities related to the project remain with the repository maintainer.
