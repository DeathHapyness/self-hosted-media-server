from pathlib import Path
from typing import Dict, List

from . import state

MIN_DISK_SPACE_DATA_GB = 2
MIN_DISK_SPACE_MEDIA_GB = 5

SERVICE_SUBDIRS = {
    "AdGuard Home": ["adguard/conf", "adguard/work"],
    "File Browser": ["filebrowser/config", "filebrowser/database"],
    "Forgejo": ["forgejo/data"],
    "Jellyfin": ["jellyfin/config", "jellyfin/cache"],
    "Jellyseerr": ["jellyseerr/config"],
    "kopia": ["kopia/config", "kopia/cache", "kopia/logs"],
    "librespeed": ["librespeed/config"],
    "Monitoring": ["monitoring/prometheus-data", "monitoring/grafana-data"],
    "Navidrome": ["navidrome/data"],
    "ntopng": ["ntopng/redis-data", "ntopng/ntopng-data"],
    "Prowlarr": ["prowlarr/config"],
    "qBittorrent": ["qbittorrent/config"],
    "Radarr": ["radarr/config"],
    "Scrutiny": ["scrutiny/config"],
    "Sonarr": ["sonarr/config"],
    "Speedtest Tracker": ["speedtest-tracker/config"],
}

MEDIA_SUBDIRS = ["filmes", "series", "musicas", "fotos", "downloads", "inbox"]

WRITABLE_MODE = 0o770


def service_dirs() -> Dict[str, List[Path]]:
    return {
        nome: [state.BASE_DIR / sub for sub in subs]
        for nome, subs in SERVICE_SUBDIRS.items()
    }


def writable_dirs() -> List[Path]:
    return [d for dirs in service_dirs().values() for d in dirs]


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

NON_SERVICE_DIRS = {
    "assets",
    "docs",
    "installer",
    "PortWatch",
    "__pycache__",
}

PROTECTED_REPO_DIRS = {
    "homepage/config",
    "mat2-web/nginx",
    "monitoring/prometheus.yml",
}
