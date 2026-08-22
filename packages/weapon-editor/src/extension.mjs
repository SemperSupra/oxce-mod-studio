import * as vscode from 'vscode';
import { collectAuthoritativeRuleEvidence, parseCompileObserverJsonl } from './engine-evidence.mjs';
import { editWeaponScalar, readWeaponDocument } from './source-editor.mjs';

let attachedEngineEvidenceText = null;
let attachedEngineEvidenceLabel = null;

function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function minimalChange(before, after) {
  if (before === after) return null;
  let start = 0;
  const maxPrefix = Math.min(before.length, after.length);
  while (start < maxPrefix && before[start] === after[start]) start += 1;

  let beforeEnd = before.length;
  let afterEnd = after.length;
  while (beforeEnd > start && afterEnd > start && before[beforeEnd - 1] === after[afterEnd - 1]) {
    beforeEnd -= 1;
    afterEnd -= 1;
  }
  return { start, beforeEnd, replacement: after.slice(start, afterEnd) };
}

async function applyScalarEdit(document, weaponId, field, value) {
  const before = document.getText();
  const after = editWeaponScalar(before, { weaponId, field, value });
  const change = minimalChange(before, after);
  if (!change) return false;

  const edit = new vscode.WorkspaceEdit();
  edit.replace(
    document.uri,
    new vscode.Range(document.positionAt(change.start), document.positionAt(change.beforeEnd)),
    change.replacement
  );
  const applied = await vscode.workspace.applyEdit(edit);
  if (!applied) throw new Error('VS Code refused the structured weapon edit.');
  return true;
}

function severityName(severity) {
  if (severity === vscode.DiagnosticSeverity.Error) return 'error';
  if (severity === vscode.DiagnosticSeverity.Warning) return 'warning';
  if (severity === vscode.DiagnosticSeverity.Information) return 'information';
  if (severity === vscode.DiagnosticSeverity.Hint) return 'hint';
  return 'unknown';
}

function diagnosticEvidenceOrigin(diagnostic) {
  const source = String(diagnostic.source ?? '').toLowerCase();
  if (source.includes('yaml')) return 'SCHEMA';
  if (source.includes('openxcom') || source.includes('ruleset') || source.includes('oxce')) return 'STATIC-SEMANTIC';
  return 'EDITOR-DIAGNOSTIC';
}

function notRunEngineEvidence() {
  return {
    evidenceOrigin: 'ENGINE-AUTHORITATIVE',
    state: 'not-run',
    message: 'No OXCE engine-authoritative evidence has been attached to this browser session yet.'
  };
}

function collectWeaponEngineEvidence(weapon) {
  if (!weapon || !attachedEngineEvidenceText) return notRunEngineEvidence();
  const evidence = collectAuthoritativeRuleEvidence(attachedEngineEvidenceText, {
    category: 'items',
    identity: weapon.id
  });
  return {
    ...evidence,
    attachedEvidence: attachedEngineEvidenceLabel
  };
}

function collectBrowserEvidence(document, weapon = null) {
  const rulesetTools = vscode.extensions.getExtension('openxcom.ruleset-tools');
  const yaml = vscode.extensions.getExtension('redhat.vscode-yaml');
  const diagnostics = vscode.languages.getDiagnostics(document.uri).map(diagnostic => ({
    evidenceOrigin: diagnosticEvidenceOrigin(diagnostic),
    severity: severityName(diagnostic.severity),
    source: diagnostic.source ?? 'unknown',
    code: typeof diagnostic.code === 'object' ? diagnostic.code?.value ?? null : diagnostic.code ?? null,
    message: diagnostic.message,
    range: {
      start: { line: diagnostic.range.start.line, character: diagnostic.range.start.character },
      end: { line: diagnostic.range.end.line, character: diagnostic.range.end.character }
    }
  }));

  return {
    source: {
      evidenceOrigin: 'SOURCE/TEXT',
      state: 'available'
    },
    providers: {
      rulesetTools: {
        id: 'openxcom.ruleset-tools',
        installed: Boolean(rulesetTools),
        active: Boolean(rulesetTools?.isActive),
        version: rulesetTools?.packageJSON?.version ?? null
      },
      yaml: {
        id: 'redhat.vscode-yaml',
        installed: Boolean(yaml),
        active: Boolean(yaml?.isActive),
        version: yaml?.packageJSON?.version ?? null
      }
    },
    diagnostics,
    engine: collectWeaponEngineEvidence(weapon)
  };
}

