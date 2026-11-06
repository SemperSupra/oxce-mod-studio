#!/usr/bin/env python3
"""Zero-dependency Linux network-namespace isolation smoke.

Does NOT launch OXCE, correct SDL_net's wildcard bind, or prove native CI safety.
Exercises a real wildcard TCP listener *inside* a new user+network namespace;
asserts it is reachable from its own loopback but not from the caller's namespace.
A blocked namespace operation returns 77, never degrades into an unsafe test.
"""
import argparse
import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import tempfile
import time


def own_ns():
    return os.readlink('/proc/self/ns/net')


def child(path):
    report = Path(path)
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
        listener.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        listener.bind(('0.0.0.0', 0))
        listener.listen(2)
        listener.settimeout(5)
        port = listener.getsockname()[1]
        # Establish that the wildcard listener does work *inside* the namespace.
        with socket.create_connection(('127.0.0.1', port), timeout=2):
            connection, _ = listener.accept()
            connection.close()
        interfaces = json.loads(subprocess.check_output(['ip', '-j', 'link', 'show'], text=True))
        names = [item['ifname'] for item in interfaces]
        routes = subprocess.check_output(['ip','-4','route','show'], text=True).strip()
        # Socket stays alive while the controlling process attempts a connection
        # from the outside network namespace.
        report.write_text(json.dumps({'port': port, 'child_namespace': own_ns(),
            'bound_host': listener.getsockname()[0], 'internal_loopback': True,
            'interfaces': names, 'routes': routes}), encoding='utf-8')
        deadline = time.monotonic() + 6
        while time.monotonic() < deadline:
            time.sleep(0.1)


def parent():
    with tempfile.TemporaryDirectory(prefix='oxce-ns-') as tmp:
        report = Path(tmp) / 'result.json'
        child_proc = subprocess.Popen(['unshare','-Urn', 'sh', '-c',
            'ip link set lo up && exec python3 "$1" --child "$2"',
            'sh', os.path.abspath(__file__), str(report)],
            stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        try:
            until = time.monotonic() + 5
            while time.monotonic() < until and not report.exists():
                if child_proc.poll() is not None:
                    out, err = child_proc.communicate(timeout=2)
                    print(json.dumps({'result':'BLOCKED_NO_USER_NETNS','exit':child_proc.returncode,
                         'stdout':out[:250],'stderr':err[:250]}))
                    return 77
                time.sleep(.05)
            if not report.exists():
                raise RuntimeError('timeout: child isolation verification did not arrive')
            info = json.loads(report.read_text(encoding='utf-8'))
            info['parent_namespace'] = own_ns()
            info['different_namespace'] = info['parent_namespace'] != info['child_namespace']
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as probe:
                probe.settimeout(.4)
                external_rc = probe.connect_ex(('127.0.0.1', info['port']))
            info['parent_socket_connect_result'] = external_rc
            info['parent_cannot_connect'] = external_rc != 0
            info['result'] = 'PASS_ISOLATED_FAKE_WILDCARD_ONLY' if (
                info['different_namespace'] and info['internal_loopback'] and info['parent_cannot_connect'] and info['interfaces'] == ['lo'] and not info['routes']) else 'FAIL_ISOLATION'
            print(json.dumps(info, sort_keys=True))
            return 0 if info['result'].startswith('PASS_') else 1
        finally:
            if child_proc.poll() is None:
                child_proc.terminate()
                try: child_proc.communicate(timeout=2)
                except subprocess.TimeoutExpired:
                    child_proc.kill();child_proc.communicate(timeout=2)

if __name__ == '__main__':
    ap = argparse.ArgumentParser()
    ap.add_argument('--child', help='internal sealed child reporting path')
    args = ap.parse_args()
    sys.exit(child(args.child) or 0 if args.child else parent())
