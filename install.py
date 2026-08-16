#!/usr/bin/env python3
"""
install.py - Instalador do Home Lab / Self-Hosted Media Server

Menu único que reúne: instalação, desinstalação e gestão do disco de mídia
(/mnt/media).

Uso:
    sudo python3 install.py
    sudo python3 install.py --dry-run   # simula sem executar nada

Requisitos:
    - Linux
    - Python 3.8+
    - Docker + Docker Compose plugin (só necessário para instalar serviços)
    - Sem dependências externas (somente biblioteca padrão)
"""

import argparse
import os
import re
import shutil
import socket
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple

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

MEDIA_SUBDIRS = ["filmes", "series", "musicas", "fotos", "downloads", "inbox"]

WRITABLE_DIRS = [d for dirs in SERVICE_DIRS.values() for d in dirs]
WRITABLE_MODE = 0o770

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

REPO_SERVICE_DIRS = [
    "adguard",
    "dozzle",
    "filebrowser",
    "jellyfin",
    "navidrome",
    "qbittorrent",
    "homepage",
]

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

# Definido em main() a partir do --dry-run. Lido pelas funções que executam
# ações (nunca pelas funções que só verificam).
DRY_RUN = False


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


def print_danger(text: str) -> None:
    print(f"{_c('[PERIGO]', Color.RED + Color.BOLD)} {text}")


class InstallError(Exception):
    """Erro crítico que deve interromper a operação atual imediatamente."""


def run(cmd, **kwargs):
    """Executa um comando e retorna CompletedProcess. Nunca levanta por código != 0."""
    return subprocess.run(cmd, capture_output=True, text=True, check=False, **kwargs)


def confirm_dangerous(explicacao: str, palavra: str = "CONFIRMO") -> bool:
    """Exige que o usuário digite uma palavra exata antes de uma ação destrutiva.

    Nunca aceita apenas 's'/'y'/Enter — é proposital, para reduzir o risco de
    o usuário confirmar sem realmente ler o aviso.
    """
    print_danger(explicacao)
    print_info(f"Isto NÃO pode ser desfeito automaticamente. Digite exatamente \"{palavra}\" para continuar.")
    try:
        resposta = input(f"> ").strip()
    except (EOFError, KeyboardInterrupt):
        print()
        return False
    return resposta == palavra


def press_enter_to_continue() -> None:
    try:
        input("\nPressione Enter para voltar ao menu...")
    except (EOFError, KeyboardInterrupt):
        print()


# --------------------------------------------------------------------------
# Verificações (somente leitura — nunca alteram o sistema)
# --------------------------------------------------------------------------


def check_os() -> None:
    if sys.platform.startswith("linux"):
        print_ok("Linux detected")
    else:
        raise InstallError(f"Sistema operacional não suportado: {sys.platform}. Este instalador requer Linux.")


def check_docker() -> None:
    if shutil.which("docker") is None:
        raise InstallError(
            "Docker não encontrado.\n"
            "    Instale com: curl -fsSL https://get.docker.com | sh\n"
            "    Ou veja: https://docs.docker.com/engine/install/\n"
            "    Fonte do script: https://github.com/docker/docker-install"
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
            "Este instalador precisa ser executado como root (sudo).\n"
            "    Execute: sudo python3 install.py"
        )
    print_ok("Running with sufficient privileges")


def check_media_mount(raise_on_fail: bool = True) -> bool:
    if not MEDIA_DIR.exists():
        if raise_on_fail:
            raise InstallError(
                f"{MEDIA_DIR} não existe. Use a opção 'Gerenciar disco de mídia' no "
                "menu principal para criar/montar, ou monte manualmente antes de continuar."
            )
        return False
    if not os.path.ismount(MEDIA_DIR):
        if raise_on_fail:
            raise InstallError(
                f"{MEDIA_DIR} existe, mas NÃO está montado como filesystem separado. "
                "Isso é intencional: instalar sem o mount ativo poderia criar arquivos "
                "no disco raiz. Use a opção 'Gerenciar disco de mídia' no menu principal, "
                "ou monte manualmente (ex: via /etc/fstab)."
            )
        return False
    print_ok(f"{MEDIA_DIR} mounted")
    return True


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


