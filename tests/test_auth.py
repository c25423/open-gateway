import os

from dotenv import load_dotenv
import httpx

from src.config import load_config

load_dotenv()
config = load_config()

base_url = f"http://{os.getenv("HOST")}:{os.getenv("PORT")}"
# base_url = "https://opengateway.nanigawarui.com"
token = config["auth"]["tokens"][0]


def test_models_endpoint():
    """Test the models endpoint with authentication"""
    url = f"{base_url}/models"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }
    
    try:
        response = httpx.get(url, headers=headers, timeout=30.0)
        print(f"Models endpoint status: {response.status_code}")
        if response.status_code == 200:
            result = response.json()
            print("Available models:")
            for model in result.get("data", []):
                print(f"  - {model['id']}")
        else:
            print(f"Error: {response.status_code}")
            print(response.text)
        return response
    except Exception as e:
        print(f"Error: {e}")
        return None


def test_models_endpoint_no_auth():
    """Test the models endpoint without authentication"""
    url = f"{base_url}/models"
    
    try:
        response = httpx.get(url, timeout=30.0)
        print(f"Models endpoint (no auth) status: {response.status_code}")
        print(response.text)
        return response
    except Exception as e:
        print(f"Error: {e}")
        return None


def test_models_endpoint_invalid_token():
    """Test the models endpoint with an invalid token"""
    url = f"{base_url}/models"
    headers = {
        "Authorization": "Bearer invalid_token"
    }
    
    try:
        response = httpx.get(url, headers=headers, timeout=30.0)
        print(f"Models endpoint (invalid token) status: {response.status_code}")
        print(response.text)
        return response
    except Exception as e:
        print(f"Error: {e}")
        return None


if __name__ == "__main__":
    print("Testing authentication...")
    print("\n1. Testing with valid token:")
    test_models_endpoint()
    
    print("\n2. Testing without token:")
    test_models_endpoint_no_auth()
    
    print("\n3. Testing with invalid token:")
    test_models_endpoint_invalid_token()