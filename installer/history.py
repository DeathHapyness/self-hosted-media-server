import json
import logging
import os
from datetime import datetime
from pathlib import Path
from typing import Optional

from . import state

HISTORY_FILE = Path(__file__).resolve().parent / "json" / "media-server-history.json"
MAX_EXECUCOES = 10000  

logger = logging.getLogger(__name__)

_historico: dict = {"execucoes": []}
_atual: Optional[dict] = None


def _agora() -> str:
    return datetime.now().strftime("%Y-%m-%d %H:%M:%S")


def _carregar() -> dict:
    if not HISTORY_FILE.exists():
        return {"execucoes": []}
    try:
        dados = json.loads(HISTORY_FILE.read_text(encoding="utf-8"))
        if isinstance(dados.get("execucoes"), list):
            return dados
    except (OSError, ValueError, AttributeError):
        pass
    backup = HISTORY_FILE.with_suffix(".json.corrompido")
    try:
        HISTORY_FILE.replace(backup)
        logger.warning("Histórico ilegível; movido para %s e recomeçado", backup)
    except OSError:
        logger.exception("Histórico ilegível e não foi possível movê-lo")
    return {"execucoes": []}


def _salvar() -> None:
    tmp = HISTORY_FILE.with_suffix(".json.tmp")
    try:
        HISTORY_FILE.parent.mkdir(parents=True, exist_ok=True)
        tmp.write_text(json.dumps(_historico, indent=2, ensure_ascii=False), encoding="utf-8")
        os.replace(tmp, HISTORY_FILE)
    except OSError:
        logger.exception("Não foi possível gravar o histórico em %s", HISTORY_FILE)


def iniciar_execucao() -> None:
    global _historico, _atual
    _historico = _carregar()
    _atual = {
        "inicio": _agora(),
        "fim": None,
        "status": "em andamento",
        "dry_run": state.DRY_RUN,
        "eventos": [],
    }
    _historico["execucoes"].append(_atual)
    del _historico["execucoes"][:-MAX_EXECUCOES]
    _salvar()
    logger.info("Histórico: %s", HISTORY_FILE)


def registrar(acao: str, **detalhes) -> None:
    if _atual is None:
        return
    _atual["eventos"].append({"hora": _agora(), "acao": acao, **detalhes})
    _salvar()


def finalizar_execucao(status: str) -> None:
    if _atual is None:
        return
    _atual["fim"] = _agora()
    _atual["status"] = status
    _salvar()