function attachEngineEvidenceText(text, label = 'attached JSONL evidence') {
  const events = parseCompileObserverJsonl(text);
  const correlationIds = [...new Set(events.map(event => event.correlationId).filter(Boolean))];
  if (correlationIds.length > 1) {
    throw new Error(`OXCE observer evidence contains multiple correlation IDs: ${correlationIds.join(', ')}.`);
  }
  attachedEngineEvidenceText = text;
  attachedEngineEvidenceLabel = label;
  return {
    evidenceOrigin: 'ENGINE-AUTHORITATIVE',
    state: 'attached',
    label,
    eventCount: events.length,
    correlationId: correlationIds[0] ?? null
  };
}

async function attachEngineEvidenceFile() {
  const picked = await vscode.window.showOpenDialog({
    canSelectMany: false,
    canSelectFiles: true,
    canSelectFolders: false,
    title: 'Attach OXCE compile observer evidence',
    filters: {
      'OXCE observer JSON Lines': ['jsonl', 'json']
    }
  });
  if (!picked?.length) return null;

  const uri = picked[0];
  const bytes = await vscode.workspace.fs.readFile(uri);
  const text = new TextDecoder('utf-8').decode(bytes);
  const result = attachEngineEvidenceText(text, uri.toString());
  const correlated = result.correlationId ? ` · correlation ${result.correlationId}` : '';
  void vscode.window.showInformationMessage(`Attached OXCE engine evidence (${result.eventCount} events${correlated}).`);
  return result;
}

function renderField(field, value) {
  const label = escapeHtml(field);
  if (typeof value === 'number') {
    return `<label class="row"><span>${label}</span><input data-field="${label}" data-kind="number" type="number" value="${escapeHtml(value)}" step="any" /></label>`;
  }
  if (typeof value === 'boolean') {
    return `<label class="row"><span>${label}</span><input data-field="${label}" data-kind="boolean" type="checkbox" ${value ? 'checked' : ''} /></label>`;
  }
  return `<label class="row"><span>${label}</span><input data-field="${label}" data-kind="string" type="text" value="${escapeHtml(value)}" /></label>`;
}

function renderEngineEvidence(engine) {
  if (engine.state === 'not-run') {
    return '<p><strong>ENGINE-AUTHORITATIVE:</strong> not run / no evidence attached.</p>';
  }

  const correlation = engine.correlationId
    ? `<p><strong>Correlation:</strong> <code>${escapeHtml(engine.correlationId)}</code></p>`
    : '<p><strong>Correlation:</strong> not present in this schema-1 trace.</p>';

  if (engine.state === 'not-found') {
    return `<p><strong>ENGINE-AUTHORITATIVE:</strong> attached evidence contains no bounded item snapshot for <code>${escapeHtml(engine.identity)}</code>.</p>${correlation}`;
  }

  const provenance = entry => {
    const source = entry?.source
      ? `<code>${escapeHtml(entry.source)}</code>`
      : escapeHtml(entry?.outcome ?? 'not reported');
    const details = [
      entry?.timestamp ? escapeHtml(entry.timestamp) : null,
      entry?.sequence != null ? `event #${escapeHtml(entry.sequence)}` : null
    ].filter(Boolean).join(' · ');
    return details ? `${source} <span class="note">(${details})</span>` : source;
  };

  return `
    <p><strong>ENGINE-AUTHORITATIVE:</strong> available · OXCE observer schema ${escapeHtml(engine.schema)}</p>
    ${correlation}
    <ul>
      <li>created by: ${provenance(engine.createdBy)}</li>
      <li>effective rule: ${provenance(engine.effectiveRule)}</li>
    </ul>`;
}

