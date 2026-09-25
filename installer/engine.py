import logging
import shutil
from typing import List

from . import state
from .output import print_fail, print_info, print_ok, print_section, print_warn
from .utils import InstallError, run

logger = logging.getLogger(__name__)


def choose_engine() -> str:
    docker_ok = shutil.which("docker") is not None
    podman_ok = shutil.which("podman") is not None

    if docker_ok and podman_ok:
        print_section("Docker e Podman detectados neste sistema.")
        print_info("1) Docker (padrão)")
        print_info("2) Podman")
        escolha = input("Qual usar? [1]: ").strip()
        return "podman" if escolha == "2" else "docker"

    if podman_ok and not docker_ok:
        return "podman"

    return "docker"


def check_engine_binary() -> None:
    binary = state.CONTAINER_ENGINE
    if shutil.which(binary) is None:
        if binary == "podman":
            raise InstallError(
                "Podman não encontrado. Instale com o gerenciador de pacotes da sua "
                "distro (ex: apt install podman) ou veja: https://podman.io/docs/installation"
            )
        raise InstallError(
            "Docker não encontrado.\n"
            "    Instale com: curl -fsSL https://get.docker.com | sh\n"
            "    Ou veja: https://docs.docker.com/engine/install/\n"
            "    Fonte do script: https://github.com/docker/docker-install"
        )
    result = run([binary, "--version"])
    if result.returncode != 0:
        raise InstallError(f"{binary} está instalado, mas '{binary} --version' falhou.")
    logger.info("Engine %s: %s", binary, result.stdout.strip())
    print_ok(f"{binary.capitalize()} installed ({result.stdout.strip()})")


def check_compose_plugin() -> None:
    engine = state.CONTAINER_ENGINE
    result = run([engine, "compose", "version"])
    if result.returncode == 0:
        state.COMPOSE_STANDALONE = False
        print_ok(f"{engine.capitalize()} Compose installed ({result.stdout.strip()})")
        return

    if engine == "podman" and shutil.which("podman-compose"):
        result = run(["podman-compose", "--version"])
        if result.returncode == 0:
            state.COMPOSE_STANDALONE = True
            print_ok(f"podman-compose installed ({result.stdout.strip()})")
            return

    raise InstallError(
        f"Compose não encontrado para {engine}. Instale o plugin '{engine} compose' "
        "ou, no caso do Podman, o pacote podman-compose."
    )


def check_engine_daemon() -> None:
    engine = state.CONTAINER_ENGINE
    result = run([engine, "info"])
    if result.returncode != 0:
        raise InstallError(
            f"{engine.capitalize()} não está rodando ou o usuário atual não tem permissão. "
            f"Tente 'sudo systemctl start {engine}' ou rode este script com sudo."
        )
    print_ok(f"{engine.capitalize()} daemon running")


def compose_base_cmd() -> List[str]:
    if state.CONTAINER_ENGINE == "podman" and state.COMPOSE_STANDALONE:
        return ["podman-compose"]
    return [state.CONTAINER_ENGINE, "compose"]


def check_optional(check_fn, label: str) -> bool:
    """Roda uma checagem; se falhar, deixa o usuário decidir se continua mesmo assim."""
    try:
        check_fn()
        return True
    except InstallError as e:
        print_fail(str(e))
        resposta = input(f"\nContinuar mesmo assim, pulando a verificação de {label}? [s/N]: ").strip().lower()
        if resposta == "s":
            logger.warning("Usuario optou por pular a verificacao de %s apos falha: %s", label, e)
            print_warn(f"Continuando sem validar {label} — por sua conta e risco.")
            return True
        return False
