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

  const engineJsonl = [
    JSON.stringify({schema: 1, kind: 'snapshot', phase: 'validate-rulesets', category: 'items', operation: 'created-by', identity: 'STR_WEB_TEST_WEAPON', source: 'Test Core', outcome: 'present'}),
    JSON.stringify({schema: 1, kind: 'snapshot', phase: 'validate-rulesets', category: 'items', operation: 'effective-rule', identity: 'STR_WEB_TEST_WEAPON', source: 'Test Balance Patch', outcome: 'present'})
  ].join('\n');

  const directEvidence = await vscode.commands.executeCommand(
    'oxceModStudio._inspectEngineEvidence',
    engineJsonl,
    'items',
    'STR_WEB_TEST_WEAPON'
  );
  requireCondition(directEvidence?.state === 'available', 'bounded engine evidence parser did not return available evidence');
  requireCondition(directEvidence?.createdBy?.source === 'Test Core', 'created-by provenance was not preserved');
  requireCondition(directEvidence?.effectiveRule?.source === 'Test Balance Patch', 'effective-rule provenance was not preserved');

  const attached = await vscode.commands.executeCommand(
    'oxceModStudio._attachEngineEvidenceText',
    engineJsonl,
    'synthetic VS Code Web evidence'
  );
  requireCondition(attached?.state === 'attached' && attached?.eventCount === 2, 'engine evidence was not attached to the web session');

  const opened = await vscode.commands.executeCommand('oxceModStudio.openWeaponEditor');
  requireCondition(opened?.weaponId === 'STR_WEB_TEST_WEAPON', `weapon inspector returned unexpected id: ${opened?.weaponId}`);
  requireCondition(opened?.kind === 'item', `weapon inspector must expose canonical kind, got ${opened?.kind}`);
  requireCondition(opened?.evidenceOrigin === 'SOURCE/TEXT', `weapon inspector must label source evidence, got ${opened?.evidenceOrigin}`);
  requireCondition(opened?.browserEvidence?.source?.evidenceOrigin === 'SOURCE/TEXT', 'browser evidence must identify source/text evidence');
  requireCondition(opened?.browserEvidence?.engine?.state === 'available', 'weapon inspector did not transition to attached authoritative evidence');
  requireCondition(opened?.browserEvidence?.engine?.createdBy?.source === 'Test Core', 'weapon inspector did not surface created-by evidence');
  requireCondition(opened?.browserEvidence?.engine?.effectiveRule?.source === 'Test Balance Patch', 'weapon inspector did not surface effective-rule evidence');
  requireCondition(opened?.browserEvidence?.providers?.rulesetTools?.id === 'openxcom.ruleset-tools', 'Ruleset Tools provider identity missing');
  requireCondition(opened?.browserEvidence?.providers?.yaml?.id === 'redhat.vscode-yaml', 'Red Hat YAML provider identity missing');
  requireCondition(
    opened?.browserEvidence?.diagnostics?.some(item => item.evidenceOrigin === 'STATIC-SEMANTIC' && item.source === 'OXCE test static'),
    'VS Code diagnostic bus evidence was not surfaced/classified as static semantic evidence'
  );

  const applied = await vscode.commands.executeCommand(
    'oxceModStudio._applyWeaponScalar',
    uri,
    'STR_WEB_TEST_WEAPON',
    'accuracySnap',
    70
  );
  requireCondition(applied === true, 'bounded source edit was not applied');

  const after = document.getText();
  const expected = before.replace('accuracySnap: 65 # preserve me', 'accuracySnap: 70 # preserve me');
  requireCondition(after === expected, 'VS Code Web edit changed more than the intended scalar source span');
  requireCondition(after.includes('futureField: 123 # must survive'), 'unknown field/comment was not preserved');

  injectedDiagnostics.dispose();

  console.log(JSON.stringify({
    result: 'pass',
    extensionId: extension.id,
    extensionVersion: extension.packageJSON.version,
    workspaceScheme: workspaceFolder.uri.scheme,
    weaponId: opened.weaponId,
    evidence: {
      source: opened.evidenceOrigin,
      staticSemanticClassificationObserved: true,
      engineBeforeAttachment: beforeEngine.engine.state,
      engineAfterAttachment: opened.browserEvidence.engine.state,
      createdBy: opened.browserEvidence.engine.createdBy.source,
      effectiveRule: opened.browserEvidence.engine.effectiveRule.source
    },
    exactSourceSpanEdit: true,
    unknownFieldPreserved: true
  }, null, 2));
}
