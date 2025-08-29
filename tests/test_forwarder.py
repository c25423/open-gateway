import json
import os

from dotenv import load_dotenv
import httpx

from src.config import load_config

load_dotenv()
config = load_config()

base_url = f"http://{os.getenv("HOST")}:{os.getenv("PORT")}"
# base_url = "https://opengateway.nanigawarui.com"
token = config["auth"]["tokens"][0]


def test_chat_completion(model: str):
    # API endpoint
    url = f"{base_url}/chat/completions"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }
    body = {
        "model": model,
        "messages": [{"role": "user", "content": "Hi"}],
    }

    try:
        # Send the request using httpx
        response = httpx.post(url, headers=headers, json=body, timeout=30.0)

        # Check if request was successful
        response.raise_for_status()
        # Parse response
        result = response.json()
        # print("Response received:")
        print(json.dumps(result, indent=2))
        # Extract and print response
        if "choices" in result and len(result["choices"]) > 0:
            assistant_reasoning = result["choices"][0]["message"].get("reasoning", None)
            print(f"Reasoning: {assistant_reasoning}")
            assistant_message = result["choices"][0]["message"]["content"]
            print(f"Answer: {assistant_message}")

        return result

    except httpx.RequestError as e:
        print(f"Request error: {e}")
        return None
    except httpx.HTTPStatusError as e:
        print(f"HTTP error: {e}")
        print(f"Response: {e.response.text}")
        return None
    except Exception as e:
        print(f"Unexpected error: {e}")
        return None


if __name__ == "__main__":
    # print("Testing openrouter:glm-4.5")
    # test_chat_completion(model="openrouter:glm-4.5")
    # print("Testing openrouter:glm-4.5:thinking")
    # test_chat_completion(model="openrouter:glm-4.5:thinking")
    # print("Testing siliconflow:glm-4.5v")
    # test_chat_completion(model="siliconflow:glm-4.5v")
    print("Testing openrouter:kimi-k2")
    test_chat_completion(model="openrouter:kimi-k2")