from pathlib import Path

DRY_RUN = False
CONTAINER_ENGINE = "docker"
COMPOSE_STANDALONE = False

DEFAULT_BASE_DIR = Path("/opt/media-server")
DEFAULT_MEDIA_DIR = Path("/mnt/media")

BASE_DIR = DEFAULT_BASE_DIR
MEDIA_DIR = DEFAULT_MEDIA_DIR
