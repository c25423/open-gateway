import logging
from pathlib import Path
from typing import Any, AsyncGenerator, Dict, List, Optional, Union

from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import StreamingResponse
import httpx
import yaml

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI()
config_path = Path("config.yaml")
if not config_path.exists():
    raise FileNotFoundError("config.yaml not found")
with open(config_path) as f:
    config: Dict[str, Any] = yaml.safe_load(f)


async def forward_chat_completion(
    request: Request,
) -> Union[StreamingResponse, Dict[str, Any]]:
    try:
        # Get incoming headers and body
        incoming_headers: Dict[str, str] = dict(request.headers)
        exclude_headers: List[str] = [
            "host",
            "content-length",
            "authorization",
            "content-type",
            "accept",
        ]
        for key in exclude_headers:
            if key in incoming_headers:
                del incoming_headers[key]
        try:
            incoming_body: Dict[str, Any] = await request.json()
        except Exception as e:
            raise HTTPException(
                status_code=400, detail=f"Invalid request body: {str(e)}"
            )

        # Get provider config and model config
        model: str = incoming_body.get("model", "")
        if ":" not in model:
            raise HTTPException(
                status_code=400,
                detail="Model must be in format {provider_name}:{model_name}",
            )
        provider_name, model_name = model.split(":", 1)
        provider_config: Optional[Dict[str, Any]] = config["oai"]["provider"].get(
            provider_name, None
        )
        if not provider_config:
            raise HTTPException(
                status_code=404, detail=f"Provider {provider_name} not found"
            )
        model_config: Optional[Dict[str, Any]] = (
            config["oai"]["model"].get(provider_name, {}).get(model_name, None)
        )
        if not model_config:
            raise HTTPException(
                status_code=404,
                detail=f"Model {model_name} not found for provider {provider_name}",
            )
        if not model_config.get("identifier", None):
            raise HTTPException(
                status_code=404,
                detail=f"Identifier not found for model {model_name}",
            )

        # Build outgoing request
        base_url: str = provider_config["base_url"].rstrip("/")
        url: str = f"{base_url}/chat/completions"
        headers: Dict[str, str] = {
            "Authorization": f"Bearer {provider_config['api_key']}",
            **model_config.get("extra_headers", {}),
            **incoming_headers,
        }
        body: Dict[str, Any] = {
            **model_config.get("extra_body", {}),
            **incoming_body,
            "model": model_config["identifier"],
        }

        # Send outgoing request
        client = httpx.AsyncClient()
        stream: bool = incoming_body.get("stream", False)
        logger.info(
            f"Forwarding request to url: {url}, model: {model_config['identifier']}, stream: {stream}"
        )
        if stream:
            async def stream_generator() -> AsyncGenerator[bytes, None]:
                async with client.stream(
                    "POST", url, headers=headers, json=body, timeout=60.0
                ) as response:
                    response.raise_for_status()
                    async for chunk in response.aiter_bytes():
                        yield chunk
                await client.aclose()

            try:
                return StreamingResponse(
                    stream_generator(), media_type="text/event-stream"
                )
            except Exception as e:
                logger.error(f"Error streaming: {e}")
                await client.aclose()
                raise HTTPException(status_code=500, detail="Streaming error")
        else:
            try:
                response: httpx.Response = await client.post(
                    url, headers=headers, json=body, timeout=60.0
                )
                response.raise_for_status()
                return response.json()
            finally:
                await client.aclose()
    except httpx.HTTPStatusError as e:
        logger.error(f"Error from provider: {str(e)}")
        raise HTTPException(status_code=e.response.status_code, detail=str(e))
    except Exception as e:
        logger.error(f"Error forwarding request: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/models")
async def list_models():
    logger.info("Incoming request /models")
    models = []
    for provider_name, models_dict in config["oai"]["model"].items():
        for model_name in models_dict.keys():
            models.append(f"{provider_name}:{model_name}")
    return {"data": [{"id": model} for model in models], "object": "list"}


@app.post("/chat/completions")
async def chat_completions(request: Request):
    logger.info("Incoming request /chat/completions")
    try:
        return await forward_chat_completion(
            request=request,
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error in chat completions: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))
