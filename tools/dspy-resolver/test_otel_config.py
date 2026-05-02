import os
import sys
import unittest
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).parent))

from otel_config import (
    _body_preview_attrs,
    _should_trace_http_body,
    clean_attributes,
    text_trace_attrs,
)


class OtelConfigTests(unittest.TestCase):
    def test_clean_attributes_drops_none_and_stringifies_complex_values(self) -> None:
        attrs = clean_attributes(
            {
                "string": "value",
                "bool": True,
                "int": 1,
                "float": 1.5,
                "none": None,
                "list": ["a", "b"],
            }
        )

        self.assertNotIn("none", attrs)
        self.assertEqual("value", attrs["string"])
        self.assertEqual(True, attrs["bool"])
        self.assertEqual(1, attrs["int"])
        self.assertEqual(1.5, attrs["float"])
        self.assertEqual("['a', 'b']", attrs["list"])

    def test_text_trace_attrs_redacts_content_by_default(self) -> None:
        with patch.dict(os.environ, {"DSPY_RESOLVER_TRACE_INCLUDE_INPUT": ""}):
            attrs = text_trace_attrs("resolver.input", "電気をつけて")

        self.assertEqual(len("電気をつけて"), attrs["resolver.input.length"])
        self.assertEqual(False, attrs["resolver.input.is_empty"])
        self.assertNotIn("resolver.input.preview", attrs)
        self.assertNotIn("resolver.input.truncated", attrs)

    def test_text_trace_attrs_can_include_truncated_preview(self) -> None:
        text = "a" * 300
        with patch.dict(os.environ, {"DSPY_RESOLVER_TRACE_INCLUDE_INPUT": "true"}):
            attrs = text_trace_attrs("resolver.input", text)

        self.assertEqual(300, attrs["resolver.input.length"])
        self.assertEqual("a" * 256, attrs["resolver.input.preview"])
        self.assertEqual(True, attrs["resolver.input.truncated"])

    def test_should_trace_http_body_only_for_lm_api_base(self) -> None:
        with patch.dict(os.environ, {"LM_API_BASE": "http://llm-swap:8080/v1"}):
            self.assertTrue(_should_trace_http_body("http://llm-swap:8080/v1/chat/completions"))
            self.assertFalse(_should_trace_http_body("http://other:8080/v1/chat/completions"))

    def test_body_preview_attrs_redacts_body_by_default(self) -> None:
        with patch.dict(os.environ, {"DSPY_RESOLVER_TRACE_INCLUDE_HTTP_BODY": ""}):
            attrs = _body_preview_attrs("llm.http.request_body", b'{"model":"x"}')

        self.assertEqual(len(b'{"model":"x"}'), attrs["llm.http.request_body.length"])
        self.assertNotIn("llm.http.request_body.preview", attrs)

    def test_body_preview_attrs_can_include_truncated_preview(self) -> None:
        body = "a" * 3000
        with patch.dict(os.environ, {"DSPY_RESOLVER_TRACE_INCLUDE_HTTP_BODY": "true"}):
            attrs = _body_preview_attrs("llm.http.request_body", body)

        self.assertEqual(3000, attrs["llm.http.request_body.length"])
        self.assertEqual("a" * 2048, attrs["llm.http.request_body.preview"])
        self.assertEqual(True, attrs["llm.http.request_body.truncated"])


if __name__ == "__main__":
    unittest.main()
