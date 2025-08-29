from pathlib import Path
from typing import Any, Dict


import yaml


def load_config(config_path: str) -> Dict[str, Any]:
    """Load configuration from the specified YAML file"""
    path = Path(config_path)
    if not path.exists():
        raise FileNotFoundError(f"Config file not found: {config_path}")
    with open(path) as f:
        config: Dict[str, Any] = yaml.safe_load(f)
    return config
