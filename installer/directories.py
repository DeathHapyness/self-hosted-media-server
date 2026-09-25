import logging
import os
import pwd
import re
from pathlib import Path
from typing import Dict, Optional, Tuple

from . import state
from .config import MEDIA_SUBDIRS, WRITABLE_MODE, service_dirs, writable_dirs
from .history import registrar
from .output import print_info, print_ok, print_section, print_warn

logger = logging.getLogger(__name__)


def _repo_root() -> Path:
    return Path(__file__).resolve().parent.parent


def _env_path() -> Path:
    return _repo_root() / ".env"


def _ler_env(path: Path) -> Dict[str, str]:
    if not path.exists():
        return {}
    valores: Dict[str, str] = {}
    for line in path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, value = line.partition("=")
        valores[key.strip()] = value.strip()
    return valores


def _detect_real_user() -> Optional[Tuple[int, int]]:
    sudo_user = os.environ.get("SUDO_USER")
    if not sudo_user:
        return None
    try:
        entry = pwd.getpwnam(sudo_user)
    except KeyError:
        return None
    return entry.pw_uid, entry.pw_gid


def _perguntar_caminho(rotulo: str, padrao: Path) -> Path:
    while True:
        try:
            resposta = input(f"{rotulo} [{padrao}]: ").strip()
        except EOFError:
            print()
            return padrao
        if not resposta:
            return padrao
        caminho = Path(resposta).expanduser()
        if caminho.is_absolute():
            return caminho
        print_warn("Use um caminho absoluto, começando com /.")


def load_env_into_state() -> None:
    valores = _ler_env(_env_path())
    if valores.get("DATA_DIR"):
        state.BASE_DIR = Path(valores["DATA_DIR"])
    if valores.get("MEDIA_DIR"):
        state.MEDIA_DIR = Path(valores["MEDIA_DIR"])


def resolve_puid_pgid() -> Optional[Tuple[int, int]]:
    valores = _ler_env(_env_path())
    try:
        return int(valores["PUID"]), int(valores["PGID"])
    except (KeyError, ValueError):
        return None


def setup_env() -> None:
    env_path = _env_path()
    example_path = _repo_root() / ".env.example"

    if env_path.exists():
        load_env_into_state()
        logger.info("Caminhos lidos do .env: dados=%s midia=%s", state.BASE_DIR, state.MEDIA_DIR)
        registrar("caminhos", origem=".env", dados=str(state.BASE_DIR), midia=str(state.MEDIA_DIR))
        print_ok(".env already present")
        print_info(f"dados/configs: {state.BASE_DIR}")
        print_info(f"biblioteca de mídia: {state.MEDIA_DIR}")
        return

    print_section("Onde este servidor vai guardar os arquivos?")
    print_info("Os padrões servem para a maioria dos casos — Enter aceita cada um.")
    base = _perguntar_caminho("Dados e configs dos serviços", state.DEFAULT_BASE_DIR)
    media = _perguntar_caminho("Biblioteca de mídia", state.DEFAULT_MEDIA_DIR)
    state.BASE_DIR = base
    state.MEDIA_DIR = media
    logger.info("Caminhos escolhidos pelo usuario: dados=%s midia=%s", base, media)
    registrar("caminhos", origem="usuario", dados=str(base), midia=str(media))

    owner = _detect_real_user()
    logger.info("PUID/PGID detectados: %s", owner)

    if state.DRY_RUN:
        print_info(f"[dry-run] criaria .env com DATA_DIR={base} e MEDIA_DIR={media}")
        if owner:
            print_info(f"[dry-run] gravaria PUID={owner[0]} PGID={owner[1]}")
        return

    if example_path.exists():
        conteudo = example_path.read_text()
    else:
        conteudo = "PUID=1000\nPGID=1000\nTZ=America/Sao_Paulo\n"

    substituicoes = {"DATA_DIR": str(base), "MEDIA_DIR": str(media)}
    if owner:
        substituicoes["PUID"] = str(owner[0])
        substituicoes["PGID"] = str(owner[1])

    for chave, valor in substituicoes.items():
        padrao = rf"(?m)^{chave}=.*$"
        if re.search(padrao, conteudo):
            conteudo = re.sub(padrao, f"{chave}={valor}", conteudo)
        else:
            if not conteudo.endswith("\n"):
                conteudo += "\n"
            conteudo += f"{chave}={valor}\n"

    env_path.write_text(conteudo)
    logger.info("Arquivo .env criado em %s", env_path)
    print_ok(f".env created (DATA_DIR={base}, MEDIA_DIR={media})")
    if owner:
        print_ok(f"PUID={owner[0]} PGID={owner[1]} detectados a partir de SUDO_USER")
    else:
        print_warn(
            "SUDO_USER não definido — PUID/PGID ficaram com os valores do exemplo. "
            "Confira com 'id -u' e 'id -g' antes de subir os serviços."
        )
    print_warn("Revise o .env criado e preencha os valores necessários antes de subir os serviços.")


def create_directories() -> None:
    print_section("Creating directories...")
    for service_name, dirs in service_dirs().items():
        created_any = False
        for d in dirs:
            existed = d.exists()
            if state.DRY_RUN:
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
        path = state.MEDIA_DIR / sub
        existed = path.exists()
        if state.DRY_RUN:
            if not existed:
                print_info(f"[dry-run] criaria {path}")
            continue
        path.mkdir(parents=True, exist_ok=True)
        if not existed:
            print_ok(f"{path} created")


def configure_permissions() -> None:
    print_section("Configuring permissions...")
    owner = resolve_puid_pgid()
    alvos = writable_dirs()
    logger.info("Ajustando permissoes de %d diretorios sob %s (owner=%s)", len(alvos), state.BASE_DIR, owner)

    if state.DRY_RUN:
        print_info(f"[dry-run] ajustaria permissões de {len(alvos)} diretórios e de {state.MEDIA_DIR}")
        print_ok("Permissions configured (simulado)")
        return

    for d in alvos:
        if not d.exists():
            continue
        os.chmod(d, WRITABLE_MODE)
        if owner:
            try:
                os.chown(d, owner[0], owner[1])
            except OSError:
                print_warn(f"Não foi possível ajustar dono de {d} para PUID/PGID do .env")

    if state.MEDIA_DIR.exists():
        os.chmod(state.MEDIA_DIR, 0o775)

    downloads_dir = state.MEDIA_DIR / "downloads"
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
