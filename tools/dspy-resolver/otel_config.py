"""OpenTelemetry helpers for the DSPy resolver."""

from __future__ import annotations

import os
import json
from contextlib import contextmanager
from typing import Any, Dict, Iterator, Mapping, Optional

try:
    from opentelemetry import trace
    from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
    from opentelemetry.instrumentation.requests import RequestsInstrumentor
    from opentelemetry.sdk.resources import SERVICE_NAME, Resource
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from opentelemetry.trace import Status, StatusCode
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
except Exception:  # pragma: no cover - allows local execution without OTEL deps.
    trace = None
    FastAPIInstrumentor = None
    RequestsInstrumentor = None
    SERVICE_NAME = "service.name"
    Resource = None
    TracerProvider = None
    BatchSpanProcessor = None
    Status = None
    StatusCode = None
    OTLPSpanExporter = None

try:
    from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
except Exception:  # pragma: no cover - httpx instrumentation is optional.
    HTTPXClientInstrumentor = None

try:
    import litellm
except Exception:  # pragma: no cover - litellm is provided through dspy.
    litellm = None


TRACER_NAME = "dspy-resolver"
DEFAULT_SERVICE_NAME = "smarthome-dspy-resolver"
_http_clients_instrumented = False
_litellm_instrumented = False


class NoopSpan:
    def add_event(self, _name: str, attributes: Optional[Mapping[str, Any]] = None) -> None:
        return None

    def record_exception(self, _exception: BaseException) -> None:
        return None

    def set_attribute(self, _key: str, _value: Any) -> None:
        return None

    def set_attributes(self, _attributes: Mapping[str, Any]) -> None:
        return None

    def set_status(self, _status: Any) -> None:
        return None


def _is_otel_disabled() -> bool:
    return os.getenv("OTEL_SDK_DISABLED", "").strip().lower() in {"true", "1", "yes", "y"}


def _has_otlp_endpoint() -> bool:
    return bool(
        os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "").strip()
        or os.getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "").strip()
    )


def _drop_empty_otel_endpoint_env() -> None:
    for key in ("OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"):
        if key in os.environ and not os.environ[key].strip():
            del os.environ[key]


def _otel_provider_is_configurable() -> bool:
    if trace is None:
        return False
    provider = trace.get_tracer_provider()
    return provider.__class__.__name__ == "ProxyTracerProvider"


def clean_attributes(attrs: Optional[Mapping[str, Any]]) -> Dict[str, Any]:
    cleaned: Dict[str, Any] = {}
    if not attrs:
        return cleaned
    for key, value in attrs.items():
        if value is None:
            continue
        if isinstance(value, (str, bool, int, float)):
            cleaned[key] = value
            continue
        cleaned[key] = str(value)
    return cleaned


def text_trace_attrs(prefix: str, value: str, max_preview_chars: int = 256) -> Dict[str, Any]:
    text = value or ""
    attrs: Dict[str, Any] = {
        f"{prefix}.length": len(text),
        f"{prefix}.is_empty": len(text.strip()) == 0,
    }
    include_input = os.getenv("DSPY_RESOLVER_TRACE_INCLUDE_INPUT", "").strip().lower() in {
        "true",
        "1",
        "yes",
        "y",
    }
    if include_input:
        attrs[f"{prefix}.preview"] = text[:max_preview_chars]
        attrs[f"{prefix}.truncated"] = len(text) > max_preview_chars
    return attrs


def _include_http_body() -> bool:
    return os.getenv("DSPY_RESOLVER_TRACE_INCLUDE_HTTP_BODY", "").strip().lower() in {
        "true",
        "1",
        "yes",
        "y",
    }


def _should_trace_http_body(url: Any) -> bool:
    api_base = os.getenv("LM_API_BASE", "").strip()
    if not api_base:
        return False
    return str(url).startswith(api_base)


def _body_preview_attrs(prefix: str, body: Any, max_preview_chars: int = 2048) -> Dict[str, Any]:
    if body is None:
        return {f"{prefix}.length": 0}
    if isinstance(body, bytes):
        raw = body
        text = raw.decode("utf-8", errors="replace")
    elif isinstance(body, str):
        raw = body.encode("utf-8")
        text = body
    else:
        text = str(body)
        raw = text.encode("utf-8")

    attrs: Dict[str, Any] = {f"{prefix}.length": len(raw)}
    if _include_http_body():
        attrs[f"{prefix}.preview"] = text[:max_preview_chars]
        attrs[f"{prefix}.truncated"] = len(text) > max_preview_chars
    return attrs


