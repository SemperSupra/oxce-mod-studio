#!/usr/bin/env python3
"""Select a PUBLIC synthetic OXCE native replay from the whole PR-vs-base diff.

This is a native-job selection only: the cache/build decision is separate. It never authenticates an engine binary,
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
        return True, 'EXPLICIT_NATIVE_REPLAY'
    if event != 'pull_request':
        raise ValueError('unsupported event: ' + event)
    # The native job must still replay *fixture-only* changes after this
    # workflow merges into the default branch. Binary build vs cache reuse is
    # a separate verified decision in the native job.
    if any(p.startswith('rdte/oxce-engine-synthetic/') or p == WORKFLOW for p in paths):
        return True, 'NATIVE_REPLAY_INPUT_CHANGED'
    return False, 'NO_NATIVE_INPUT_CHANGED'


class SourceSelectorTests(unittest.TestCase):
    def test_engine_source_change(self):
        self.assertTrue(decide('pull_request', [ENGINE + 'main.cpp'])[0])

    def test_workflow_contract_change(self):
        self.assertTrue(decide('pull_request', [WORKFLOW])[0])

    def test_selector_script_change(self):
        self.assertTrue(decide('pull_request', [SELECTOR])[0])

    def test_fixture_only_runs_native_loader_without_demanding_recompile(self):
        for name in ('generate.mjs','generate-legacy.mjs','generate-font.mjs','README.md',
                     'classify.mjs','test-generated-master.mjs'):
            with self.subTest(name=name):
                self.assertEqual(
                    decide('pull_request',['rdte/oxce-engine-synthetic/'+name]),
                    (True,'NATIVE_REPLAY_INPUT_CHANGED'))

    def test_unrelated_similar_prefix_is_not_native_input(self):
        self.assertFalse(decide('pull_request', ['rdte/oxce-engine-synthetic-fake/file'])[0])
        self.assertFalse(decide('pull_request', ['docs/unrelated.md'])[0])

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
        f.write('Native loader replay scheduled: ' + str(needed).lower() + '\n')
        if not needed:
            f.write('NATIVE REPLAY NOT SCHEDULED; no matching PR source changes.\n')
    print(reason + ': needed=' + str(needed).lower())


if __name__ == '__main__':
    import sys
    if sys.argv[1:] == ['--self-test']:
        unittest.main(argv=[sys.argv[0]], verbosity=2)
    else:
        cli()
