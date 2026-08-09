#!/usr/bin/env python3
"""
install.py - Instalador do Home Lab / Self-Hosted Media Server

Prepara um servidor Linux do zero para rodar os serviços do projeto via
Docker Compose (AdGuard Home, Dozzle, File Browser, Jellyfin, Navidrome,
qBittorrent, Homepage), sem apagar ou sobrescrever configurações existentes.

Uso:
    sudo python3 install.py

Requisitos:
    - Linux
    - Python 3.8+
    - Docker + Docker Compose plugin
    - /mnt/media montado
    - Sem dependências externas (somente biblioteca padrão)
"""

import os
import re
import shutil
import socket
import subprocess
import sys
from pathlib import Path
from typing import Dict, Optional, Tuple

# --------------------------------------------------------------------------
# Banner
# --------------------------------------------------------------------------

ASCII_ART = r"""  ___           _        _           _              ____       _  __       _   _           _           _
 |_ _|_ __  ___| |_ __ _| | __ _  __| | ___  _ __  / ___|  ___| |/ _|     | | | | ___  ___| |_ ___  __| |
 | || '_ \/ __| __/ _` | |/ _` |/ _` |/ _ \| '__| \___ \ / _ \ | |_ _____| |_| |/ _ \/ __| __/ _ \/ _` |
 | || | | \__ \ || (_| | | (_| | (_| | (_) | |     ___) |  __/ |  _|_____|  _  | (_) \__ \ ||  __/ (_| |
|___|_| |_|___/\__\__,_|_|\__,_|\__,_|\___/|_|    |____/ \___|_|_|       |_| |_|\___/|___/\__\___|\__,_|
"""

# --------------------------------------------------------------------------
# Configuração
# --------------------------------------------------------------------------

BASE_DIR = Path("/opt/media-server")
MEDIA_DIR = Path("/mnt/media")

# Espaço mínimo livre exigido (em GB) em /opt e em /mnt/media
MIN_DISK_SPACE_OPT_GB = 2
MIN_DISK_SPACE_MEDIA_GB = 5

# Diretórios criados por serviço (relativos a BASE_DIR)
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
    "Jellyfin": [
        BASE_DIR / "jellyfin" / "config",
        BASE_DIR / "jellyfin" / "cache",
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
}

# Subpastas de mídia esperadas (não sensíveis, apenas garantidas se o mount existir)
MEDIA_SUBDIRS = ["filmes", "series", "musicas", "fotos", "downloads", "inbox"]

# Diretórios de /opt/media-server que precisam ser graváveis pelos containers
WRITABLE_DIRS = [d for dirs in SERVICE_DIRS.values() for d in dirs]
WRITABLE_MODE = 0o770

# Portas host esperadas (usadas como fallback caso não seja possível extrair
# do docker-compose.yml). Ajuste conforme o seu compose real.
DEFAULT_EXPECTED_PORTS = {
    53: "AdGuard Home (DNS)",
    3000: "AdGuard Home (setup UI) / Homepage",
    8081: "Dozzle",
    8082: "File Browser",
    8096: "Jellyfin",
    4533: "Navidrome",
    8080: "qBittorrent",
}

COMPOSE_CANDIDATES = ["docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"]

# --------------------------------------------------------------------------
# Saída no terminal
# --------------------------------------------------------------------------

class Color:
    GREEN = "\033[92m"
    RED = "\033[91m"
    YELLOW = "\033[93m"
    CYAN = "\033[96m"
    BOLD = "\033[1m"
    RESET = "\033[0m"

USE_COLOR = sys.stdout.isatty()


def _c(text: str, color: str) -> str:
    if not USE_COLOR:
        return text
    return f"{color}{text}{Color.RESET}"


def print_header(text: str) -> None:
    line = "=" * 40
    print(f"\n{_c(line, Color.CYAN)}")
    print(_c(text.center(40), Color.CYAN + Color.BOLD))
    print(f"{_c(line, Color.CYAN)}\n")


