import * as vscode from 'vscode';

function requireCondition(condition, message) {
  if (!condition) throw new Error(message);
}

export async function run() {
  const extension = vscode.extensions.getExtension('sempersupra.oxce-mod-studio');
  requireCondition(extension, 'OXCE Mod Studio development extension was not loaded');
  requireCondition(extension.packageJSON.browser === './dist/extension.js', `unexpected browser entrypoint: ${extension.packageJSON.browser}`);
  requireCondition(extension.packageJSON.capabilities?.virtualWorkspaces === true, 'extension must explicitly support virtual workspaces');
  await extension.activate();
  requireCondition(extension.isActive, 'OXCE Mod Studio failed to activate in VS Code Web');

  const commands = await vscode.commands.getCommands(true);
  requireCondition(commands.includes('oxceModStudio.openWeaponEditor'), 'public weapon inspector command not registered');
  requireCondition(commands.includes('oxceModStudio.attachEngineEvidence'), 'public engine evidence attachment command not registered');
  requireCondition(commands.includes('oxceModStudio._applyWeaponScalar'), 'testable bounded scalar command not registered');
  requireCondition(commands.includes('oxceModStudio._inspectBrowserEvidence'), 'browser evidence adapter command not registered');
  requireCondition(commands.includes('oxceModStudio._inspectEngineEvidence'), 'engine evidence adapter command not registered');
  requireCondition(commands.includes('oxceModStudio._attachEngineEvidenceText'), 'test engine evidence attachment command not registered');

  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  requireCondition(workspaceFolder, 'expected one web-test workspace folder');
  requireCondition(workspaceFolder.uri.scheme === 'vscode-test-web', `expected vscode-test-web workspace, got ${workspaceFolder.uri.scheme}`);

  const uri = vscode.Uri.joinPath(workspaceFolder.uri, 'weapon.rul');
  const document = await vscode.workspace.openTextDocument(uri);
  await vscode.window.showTextDocument(document);
  const before = document.getText();

  const injectedDiagnostics = vscode.languages.createDiagnosticCollection('oxce-mod-studio-test-static');
  const diagnostic = new vscode.Diagnostic(
    new vscode.Range(new vscode.Position(0, 0), new vscode.Position(0, 1)),
    'test static diagnostic from the VS Code diagnostic bus',
    vscode.DiagnosticSeverity.Warning
  );
  diagnostic.source = 'OXCE test static';
  injectedDiagnostics.set(uri, [diagnostic]);

  const beforeEngine = await vscode.commands.executeCommand('oxceModStudio._inspectBrowserEvidence', uri);
  requireCondition(beforeEngine?.engine?.evidenceOrigin === 'ENGINE-AUTHORITATIVE', 'engine evidence class must be explicit before attachment');
  requireCondition(beforeEngine?.engine?.state === 'not-run', 'engine evidence must begin explicitly not-run');

  // This two-event trace is a browser-contract fixture only. The same provenance
  // shape has been proven separately by a real isolated OXCE run; public Web CI
  // must not masquerade this synthetic transport test as that semantic proof.
  const correlationId = 'web-test-correlation-1';
  const engineJsonl = [
    JSON.stringify({schema: 1, timestamp: '2026-08-22T16:47:51.147Z', correlation_id: correlationId, sequence: 38, kind: 'snapshot', phase: 'validate-rulesets', category: 'items', operation: 'created-by', identity: 'STR_PISTOL', source: 'xcom1', outcome: 'present'}),
    JSON.stringify({schema: 1, timestamp: '2026-08-22T16:47:51.147Z', correlation_id: correlationId, sequence: 39, kind: 'snapshot', phase: 'validate-rulesets', category: 'items', operation: 'effective-rule', identity: 'STR_PISTOL', source: 'sempersupra-studio-mvp', outcome: 'present'})
  ].join('\n');

  const directEvidence = await vscode.commands.executeCommand(
    'oxceModStudio._inspectEngineEvidence',
    engineJsonl,
    'items',
    'STR_PISTOL'
  );
  requireCondition(directEvidence?.state === 'available', 'bounded engine evidence parser did not return available evidence');
  requireCondition(directEvidence?.correlationId === correlationId, 'engine correlation ID was not preserved');
  requireCondition(directEvidence?.createdBy?.source === 'xcom1', 'created-by provenance was not preserved');
  requireCondition(directEvidence?.createdBy?.sequence === 38, 'created-by sequence was not preserved');
  requireCondition(directEvidence?.createdBy?.timestamp === '2026-08-22T16:47:51.147Z', 'created-by timestamp was not preserved');
  requireCondition(directEvidence?.effectiveRule?.source === 'sempersupra-studio-mvp', 'effective-rule provenance was not preserved');

  const attached = await vscode.commands.executeCommand(
    'oxceModStudio._attachEngineEvidenceText',
    engineJsonl,
    'synthetic VS Code Web evidence matching the real MVP provenance shape'
  );
  requireCondition(attached?.state === 'attached' && attached?.eventCount === 2, 'engine evidence was not attached to the web session');

  const opened = await vscode.commands.executeCommand('oxceModStudio.openWeaponEditor');
  requireCondition(opened?.weaponId === 'STR_PISTOL', `weapon inspector returned unexpected id: ${opened?.weaponId}`);
  requireCondition(opened?.kind === 'item', `weapon inspector must expose canonical kind, got ${opened?.kind}`);
  requireCondition(opened?.evidenceOrigin === 'SOURCE/TEXT', `weapon inspector must label source evidence, got ${opened?.evidenceOrigin}`);
  requireCondition(opened?.browserEvidence?.source?.evidenceOrigin === 'SOURCE/TEXT', 'browser evidence must identify source/text evidence');
  requireCondition(opened?.browserEvidence?.engine?.state === 'available', 'weapon inspector did not transition to attached authoritative evidence');
  requireCondition(opened?.browserEvidence?.engine?.correlationId === correlationId, 'weapon inspector did not retain engine correlation ID');
  requireCondition(opened?.browserEvidence?.engine?.createdBy?.source === 'xcom1', 'weapon inspector did not surface created-by evidence');
  requireCondition(opened?.browserEvidence?.engine?.effectiveRule?.source === 'sempersupra-studio-mvp', 'weapon inspector did not surface effective-rule evidence');
  requireCondition(opened?.browserEvidence?.providers?.rulesetTools?.id === 'openxcom.ruleset-tools', 'Ruleset Tools provider identity missing');
  requireCondition(opened?.browserEvidence?.providers?.yaml?.id === 'redhat.vscode-yaml', 'Red Hat YAML provider identity missing');
  requireCondition(
    opened?.browserEvidence?.diagnostics?.some(item => item.evidenceOrigin === 'STATIC-SEMANTIC' && item.source === 'OXCE test static'),
    'VS Code diagnostic bus evidence was not surfaced/classified as static semantic evidence'
  );

  const applied = await vscode.commands.executeCommand(
    'oxceModStudio._applyWeaponScalar',
    uri,
    'STR_PISTOL',
    'accuracySnap',
    62
  );
  requireCondition(applied === true, 'bounded source edit was not applied');

  const after = document.getText();
  const expected = before.replace('accuracySnap: 61', 'accuracySnap: 62');
  requireCondition(after === expected, 'VS Code Web edit changed more than the intended scalar source span');
  requireCondition(after.includes('STR_PISTOL_CLIP'), 'unrelated ammo override did not survive');
  requireCondition(after.includes('costBuy: 25'), 'unrelated ammo scalar did not survive');
  requireCondition(!before.includes('battleType:'), 'MVP Web fixture must remain a genuine partial override without repeated battleType');

  injectedDiagnostics.dispose();

  console.log(JSON.stringify({
    result: 'pass',
    extensionId: extension.id,
    extensionVersion: extension.packageJSON.version,
    workspaceScheme: workspaceFolder.uri.scheme,
    weaponId: opened.weaponId,
    partialOverrideWithoutBattleType: true,
    evidence: {
      source: opened.evidenceOrigin,
      staticSemanticClassificationObserved: true,
      engineBeforeAttachment: beforeEngine.engine.state,
      engineAfterAttachment: opened.browserEvidence.engine.state,
      correlationId: opened.browserEvidence.engine.correlationId,
      createdBy: opened.browserEvidence.engine.createdBy.source,
      effectiveRule: opened.browserEvidence.engine.effectiveRule.source,
      browserEvidenceFixture: 'synthetic-transport-contract-only'
    },
    exactSourceSpanEdit: true,
    unrelatedAmmoOverridePreserved: true
  }, null, 2));
}
