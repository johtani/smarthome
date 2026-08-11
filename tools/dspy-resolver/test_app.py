"""Tests for parse_command_list and build_catalog_text in app.py."""

from __future__ import annotations

import pytest

from app import (
    MUSIC_COMMAND_ROUTING_POLICY,
    CommandEntry,
    ResolveResponse,
    build_catalog_text,
    parse_command_list,
)


COMMAND_LIST_WITH_ARGS = """\
  search and play : Search Music by keyword And play
    args: keyword(required), type(optional,prefix=type:,enum=artist|album|track|genre)
  start music : Play music randomly by playlist, artist, or genre
    args: mode(optional,enum=artist|genre)
  light on : turn on the light
"""

COMMAND_LIST_NO_ARGS = """\
  light on : turn on the light
  light off : turn off the light
"""

COMMAND_LIST_WITH_SHORTNAME = """\
  search and play [search play]: Search Music by keyword And play
    args: keyword(required), type(optional,prefix=type:,enum=artist|album|track|genre)
"""


class TestParseCommandList:
    def test_command_with_args_captures_args_text(self):
        entries = parse_command_list(COMMAND_LIST_WITH_ARGS)
        search_play = next(e for e in entries if e.name == "search and play")
        assert search_play.args_text == "keyword(required), type(optional,prefix=type:,enum=artist|album|track|genre)"

    def test_command_with_args_optional_mode(self):
        entries = parse_command_list(COMMAND_LIST_WITH_ARGS)
        start_music = next(e for e in entries if e.name == "start music")
        assert start_music.args_text == "mode(optional,enum=artist|genre)"

    def test_command_without_args_has_empty_args_text(self):
        entries = parse_command_list(COMMAND_LIST_WITH_ARGS)
        light_on = next(e for e in entries if e.name == "light on")
        assert light_on.args_text == ""

    def test_parses_all_commands(self):
        entries = parse_command_list(COMMAND_LIST_WITH_ARGS)
        names = [e.name for e in entries]
        assert names == ["search and play", "start music", "light on"]

    def test_no_args_commands(self):
        entries = parse_command_list(COMMAND_LIST_NO_ARGS)
        assert len(entries) == 2
        assert all(e.args_text == "" for e in entries)

    def test_shortname_stripped_from_command_name(self):
        entries = parse_command_list(COMMAND_LIST_WITH_SHORTNAME)
        assert len(entries) == 1
        assert entries[0].name == "search and play"
        assert entries[0].args_text == "keyword(required), type(optional,prefix=type:,enum=artist|album|track|genre)"

    def test_empty_input_returns_empty(self):
        assert parse_command_list("") == []

    def test_args_line_without_preceding_command_is_ignored(self):
        command_list = "    args: orphan(required)\n  light on : turn on the light\n"
        entries = parse_command_list(command_list)
        assert len(entries) == 1
        assert entries[0].args_text == ""


class TestBuildCatalogText:
    def test_entry_with_args_text_includes_args(self):
        entries = [
            CommandEntry(
                name="search and play",
                description="Search Music by keyword And play",
                args_text="keyword(required), type(optional,prefix=type:,enum=artist|album|track|genre)",
            )
        ]
        result = build_catalog_text(entries)
        assert result == (
            "- search and play: Search Music by keyword And play"
            " [args: keyword(required,no-prefix), type(optional,prefix=type:,enum=artist|album|track|genre)]"
        )

    def test_entry_without_args_text_excludes_args(self):
        entries = [CommandEntry(name="light on", description="turn on the light")]
        result = build_catalog_text(entries)
        assert result == "- light on: turn on the light"

    def test_mixed_entries(self):
        entries = [
            CommandEntry(
                name="search and play",
                description="Search Music by keyword And play",
                args_text="keyword(required)",
            ),
            CommandEntry(name="light on", description="turn on the light"),
        ]
        lines = build_catalog_text(entries).splitlines()
        assert lines[0] == "- search and play: Search Music by keyword And play [args: keyword(required,no-prefix)]"
        assert lines[1] == "- light on: turn on the light"

    def test_empty_entries_returns_empty_string(self):
        assert build_catalog_text([]) == ""


def test_music_command_routing_policy_distinguishes_search_from_random_playback():
    assert "select 'search and play'" in MUSIC_COMMAND_ROUTING_POLICY
    assert "Select 'start music' only" in MUSIC_COMMAND_ROUTING_POLICY
    assert "random-playback grouping modes" in MUSIC_COMMAND_ROUTING_POLICY


def test_resolve_response_includes_model_and_version_metadata():
    response = ResolveResponse(
        command="start ps5",
        args="",
        thought="PS5 is explicit",
        model="lfm2.5-2.6B",
        prompt_version="prompt-v2",
        artifact_version="artifact-v3",
        dataset_version="dataset-v4",
    )

    assert response.model == "lfm2.5-2.6B"
    assert response.prompt_version == "prompt-v2"
    assert response.artifact_version == "artifact-v3"
    assert response.dataset_version == "dataset-v4"
