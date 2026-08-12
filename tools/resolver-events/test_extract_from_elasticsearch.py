import importlib.util
import json
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("extract-from-elasticsearch.py")
SPEC = importlib.util.spec_from_file_location("extract_from_elasticsearch", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC and SPEC.loader
sys.modules[SPEC.name] = MODULE
SPEC.loader.exec_module(MODULE)


def attr(key, value):
    kind = "boolValue" if isinstance(value, bool) else "stringValue"
    return {"key": key, "value": {kind: value}}


class ExtractFromElasticsearchTest(unittest.TestCase):
    def fixture(self):
        return {
            "resourceSpans": [{
                "resource": {"attributes": [attr("service.name", "smarthome")]},
                "scopeSpans": [{"spans": [{
                    "traceId": "trace-1",
                    "attributes": [attr("resolver.request_id", "req-1")],
                    "events": [
                        {"name": "dspy.request", "timeUnixNano": "10", "attributes": [
                            attr("dspy.request.utterance.preview", "照明をつけて\nください"),
                            attr("dspy.request.utterance.truncated", True),
                            attr("dspy.lm.model", "test-model"),
                            attr("resolver.artifact_version", "artifact-1"),
                        ]},
                        {"name": "resolver.decision", "timeUnixNano": "20", "attributes": [
                            attr("resolver.request_id", "req-1"),
                            attr("resolver.resolved_command", "light on"),
                            attr("resolver.resolved_args", "room=living"),
                            attr("resolver.prompt_version", "prompt-1"),
                            attr("resolver.dataset_version", "dataset-1"),
                            attr("resolver.schema_version", "v2"),
                        ]},
                        {"name": "resolver.feedback", "timeUnixNano": "30", "attributes": [
                            attr("resolver.request_id", "req-1"), attr("feedback.label", "correct")
                        ]},
                    ],
                }]}],
            }]
        }

    def test_normalizes_joined_events_and_preserves_multiline_input(self):
        events = list(MODULE.events_from_source(self.fixture()))
        rows = MODULE.normalize(events)
        self.assertEqual(1, len(rows))
        row = rows[0]
        self.assertEqual("req-1", row.group_id)
        self.assertEqual("照明をつけて\nください", row.input_text)
        self.assertEqual("dspy.request.utterance.preview", row.input_text_source)
        self.assertTrue(row.input_text_truncated)
        self.assertEqual("light on", row.candidate_command)
        self.assertEqual("test-model", row.llm_model)
        self.assertEqual("correct", row.feedback_label)

    def test_input_priority_prefers_full_dspy_text(self):
        events = list(MODULE.events_from_source(self.fixture()))
        events[0]["attrs"]["dspy.request.text"] = "完全な入力"
        row = MODULE.normalize(events)[0]
        self.assertEqual("完全な入力", row.input_text)
        self.assertEqual("dspy.request.text", row.input_text_source)
        self.assertFalse(row.input_text_truncated)

    def test_extracts_input_from_llm_body(self):
        event = {"trace_id": "t", "service_name": "svc", "timestamp": "1", "name": "llm.request", "attrs": {
            "llm.request_body.preview": json.dumps({"messages": [{"role": "user", "content": "再生して"}], "text": "再生して"}, ensure_ascii=False),
            "llm.request_body.truncated": False,
        }}
        row = MODULE.normalize([event])[0]
        self.assertEqual("再生して", row.input_text)
        self.assertEqual("llm.request_body.preview", row.input_text_source)

    def test_normalizes_elasticsearch_span_document_without_events(self):
        source = {
            "@timestamp": "2026-08-09T03:50:10Z",
            "trace_id": "trace-es",
            "resource": {"attributes": {"service.name": "smarthome"}},
            "name": "dspy.Resolve",
            "attributes": {
                "dspy.request.text": "TM Networkを再生して",
                "dspy.request.prompt_version": "v3",
                "dspy.response_body": json.dumps({
                    "command": "search and play", "args": "keyword:TM Network",
                    "model": "model-1", "artifact_version": "artifact-2", "dataset_version": "dataset-2",
                }),
            },
        }
        rows = MODULE.normalize(MODULE.events_from_source(source))
        self.assertEqual(1, len(rows))
        self.assertEqual("TM Networkを再生して", rows[0].input_text)
        self.assertEqual("search and play", rows[0].candidate_command)
        self.assertEqual("keyword:TM Network", rows[0].candidate_args)
        self.assertEqual("model-1", rows[0].llm_model)

    def test_infers_feedback_from_elasticsearch_span_attributes(self):
        source = {
            "@timestamp": "2026-08-09T03:50:10Z", "trace_id": "feedback-trace",
            "resource": {"attributes": {"service.name": "smarthome"}},
            "name": "ResolverFeedback.Record",
            "attributes": {"resolver.request_id": "req-1", "feedback.label": "correct", "resolver.resolved_command": "search and play"},
        }
        event = list(MODULE.events_from_source(source))[0]
        self.assertEqual("resolver.feedback", event["name"])
        self.assertEqual("correct", MODULE.normalize([event])[0].feedback_label)

    def test_writes_jsonl_csv_and_audits(self):
        rows = MODULE.normalize(MODULE.events_from_source(self.fixture()))
        meta = {"from": "2026-01-01T00:00:00Z", "to": "2026-01-02T00:00:00Z", "index": "traces-*", "service": "smarthome", "page_size": 100, "documents": 1, "events": 3}
        report = MODULE.audit(rows, meta)
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            MODULE.write_outputs(output, rows, report)
            jsonl_row = json.loads((output / "review-candidates.jsonl").read_text(encoding="utf-8"))
            self.assertEqual("照明をつけて\nください", jsonl_row["input_text"])
            self.assertTrue((output / "review-candidates.csv").exists())
            self.assertEqual(0, json.loads((output / "audit.json").read_text(encoding="utf-8"))["missing"]["llm_model"])
            self.assertTrue((output / "audit.md").exists())

    def test_elasticsearch_uses_pit_and_search_after(self):
        client = object.__new__(MODULE.ElasticsearchClient)
        client.index = "traces-*"
        calls = []

        def request(path, body=None, method="POST"):
            calls.append((path, body, method))
            if path.endswith("/_pit?keep_alive=2m&expand_wildcards=all"):
                return {"id": "pit-1"}
            if method == "DELETE":
                return {"succeeded": True}
            if "search_after" not in body:
                return {"pit_id": "pit-2", "hits": {"hits": [
                    {"_source": {"@timestamp": "2026-01-01T00:00:00Z"}, "sort": ["2026-01-01T00:00:00Z", 1]}
                ]}}
            return {"hits": {"hits": []}}

        client.request = request
        hits = list(client.hits("2026-01-01T00:00:00Z", "2026-01-02T00:00:00Z", "smarthome", 1))
        self.assertEqual(1, len(hits))
        self.assertEqual({"term": {"service.name": "smarthome"}}, calls[1][1]["query"]["bool"]["filter"][1])
        self.assertEqual(["2026-01-01T00:00:00Z", 1], calls[2][1]["search_after"])
        self.assertEqual(("/_pit", {"id": "pit-2"}, "DELETE"), calls[-1])


if __name__ == "__main__":
    unittest.main()
