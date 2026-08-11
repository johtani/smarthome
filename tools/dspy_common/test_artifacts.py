import json

from dspy_common import artifacts


class FakeProgram:
    loaded_path = ""

    def load(self, path):
        self.loaded_path = path


def write_artifact(tmp_path, **overrides):
    manifest = {
        "program_schema_version": artifacts.PROGRAM_SCHEMA_VERSION,
        "model": "openai/lfm2.5-2.6B",
        "prompt_version": "prompt-v2",
        "artifact_version": "lfm-v2",
        "dataset_version": "dataset-v3",
        "gate_passed": True,
        **overrides,
    }
    (tmp_path / artifacts.MANIFEST_FILE).write_text(json.dumps(manifest), encoding="utf-8")
    (tmp_path / artifacts.PROGRAM_FILE).write_text("{}", encoding="utf-8")


def test_loads_compatible_artifact(tmp_path, monkeypatch):
    write_artifact(tmp_path)
    monkeypatch.setattr(artifacts, "ResolverProgram", FakeProgram)

    result = artifacts.load_artifact(
        tmp_path,
        expected_model="openai/lfm2.5-2.6B",
        expected_prompt_version="prompt-v2",
    )

    assert result.loaded is True
    assert result.artifact_version == "lfm-v2"
    assert result.dataset_version == "dataset-v3"
    assert result.program.loaded_path.endswith(artifacts.PROGRAM_FILE)


def test_model_mismatch_falls_back(tmp_path, monkeypatch):
    write_artifact(tmp_path)
    monkeypatch.setattr(artifacts, "ResolverProgram", FakeProgram)

    result = artifacts.load_artifact(
        tmp_path,
        expected_model="openai/qwen2.5:14b",
        expected_prompt_version="prompt-v2",
    )

    assert result.loaded is False
    assert result.error == "model mismatch"


def test_schema_mismatch_falls_back(tmp_path, monkeypatch):
    write_artifact(tmp_path, program_schema_version="old-schema")
    monkeypatch.setattr(artifacts, "ResolverProgram", FakeProgram)

    result = artifacts.load_artifact(
        tmp_path,
        expected_model="openai/lfm2.5-2.6B",
        expected_prompt_version="prompt-v2",
    )

    assert result.loaded is False
    assert result.error == "program schema version mismatch"


def test_prompt_mismatch_falls_back(tmp_path, monkeypatch):
    write_artifact(tmp_path)
    monkeypatch.setattr(artifacts, "ResolverProgram", FakeProgram)

    result = artifacts.load_artifact(
        tmp_path,
        expected_model="openai/lfm2.5-2.6B",
        expected_prompt_version="prompt-v3",
    )

    assert result.loaded is False
    assert result.error == "prompt version mismatch"


def test_corrupt_manifest_falls_back(tmp_path, monkeypatch):
    (tmp_path / artifacts.MANIFEST_FILE).write_text("not-json", encoding="utf-8")
    monkeypatch.setattr(artifacts, "ResolverProgram", FakeProgram)

    result = artifacts.load_artifact(
        tmp_path,
        expected_model="openai/lfm2.5-2.6B",
        expected_prompt_version="prompt-v2",
    )

    assert result.loaded is False
    assert result.error


def test_gate_failure_falls_back(tmp_path, monkeypatch):
    write_artifact(tmp_path, gate_passed=False)
    monkeypatch.setattr(artifacts, "ResolverProgram", FakeProgram)

    result = artifacts.load_artifact(
        tmp_path,
        expected_model="openai/lfm2.5-2.6B",
        expected_prompt_version="prompt-v2",
    )

    assert result.loaded is False
    assert "evaluation gate" in result.error
