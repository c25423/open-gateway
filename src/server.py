import logging
from typing import Optional

from dotenv import load_dotenv
from fastapi import FastAPI, Request, Depends
from fastapi.security import HTTPAuthorizationCredentials
from contextlib import asynccontextmanager
import httpx

from auth import validate_token_dependency
from config import load_config
from forwarder import forward_chat_completion

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global HTTP client instance
client: Optional[httpx.AsyncClient] = None

# Load env
load_dotenv()

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
async def list_models(
    credentials: HTTPAuthorizationCredentials = Depends(auth_dependency),
):
    logger.info("Incoming request /models")
    models = []
    for provider_name, models_dict in config["oai"]["model"].items():
        for model_name in models_dict.keys():
            models.append(f"{provider_name}:{model_name}")
    return {"data": [{"id": model} for model in models], "object": "list"}


@app.post("/chat/completions")
async def chat_completions(
    request: Request,
    credentials: HTTPAuthorizationCredentials = Depends(auth_dependency),
):
    logger.info("Incoming request /chat/completions")
    return await forward_chat_completion(request=request, config=config, client=client)


def main():
    import argparse
    import os
    import uvicorn

    # Set up argument parser
    parser = argparse.ArgumentParser(description="Run the Open Gateway API server")

    # Add arguments with environment variable defaults
    parser.add_argument(
        "--host",
        default=os.getenv("HOST", "0.0.0.0"),
        help="Host to bind to (default: HOST env var or 0.0.0.0)",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=int(os.getenv("PORT", "4283")),
        help="Port to bind to (default: PORT env var or 4283)",
    )
    parser.add_argument(
        "--reload",
        action="store_true",
        default=os.getenv("RELOAD", "false").lower() == "true",
        help="Enable auto-reload (default: RELOAD env var or false)",
    )
    parser.add_argument(
        "--log-level",
        default=os.getenv("LOG_LEVEL", "info"),
        choices=["debug", "info", "warning", "error", "critical"],
        help="Log level (default: LOG_LEVEL env var or info)",
    )

    args = parser.parse_args()

    logger.info(f"HOST: {args.host}")
    logger.info(f"PORT: {args.port}")

    # Run the server with the parsed arguments
    uvicorn.run(
        "server:app",
        host=args.host,
        port=args.port,
        reload=args.reload,
        log_level=args.log_level,
    )


if __name__ == "__main__":
    main()
