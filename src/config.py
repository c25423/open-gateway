from pathlib import Path
from typing import Any, Dict


import yaml


def load_config() -> Dict[str, Any]:
    """Load configuration from config.yaml"""
    config_path = Path("config.yaml")
    if not config_path.exists():
        raise FileNotFoundError("config.yaml not found")
    with open(config_path) as f:
        config: Dict[str, Any] = yaml.safe_load(f)
    return config
