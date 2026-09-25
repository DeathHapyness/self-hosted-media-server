import logging
import os
import re
from pathlib import Path
from typing import Dict, List

from . import state
from .config import COMPOSE_CANDIDATES, NON_SERVICE_DIRS, PROTECTED_REPO_DIRS
from .engine import compose_base_cmd
from .output import print_fail, print_info, print_ok, print_section, print_warn
from .utils import InstallError, run

logger = logging.getLogger(__name__)


def compose_env() -> Dict[str, str]:
    return {
        **os.environ,
        "DATA_DIR": str(state.BASE_DIR),
        "MEDIA_DIR": str(state.MEDIA_DIR),
    }


def find_compose_files() -> Dict[str, Path]:
    repo_root = Path(__file__).resolve().parent.parent
    found: Dict[str, Path] = {}
    for service_dir in sorted(repo_root.iterdir()):
        if not service_dir.is_dir():
            continue
        if service_dir.name.startswith(".") or service_dir.name in NON_SERVICE_DIRS:
            continue
        for candidate_name in COMPOSE_CANDIDATES:
            candidate = service_dir / candidate_name
            if candidate.exists():
                found[service_dir.name] = candidate
                break
    logger.info("Servicos descobertos (%d): %s", len(found), ", ".join(sorted(found)))
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


def extract_data_dirs_from_compose(compose_path: Path) -> List[Path]:
    repo_root = Path(__file__).resolve().parent.parent
    service_dir = compose_path.parent
    protegidos = {(repo_root / p).resolve() for p in PROTECTED_REPO_DIRS}

    try:
        text = compose_path.read_text()
    except OSError:
        return []

    dirs: List[Path] = []
    in_volumes = False
    volumes_indent = 0

    for raw_line in text.splitlines():
        stripped = raw_line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        indent = len(raw_line) - len(raw_line.lstrip())

        if re.match(r"^volumes:\s*$", stripped):
            in_volumes = True
            volumes_indent = indent
            continue

        if not in_volumes:
            continue

        if not stripped.startswith("- "):
            if indent <= volumes_indent:
                in_volumes = False
            continue

        host = stripped[2:].split(":", 1)[0].strip()
        if "$" in host or not host.startswith("./"):
            continue

        alvo = (service_dir / host[2:]).resolve()
        if alvo in (service_dir.resolve(), repo_root) or alvo in protegidos:
            continue
        if repo_root not in alvo.parents or not alvo.is_dir():
            continue
        dirs.append(alvo)

    return dirs


def collect_data_dirs(compose_files: Dict[str, Path]) -> List[Path]:
    alvos: List[Path] = []
    if state.BASE_DIR.is_dir():
        alvos.append(state.BASE_DIR.resolve())
    for compose_path in compose_files.values():
        alvos.extend(extract_data_dirs_from_compose(compose_path))

    unicos: List[Path] = []
    for alvo in sorted(set(alvos)):
        if any(pai in alvo.parents for pai in unicos):
            continue
        unicos.append(alvo)
    logger.info("Pastas de dados candidatas a remocao (%d): %s",
                len(unicos), ", ".join(str(u) for u in unicos))
    return unicos


def escolher_servicos(compose_files: Dict[str, Path]) -> Dict[str, Path]:
    nomes = sorted(compose_files)
    print_section("Quais serviços instalar?")
    print_info("0) Todos")
    for i, nome in enumerate(nomes, start=1):
        print_info(f"{i}) {nome}")

    escolha = input("\nNúmeros separados por vírgula (vazio = todos): ").strip()
    if not escolha or escolha == "0":
        return compose_files

    selecionados: Dict[str, Path] = {}
    for parte in escolha.split(","):
        parte = parte.strip()
        if not parte.isdigit():
            continue
        idx = int(parte) - 1
        if 0 <= idx < len(nomes):
            nome = nomes[idx]
            selecionados[nome] = compose_files[nome]

    if not selecionados:
        print_warn("Nenhum serviço válido selecionado — instalando todos.")
        return compose_files

    return selecionados


def validate_compose(compose_files: Dict[str, Path]) -> Dict[str, Path]:
    print_section("Validating compose files...")
    failed = []
    for service_name, compose_path in compose_files.items():
        result = run([*compose_base_cmd(), "-f", str(compose_path), "config"], env=compose_env())
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
    cmd = compose_base_cmd()

    if state.DRY_RUN:
        for service_name, compose_path in compose_files.items():
            print_info(f"[dry-run] rodaria: {' '.join(cmd)} -f {compose_path} -p {service_name} up -d")
        return

    failed = []
    for service_name, compose_path in compose_files.items():
        result = run([*cmd, "-f", str(compose_path), "-p", service_name, "up", "-d"], env=compose_env())
        if result.returncode != 0:
            failed.append((service_name, result.stderr.strip()))
        else:
            print_ok(f"{service_name} container(s) started")

    if failed:
        for service_name, error in failed:
            print_fail(f"{service_name}: falha ao iniciar")
            print_info(error)
        raise InstallError(f"Verifique os logs acima e '{' '.join(cmd)} logs' de cada serviço para mais detalhes.")


def stop_services(compose_files: Dict[str, Path]) -> None:
    print_section("Parando serviços...")
    if not compose_files:
        print_warn("Nenhum compose encontrado — nada para parar.")
        return

    cmd = compose_base_cmd()

    if state.DRY_RUN:
        for service_name, compose_path in compose_files.items():
            print_info(f"[dry-run] rodaria: {' '.join(cmd)} -f {compose_path} -p {service_name} down")
        return

    for service_name, compose_path in compose_files.items():
        result = run([*cmd, "-f", str(compose_path), "-p", service_name, "down"], env=compose_env())
        if result.returncode != 0:
            print_warn(f"{service_name}: {result.stderr.strip()}")
        else:
            print_ok(f"{service_name} parado")
