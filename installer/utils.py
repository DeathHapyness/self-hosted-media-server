import subprocess

from .output import print_danger, print_info


class InstallError(Exception):
    """Erro crítico que deve interromper a operação atual imediatamente."""


def run(cmd, **kwargs):
    """Executa um comando e retorna CompletedProcess. Nunca levanta por código != 0."""
    return subprocess.run(cmd, capture_output=True, text=True, check=False, **kwargs)


def confirm_dangerous(explicacao: str, palavra: str = "CONFIRMO") -> bool:
    """Exige que o usuário digite uma palavra exata antes de uma ação destrutiva."""
    print_danger(explicacao)
    print_info(f"Isto NÃO pode ser desfeito automaticamente. Digite exatamente \"{palavra}\" para continuar.")
    try:
        resposta = input("> ").strip()
    except (EOFError, KeyboardInterrupt):
        print()
        return False
    return resposta == palavra


def press_enter_to_continue() -> None:
    try:
        input("\nPressione Enter para voltar ao menu...")
    except (EOFError, KeyboardInterrupt):
        print()
