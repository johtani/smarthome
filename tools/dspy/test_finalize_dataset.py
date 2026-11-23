import json

import pytest

import finalize_dataset as dataset
from optimize_and_evaluate import read_manifest


def reviewed(i):
    return dict(case_id=str(i), input_text=f"example {i}", group_id=f"group-{i}",
                expected_command="start ps5", expected_args="", source="production_log",
                schema_version="resolver-reviewed/v1", reviewed=True, review_status="accepted",
                reviewed_at="2026-01-01T01:01:00Z")


def invoke(tmp_path, rows):
    source = tmp_path / "reviewed.jsonl"
    source.write_bytes(dataset.jsonl(rows))
    return dataset.main(["--reviewed", str(source), "--config-dir", str(dataset.ROOT / "config"),
                         "--output-dir", str(tmp_path / "result"), "--model", "qwen", "--model", "lfm",
                         "--prompt-version", "v1", "--from", "2026-01-01T00:00:00Z",
                         "--to", "2026-02-01T00:00:00Z"])


@pytest.mark.parametrize("change", [
    dict(reviewed=False), dict(review_status="pending"), dict(expected_args=None),
    dict(feedback_label="incorrect"), dict(input_text="api_key=example"),
    dict(feedback_label="incorrect", review_status="corrected", candidate_command="start ps5"),
    dict(feedback_label="incorrect", feedback_correction=[]),
    dict(expected_command="unknown-command"), dict(expected_command="search and play", expected_args=""),
    dict(expected_command="start music", expected_args="mode:random"),
])
def test_invalid_rows_only_publish_reports(tmp_path, change):
    rows = [reviewed(i) for i in range(4)]
    rows[0].update(change)
    assert invoke(tmp_path, rows) == 1
    folder = tmp_path / "result"
    assert {p.name for p in folder.iterdir()} == {"validation.md", "validation.json"}
    assert json.loads((folder / "validation.json").read_bytes())["errors"]


def test_common_manifest_and_tamper_detection(tmp_path):
    rows = [reviewed(i) for i in range(10)]
    rows[1].update(expected_command="")
    assert invoke(tmp_path, rows) == 0
    path = tmp_path / "result/evaluation-manifest.json"
    manifest, splits, _ = read_manifest(path, "qwen")
    assert read_manifest(path, "lfm")[1] == splits
    assert manifest["splits"]["test"]["unresolved"] == 1
    with pytest.raises(ValueError, match="model"):
        read_manifest(path, "other")
    with pytest.raises(SystemExit):
        invoke(tmp_path, rows)
    original = path.read_bytes()
    manifest["dataset_version"] = "sha256:stale"
    path.write_text(json.dumps(manifest), encoding="utf-8")
    with pytest.raises(ValueError, match="version mismatch"):
        read_manifest(path, "qwen")
    path.write_bytes(original)
    (path.parent / "test.jsonl").write_bytes(b"")
    with pytest.raises(ValueError, match="hash mismatch"):
        read_manifest(path, "lfm")


def test_groups_join_before_deduplication_and_fixture_exclusion():
    rows = [reviewed(i) for i in range(10)]
    rows[1]["input_text"] = "  ＥＸＡＭＰＬＥ   0"
    rows[2]["group_id"] = rows[1]["group_id"]
    rows[3]["input_text"] = "PS5やるぞ"
    aliases = {rows[4]["group_id"]: rows[3]["group_id"]}
    fixtures = dataset.load_fixtures()
    splits, duplicates, overlaps = dataset.split_rows(rows, fixtures, aliases, 42)
    assert (splits, duplicates, overlaps) == dataset.split_rows(rows[::-1], fixtures, aliases, 42)
    assert duplicates == [["1", "0"]]
    assert set(overlaps) == {"3", "4"}
    assert any({"0", "2"} <= {r["case_id"] for r in split} for split in splits.values())
    rows[1]["expected_command"] = "light on"
    with pytest.raises(ValueError, match="conflicting labels"):
        dataset.split_rows(rows, fixtures, aliases, 42)