def find_compose_files() -> Dict[str, Path]:
    script_dir = Path(__file__).resolve().parent
    found: Dict[str, Path] = {}
    for service_dir_name in REPO_SERVICE_DIRS:
        service_dir = script_dir / service_dir_name
        if not service_dir.is_dir():
            continue
        for candidate_name in COMPOSE_CANDIDATES:
            candidate = service_dir / candidate_name
            if candidate.exists():
                found[service_dir_name] = candidate
                break
    return found


def extract_ports_from_compose(compose_path: Path) -> Dict[int, str]:
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
    compose_files = find_compose_files()
    expected: Dict[int, str] = {}
    for service_name, compose_path in compose_files.items():
        service_ports = extract_ports_from_compose(compose_path)
        for port, container_name in service_ports.items():
            expected[port] = f"{service_name}/{container_name}"

    if not expected:
        print_warn("Não foi possível extrair portas dos compose files, usando lista padrão.")
        expected = {p: svc for p, svc in DEFAULT_EXPECTED_PORTS.items()}

    busy = []
    for port, service in sorted(expected.items()):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(0.5)
            in_use = s.connect_ex(("127.0.0.1", port)) == 0
        if in_use:
            busy.append((port, service))

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
# Gestão do disco de mídia (/mnt/media)
# --------------------------------------------------------------------------


def list_unmounted_block_devices() -> List[Tuple[str, str]]:
    """Lista partições/discos que existem mas não estão montados em lugar nenhum.

    Retorna lista de tuplas (nome_dispositivo, tamanho_legivel).
    Usa 'lsblk', que já vem por padrão em praticamente toda distro Linux.
    """
    result = run(["lsblk", "-rno", "NAME,SIZE,TYPE,MOUNTPOINT"])
    if result.returncode != 0:
        return []

    devices = []
    for line in result.stdout.splitlines():
        parts = line.split()
        if len(parts) < 3:
            continue
        name, size, dev_type = parts[0], parts[1], parts[2]
        mountpoint = parts[3] if len(parts) > 3 else ""
        if dev_type not in ("part", "disk"):
            continue
        if mountpoint:
            continue
        devices.append((f"/dev/{name}", size))
    return devices


def mount_real_device_menu() -> None:
    print_section("Discos/partições disponíveis (sem uso no momento):")
    devices = list_unmounted_block_devices()
    if not devices:
        print_warn("Nenhum disco/partição livre foi encontrado (ou 'lsblk' não está disponível).")
        return

    for i, (name, size) in enumerate(devices, start=1):
        print_info(f"{i}) {name} ({size})")

    escolha = input("\nDigite o número do dispositivo (ou vazio para cancelar): ").strip()
    if not escolha:
        print_info("Cancelado.")
        return

    try:
        idx = int(escolha) - 1
        device_path, device_size = devices[idx]
    except (ValueError, IndexError):
        print_fail("Opção inválida.")
        return

    aviso = (
        f"Você está prestes a FORMATAR {device_path} ({device_size}) como ext4.\n"
        f"    TODOS OS DADOS atualmente nesse dispositivo serão APAGADOS PERMANENTEMENTE.\n"
        f"    Confirme que {device_path} é realmente o disco certo antes de continuar."
    )
    if not confirm_dangerous(aviso, palavra=device_path):
        print_info("Operação cancelada — nada foi alterado.")
        return

    if DRY_RUN:
        print_ok(f"[dry-run] formataria {device_path} como ext4 e montaria em {MEDIA_DIR}")
        return

    MEDIA_DIR.mkdir(parents=True, exist_ok=True)

    print_section(f"Formatando {device_path}...")
    result = run(["mkfs.ext4", "-F", device_path])
    if result.returncode != 0:
        print_fail("Falha ao formatar:")
        print_info(result.stderr.strip())
        return
    print_ok("Formatado com sucesso")

    result = run(["mount", device_path, str(MEDIA_DIR)])
    if result.returncode != 0:
        print_fail("Falha ao montar:")
        print_info(result.stderr.strip())
        return
    print_ok(f"{device_path} montado em {MEDIA_DIR}")

    _adicionar_fstab(f"{device_path} {MEDIA_DIR} ext4 defaults 0 2")


