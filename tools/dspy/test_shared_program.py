import json
from pathlib import Path
from types import SimpleNamespace

from dspy_common.resolver_program import (
    ResolverProgram as SharedResolverProgram,
    format_command_catalog,
)
from optimize_and_evaluate import (
    ResolverProgram as BatchResolverProgram,
    to_examples,
    catalog_errors,
    evaluate,
)


def test_optimization_imports_shared_resolver_program():
    assert BatchResolverProgram is SharedResolverProgram


def test_training_examples_match_shared_signature_fields():
    fixture_path = Path(__file__).with_name("regression_cases.jsonl")
    rows = [json.loads(line) for line in fixture_path.read_text(encoding="utf-8").splitlines() if line]
    examples = to_examples(
        rows,
        "- start ps5: Actions before starting PS5",
        "regression-v1",
    )

    example = examples[0]
    assert example.utterance == "PS5やるぞ"
    assert example.selected_command == "start ps5"
    assert example.selected_args == ""
    assert set(example.inputs().keys()) == {"utterance", "command_catalog", "prompt_version"}


def test_offline_catalog_uses_production_prompt_format():
    catalog = format_command_catalog(
        [
            {
                "name": "start ps5",
                "description": "Actions before starting PS5",
                "args": "",
            }
        ]
    )

    assert catalog == "- start ps5: Actions before starting PS5"


def test_catalog_errors_use_frozen_schema_and_go_validation():
    catalog = json.loads(Path(__file__).with_name("command_catalog.sample.json").read_bytes())
    pairs = [
        ("search and play", "type:artist"),
        ("search and play", "Meja type:artist"),
        ("search and play", "Meja type:video"),
        ("search and play", "keyword:Meja"),
        ("start music", "artist"),
        ("start music", "random"),
        ("unknown", ""),
        ("", ""),
        ("", "artist"),
    ]
    labels = [{"command": command, "args": args} for command, args in pairs]
    assert catalog_errors(catalog, labels) == [
        "missing_required_argument", "", "invalid_enum_value", "invalid_prefix",
        "", "invalid_enum_value", "unknown_command", "", "invalid_unresolved",
    ]


def test_evaluation_counts_missing_required_keyword():
    catalog = json.loads(Path(__file__).with_name("command_catalog.sample.json").read_bytes())
    rows = [{"input_text": "artist", "expected_command": "search and play", "expected_args": "artist"}]

    def predict(**kwargs):
        return SimpleNamespace(selected_command="search and play", selected_args="type:artist")

    assert evaluate(predict, rows, catalog, "test")["catalog_violation_rate"] == 1.0