function renderDiagnostics(evidence) {
  const providerRows = [evidence.providers.rulesetTools, evidence.providers.yaml]
    .map(provider => `<li><code>${escapeHtml(provider.id)}</code>: ${provider.installed ? `installed${provider.version ? ` ${escapeHtml(provider.version)}` : ''}${provider.active ? ', active' : ', not active'}` : 'not installed'}</li>`)
    .join('');

  const diagnosticRows = evidence.diagnostics.length === 0
    ? '<li>No diagnostics currently reported for this document.</li>'
    : evidence.diagnostics.map(diagnostic => `
      <li>
        <strong>${escapeHtml(diagnostic.evidenceOrigin)}</strong>
        · ${escapeHtml(diagnostic.severity)}
        · ${escapeHtml(diagnostic.source)}
        — ${escapeHtml(diagnostic.message)}
      </li>`).join('');

  return `
    <section class="evidence-block">
      <h2>Evidence</h2>
      <p><strong>SOURCE/TEXT:</strong> authoritative source projection available.</p>
      <p><strong>Providers:</strong></p>
      <ul>${providerRows}</ul>
      <p><strong>Current document diagnostics:</strong></p>
      <ul>${diagnosticRows}</ul>
      ${renderEngineEvidence(evidence.engine)}
    </section>`;
}

function renderWebview(webview, weapon, evidence) {
  const rows = Object.entries(weapon.fields).map(([field, value]) => renderField(field, value)).join('\n');
  const nonce = Math.random().toString(36).slice(2);
  return `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}';" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>OXCE Weapon Inspector</title>
  <style>
    body { font-family: var(--vscode-font-family); color: var(--vscode-foreground); padding: 16px; }
    h1 { font-size: 1.2rem; margin: 0 0 4px; }
    h2 { font-size: 1rem; margin: 0 0 8px; }
    code { font-family: var(--vscode-editor-font-family); }
    .identity, .note { color: var(--vscode-descriptionForeground); font-size: 0.9rem; }
    .identity { margin-bottom: 14px; }
    .evidence-block { margin: 10px 0 16px; padding: 10px; border-left: 2px solid var(--vscode-focusBorder); background: var(--vscode-editorWidget-background); }
    .evidence-block p { margin: 6px 0; }
    .evidence-block ul { margin: 4px 0 8px 20px; padding: 0; }
    .row { display: grid; grid-template-columns: minmax(130px, 1fr) minmax(100px, 1fr); gap: 12px; align-items: center; margin: 8px 0; }
    input { box-sizing: border-box; width: 100%; color: var(--vscode-input-foreground); background: var(--vscode-input-background); border: 1px solid var(--vscode-input-border, transparent); padding: 5px 7px; }
    input[type=checkbox] { width: auto; justify-self: start; }
    #status { min-height: 1.2em; margin-top: 12px; }
  </style>
</head>
<body>
  <h1>${escapeHtml(weapon.id)}</h1>
  <div class="identity"><code>${escapeHtml(weapon.id)}</code> · ${escapeHtml(weapon.kind)} · canonical engine identity</div>
  ${renderDiagnostics(evidence)}
  <p class="note">Edits only existing supported scalar fields in the open .rul file. Raw YAML remains authoritative; unknown fields and surrounding formatting are left untouched.</p>
  ${rows || '<p>No supported existing scalar fields are present.</p>'}
  <div id="status" role="status"></div>
  <script nonce="${nonce}">
    const vscode = acquireVsCodeApi();
    const status = document.getElementById('status');
    document.querySelectorAll('input[data-field]').forEach(input => {
      input.addEventListener('change', () => {
        let value;
        if (input.dataset.kind === 'number') {
          value = Number(input.value);
          if (!Number.isFinite(value)) {
            status.textContent = 'Enter a finite number.';
            return;
          }
        } else if (input.dataset.kind === 'boolean') {
          value = input.checked;
        } else {
          value = input.value;
        }
        status.textContent = 'Applying source-span edit…';
        vscode.postMessage({ type: 'editScalar', field: input.dataset.field, value });
      });
    });
    window.addEventListener('message', event => {
      if (event.data?.type === 'editResult') status.textContent = event.data.message;
    });
  </script>
</body>
</html>`;
}

