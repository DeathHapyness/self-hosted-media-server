import shlex
from pathlib import Path
from typing import List, Tuple

from . import state
from .checks import check_media_mount
from .config import MEDIA_DIR
from .history import registrar
from .output import Color, _c, print_fail, print_header, print_info, print_ok, print_section, print_warn
from .utils import confirm_dangerous, press_enter_to_continue, run


def list_unmounted_block_devices() -> List[Tuple[str, str]]:
    # -P gera KEY="valor", então colunas vazias (ex: MOUNTPOINT) não desalinham o resto.
    result = run(["lsblk", "-Pno", "NAME,SIZE,TYPE,MOUNTPOINT,PKNAME"])
    if result.returncode != 0:
        return []

    rows = []
    for line in result.stdout.splitlines():
        rows.append(dict(item.split("=", 1) for item in shlex.split(line)))

    # Um disco com partições (ou uma partição com LVM/cripto dentro) nunca é "livre",
    # mesmo sem mountpoint próprio: pode ser o disco do sistema.
    has_children = {row["PKNAME"] for row in rows if row.get("PKNAME")}

    devices = []
    for row in rows:
        if row.get("TYPE") not in ("part", "disk"):
            continue
        if row.get("MOUNTPOINT") or row["NAME"] in has_children:
            continue
        devices.append((f"/dev/{row['NAME']}", row.get("SIZE", "?")))
    return devices


def mount_real_device_menu() -> None:
    print_section("Discos/partições disponíveis (sem uso no momento):")
    devices = list_unmounted_block_devices()
    if not devices:
        print_warn("Nenhum disco/partição livre foi encontrado (ou 'lsblk' não está disponível).")
        return

    for i, (name, size) in enumerate(devices, start=1):
        print_info(f"{i}) {name} ({size})")

    escolha = input("\nDigite o número do dispositivo (ou vazio para cancelar): ").strip()
    if not escolha:
        print_info("Cancelado.")
        return

    try:
        idx = int(escolha) - 1
        device_path, device_size = devices[idx]
    except (ValueError, IndexError):
        print_fail("Opção inválida.")
        return

    aviso = (
        f"Você está prestes a FORMATAR {device_path} ({device_size}) como ext4.\n"
        f"    TODOS OS DADOS atualmente nesse dispositivo serão APAGADOS PERMANENTEMENTE.\n"
        f"    Confirme que {device_path} é realmente o disco certo antes de continuar."
    )
    if not confirm_dangerous(aviso, palavra=device_path):
        print_info("Operação cancelada — nada foi alterado.")
        registrar("formatar_disco", dispositivo=device_path, resultado="cancelado")
        return

    if state.DRY_RUN:
        print_ok(f"[dry-run] formataria {device_path} como ext4 e montaria em {MEDIA_DIR}")
        registrar("formatar_disco", dispositivo=device_path, resultado="dry-run")
        return

    MEDIA_DIR.mkdir(parents=True, exist_ok=True)

    print_section(f"Formatando {device_path}...")
    result = run(["mkfs.ext4", "-F", device_path])
    if result.returncode != 0:
        print_fail("Falha ao formatar:")
        print_info(result.stderr.strip())
        registrar("formatar_disco", dispositivo=device_path, resultado="falha ao formatar", erro=result.stderr.strip())
        return
    print_ok("Formatado com sucesso")

    result = run(["mount", device_path, str(MEDIA_DIR)])
    if result.returncode != 0:
        print_fail("Falha ao montar:")
        print_info(result.stderr.strip())
        registrar("formatar_disco", dispositivo=device_path, resultado="falha ao montar", erro=result.stderr.strip())
        return
    print_ok(f"{device_path} montado em {MEDIA_DIR}")
    registrar("formatar_disco", dispositivo=device_path, resultado=f"montado em {MEDIA_DIR}")

    _adicionar_fstab(f"{device_path} {MEDIA_DIR} ext4 defaults 0 2")


