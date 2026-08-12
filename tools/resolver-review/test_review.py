import importlib.util
import json
import tempfile
import threading
import unittest
import urllib.error
import urllib.request
from pathlib import Path

SCRIPT = Path(__file__).with_name("review.py")
SPEC = importlib.util.spec_from_file_location("resolver_review", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC and SPEC.loader
SPEC.loader.exec_module(MODULE)


class ReviewStoreTest(unittest.TestCase):
    def candidates(self):
        return [
            {"case_id": "a", "group_id": "g1", "source": "fixture", "input_text": "音楽", "candidate_command": "start music", "candidate_args": ""},
            {"case_id": "b", "group_id": "g2", "source": "synthetic", "input_text": "音楽", "candidate_command": "bad", "candidate_args": "x", "feedback_label": "incorrect"},
        ]

    def catalog(self):
        return [{"name": "start music", "description": "", "args": ""}, {"name": "stop music", "description": "", "args": ""}]

    def test_accept_saves_candidate_as_human_confirmed_expected_value(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "reviewed.jsonl"
            store = MODULE.ReviewStore(self.candidates(), self.catalog(), output)
            store.update("a", {"review_status": "accepted", "expected_command": "start music", "expected_args": "", "review_note": "確認済み"})
            row = json.loads(output.read_text(encoding="utf-8").splitlines()[0])
            self.assertTrue(row["reviewed"])
            self.assertEqual("accepted", row["review_status"])
            self.assertEqual("start music", row["expected_command"])
            self.assertTrue(row["reviewed_at"])

    def test_excluded_and_pending_rows_never_keep_expected_values(self):
        with tempfile.TemporaryDirectory() as directory:
            store = MODULE.ReviewStore(self.candidates(), self.catalog(), Path(directory) / "reviewed.jsonl")
            store.update("a", {"review_status": "excluded", "expected_command": "start music", "expected_args": "bad", "review_note": ""})
            self.assertEqual("", store.reviews["a"]["expected_command"])
            self.assertEqual("", store.reviews["a"]["expected_args"])

    def test_rejects_expected_command_outside_catalog(self):
        with tempfile.TemporaryDirectory() as directory:
            store = MODULE.ReviewStore(self.candidates(), self.catalog(), Path(directory) / "reviewed.jsonl")
            with self.assertRaisesRegex(ValueError, "not in catalog"):
                store.update("a", {"review_status": "corrected", "expected_command": "unknown"})

    def test_resume_and_warnings(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "reviewed.jsonl"
            store = MODULE.ReviewStore(self.candidates(), self.catalog(), output)
            store.update("a", {"review_status": "accepted", "expected_command": "start music"})
            resumed = MODULE.ReviewStore(self.candidates(), self.catalog(), output)
            payload = resumed.payload()
            self.assertEqual("accepted", payload["rows"][0]["review_status"])
            self.assertTrue(payload["rows"][0]["warnings"]["duplicate"])
            self.assertTrue(payload["rows"][1]["warnings"]["candidate_catalog_violation"])
            self.assertTrue(payload["rows"][1]["warnings"]["incorrect_without_correction"])

    def test_http_rejects_non_object_payload(self):
        with tempfile.TemporaryDirectory() as directory:
            store = MODULE.ReviewStore(self.candidates(), self.catalog(), Path(directory) / "reviewed.jsonl")
            server = MODULE.HTTPServer(("127.0.0.1", 0), MODULE.handler_for(store, b"test"))
            thread = threading.Thread(target=server.serve_forever)
            thread.start()
            try:
                request = urllib.request.Request(
                    f"http://127.0.0.1:{server.server_port}/api/review",
                    data=b"[]",
                    headers={"Content-Type": "application/json"},
                    method="POST",
                )
                with self.assertRaises(urllib.error.HTTPError) as caught:
                    urllib.request.urlopen(request)
                self.assertEqual(400, caught.exception.code)
            finally:
                server.shutdown()
                server.server_close()
                thread.join()


if __name__ == "__main__":
    unittest.main()
