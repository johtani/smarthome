import json
from pathlib import Path

from dspy_common.resolver_program import (
    ResolverProgram as SharedResolverProgram,
    format_command_catalog,
)
from optimize_and_evaluate import (
    ResolverProgram as BatchResolverProgram,
    to_examples,
    violates_catalog,
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


def test_catalog_violation_detects_invalid_prefix_and_enum():
    catalog = (
        "- start music: random playback [args: mode(optional,enum=artist|genre,no-prefix)]\n"
        "- search and play: targeted playback "
        "[args: keyword(required,no-prefix), type(optional,prefix=type:,enum=artist|album)]"
    )

    assert violates_catalog("start music", "mode:random", catalog) is True
    assert violates_catalog("start music", "random", catalog) is True
    assert violates_catalog("start music", "artist", catalog) is False
    assert violates_catalog("search and play", "Meja type:artist", catalog) is False
    assert violates_catalog("search and play", "Meja type:genre", catalog) is True
    assert violates_catalog("unknown", "", catalog) is True
