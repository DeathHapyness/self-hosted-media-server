import logging
import shutil

from . import state
from .checks import check_disk_space, check_os, check_ports, resolve_media_mount_interactive
from .compose import escolher_servicos, find_compose_files, start_services, stop_services, validate_compose
from .config import BASE_DIR, MEDIA_DIR, REPO_SERVICE_DIRS
from .directories import configure_permissions, create_directories, ensure_env_file
from .engine import check_compose_plugin, check_engine_binary, check_engine_daemon, check_optional, choose_engine
from .history import registrar
from .media import media_mount_menu
from .output import Color, _c, print_fail, print_header, print_info, print_ok, print_warn
from .utils import InstallError, confirm_dangerous, press_enter_to_continue

logger = logging.getLogger(__name__)

OPCOES_MENU = {"1": "instalar", "2": "desinstalar", "3": "disco de mídia", "4": "sair"}


def _interromper_instalacao(motivo: str) -> None:
    print(f"\n{_c('Instalação interrompida.', Color.RED + Color.BOLD)}\n")
    logger.warning("Instalação interrompida: %s", motivo)
    registrar("instalacao", resultado="interrompida", motivo=motivo)


def fluxo_instalar() -> None:
    print_header("INSTALAÇÃO")
    if state.DRY_RUN:
        print_warn("Modo --dry-run ativo: nenhuma alteração real será feita.")

    state.CONTAINER_ENGINE = choose_engine()
    registrar("instalacao", resultado="iniciada", engine=state.CONTAINER_ENGINE)

    try:
        check_os()

        engine_checks = (
            (check_engine_binary, state.CONTAINER_ENGINE),
            (check_compose_plugin, f"{state.CONTAINER_ENGINE} compose"),
            (check_engine_daemon, f"{state.CONTAINER_ENGINE} daemon"),
        )
        for check_fn, label in engine_checks:
            if not check_optional(check_fn, label):
                _interromper_instalacao(f"verificação de {label} falhou")
                return

        if not resolve_media_mount_interactive():
            _interromper_instalacao(f"usuário não quis continuar sem {MEDIA_DIR} montado")
            return

        check_disk_space()

        todos_compose = find_compose_files()
        if not todos_compose:
            raise InstallError(
                "Nenhum arquivo docker-compose.yml/yaml encontrado dentro das pastas de "
                f"serviço ({', '.join(REPO_SERVICE_DIRS)})."
            )
        faltando = [name for name in REPO_SERVICE_DIRS if name not in todos_compose]
        if faltando:
            print_warn(f"Sem compose encontrado para: {', '.join(faltando)} (pulando esses serviços)")
            logger.warning("Compose não encontrado para: %s", ", ".join(faltando))
        compose_files = escolher_servicos(todos_compose)
        registrar("servicos_escolhidos", servicos=list(compose_files))
        check_ports(compose_files)

        create_directories()
        configure_permissions()
        ensure_env_file()

        compose_files = validate_compose(compose_files)
        start_services(compose_files)

    except InstallError as e:
        print_fail(str(e))
        _interromper_instalacao(str(e))
        return

    registrar("instalacao", resultado="concluida", servicos=list(compose_files))
    print_header("INSTALLATION COMPLETE")
    logger.info("Instalação concluída")


def fluxo_desinstalar() -> None:
    print_header("DESINSTALAÇÃO")

    state.CONTAINER_ENGINE = choose_engine()
    registrar("desinstalacao", resultado="iniciada", engine=state.CONTAINER_ENGINE)

    compose_files = find_compose_files()
    stop_services(compose_files)
    registrar("containers_parados", servicos=list(compose_files))

    print_info(f"\n{MEDIA_DIR} NUNCA é tocado por esta opção — seus arquivos de mídia estão seguros.")

    resposta = input(
        f"\nTambém remover as pastas de configuração em {BASE_DIR}? "
        "Isso apaga bancos de dados e configs dos serviços. [s/N]: "
    ).strip().lower()

    if resposta != "s":
        print_info("Configurações mantidas. Apenas os containers foram parados.")
        registrar("desinstalacao", resultado="concluida", configs="mantidas")
        return

    aviso = (
        f"Isso vai apagar PERMANENTEMENTE tudo dentro de {BASE_DIR}: configs, "
        f"bancos de dados (File Browser, Navidrome, etc.) e caches.\n"
        f"    {MEDIA_DIR} continua intocado."
    )
    if not confirm_dangerous(aviso):
        print_info("Operação cancelada — configurações mantidas.")
        registrar("desinstalacao", resultado="concluida", configs="mantidas (confirmação cancelada)")
        return

    if state.DRY_RUN:
        print_info(f"[dry-run] removeria {BASE_DIR}")
        registrar("desinstalacao", resultado="concluida", configs=f"[dry-run] removeria {BASE_DIR}")
        return

    logger.warning("Removendo %s", BASE_DIR)
    shutil.rmtree(BASE_DIR, ignore_errors=True)
    print_ok(f"{BASE_DIR} removido")
    registrar("desinstalacao", resultado="concluida", configs=f"{BASE_DIR} removido")


def main_menu() -> None:
    while True:
        print_header("MEDIA SERVER INSTALLER")
        if state.DRY_RUN:
            print_warn("Modo --dry-run ativo\n")

        print_info("1) Instalar / atualizar serviços")
        print_info("2) Desinstalar serviços")
        print_info("3) Gerenciar disco de mídia (/mnt/media)")
        print_info("4) Sair")

        try:
            escolha = input("\nEscolha uma opção: ").strip()
            logger.info("Opção escolhida: %s", escolha)
        except (EOFError, KeyboardInterrupt):
            print()
            registrar("menu", opcao=None, significado="sair (Ctrl+C)")
            return

        registrar("menu", opcao=escolha, significado=OPCOES_MENU.get(escolha, "inválida"))

        if escolha == "1":
            fluxo_instalar()
            press_enter_to_continue()
        elif escolha == "2":
            fluxo_desinstalar()
            press_enter_to_continue()
        elif escolha == "3":
            media_mount_menu()
        elif escolha == "4":
            print_info("Até mais!")
            return
        else:
            print_fail("Opção inválida.")
