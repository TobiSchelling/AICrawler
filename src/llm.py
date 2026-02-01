"""LLM provider abstraction for AICrawler."""

import json
import logging
import os
from abc import ABC, abstractmethod

import httpx

logger = logging.getLogger(__name__)


class LLMProvider(ABC):
    """Abstract base class for LLM providers."""

    @abstractmethod
    def generate(self, prompt: str, max_tokens: int = 1024) -> str | None:
        """Generate a response from the LLM."""

    @abstractmethod
    def is_configured(self) -> bool:
        """Check if the provider is properly configured."""


class OllamaProvider(LLMProvider):
    """Ollama local LLM provider."""

    def __init__(self, model: str = "qwen2.5:7b", base_url: str = "http://localhost:11434"):
        self.model = model
        self.base_url = base_url
        self._available: bool | None = None

    def is_configured(self) -> bool:
        if self._available is not None:
            return self._available

        try:
            with httpx.Client(timeout=5.0) as client:
                response = client.get(f"{self.base_url}/api/tags")
                if response.status_code == 200:
                    models = response.json().get("models", [])
                    model_names = [m.get("name", "") for m in models]
                    model_base = self.model.split(":")[0]
                    self._available = any(model_base in name for name in model_names)
                    if not self._available:
                        logger.warning(
                            "Ollama model '%s' not found. Available: %s",
                            self.model,
                            model_names,
                        )
                    return self._available
        except httpx.ConnectError:
            logger.warning("Ollama not running at %s", self.base_url)
            self._available = False
        except Exception as e:
            logger.warning("Error checking Ollama: %s", e)
            self._available = False

        return self._available

    def generate(self, prompt: str, max_tokens: int = 1024) -> str | None:
        try:
            with httpx.Client(timeout=120.0) as client:
                response = client.post(
                    f"{self.base_url}/api/chat",
                    json={
                        "model": self.model,
                        "messages": [{"role": "user", "content": prompt}],
                        "stream": False,
                        "options": {
                            "num_predict": max_tokens,
                            "temperature": 0.3,
                        },
                    },
                )
                response.raise_for_status()
                return response.json()["message"]["content"]
        except Exception as e:
            logger.error("Ollama API error: %s", e)
            return None


class OpenAIProvider(LLMProvider):
    """OpenAI API provider."""

    def __init__(self, model: str = "gpt-4o-mini", api_key_env: str = "OPENAI_API_KEY"):
        self.model = model
        self.api_key = os.environ.get(api_key_env)
        self._client = None

        if self.api_key:
            from openai import OpenAI

            self._client = OpenAI(api_key=self.api_key)
        else:
            logger.debug("OpenAI API key not found in %s", api_key_env)

    def is_configured(self) -> bool:
        return self._client is not None

    def generate(self, prompt: str, max_tokens: int = 1024) -> str | None:
        if not self._client:
            return None

        try:
            response = self._client.chat.completions.create(
                model=self.model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=max_tokens,
                temperature=0.3,
            )
            return response.choices[0].message.content
        except Exception as e:
            logger.error("OpenAI API error: %s", e)
            return None


def create_provider(config: dict) -> LLMProvider | None:
    """Create LLM provider based on configuration."""
    summ_config = config.get("summarization", {})
    provider_name = summ_config.get("provider", "ollama").lower()

    if provider_name == "ollama":
        model = summ_config.get("model", "qwen2.5:7b")
        base_url = summ_config.get("ollama_url", "http://localhost:11434")
        provider = OllamaProvider(model=model, base_url=base_url)
        if provider.is_configured():
            logger.info("Using Ollama with model: %s", model)
            return provider
        logger.info("Ollama not available, trying OpenAI fallback...")
        provider_name = "openai"

    if provider_name == "openai":
        model = summ_config.get("openai_model", summ_config.get("model", "gpt-4o-mini"))
        api_key_env = summ_config.get("api_key_env", "OPENAI_API_KEY")
        provider = OpenAIProvider(model=model, api_key_env=api_key_env)
        if provider.is_configured():
            logger.info("Using OpenAI with model: %s", model)
            return provider

    logger.error("No LLM provider available. Check Ollama is running or set OPENAI_API_KEY.")
    return None


def parse_json_response(text: str) -> dict | None:
    """Parse a JSON response from an LLM, handling markdown code blocks."""
    text = text.strip()

    # Strip markdown code fences
    if text.startswith("```"):
        lines = text.split("\n")
        end_idx = len(lines) - 1
        for i in range(len(lines) - 1, 0, -1):
            if lines[i].strip() == "```":
                end_idx = i
                break
        text = "\n".join(lines[1:end_idx])

    try:
        return json.loads(text)
    except json.JSONDecodeError as e:
        logger.warning("Failed to parse LLM response as JSON: %s", e)
        logger.debug("Response was: %s", text[:200])
        return None
