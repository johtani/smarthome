import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from dspy_common.lm_config import build_lm_config


class LmConfigTests(unittest.TestCase):
    def test_prefers_lm_api_key_over_openai_api_key(self) -> None:
        conf = build_lm_config(
            {
                "MODEL": "openai/qwen2.5:14b",
                "LM_API_KEY": "lm-key",
                "OPENAI_API_KEY": "openai-key",
            }
        )
        self.assertEqual("lm-key", conf["kwargs"]["api_key"])
        self.assertEqual("LM_API_KEY", conf["health"]["api_key_source"])

    def test_fallback_to_openai_api_key(self) -> None:
        conf = build_lm_config(
            {
                "MODEL": "openai/qwen2.5:14b",
                "OPENAI_API_KEY": "openai-key",
            }
        )
        self.assertEqual("openai-key", conf["kwargs"]["api_key"])
        self.assertEqual("OPENAI_API_KEY", conf["health"]["api_key_source"])

    def test_cli_api_key_overrides_environment_keys(self) -> None:
        conf = build_lm_config(
            {
                "LM_API_KEY": "lm-key",
                "OPENAI_API_KEY": "openai-key",
            },
            overrides={"api_key": "cli-key"},
        )
        self.assertEqual("cli-key", conf["kwargs"]["api_key"])
        self.assertEqual("cli", conf["health"]["api_key_source"])

    def test_cli_overrides_lm_options(self) -> None:
        conf = build_lm_config(
            {
                "MODEL": "openai/env-model",
                "LM_API_BASE": "http://env.example/v1",
                "LM_MODEL_TYPE": "env-chat",
                "LM_TEMPERATURE": "0.9",
                "LM_MAX_TOKENS": "128",
            },
            overrides={
                "model": "openai/cli-model",
                "api_base": "http://cli.example/v1",
                "model_type": "chat",
                "temperature": 0.25,
                "max_tokens": 512,
            },
        )
        self.assertEqual("openai/cli-model", conf["model"])
        self.assertEqual("http://cli.example/v1", conf["kwargs"]["api_base"])
        self.assertEqual("chat", conf["kwargs"]["model_type"])
        self.assertEqual(0.25, conf["kwargs"]["temperature"])
        self.assertEqual(512, conf["kwargs"]["max_tokens"])

    def test_parses_temperature_and_max_tokens(self) -> None:
        conf = build_lm_config(
            {
                "MODEL": "openai/qwen2.5:14b",
                "LM_TEMPERATURE": "0.25",
                "LM_MAX_TOKENS": "512",
            }
        )
        self.assertEqual(0.25, conf["kwargs"]["temperature"])
        self.assertEqual(512, conf["kwargs"]["max_tokens"])
        self.assertEqual(0.25, conf["health"]["temperature"])
        self.assertEqual(512, conf["health"]["max_tokens"])

    def test_invalid_temperature_raises(self) -> None:
        with self.assertRaises(ValueError):
            build_lm_config({"LM_TEMPERATURE": "abc"})

    def test_invalid_max_tokens_raises(self) -> None:
        with self.assertRaises(ValueError):
            build_lm_config({"LM_MAX_TOKENS": "abc"})


if __name__ == "__main__":
    unittest.main()
