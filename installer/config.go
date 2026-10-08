package installer

import (
	"path/filepath"
)

var SERVICE_SUBDIRS = map[string][]string{
	"AdGuard Home":      {"adguard/conf", "adguard/work"},
	"File Browser":      {"filebrowser/config", "filebrowser/database"},
	"Forgejo":           {"forgejo/data"},
	"Jellyfin":          {"jellyfin/config", "jellyfin/cache"},
	"Jellyseerr":        {"jellyseerr/config"},
	"kopia":             {"kopia/config", "kopia/cache", "kopia/logs"},
	"librespeed":        {"librespeed/config"},
	"Monitoring":        {"monitoring/prometheus-data", "monitoring/grafana-data"},
	"Navidrome":         {"navidrome/data"},
	"ntopng":            {"ntopng/redis-data", "ntopng/ntopng-data"},
	"Prowlarr":          {"prowlarr/config"},
	"qBittorrent":       {"qbittorrent/config"},
	"Radarr":            {"radarr/config"},
	"Scrutiny":          {"scrutiny/config"},
	"Sonarr":            {"sonarr/config"},
	"Speedtest Tracker": {"speedtest-tracker/config"},
}

var MEDIA_SUBDIRS = []string{
	"filmes",
	"series",
	"musicas",
	"fotos",
	"downloads",
	"inbox",
}

var WRITABLE_MODE int = 0o770

func ServiceDirs() map[string][]string {
	resultado := make(map[string][]string)
	for nome, subs := range SERVICE_SUBDIRS {
		for _, sub := range subs {
			resultado[nome] = append(resultado[nome], filepath.Join(DefaultBaseDir, sub))
		}
	}
	return resultado
}

func writable_dirs() []string {
	var resultado []string
	for _, dirs := range ServiceDirs() {
		resultado = append(resultado, dirs...)
	}
	return resultado
}

var DEFAULT_EXPECTED_PORTS = map[int]string{
	53:    "AdGuard Home (DNS)",
	2283:  "immich-app",
	3000:  "AdGuard Home (setup UI) / Grafana",
	3001:  "Homepage",
	3044:  "forgejo",
	3050:  "Juice Shop",
	3100:  "ntopng",
	4533:  "Navidrome",
	5055:  "Jellyseerr",
	5678:  "n8n",
	7878:  "Radarr",
	8080:  "qBittorrent",
	8081:  "Dozzle",
	8082:  "File Browser",
	8085:  "cAdvisor",
	8090:  "Scrutiny",
	8096:  "Jellyfin",
	8181:  "AdGuard Home (UI)",
	8282:  "MAT2 Web",
	8765:  "Speedtest Tracker",
	8989:  "Sonarr",
	9090:  "Prometheus",
	9100:  "Node Exporter",
	9696:  "Prowlarr",
	51515: "kopia",
	61208: "glances",
}

var COMPOSE_CANDIDATES = []string{
	"docker-compose.yml",
	"docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
}

var PROTECTED_REPO_DIRS = []string{
	"homepage/config",
	"mat2-web/nginx",
	"monitoring/prometheus.yml",
}
