import json
import logging
from pathlib import Path

LOG_FILE = Path("/var/log/media-server-installer.log")


class JsonFormatter(logging.Formatter):
    def format(self, record: logging.LogRecord) -> str:
        dados = {
            "hora": self.formatTime(record, "%Y-%m-%d %H:%M:%S"),
            "nivel": record.levelname,
            "modulo": record.name,
            "mensagem": record.getMessage(),
        }
        if record.exc_info:
            dados["erro"] = self.formatException(record.exc_info)
        return json.dumps(dados, ensure_ascii=False)


def setup_logging(log_file: Path = LOG_FILE) -> None:
    handler = logging.FileHandler(log_file, encoding="utf-8")
    handler.setFormatter(JsonFormatter())
    logging.basicConfig(handlers=[handler], level=logging.DEBUG)
