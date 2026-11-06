#!/bin/sh
set -eu
cd "$(dirname "$0")"
sh scripts-verify-b41.sh
PYTHONPATH=tools python3 -m unittest -v tools/test_native_gate.py
printf 'B42 SOURCE VERIFICATION PASS; native gameplay not run.\n'
