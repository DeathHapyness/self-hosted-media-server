from pathlib import Path

BASE_DIR = Path("/opt/media-server")
MEDIA_DIR = Path("/mnt/media")

MIN_DISK_SPACE_OPT_GB = 2
MIN_DISK_SPACE_MEDIA_GB = 5

SERVICE_DIRS = {
    "AdGuard Home": [
        BASE_DIR / "adguard" / "work",
        BASE_DIR / "adguard" / "conf",
    ],
    "Dozzle": [
        BASE_DIR / "dozzle",
    ],
    "File Browser": [
        BASE_DIR / "filebrowser" / "config",
        BASE_DIR / "filebrowser" / "database",
    ],
    "Forgejo": [
        BASE_DIR / "forgejo" / "data",
    ],
    "glances": [
        BASE_DIR / "glances" / "config",
    ],
    "Jellyfin": [
        BASE_DIR / "jellyfin" / "config",
        BASE_DIR / "jellyfin" / "cache",
    ],
    "Jellyseerr": [
        BASE_DIR / "jellyseerr" / "config",
    ],
    "Navidrome": [
        BASE_DIR / "navidrome" / "data",
    ],
    "qBittorrent": [
        BASE_DIR / "qbittorrent" / "config",
    ],
    "Homepage": [
        BASE_DIR / "homepage" / "config",
    ],
    "immich-app": [
        BASE_DIR / "immich-app" / "config",
    ],
    "Prowlarr": [
        BASE_DIR / "prowlarr" / "config",
    ],
    "Radarr": [
        BASE_DIR / "radarr" / "config",
    ],
    "Sonarr": [
        BASE_DIR / "sonarr" / "config",
    ],
    "Scrutiny": [
        BASE_DIR / "scrutiny" / "config",
    ],
    "Monitoring": [
        BASE_DIR / "monitoring" / "prometheus-data",
        BASE_DIR / "monitoring" / "grafana-data",
    ],
    "ntopng": [
        BASE_DIR / "ntopng" / "redis-data",
        BASE_DIR / "ntopng" / "ntopng-data",
    ],
    "Speedtest Tracker": [
        BASE_DIR / "speedtest-tracker" / "config",
    ],
}

MEDIA_SUBDIRS = ["filmes", "series", "musicas", "fotos", "downloads", "inbox"]

WRITABLE_DIRS = [d for dirs in SERVICE_DIRS.values() for d in dirs]
WRITABLE_MODE = 0o770

DEFAULT_EXPECTED_PORTS = {
    53: "AdGuard Home (DNS)",
    2283: "immich-app",
    3000: "AdGuard Home (setup UI) / Grafana",
    3001: "Homepage",
    3044: "forgejo",
    3050: "Juice Shop",
    3100: "ntopng",
    4533: "Navidrome",
    5055: "Jellyseerr",
    5678: "n8n",
    7878: "Radarr",
    8080: "qBittorrent",
    8081: "Dozzle",
    8082: "File Browser",
    8085: "cAdvisor",
    8090: "Scrutiny",
    8096: "Jellyfin",
    8181: "AdGuard Home (UI)",
    8282: "MAT2 Web",
    8765: "Speedtest Tracker",
    8989: "Sonarr",
    9090: "Prometheus",
    9100: "Node Exporter",
    9696: "Prowlarr",
    51515: "kopia",
    61208: "glances",
}

COMPOSE_CANDIDATES = ["docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"]

REPO_SERVICE_DIRS = [
    "adguard",
    "dozzle",
    "filebrowser",
    "forgejo",
    "homepage",
    "immich-app",
    "jellyfin",
    "jellyseerr",
    "juice-shop",
    "kopia",
    "librespeed",
    "mat2-web",
    "monitoring",
    "n8n",
    "navidrome",
    "ntopng",
    "prowlarr",
    "qbittorrent",
    "radarr",
    "scrutiny",
    "sonarr",
    "speedtest-tracker",
]