def create_virtual_disk_menu() -> None:
    print_section("Criar disco virtual (arquivo de imagem) para /mnt/media")
    print_info("Isso ocupa espaço no disco atual — só recomendado para testes ou quando")
    print_info("não há disco físico disponível. Para produção, prefira um disco real.")

    tamanho = input("\nTamanho em GB (padrão 20): ").strip() or "20"
    try:
        tamanho_gb = int(tamanho)
        if tamanho_gb <= 0:
            raise ValueError
    except ValueError:
        print_fail("Tamanho inválido.")
        return

    img_path = Path("/var/media-disk.img")
    aviso = (
        f"Será criado um arquivo de {tamanho_gb}GB em {img_path} e montado em {MEDIA_DIR}.\n"
        f"    Se já existir um disco virtual anterior nesse caminho, ele será sobrescrito."
    )
    if not confirm_dangerous(aviso):
        print_info("Operação cancelada — nada foi alterado.")
        return

    if DRY_RUN:
        print_ok(f"[dry-run] criaria {img_path} ({tamanho_gb}GB) e montaria em {MEDIA_DIR}")
        return

    MEDIA_DIR.mkdir(parents=True, exist_ok=True)

    print_section("Criando arquivo de disco virtual...")
    result = run(["fallocate", "-l", f"{tamanho_gb}G", str(img_path)])
    if result.returncode != 0:
        # fallocate pode não funcionar em alguns filesystems — usa dd como fallback
        result = run(["dd", "if=/dev/zero", f"of={img_path}", "bs=1M", f"count={tamanho_gb * 1024}"])
        if result.returncode != 0:
            print_fail("Falha ao criar o arquivo de disco virtual:")
            print_info(result.stderr.strip())
            return
    print_ok(f"Arquivo criado ({tamanho_gb}GB)")

    result = run(["mkfs.ext4", "-F", str(img_path)])
    if result.returncode != 0:
        print_fail("Falha ao formatar:")
        print_info(result.stderr.strip())
        return
    print_ok("Formatado como ext4")

    result = run(["mount", "-o", "loop", str(img_path), str(MEDIA_DIR)])
    if result.returncode != 0:
        print_fail("Falha ao montar:")
        print_info(result.stderr.strip())
        return
    print_ok(f"Montado em {MEDIA_DIR}")

    _adicionar_fstab(f"{img_path} {MEDIA_DIR} ext4 loop 0 0")


def _adicionar_fstab(linha: str) -> None:
    fstab = Path("/etc/fstab")
    conteudo = fstab.read_text() if fstab.exists() else ""
    if linha in conteudo:
        print_ok("/etc/fstab já contém essa entrada")
        return
    resposta = input("\nAdicionar entrada em /etc/fstab para sobreviver a um reboot? [s/N]: ").strip().lower()
    if resposta != "s":
        print_info("Pulado — o mount não sobrevive a um reboot até você adicionar manualmente.")
        return
    with fstab.open("a") as f:
        f.write(f"\n{linha}\n")
    print_ok("/etc/fstab atualizado")


def media_mount_menu() -> None:
    while True:
        print_header("GESTÃO DO DISCO DE MÍDIA")
        montado = check_media_mount(raise_on_fail=False)
        status = _c("montado", Color.GREEN) if montado else _c("NÃO montado", Color.RED)
        print(f"Status atual de {MEDIA_DIR}: {status}\n")

        print_info("1) Já montei manualmente — só verificar")
        print_info("2) Usar um disco/partição existente (formata e monta — APAGA DADOS do disco escolhido)")
        print_info("3) Criar um disco virtual (arquivo, menos arriscado, ocupa espaço do disco atual)")
        print_info("4) Voltar ao menu principal")

        escolha = input("\nEscolha uma opção: ").strip()
        if escolha == "1":
            check_media_mount(raise_on_fail=False)
            press_enter_to_continue()
        elif escolha == "2":
            mount_real_device_menu()
            press_enter_to_continue()
        elif escolha == "3":
            create_virtual_disk_menu()
            press_enter_to_continue()
        elif escolha == "4":
            return
        else:
            print_fail("Opção inválida.")


# --------------------------------------------------------------------------
# Criação de diretórios e permissões
# --------------------------------------------------------------------------


