#!/usr/bin/env python3
"""Select a PUBLIC synthetic OXCE source build from the whole PR-vs-base diff.

This is a source-build decision only: it never authenticates an engine binary,
licenses retail assets, launches an engine, or approves gameplay.
"""
import argparse
from pathlib import Path
import subprocess
import tempfile
import unittest

ENGINE = 'rdte/oxce-engine-synthetic/engine-source/'
WORKFLOW = '.github/workflows/oxce-public-synthetic-rdte.yml'
SELECTOR = 'rdte/oxce-engine-synthetic/engine_selector.py'


def decide(event, paths):
    if event == 'workflow_dispatch':
        return True, 'EXPLICIT_MANUAL_BUILD'
    if event != 'pull_request':
        raise ValueError('unsupported event: ' + event)
    if any(p.startswith(ENGINE) or p in (WORKFLOW, SELECTOR) for p in paths):
        return True, 'ENGINE_OR_QUALIFICATION_INPUT_CHANGED'
    return False, 'SOURCE_INPUT_UNCHANGED_REPLAY_REQUIRED'


class SourceSelectorTests(unittest.TestCase):
    def test_engine_source_change(self):
        self.assertTrue(decide('pull_request', [ENGINE + 'main.cpp'])[0])

    def test_workflow_contract_change(self):
        self.assertTrue(decide('pull_request', [WORKFLOW])[0])

    def test_selector_script_change(self):
        self.assertTrue(decide('pull_request', [SELECTOR])[0])

    def test_fixture_only_does_not_recompile(self):
        self.assertEqual(decide('pull_request', ['rdte/oxce-engine-synthetic/generate.mjs']),
                         (False, 'SOURCE_INPUT_UNCHANGED_REPLAY_REQUIRED'))

    def test_similar_prefix_not_engine_source(self):
        self.assertFalse(decide('pull_request', [ENGINE[:-1] + '-fake/file'])[0])

    def test_manual_run_forces_build(self):
        self.assertTrue(decide('workflow_dispatch', [])[0])

    def test_unknown_event_fails_closed(self):
        with self.assertRaises(ValueError):
            decide('push', [])

    def test_full_pr_tree_diff_catches_historical_input_change(self):
        with tempfile.TemporaryDirectory() as td:
            repo = Path(td)
            def git(*args):
                return subprocess.check_output(['git', '-C', td, *args], text=True).strip()
            git('init', '-q')
            git('config', 'user.name', 'Fixture')
            git('config', 'user.email', 'fixture@example.invalid')
            (repo / 'README.md').write_text('base\n')
            git('add', '.'); git('commit', '-qm', 'base')
            base = git('rev-parse', 'HEAD')
            d = repo / ENGINE
            d.mkdir(parents=True)
            (d / 'main.cpp').write_text('int main() { return 0; }\n')
            git('add', '.'); git('commit', '-qm', 'engine source')
            (repo / 'README.md').write_text('fixture-only last commit\n')
            git('add', '.'); git('commit', '-qm', 'fixture docs')
            last_only = git('diff-tree', '--no-commit-id', '--name-only', '-r', 'HEAD').splitlines()
            entire_pr = git('diff', '--name-only', '--no-renames', base, 'HEAD').splitlines()
            self.assertFalse(decide('pull_request', last_only)[0])
            self.assertTrue(decide('pull_request', entire_pr)[0])


def cli():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--event', required=True)
    p.add_argument('--paths', type=Path)
    p.add_argument('--output', type=Path, required=True)
    p.add_argument('--summary', type=Path, required=True)
    args = p.parse_args()
    if args.event == 'pull_request' and args.paths is None:
        p.error('--paths required for pull_request; no HEAD-only fallback')
    paths = args.paths.read_text().splitlines() if args.paths is not None else []
    needed, reason = decide(args.event, paths)
    with args.output.open('a', encoding='utf-8') as f:
        f.write('needed=' + str(needed).lower() + '\n')
    with args.summary.open('a', encoding='utf-8') as f:
        f.write('OXCE public synthetic source selector: ' + reason + '\n')
        f.write('Engine build scheduled: ' + str(needed).lower() + '\n')
        if not needed:
            f.write('NO ENGINE EXECUTED; source preflight/immutable replay must qualify separately.\n')
    print(reason + ': needed=' + str(needed).lower())


if __name__ == '__main__':
    import sys
    if sys.argv[1:] == ['--self-test']:
        unittest.main(argv=[sys.argv[0]], verbosity=2)
    else:
        cli()
