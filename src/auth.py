from fastapi import Request, HTTPException, Depends
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
import logging

logger = logging.getLogger(__name__)

security = HTTPBearer()


def validate_token(credentials: HTTPAuthorizationCredentials, config: dict) -> bool:
    """
    Validate the bearer token against the tokens in config

    Args:
        credentials: HTTP authorization credentials containing the bearer token
        config: Configuration dictionary containing auth tokens

    Returns:
        bool: True if token is valid

    Raises:
        HTTPException: If token is invalid or missing
    """
    token = credentials.credentials

    # Get valid tokens from config
    valid_tokens = config.get("auth", {}).get("tokens", [])

    # Check if token is in the valid tokens list
    if token in valid_tokens:
        return True
    else:
        logger.warning(f"Invalid token attempted: {token}")
        raise HTTPException(
            status_code=401,
            detail="Invalid authentication credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )


def validate_token_dependency(config: dict):
    async def validate(credentials: HTTPAuthorizationCredentials = Depends(security)):
        validate_token(credentials=credentials, config=config)
        return credentials  # Return credentials if validation passes
    return validate
