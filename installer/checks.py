import logging
import os
import shutil
import socket
import sys
from pathlib import Path
from typing import Dict

from . import state
from .compose import extract_ports_from_compose
from .config import DEFAULT_EXPECTED_PORTS, MIN_DISK_SPACE_DATA_GB, MIN_DISK_SPACE_MEDIA_GB
from .output import print_fail, print_info, print_ok, print_warn
from .utils import InstallError

logger = logging.getLogger(__name__)


def check_os() -> None:
    if sys.platform.startswith("linux"):
        logger.info("Sistema operacional: %s", sys.platform)
        print_ok("Linux detected")
    else:
        raise InstallError(f"Sistema operacional não suportado: {sys.platform}. Este instalador requer Linux.")


def check_privileges() -> None:
    if os.geteuid() != 0:
        raise InstallError(
            "Este instalador precisa ser executado como root (sudo).\n"
            "    Execute: sudo python3 install.py"
        )
    print_ok("Running with sufficient privileges")


def check_media_mount(raise_on_fail: bool = True) -> bool:
    if not state.MEDIA_DIR.exists():
        if raise_on_fail:
            raise InstallError(
                f"{state.MEDIA_DIR} não existe. Use a opção 'Gerenciar disco de mídia' no "
                "menu principal para criar/montar, ou monte manualmente antes de continuar."
            )
        return False
    if not os.path.ismount(state.MEDIA_DIR):
        if raise_on_fail:
            raise InstallError(
                f"{state.MEDIA_DIR} existe, mas NÃO está montado como filesystem separado. "
                "Isso é intencional: instalar sem o mount ativo poderia criar arquivos "
                "no disco raiz. Use a opção 'Gerenciar disco de mídia' no menu principal, "
                "ou monte manualmente (ex: via /etc/fstab)."
            )
        return False
    print_ok(f"{state.MEDIA_DIR} mounted")
    return True


def resolve_media_mount_interactive() -> bool:
    """Um disco dedicado para a mídia é opcional: fica a critério do usuário."""
    if check_media_mount(raise_on_fail=False):
        return True
    print_warn(
        f"{state.MEDIA_DIR} não existe ou não é um mount separado. Ter um disco dedicado "
        "é opcional — você pode seguir usando esse caminho como uma pasta comum no disco raiz."
    )
    resposta = input("Continuar mesmo assim? [s/N]: ").strip().lower()
    return resposta == "s"


def _primeiro_existente(path: Path) -> Path:
    while not path.exists() and path != path.parent:
        path = path.parent
    return path


def check_disk_space() -> None:
    data_check_path = _primeiro_existente(state.BASE_DIR)
    data_free_gb = shutil.disk_usage(data_check_path).free / (1024 ** 3)
    if data_free_gb < MIN_DISK_SPACE_DATA_GB:
        raise InstallError(
            f"Espaço insuficiente em {data_check_path}: {data_free_gb:.1f}GB livres, "
            f"mínimo exigido {MIN_DISK_SPACE_DATA_GB}GB."
        )

    logger.info("Espaco livre em %s: %.1fGB", data_check_path, data_free_gb)
    media_check_path = _primeiro_existente(state.MEDIA_DIR)
    media_free_gb = shutil.disk_usage(media_check_path).free / (1024 ** 3)
    logger.info("Espaco livre em %s: %.1fGB", media_check_path, media_free_gb)
    if media_free_gb < MIN_DISK_SPACE_MEDIA_GB:
        print_warn(
            f"Pouco espaço livre em {media_check_path}: {media_free_gb:.1f}GB "
            f"(recomendado: {MIN_DISK_SPACE_MEDIA_GB}GB+). Continuando mesmo assim."
        )
    else:
        print_ok(f"Storage available ({media_free_gb:.1f}GB free on {media_check_path})")


def check_ports(compose_files: Dict[str, Path]) -> None:
    expected: Dict[int, str] = {}
    for service_name, compose_path in compose_files.items():
        service_ports = extract_ports_from_compose(compose_path)
        for port, container_name in service_ports.items():
            expected[port] = f"{service_name}/{container_name}"

    if not expected:
        print_warn("Não foi possível extrair portas dos compose files, usando lista padrão.")
        expected = dict(DEFAULT_EXPECTED_PORTS)

    busy = []
    for port, service in sorted(expected.items()):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(0.5)
            in_use = s.connect_ex(("127.0.0.1", port)) == 0
        if in_use:
            busy.append((port, service))

    if busy:
        logger.warning("Portas ocupadas: %s", ", ".join(f"{p} ({s})" for p, s in busy))
        print_fail("Portas em uso detectadas:")
        for port, service in busy:
            print_info(f"porta {port} ({service}) já está em uso")
        raise InstallError(
            "Libere as portas acima (pare o serviço que as está usando) ou ajuste "
            "o docker-compose.yml antes de continuar."
        )
    print_ok("Required ports available")
