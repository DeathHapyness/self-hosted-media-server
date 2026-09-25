import logging
import sys

logger = logging.getLogger(__name__)


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
    logger.error(text)


def print_warn(text: str) -> None:
    print(f"{_c('[!]', Color.YELLOW)} {text}")


def print_info(text: str) -> None:
    print(f"    {text}")


def print_danger(text: str) -> None:
    print(f"{_c('[PERIGO]', Color.RED + Color.BOLD)} {text}")
