#!/usr/bin/env python3
"""Read-only Python reuse bridge for upstream co-op `harness.GameClient`.

The `GameClient` class is supplied by the pinned upstream checkout; this module
only asks the Go catalog to prepare validated read operations and passes them
to GameClient.ok(). No independent socket client, command parser or engine model.
Do not call against an unsealed production or public TestServer listener.
"""
import json
import subprocess

READ_SET = ('game.inspect_state', 'ui.inspect_screen', 'coop.inspect_checksum')


def read_with_upstream_client(client, catalogcheck, selected=READ_SET):
    """Requires an already-connected upstream GameClient; records plain receipts."""
    receipts = []
    for capability in selected:
        p = subprocess.run([catalogcheck, '-read-id', capability], check=True,
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=10)
        request = json.loads(p.stdout)
        if not isinstance(request, dict) or not isinstance(request.get('cmd'), str):
            raise RuntimeError('catalog returned invalid native request')
        # `ok` executes the existing main-thread TestServer handler.
        response = client.ok(request)
        if not isinstance(response, dict) or response.get('ok') is not True:
            raise RuntimeError('native response not acknowledged')
        receipts.append({'capability': capability, 'native_command': request['cmd'],
                         'response': response, 'evidence_class': 'RESPONSE_FROM_CALLER_NOT_ATTESTED'})
    return receipts