async function chooseWeapon(document) {
  const { weapons } = readWeaponDocument(document.getText());
  if (weapons.length === 0) throw new Error('No battleType 1 weapon exists in the active ruleset document.');
  if (weapons.length === 1) return weapons[0];
  const picked = await vscode.window.showQuickPick(
    weapons.map(weapon => ({ label: weapon.id, description: `${weapon.kind} · canonical ID`, weapon })),
    { title: 'Choose OXCE weapon' }
  );
  return picked?.weapon ?? null;
}

async function openWeaponEditor() {
  const editor = vscode.window.activeTextEditor;
  if (!editor || !editor.document.uri.path.toLowerCase().endsWith('.rul')) {
    throw new Error('Open an OXCE .rul document before opening the weapon inspector.');
  }

  const weapon = await chooseWeapon(editor.document);
  if (!weapon) return null;
  const panel = vscode.window.createWebviewPanel(
    'oxceModStudio.weaponEditor',
    `OXCE Weapon: ${weapon.id}`,
    vscode.ViewColumn.Beside,
    { enableScripts: true, retainContextWhenHidden: false }
  );
  const evidence = collectBrowserEvidence(editor.document, weapon);
  panel.webview.html = renderWebview(panel.webview, weapon, evidence);

  panel.webview.onDidReceiveMessage(async message => {
    if (message?.type !== 'editScalar') return;
    try {
      await applyScalarEdit(editor.document, weapon.id, message.field, message.value);
      const refreshed = readWeaponDocument(editor.document.getText()).weapons.find(item => item.id === weapon.id);
      if (refreshed) panel.webview.html = renderWebview(panel.webview, refreshed, collectBrowserEvidence(editor.document, refreshed));
      await panel.webview.postMessage({ type: 'editResult', message: 'Applied to authoritative YAML source.' });
    } catch (error) {
      await panel.webview.postMessage({ type: 'editResult', message: String(error.message ?? error) });
    }
  });

  return {
    weaponId: weapon.id,
    kind: weapon.kind,
    evidenceOrigin: weapon.evidenceOrigin,
    fields: weapon.fields,
    browserEvidence: evidence
  };
}

export function activate(context) {
  context.subscriptions.push(
    vscode.commands.registerCommand('oxceModStudio.openWeaponEditor', openWeaponEditor),
    vscode.commands.registerCommand('oxceModStudio.attachEngineEvidence', attachEngineEvidenceFile),
    vscode.commands.registerCommand('oxceModStudio._applyWeaponScalar', async (uri, weaponId, field, value) => {
      const document = await vscode.workspace.openTextDocument(uri);
      return applyScalarEdit(document, weaponId, field, value);
    }),
    vscode.commands.registerCommand('oxceModStudio._inspectBrowserEvidence', async uri => {
      const document = await vscode.workspace.openTextDocument(uri);
      return collectBrowserEvidence(document);
    }),
    vscode.commands.registerCommand('oxceModStudio._inspectEngineEvidence', (jsonl, category, identity) =>
      collectAuthoritativeRuleEvidence(jsonl, { category, identity })
    ),
    vscode.commands.registerCommand('oxceModStudio._attachEngineEvidenceText', (jsonl, label) =>
      attachEngineEvidenceText(jsonl, label)
    )
  );
}

export function deactivate() {}