def _redact_sensitive(value: Any) -> Any:
    if isinstance(value, Mapping):
        redacted: Dict[str, Any] = {}
        for key, item in value.items():
            key_text = str(key)
            if key_text.lower() in {"api_key", "authorization", "token", "access_token", "secret_key"}:
                redacted[key_text] = "[redacted]"
            else:
                redacted[key_text] = _redact_sensitive(item)
        return redacted
    if isinstance(value, list):
        return [_redact_sensitive(item) for item in value]
    if isinstance(value, tuple):
        return [_redact_sensitive(item) for item in value]
    return value


def _json_payload_preview_attrs(prefix: str, payload: Any, max_preview_chars: int = 4096) -> Dict[str, Any]:
    try:
        body = json.dumps(_redact_sensitive(payload), ensure_ascii=False, sort_keys=True, default=str)
    except Exception:
        body = str(payload)
    return _body_preview_attrs(prefix, body, max_preview_chars)


def _litellm_input_callback(kwargs: Mapping[str, Any]) -> None:
    if trace is None or not _include_http_body():
        return

    span = trace.get_current_span()
    additional_args = kwargs.get("additional_args", {})
    complete_input = None
    if isinstance(additional_args, Mapping):
        complete_input = additional_args.get("complete_input_dict")
    payload = complete_input if complete_input is not None else kwargs
    span.add_event(
        "llm.request",
        attributes=clean_attributes(
            {
                "llm.request.source": "litellm.input_callback",
                **_json_payload_preview_attrs("llm.request_body", payload),
            }
        ),
    )


def _requests_request_hook(span: Any, request: Any) -> None:
    if span is None or not _should_trace_http_body(getattr(request, "url", "")):
        return
    span.set_attributes(clean_attributes(_body_preview_attrs("llm.http.request_body", getattr(request, "body", None))))


def _httpx_request_hook(span: Any, request: Any) -> None:
    if span is None or not _should_trace_http_body(getattr(request, "url", "")):
        return
    try:
        body = getattr(request, "content", None)
    except Exception:
        body = None
    span.set_attributes(clean_attributes(_body_preview_attrs("llm.http.request_body", body)))


def setup_otel(service_name: str = DEFAULT_SERVICE_NAME) -> None:
    if _is_otel_disabled() or trace is None:
        return

    _drop_empty_otel_endpoint_env()
    instrument_http_clients()
    instrument_litellm()
    if not _has_otlp_endpoint():
        return
    if not _otel_provider_is_configurable():
        return
    if Resource is None or TracerProvider is None or BatchSpanProcessor is None or OTLPSpanExporter is None:
        return

    resolved_service_name = os.getenv("OTEL_SERVICE_NAME", "").strip() or service_name
    resource = Resource.create({SERVICE_NAME: resolved_service_name})
    provider = TracerProvider(resource=resource)
    provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter()))
    trace.set_tracer_provider(provider)


def instrument_http_clients() -> None:
    global _http_clients_instrumented
    if _http_clients_instrumented:
        return
    if RequestsInstrumentor is not None:
        RequestsInstrumentor().instrument(request_hook=_requests_request_hook)
    if HTTPXClientInstrumentor is not None:
        HTTPXClientInstrumentor().instrument(request_hook=_httpx_request_hook)
    _http_clients_instrumented = True


def instrument_litellm() -> None:
    global _litellm_instrumented
    if _litellm_instrumented or litellm is None:
        return

    callbacks = list(getattr(litellm, "input_callback", []) or [])
    if _litellm_input_callback not in callbacks:
        callbacks.append(_litellm_input_callback)
        litellm.input_callback = callbacks
    _litellm_instrumented = True


def instrument_fastapi_app(app: Any) -> None:
    if FastAPIInstrumentor is None or _is_otel_disabled():
        return
    FastAPIInstrumentor.instrument_app(app)


@contextmanager
def start_span(name: str, attributes: Optional[Mapping[str, Any]] = None) -> Iterator[Any]:
    if trace is None or _is_otel_disabled():
        yield NoopSpan()
        return
    tracer = trace.get_tracer(TRACER_NAME)
    with tracer.start_as_current_span(name, attributes=clean_attributes(attributes)) as span:
        yield span


def add_event(span: Any, name: str, attributes: Optional[Mapping[str, Any]] = None) -> None:
    span.add_event(name, attributes=clean_attributes(attributes))


def set_error(span: Any, err: BaseException | str) -> None:
    message = str(err)
    if isinstance(err, BaseException):
        span.record_exception(err)
    if Status is not None and StatusCode is not None:
        span.set_status(Status(StatusCode.ERROR, message))
    span.set_attribute("error.message", message)