def create_directories() -> None:
    print_section("Creating directories...")
    for service_name, dirs in SERVICE_DIRS.items():
        created_any = False
        for d in dirs:
            existed = d.exists()
            if DRY_RUN:
                if not existed:
                    created_any = True
                    print_info(f"[dry-run] criaria {d}")
                continue
            d.mkdir(parents=True, exist_ok=True)
            if not existed:
                created_any = True
        if created_any:
            print_ok(f"{service_name} directories")
        else:
            print_ok(f"{service_name} directories (already existed)")

    for sub in MEDIA_SUBDIRS:
        path = MEDIA_DIR / sub
        existed = path.exists()
        if DRY_RUN:
            if not existed:
                print_info(f"[dry-run] criaria /mnt/media/{sub}")
            continue
        path.mkdir(parents=True, exist_ok=True)
        if not existed:
            print_ok(f"/mnt/media/{sub} created")


def resolve_puid_pgid() -> Optional[Tuple[int, int]]:
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

    if DRY_RUN:
        print_info(f"[dry-run] ajustaria permissões de {len(WRITABLE_DIRS)} diretórios e de {MEDIA_DIR}")
        print_ok("Permissions configured (simulado)")
        return

    for d in WRITABLE_DIRS:
        os.chmod(d, WRITABLE_MODE)
        if owner:
            try:
                os.chown(d, owner[0], owner[1])
            except OSError:
                print_warn(f"Não foi possível ajustar dono de {d} para PUID/PGID do .env")

    os.chmod(MEDIA_DIR, 0o775)

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


def ensure_env_file() -> None:
    script_dir = Path(__file__).resolve().parent
    env_path = script_dir / ".env"
    example_path = script_dir / ".env.example"

    if env_path.exists():
        print_ok(".env already present")
        return

    if not example_path.exists():
        print_warn(
            "Nenhum .env ou .env.example encontrado. Se algum serviço exigir "
            "credenciais, crie um .env manualmente antes de continuar."
        )
        return

    if DRY_RUN:
        print_info("[dry-run] copiaria .env.example para .env")
        return

    shutil.copy(example_path, env_path)
    print_ok(".env created from .env.example")
    print_warn("Revise o .env criado e preencha os valores necessários antes de subir os serviços.")


# --------------------------------------------------------------------------
# Docker Compose
# --------------------------------------------------------------------------


def validate_compose() -> Dict[str, Path]:
    print_section("Validating Docker Compose...")
    compose_files = find_compose_files()
    if not compose_files:
        raise InstallError(
            "Nenhum arquivo docker-compose.yml/yaml encontrado dentro das pastas de "
            f"serviço ({', '.join(REPO_SERVICE_DIRS)})."
        )

    missing = [name for name in REPO_SERVICE_DIRS if name not in compose_files]
    if missing:
        print_warn(f"Sem compose encontrado para: {', '.join(missing)} (pulando esses serviços)")

    failed = []
    for service_name, compose_path in compose_files.items():
        result = run(["docker", "compose", "-f", str(compose_path), "config"])
        if result.returncode != 0:
            failed.append((service_name, result.stderr.strip()))
        else:
            print_ok(f"{service_name} compose válido")

    if failed:
        for service_name, error in failed:
            print_fail(f"{service_name}: configuração inválida")
            print_info(error)
        raise InstallError("Corrija os compose files acima e execute o instalador novamente.")

    return compose_files


def start_services(compose_files: Dict[str, Path]) -> None:
    print_section("Starting services...")
    if DRY_RUN:
        for service_name, compose_path in compose_files.items():
            print_info(f"[dry-run] rodaria: docker compose -f {compose_path} -p {service_name} up -d")
        return

    failed = []
    for service_name, compose_path in compose_files.items():
        result = run(["docker", "compose", "-f", str(compose_path), "-p", service_name, "up", "-d"])
        if result.returncode != 0:
            failed.append((service_name, result.stderr.strip()))
        else:
            print_ok(f"{service_name} container(s) started")

    if failed:
        for service_name, error in failed:
            print_fail(f"{service_name}: falha ao iniciar")
            print_info(error)
        raise InstallError("Verifique os logs acima e 'docker compose logs' de cada serviço para mais detalhes.")