def print_section(text: str) -> None:
    print(f"\n{_c(text, Color.BOLD)}")


def print_ok(text: str) -> None:
    print(f"{_c('[✓]', Color.GREEN)} {text}")


def print_fail(text: str) -> None:
    print(f"{_c('[✗]', Color.RED)} {text}")


def print_warn(text: str) -> None:
    print(f"{_c('[!]', Color.YELLOW)} {text}")


def print_info(text: str) -> None:
    print(f"    {text}")


class InstallError(Exception):
    """Erro crítico que deve interromper a instalação imediatamente."""


def run(cmd, **kwargs):
    """Executa um comando e retorna CompletedProcess. Nunca levanta por código != 0."""
    return subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        check=False,
        **kwargs,
    )


# --------------------------------------------------------------------------
# Verificações
# --------------------------------------------------------------------------

def check_os() -> None:
    if sys.platform.startswith("linux"):
        print_ok("Linux detected")
    else:
        raise InstallError(f"Sistema operacional não suportado: {sys.platform}. Este instalador requer Linux.")


def check_docker() -> None:
    if shutil.which("docker") is None:
        raise InstallError(
            "Docker não encontrado. Instale o Docker antes de continuar: "
            "https://docs.docker.com/engine/install/"
        )
    result = run(["docker", "--version"])
    if result.returncode != 0:
        raise InstallError("Docker está instalado, mas 'docker --version' falhou.")
    print_ok(f"Docker installed ({result.stdout.strip()})")


def check_docker_compose() -> None:
    result = run(["docker", "compose", "version"])
    if result.returncode != 0:
        raise InstallError(
            "Docker Compose (plugin 'docker compose') não encontrado. "
            "Instale o plugin oficial do Docker Compose v2."
        )
    print_ok(f"Docker Compose installed ({result.stdout.strip()})")


def check_docker_daemon() -> None:
    result = run(["docker", "info"])
    if result.returncode != 0:
        raise InstallError(
            "Docker daemon não está rodando ou o usuário atual não tem permissão. "
            "Tente 'sudo systemctl start docker' ou rode este script com sudo."
        )
    print_ok("Docker daemon running")


def check_privileges() -> None:
    if os.geteuid() != 0:
        raise InstallError(
            "Este instalador precisa ser executado como root (sudo), pois cria "
            "diretórios em /opt e ajusta permissões.\n"
            "    Execute: sudo python3 install.py"
        )
    print_ok("Running with sufficient privileges")


def check_media_mount() -> None:
    if not MEDIA_DIR.exists():
        raise InstallError(
            f"{MEDIA_DIR} não existe. Crie o ponto de montagem e monte o disco de "
            "mídia antes de continuar."
        )
    if not os.path.ismount(MEDIA_DIR):
        raise InstallError(
            f"{MEDIA_DIR} existe, mas NÃO está montado como filesystem separado. "
            "Isso é intencional: instalar sem o mount ativo poderia criar arquivos "
            "no disco raiz. Monte o disco de mídia (ex: via /etc/fstab) e execute o "
            "instalador novamente."
        )
    print_ok(f"{MEDIA_DIR} mounted")


def check_disk_space() -> None:
    opt_check_path = BASE_DIR if BASE_DIR.exists() else BASE_DIR.parent
    opt_free_gb = shutil.disk_usage(opt_check_path).free / (1024 ** 3)
    if opt_free_gb < MIN_DISK_SPACE_OPT_GB:
        raise InstallError(
            f"Espaço insuficiente em {opt_check_path}: {opt_free_gb:.1f}GB livres, "
            f"mínimo exigido {MIN_DISK_SPACE_OPT_GB}GB."
        )

    media_free_gb = shutil.disk_usage(MEDIA_DIR).free / (1024 ** 3)
    if media_free_gb < MIN_DISK_SPACE_MEDIA_GB:
        print_warn(
            f"Pouco espaço livre em {MEDIA_DIR}: {media_free_gb:.1f}GB "
            f"(recomendado: {MIN_DISK_SPACE_MEDIA_GB}GB+). Continuando mesmo assim."
        )
    else:
        print_ok(f"Storage available ({media_free_gb:.1f}GB free on {MEDIA_DIR})")


