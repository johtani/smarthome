from types import SimpleNamespace

import pytest

from compare_resolvers import load_splits, summarize
from optimize_and_evaluate import metric


@pytest.mark.parametrize("expected, predicted, accepted", [
    (("start ps5", ""), ("start ps5", ""), True),
    (("start ps5", ""), ("stop music", ""), False),
    (("search and play", "Meja"), ("search and play", "AC/DC"), False),
    (("", ""), ("", "Meja"), False),
    (("", ""), ("", ""), True),
])
def test_bootstrap_example_acceptance(expected, predicted, accepted):
    def prediction(pair):
        return SimpleNamespace(selected_command=pair[0], selected_args=pair[1])
    assert metric(prediction(expected), prediction(predicted)) is accepted


def test_splits_disallow_leaked_group(tmp_path):
    path = tmp_path / "split.jsonl"
    path.write_text('{"group_id":"g","input_text":"hello","expected_command":"","expected_args":""}')
    with pytest.raises(ValueError, match="group shared across splits"):
        load_splits({"train": path, "test": path})


def test_scores_use_matching_pair_and_unsupported_denominator():
    def row(cmd, args, raw_cmd, raw_args):
        return dict(expected_command=cmd, expected_args=args, decisions={
            "legacy": {"Raw": {"command": raw_cmd, "args": raw_args}}})
    rows = [row("", "", "start ps5", ""), row("", "", "", ""),
            row("search and play", "Meja", "search and play", "")]
    assert summarize(rows, "legacy", "Raw") == dict(total=3, exact=1, accuracy=1/3,
                                                     false_positives=1, unsupported=2, unresolved=1, invalid_output=0)
    rows[0]["decisions"]["legacy"]["Raw"].update(command="", error="invalid args")
    summary = summarize(rows, "legacy", "Raw")
    assert (summary["exact"], summary["unresolved"], summary["invalid_output"]) == (1, 1, 1)