def create_virtual_disk_menu() -> None:
    print_section("Criar disco virtual (arquivo de imagem) para /mnt/media")
    print_info("Isso ocupa espaço no disco atual — só recomendado para testes ou quando")
    print_info("não há disco físico disponível. Para produção, prefira um disco real.")

    tamanho = input("\nTamanho em GB (padrão 20): ").strip() or "20"
    try:
        tamanho_gb = int(tamanho)
        if tamanho_gb <= 0:
            raise ValueError
    except ValueError:
        print_fail("Tamanho inválido.")
        return

    img_path = Path("/var/media-disk.img")
    aviso = (
        f"Será criado um arquivo de {tamanho_gb}GB em {img_path} e montado em {MEDIA_DIR}.\n"
        f"    Se já existir um disco virtual anterior nesse caminho, ele será sobrescrito."
    )
    if not confirm_dangerous(aviso):
        print_info("Operação cancelada — nada foi alterado.")
        registrar("criar_disco_virtual", tamanho_gb=tamanho_gb, resultado="cancelado")
        return

    if state.DRY_RUN:
        print_ok(f"[dry-run] criaria {img_path} ({tamanho_gb}GB) e montaria em {MEDIA_DIR}")
        registrar("criar_disco_virtual", tamanho_gb=tamanho_gb, resultado="dry-run")
        return

    MEDIA_DIR.mkdir(parents=True, exist_ok=True)

    print_section("Criando arquivo de disco virtual...")
    result = run(["fallocate", "-l", f"{tamanho_gb}G", str(img_path)])
    if result.returncode != 0:
        result = run(["dd", "if=/dev/zero", f"of={img_path}", "bs=1M", f"count={tamanho_gb * 1024}"])
        if result.returncode != 0:
            print_fail("Falha ao criar o arquivo de disco virtual:")
            print_info(result.stderr.strip())
            registrar("criar_disco_virtual", tamanho_gb=tamanho_gb, resultado="falha ao criar arquivo",
                      erro=result.stderr.strip())
            return
    print_ok(f"Arquivo criado ({tamanho_gb}GB)")

    result = run(["mkfs.ext4", "-F", str(img_path)])
    if result.returncode != 0:
        print_fail("Falha ao formatar:")
        print_info(result.stderr.strip())
        registrar("criar_disco_virtual", tamanho_gb=tamanho_gb, resultado="falha ao formatar", erro=result.stderr.strip())
        return
    print_ok("Formatado como ext4")

    result = run(["mount", "-o", "loop", str(img_path), str(MEDIA_DIR)])
    if result.returncode != 0:
        print_fail("Falha ao montar:")
        print_info(result.stderr.strip())
        registrar("criar_disco_virtual", tamanho_gb=tamanho_gb, resultado="falha ao montar", erro=result.stderr.strip())
        return
    print_ok(f"Montado em {MEDIA_DIR}")
    registrar("criar_disco_virtual", tamanho_gb=tamanho_gb, resultado=f"montado em {MEDIA_DIR}")

    _adicionar_fstab(f"{img_path} {MEDIA_DIR} ext4 loop 0 0")


def _adicionar_fstab(linha: str) -> None:
    fstab = Path("/etc/fstab")
    conteudo = fstab.read_text() if fstab.exists() else ""
    if linha in conteudo:
        print_ok("/etc/fstab já contém essa entrada")
        return
    resposta = input("\nAdicionar entrada em /etc/fstab para sobreviver a um reboot? [s/N]: ").strip().lower()
    if resposta != "s":
        print_info("Pulado — o mount não sobrevive a um reboot até você adicionar manualmente.")
        registrar("fstab", linha=linha, resultado="pulado")
        return
    with fstab.open("a") as f:
        f.write(f"\n{linha}\n")
    print_ok("/etc/fstab atualizado")
    registrar("fstab", linha=linha, resultado="adicionada")


def media_mount_menu() -> None:
    while True:
        print_header("GESTÃO DO DISCO DE MÍDIA")
        montado = check_media_mount(raise_on_fail=False)
        status = _c("montado", Color.GREEN) if montado else _c("NÃO montado", Color.RED)
        print(f"Status atual de {MEDIA_DIR}: {status}\n")

        print_info("1) Já montei manualmente — só verificar")
        print_info("2) Usar um disco/partição existente (formata e monta — APAGA DADOS do disco escolhido)")
        print_info("3) Criar um disco virtual (arquivo, menos arriscado, ocupa espaço do disco atual)")
        print_info("4) Voltar ao menu principal")

        escolha = input("\nEscolha uma opção: ").strip()
        if escolha == "1":
            check_media_mount(raise_on_fail=False)
            press_enter_to_continue()
        elif escolha == "2":
            mount_real_device_menu()
            press_enter_to_continue()
        elif escolha == "3":
            create_virtual_disk_menu()
            press_enter_to_continue()
        elif escolha == "4":
            return
        else:
            print_fail("Opção inválida.")
