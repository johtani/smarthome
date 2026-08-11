import unittest

from dspy_common.reporting import build_metadata_breakdown


class ReportingTest(unittest.TestCase):
    def test_build_metadata_breakdown_groups_model_and_prompt_version(self) -> None:
        rows = [
            {
                "llm_model": "qwen",
                "prompt_version": "v1",
                "expected_command": "light on",
                "pred_command": "light on",
                "expected_args": "",
                "pred_args": "",
            },
            {
                "llm_model": "qwen",
                "prompt_version": "v2",
                "expected_command": "start ps5",
                "pred_command": "start music",
                "expected_args": "",
                "pred_args": "mode:random",
            },
            {
                "llm_model": "",
                "prompt_version": "",
                "expected_command": "health",
                "pred_command": "health",
                "expected_args": "",
                "pred_args": "",
            },
        ]

        breakdown = build_metadata_breakdown(rows)

        self.assertEqual(breakdown["by_llm_model"]["qwen"]["count"], 2)
        self.assertEqual(breakdown["by_llm_model"]["qwen"]["command_accuracy"], 0.5)
        self.assertEqual(breakdown["by_prompt_version"]["v2"]["arg_accuracy"], 0.0)
        self.assertEqual(breakdown["by_llm_model"]["(unknown)"]["count"], 1)


if __name__ == "__main__":
    unittest.main()
