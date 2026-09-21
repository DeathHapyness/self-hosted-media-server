#!/usr/bin/env python3
"""
install.py - Instalador do Home Lab / Self-Hosted Media Server

Menu único que reúne: instalação, desinstalação e gestão do disco de mídia
(/mnt/media). A lógica fica em installer/, pacote ao lado deste arquivo.

Uso:
    sudo python3 install.py
    sudo python3 install.py --dry-run   # simula sem executar nada

Requisitos:
    - Linux
    - Python 3.8+
    - Docker ou Podman, com o respectivo plugin/ferramenta de compose
      (só necessário para instalar serviços; o instalador pergunta qual usar)
    - Sem dependências externas (somente biblioteca padrão)
"""

import argparse
import sys

from installer import state
from installer.checks import check_privileges
from installer.flows import main_menu
from installer.output import print_fail, print_warn
from installer.utils import InstallError

ASCII_ART = r"""  ___           _        _           _              ____       _  __       _   _           _           _
 |_ _|_ __  ___| |_ __ _| | __ _  __| | ___  _ __  / ___|  ___| |/ _|     | | | | ___  ___| |_ ___  __| |
 | || '_ \/ __| __/ _` | |/ _` |/ _` |/ _ \| '__| \___ \ / _ \ | |_ _____| |_| |/ _ \/ __| __/ _ \/ _` |
 | || | | \__ \ || (_| | | (_| | (_| | (_) | |     ___) |  __/ |  _|_____|  _  | (_) \__ \ ||  __/ (_| |
|___|_| |_|___/\__\__,_|_|\__,_|\__,_|\___/|_|    |____/ \___|_|_|       |_| |_|\___/|___/\__\___|\__,_|
"""

ADVERTENCIA = """\033[1;31mADVERTÊNCIA:\033[0m Este script é fornecido "como está" e não se responsabiliza por quaisquer danos ou perda de dados. Use por sua própria conta e risco."""


def main() -> int:
    parser = argparse.ArgumentParser(description="Instalador do Home Lab / Self-Hosted Media Server")
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simula todas as ações sem executar nada de fato (nenhum arquivo criado, nenhum comando destrutivo rodado).",
    )
    args = parser.parse_args()
    state.DRY_RUN = args.dry_run

    print(ASCII_ART)
    print(ADVERTENCIA)

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