def find_compose_file() -> Optional[Path]:
    script_dir = Path(__file__).resolve().parent
    for name in COMPOSE_CANDIDATES:
        candidate = script_dir / name
        if candidate.exists():
            return candidate
    return None


def extract_ports_from_compose(compose_path: Path) -> Dict[int, str]:
    """Extrai portas host (formato 'HOST:CONTAINER') do compose via regex.

    Evita dependência de PyYAML. Não é um parser YAML completo, mas cobre o
    formato padrão de 'ports:' usado neste projeto.
    """
    ports: Dict[int, str] = {}
    try:
        text = compose_path.read_text()
    except OSError:
        return ports

    current_service = None
    for raw_line in text.splitlines():
        service_match = re.match(r"^\s{2}([a-zA-Z0-9_-]+):\s*$", raw_line)
        if service_match:
            current_service = service_match.group(1)
            continue
        port_match = re.search(r'-\s*"?(\d+):\d+(?:/\w+)?"?', raw_line)
        if port_match:
            ports[int(port_match.group(1))] = current_service or "desconhecido"
    return ports


def check_ports() -> None:
    compose_path = find_compose_file()
    expected = extract_ports_from_compose(compose_path) if compose_path else {}
    if not expected:
        print_warn("Não foi possível extrair portas do docker-compose.yml, usando lista padrão.")
        expected = {p: svc for p, svc in DEFAULT_EXPECTED_PORTS.items()}

    busy = []
    for port, service in sorted(expected.items()):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(0.5)
            in_use = s.connect_ex(("127.0.0.1", port)) == 0
        if in_use:
            busy.append((port, service))
from __future__ import annotations
    if busy:
        print_fail("Portas em uso detectadas:")
        for port, service in busy:
            print_info(f"porta {port} ({service}) já está em uso")
        raise InstallError(
            "Libere as portas acima (pare o serviço que as está usando) ou ajuste "
            "o docker-compose.yml antes de continuar."
        )
    print_ok("Required ports available")


# --------------------------------------------------------------------------
# Criação de diretórios e permissões
# --------------------------------------------------------------------------

def create_directories() -> None:
    print_section("Creating directories...")
    for service_name, dirs in SERVICE_DIRS.items():
        created_any = False
        for d in dirs:
            existed = d.exists()
            d.mkdir(parents=True, exist_ok=True)
            if not existed:
                created_any = True
        if created_any:
            print_ok(f"{service_name} directories")
        else:
            print_ok(f"{service_name} directories (already existed)")

    # Subpastas de mídia: só cria se o mount já foi validado em check_media_mount()
    for sub in MEDIA_SUBDIRS:
        path = MEDIA_DIR / sub
        existed = path.exists()
        path.mkdir(parents=True, exist_ok=True)
        if not existed:
            print_ok(f"/mnt/media/{sub} created")


def resolve_puid_pgid() -> Optional[Tuple[int, int]]:
    """Lê PUID/PGID de um .env na raiz do projeto, se existir."""
    script_dir = Path(__file__).resolve().parent
    env_path = script_dir / ".env"
    if not env_path.exists():
        return None
    values = {}
    for line in env_path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, value = line.partition("=")
        values[key.strip()] = value.strip()
    if "PUID" in values and "PGID" in values:
        try:
            return int(values["PUID"]), int(values["PGID"])
        except ValueError:
            return None
    return None


