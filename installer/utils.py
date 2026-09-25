import logging
import os
import subprocess
from pathlib import Path

from .output import print_danger, print_info

logger = logging.getLogger(__name__)


class InstallError(Exception):
    """Erro crítico que deve interromper a operação atual imediatamente."""


def run(cmd, **kwargs):
    comando = " ".join(cmd)
    logger.info("Executando: %s", comando)
    result = subprocess.run(cmd, capture_output=True, text=True, check=False, **kwargs)
    if result.returncode != 0:
        logger.error("Falhou (código %d): %s: %s", result.returncode, comando, result.stderr.strip())
    elif result.stdout.strip():
        logger.debug("stdout: %s", result.stdout.strip())
    return result


def confirm_dangerous(explicacao: str, palavra: str = "CONFIRMO") -> bool:
    print_danger(explicacao)
    print_info(f"Isto NÃO pode ser desfeito automaticamente. Digite exatamente \"{palavra}\" para continuar.")
    try:
        resposta = input("> ").strip()
    except (EOFError, KeyboardInterrupt):
        print()
        resposta = ""
    confirmado = resposta == palavra
    if confirmado:
        logger.warning("Ação destrutiva CONFIRMADA pelo usuário: %s", explicacao)
    else:
        logger.info("Ação destrutiva cancelada pelo usuário")
    return confirmado


def tamanho_legivel(path: Path) -> str:
    total = 0
    for raiz, _dirs, arquivos in os.walk(path, onerror=lambda _e: None):
        for nome in arquivos:
            try:
                total += os.path.getsize(os.path.join(raiz, nome))
            except OSError:
                continue
    tamanho = float(total)
    for unidade in ("B", "KB", "MB", "GB", "TB"):
        if tamanho < 1024:
            return f"{tamanho:.0f}{unidade}"
        tamanho /= 1024
    return f"{tamanho:.0f}PB"


def press_enter_to_continue() -> None:
    try:
        input("\nPressione Enter para voltar ao menu...")
    except (EOFError, KeyboardInterrupt):
        print()
