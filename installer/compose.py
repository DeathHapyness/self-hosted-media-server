import re
from pathlib import Path
from typing import Dict

from . import state
from .config import COMPOSE_CANDIDATES, REPO_SERVICE_DIRS
from .engine import compose_base_cmd
from .output import print_fail, print_info, print_ok, print_section, print_warn
from .utils import InstallError, run


def find_compose_files() -> Dict[str, Path]:
    repo_root = Path(__file__).resolve().parent.parent
    found: Dict[str, Path] = {}
    for service_dir_name in REPO_SERVICE_DIRS:
        service_dir = repo_root / service_dir_name
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
        result = run([*compose_base_cmd(), "-f", str(compose_path), "config"])
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
        result = run([*cmd, "-f", str(compose_path), "-p", service_name, "up", "-d"])
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
        result = run([*cmd, "-f", str(compose_path), "-p", service_name, "down"])
        if result.returncode != 0:
            print_warn(f"{service_name}: {result.stderr.strip()}")
        else:
            print_ok(f"{service_name} parado")