def configure_permissions() -> None:
    print_section("Configuring permissions...")
    owner = resolve_puid_pgid()

    for d in WRITABLE_DIRS:
        os.chmod(d, WRITABLE_MODE)
        if owner:
            try:
                os.chown(d, owner[0], owner[1])
            except OSError:
                print_warn(f"Não foi possível ajustar dono de {d} para PUID/PGID do .env")

    # File Browser precisa de leitura/escrita em /mnt/media inteiro
    os.chmod(MEDIA_DIR, 0o775)

    # qBittorrent precisa de leitura/escrita em downloads
    downloads_dir = MEDIA_DIR / "downloads"
    if downloads_dir.exists():
        os.chmod(downloads_dir, 0o775)

    if owner:
        print_ok(f"Permissions configured (PUID={owner[0]} PGID={owner[1]})")
    else:
        print_ok("Permissions configured")
        print_warn(
            "Nenhum PUID/PGID encontrado em .env — mantendo dono atual dos "
            "diretórios. Ajuste manualmente se os containers rodarem com outro "
            "usuário."
        )
    print_info("Jellyfin e Navidrome devem montar a mídia como somente leitura (:ro) no docker-compose.yml")


# --------------------------------------------------------------------------
# .env
# --------------------------------------------------------------------------

def ensure_env_file() -> None:
    script_dir = Path(__file__).resolve().parent
    env_path = script_dir / ".env"
    example_path = script_dir / ".env.example"

    if env_path.exists():
        print_ok(".env already present")
        return

    if example_path.exists():
        shutil.copy(example_path, env_path)
        print_ok(".env created from .env.example")
        print_warn("Revise o .env criado e preencha os valores necessários antes de subir os serviços.")
    else:
        print_warn(
            "Nenhum .env ou .env.example encontrado. Se algum serviço exigir "
            "credenciais, crie um .env manualmente antes de continuar."
        )


# --------------------------------------------------------------------------
# Docker Compose
# --------------------------------------------------------------------------

def validate_compose() -> Path:
    print_section("Validating Docker Compose...")
    compose_path = find_compose_file()
    if compose_path is None:
        raise InstallError(
            "Nenhum arquivo docker-compose.yml/yaml encontrado ao lado de install.py."
        )

    result = run(["docker", "compose", "-f", str(compose_path), "config"])
    if result.returncode != 0:
        print_fail("Configuração do Docker Compose inválida:")
        print_info(result.stderr.strip())
        raise InstallError("Corrija o docker-compose.yml e execute o instalador novamente.")

    print_ok("Configuration valid")
    return compose_path


def start_services(compose_path: Path) -> None:
    print_section("Starting services...")
    result = run(["docker", "compose", "-f", str(compose_path), "up", "-d"])
    if result.returncode != 0:
        print_fail("Falha ao iniciar os containers:")
        print_info(result.stderr.strip())
        raise InstallError("Verifique os logs acima e execute 'docker compose logs' para mais detalhes.")
    print_ok("Containers started")


# --------------------------------------------------------------------------
# Main
# --------------------------------------------------------------------------

def main() -> int:
    print(ASCII_ART)
    print_header("MEDIA SERVER INSTALLER")

    steps = [
        check_privileges,
        check_os,
        check_docker,
        check_docker_compose,
        check_docker_daemon,
        check_media_mount,
        check_disk_space,
        check_ports,
    ]

    try:
        for step in steps:
            step()

        create_directories()
        configure_permissions()
        ensure_env_file()
        compose_path = validate_compose()
        start_services(compose_path)

    except InstallError as e:
        print_fail(str(e))
        print(f"\n{_c('Instalação interrompida.', Color.RED + Color.BOLD)}\n")
        return 1
    except KeyboardInterrupt:
        print("\n")
        print_warn("Instalação cancelada pelo usuário.")
        return 130

    print_header("INSTALLATION COMPLETE")
    return 0


if __name__ == "__main__":
    sys.exit(main())
