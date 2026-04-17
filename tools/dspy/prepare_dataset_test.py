import csv
import tempfile
import unittest
from pathlib import Path

from prepare_dataset import build_dataset_rows, split_correction_target


class PrepareDatasetTest(unittest.TestCase):
    def test_split_correction_target_prefers_longest_command_match(self) -> None:
        cmd, args = split_correction_target(
            "search and play 宇多田ヒカル",
            {"search", "search and play", "light on"},
        )
        self.assertEqual(cmd, "search and play")
        self.assertEqual(args, "宇多田ヒカル")

    def test_build_dataset_rows_uses_feedback_correction_for_unresolved(self) -> None:
        rows = [
            {
                "resolver_request_id": "req-known",
                "event_name": "resolver.decision",
                "resolver_resolved_command": "search and play",
                "resolver_resolved_args": "",
                "feedback_label": "",
                "feedback_correction": "",
                "input_text": "known sample",
            },
            {
                "resolver_request_id": "req-unresolved",
                "event_name": "resolver.decision",
                "resolver_resolved_command": "",
                "resolver_resolved_args": "",
                "feedback_label": "",
                "feedback_correction": "",
                "input_text": "ヒッキー再生して",
            },
            {
                "resolver_request_id": "req-unresolved",
                "event_name": "resolver.feedback",
                "resolver_resolved_command": "",
                "resolver_resolved_args": "",
                "feedback_label": "incorrect",
                "feedback_correction": "search and play 宇多田ヒカル",
                "input_text": "",
            },
        ]

        with tempfile.TemporaryDirectory() as tmpdir:
            csv_path = Path(tmpdir) / "resolver-events.csv"
            with csv_path.open("w", encoding="utf-8", newline="") as f:
                writer = csv.DictWriter(f, fieldnames=list(rows[0].keys()))
                writer.writeheader()
                writer.writerows(rows)

            dataset = build_dataset_rows(csv_path)

        by_request = {row["request_id"]: row for row in dataset}
        unresolved = by_request["req-unresolved"]
        self.assertEqual(unresolved["expected_command"], "search and play")
        self.assertEqual(unresolved["expected_args"], "宇多田ヒカル")
        self.assertEqual(unresolved["feedback_label"], "incorrect")


if __name__ == "__main__":
    unittest.main()
