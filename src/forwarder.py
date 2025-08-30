import logging
from typing import Any, AsyncGenerator, Dict, List, Optional, Union

from fastapi import Request, HTTPException
from fastapi.responses import StreamingResponse
import httpx

logger = logging.getLogger(__name__)


async def forward_chat_completion(
    request: Request,
    config: Dict[str, Any],
    client: httpx.AsyncClient
) -> Union[StreamingResponse, Dict[str, Any]]:
    try:
        logger.info(f"Incoming request URL: {str(request.url)}")

        # Get incoming headers and body
        incoming_headers: Dict[str, str] = dict(request.headers)
        logger.info(f"Incoming request headers: {incoming_headers}")
        exclude_headers: List[str] = [
            "accept",
            "accept-encoding",
            "authorization",
            "host",
            "content-length",
            "content-type",
        ]
        for key in exclude_headers:
            if key in incoming_headers:
                del incoming_headers[key]
        try:
            incoming_body: Dict[str, Any] = await request.json()
        except Exception as e:
            logger.error(f"Invalid request body: {type(e).__name__}: {str(e)}")
            raise HTTPException(
                status_code=400, detail=f"Invalid request body: {str(e)}"
            )
        logger.info(f"Incoming request body: {incoming_body}")

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
            **model_config.get("extra_headers", {}),
            **incoming_headers,
            "accept": "*/*",
            "authorization": f"Bearer {provider_config['api_key']}",
            "content-type": "application/json"
        }
        body: Dict[str, Any] = {
            **model_config.get("extra_body", {}),
            **incoming_body,
            "model": model_config["identifier"],
        }

        # Send outgoing request using shared client
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
                        # logger.info(f"Reveived from url: {url}, model: {model_config['identifier']}: {chunk}")
                        yield chunk

            try:
                return StreamingResponse(
                    stream_generator(), media_type="text/event-stream"
                )
            except Exception as e:
                logger.error(f"Streaming error: {type(e).__name__}: {str(e)}")
                raise HTTPException(
                    status_code=500, detail=f"Streaming error: {str(e)}"
                )
        else:
            try:
                response: httpx.Response = await client.post(
                    url, headers=headers, json=body, timeout=60.0
                )
                response.raise_for_status()
                return response.json()
            except Exception as e:
                logger.error(f"Non-streaming error: {type(e).__name__}: {str(e)}")
                raise HTTPException(
                    status_code=500, detail=f"Non-streaming error: {str(e)}"
                )
    except httpx.HTTPStatusError as e:
        logger.error(f"Provider error: {e.response.status_code} - {e.response.text}")
        raise HTTPException(
            status_code=e.response.status_code,
            detail=f"Provider error: {e.response.text}",
        )
    except Exception as e:
        logger.error(f"Internal error: {type(e).__name__}: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail="Internal server error")