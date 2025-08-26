import logging
from typing import Optional

from fastapi import FastAPI, Request, Depends
from fastapi.security import HTTPAuthorizationCredentials
from contextlib import asynccontextmanager
import httpx

from src.config import load_config
from src.forwarder import forward_chat_completion
from src.auth import validate_token_dependency

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global HTTP client instance
client: Optional[httpx.AsyncClient] = None

# Load configuration
config = load_config()

# Create the auth dependency with config
auth_dependency = validate_token_dependency(config)


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup: create the HTTP client
    global client
    client = httpx.AsyncClient()
    yield
    # Shutdown: close the HTTP client
    if client:
        await client.aclose()


# Create FastAPI app with lifespan
app = FastAPI(lifespan=lifespan)


@app.get("/models")
async def list_models(credentials: HTTPAuthorizationCredentials = Depends(auth_dependency)):
    logger.info("Incoming request /models")
    models = []
    for provider_name, models_dict in config["oai"]["model"].items():
        for model_name in models_dict.keys():
            models.append(f"{provider_name}:{model_name}")
    return {"data": [{"id": model} for model in models], "object": "list"}


@app.post("/chat/completions")
async def chat_completions(
    request: Request, credentials: HTTPAuthorizationCredentials = Depends(auth_dependency)
):
    logger.info("Incoming request /chat/completions")
    return await forward_chat_completion(request=request, config=config, client=client)
