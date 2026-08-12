import importlib.util
import json
import tempfile
import unittest
from pathlib import Path

MODULE_PATH = Path(__file__).with_name("generate.py")
SPEC = importlib.util.spec_from_file_location("resolver_synthetic", MODULE_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC.loader
SPEC.loader.exec_module(MODULE)


class GenerateTest(unittest.TestCase):
    def config(self):
        return {"groups": [{
            "expected_command": "start ps5",
            "expected_args": "",
            "targets": ["PS5", "プレステ5"],
            "variants": [{"scenario": "normal", "templates": ["{target}を起動して"]}],
        }]}

    def test_generates_stable_unreviewed_groups(self):
        first = MODULE.generate(self.config(), {"start ps5"})
        second = MODULE.generate(self.config(), {"start ps5"})
        self.assertEqual(first, second)
        self.assertEqual(2, len(first))
        self.assertEqual(first[0]["group_id"], first[1]["group_id"])
        self.assertNotEqual(first[0]["case_id"], first[1]["case_id"])
        self.assertEqual("synthetic_template", first[0]["source"])
        self.assertEqual("resolver-synthetic/v1", first[0]["generator"])
        self.assertFalse(first[0]["reviewed"])
        self.assertEqual("unreviewed", first[0]["review_status"])

    def test_rejects_command_outside_catalog(self):
        with self.assertRaisesRegex(ValueError, "not in catalog"):
            MODULE.generate(self.config(), {"help"})

    def test_variant_can_override_expected_target(self):
        config = self.config()
        config["groups"][0]["variants"].append({
            "scenario": "negation",
            "expected_command": "",
            "expected_args": "",
            "templates": ["{target}は起動しないで"],
        })
        rows = MODULE.generate(config, {"start ps5"})
        negatives = [row for row in rows if row["scenario"] == "negation"]
        self.assertEqual(2, len(negatives))
        self.assertTrue(all(row["expected_command"] == "" for row in negatives))
        self.assertEqual(rows[0]["group_id"], negatives[0]["group_id"])

    def test_output_is_reproducible_jsonl(self):
        rows = MODULE.generate(self.config(), {"start ps5"})
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "candidates.jsonl"
            MODULE.write_jsonl(output, rows)
            first = output.read_bytes()
            MODULE.write_jsonl(output, rows)
            self.assertEqual(first, output.read_bytes())
            self.assertEqual(rows, [json.loads(line) for line in output.read_text(encoding="utf-8").splitlines()])


if __name__ == "__main__":
    unittest.main()