def stop_services(compose_files: Dict[str, Path]) -> None:
    print_section("Parando serviços...")
    if not compose_files:
        print_warn("Nenhum compose encontrado — nada para parar.")
        return

    if DRY_RUN:
        for service_name, compose_path in compose_files.items():
            print_info(f"[dry-run] rodaria: docker compose -f {compose_path} -p {service_name} down")
        return

    for service_name, compose_path in compose_files.items():
        result = run(["docker", "compose", "-f", str(compose_path), "-p", service_name, "down"])
        if result.returncode != 0:
            print_warn(f"{service_name}: {result.stderr.strip()}")
        else:
            print_ok(f"{service_name} parado")


# --------------------------------------------------------------------------
# Fluxo: instalar
# --------------------------------------------------------------------------


def fluxo_instalar() -> None:
    print_header("INSTALAÇÃO")
    if DRY_RUN:
        print_warn("Modo --dry-run ativo: nenhuma alteração real será feita.")

    steps = [
        check_os,
        check_docker,
        check_docker_compose,
        check_docker_daemon,
        lambda: check_media_mount(raise_on_fail=True),
        check_disk_space,
        check_ports,
    ]

    try:
        for step in steps:
            step()

        create_directories()
        configure_permissions()
        ensure_env_file()
        compose_files = validate_compose()
        start_services(compose_files)

    except InstallError as e:
        print_fail(str(e))
        print(f"\n{_c('Instalação interrompida.', Color.RED + Color.BOLD)}\n")
        return

    print_header("INSTALLATION COMPLETE")


# --------------------------------------------------------------------------
# Fluxo: desinstalar
# --------------------------------------------------------------------------


def fluxo_desinstalar() -> None:
    print_header("DESINSTALAÇÃO")
    compose_files = find_compose_files()

    stop_services(compose_files)

    print_info(f"\n{MEDIA_DIR} NUNCA é tocado por esta opção — seus arquivos de mídia estão seguros.")

    resposta = input(
        f"\nTambém remover as pastas de configuração em {BASE_DIR}? "
        "Isso apaga bancos de dados e configs dos serviços. [s/N]: "
    ).strip().lower()

    if resposta != "s":
        print_info("Configurações mantidas. Apenas os containers foram parados.")
        return

    aviso = (
        f"Isso vai apagar PERMANENTEMENTE tudo dentro de {BASE_DIR}: configs, "
        f"bancos de dados (File Browser, Navidrome, etc.) e caches.\n"
        f"    {MEDIA_DIR} continua intocado."
    )
    if not confirm_dangerous(aviso):
        print_info("Operação cancelada — configurações mantidas.")
        return

    if DRY_RUN:
        print_info(f"[dry-run] removeria {BASE_DIR}")
        return

    shutil.rmtree(BASE_DIR, ignore_errors=True)
    print_ok(f"{BASE_DIR} removido")


# --------------------------------------------------------------------------
# Menu principal
# --------------------------------------------------------------------------


def main_menu() -> None:
    while True:
        print_header("MEDIA SERVER INSTALLER")
        if DRY_RUN:
            print_warn("Modo --dry-run ativo\n")

        print_info("1) Instalar / atualizar serviços")
        print_info("2) Desinstalar serviços")
        print_info("3) Gerenciar disco de mídia (/mnt/media)")
        print_info("4) Sair")

        try:
            escolha = input("\nEscolha uma opção: ").strip()
        except (EOFError, KeyboardInterrupt):
            print()
            return

        if escolha == "1":
            fluxo_instalar()
            press_enter_to_continue()
        elif escolha == "2":
            fluxo_desinstalar()
            press_enter_to_continue()
        elif escolha == "3":
            media_mount_menu()
        elif escolha == "4":
            print_info("Até mais!")
            return
        else:
            print_fail("Opção inválida.")


# --------------------------------------------------------------------------
# Main
# --------------------------------------------------------------------------


def main() -> int:
    global DRY_RUN

    parser = argparse.ArgumentParser(description="Instalador do Home Lab / Self-Hosted Media Server")
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simula todas as ações sem executar nada de fato (nenhum arquivo criado, nenhum comando destrutivo rodado).",
    )
    args = parser.parse_args()
    DRY_RUN = args.dry_run

    print(ASCII_ART)

    try:
        check_privileges()
    except InstallError as e:
        print_fail(str(e))
        return 1

    try:
        main_menu()
    except KeyboardInterrupt:
        print("\n")
        print_warn("Encerrado pelo usuário.")
        return 130

    return 0


if __name__ == "__main__":
    sys.exit(main())