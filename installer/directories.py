import os
import shutil
from pathlib import Path
from typing import Optional, Tuple

from . import state
from .config import MEDIA_DIR, MEDIA_SUBDIRS, SERVICE_DIRS, WRITABLE_DIRS, WRITABLE_MODE
from .output import print_info, print_ok, print_section, print_warn


def create_directories() -> None:
    print_section("Creating directories...")
    for service_name, dirs in SERVICE_DIRS.items():
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
        path = MEDIA_DIR / sub
        existed = path.exists()
        if state.DRY_RUN:
            if not existed:
                print_info(f"[dry-run] criaria /mnt/media/{sub}")
            continue
        path.mkdir(parents=True, exist_ok=True)
        if not existed:
            print_ok(f"/mnt/media/{sub} created")


def resolve_puid_pgid() -> Optional[Tuple[int, int]]:
    repo_root = Path(__file__).resolve().parent.parent
    env_path = repo_root / ".env"
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

    if state.DRY_RUN:
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
    repo_root = Path(__file__).resolve().parent.parent
    env_path = repo_root / ".env"
    example_path = repo_root / ".env.example"

    if env_path.exists():
        print_ok(".env already present")
        return

    if not example_path.exists():
        print_warn(
            "Nenhum .env ou .env.example encontrado. Se algum serviço exigir "
            "credenciais, crie um .env manualmente antes de continuar."
        )
        return

    if state.DRY_RUN:
        print_info("[dry-run] copiaria .env.example para .env")
        return

    shutil.copy(example_path, env_path)
    print_ok(".env created from .env.example")
    print_warn("Revise o .env criado e preencha os valores necessários antes de subir os serviços.")
