import * as vscode from 'vscode';

function requireCondition(condition, message) {
  if (!condition) throw new Error(message);
}

async function sleep(ms) {
  await new Promise(resolve => setTimeout(resolve, ms));
}

async function waitForDiagnostics(uri, timeoutMs = 30000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const diagnostics = vscode.languages.getDiagnostics(uri);
    if (diagnostics.length > 0) return diagnostics;
    await sleep(250);
  }
  return vscode.languages.getDiagnostics(uri);
}

export async function run() {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  requireCondition(workspaceFolder, 'expected one provider-integration workspace folder');
  requireCondition(workspaceFolder.uri.scheme === 'vscode-test-web', `expected vscode-test-web workspace, got ${workspaceFolder.uri.scheme}`);

  await vscode.workspace.getConfiguration('yaml').update('schemaStore.enable', false, vscode.ConfigurationTarget.Workspace);
  await vscode.workspace.getConfiguration('yaml').update('kubernetesCRDStore.enable', false, vscode.ConfigurationTarget.Workspace);
  await vscode.workspace.getConfiguration('openxcom.ruleset').update('validator', 'oxce', vscode.ConfigurationTarget.Workspace);

  const yaml = vscode.extensions.getExtension('redhat.vscode-yaml');
  requireCondition(yaml, 'redhat.vscode-yaml was not loaded into the web host');
  requireCondition(yaml.packageJSON.version === '1.24.0', `expected redhat.vscode-yaml 1.24.0, got ${yaml.packageJSON.version ?? 'unknown'}`);
  await yaml.activate();

  const rulesetTools = vscode.extensions.getExtension('openxcom.ruleset-tools');
  requireCondition(rulesetTools, 'OpenXcom Ruleset Tools was not loaded into the web host');
  requireCondition(rulesetTools.packageJSON.version === '0.9.45', `expected Ruleset Tools 0.9.45, got ${rulesetTools.packageJSON.version ?? 'unknown'}`);
  await rulesetTools.activate();
  requireCondition(rulesetTools.isActive, 'Ruleset Tools failed to activate in the browser host');

  const studio = vscode.extensions.getExtension('sempersupra.oxce-mod-studio');
  requireCondition(studio, 'OXCE Mod Studio development extension was not loaded');
  await studio.activate();
  requireCondition(studio.isActive, 'OXCE Mod Studio failed to activate in the browser host');

  const uri = vscode.Uri.joinPath(workspaceFolder.uri, 'provider-invalid.rul');
  const document = await vscode.workspace.openTextDocument(uri);
  await vscode.window.showTextDocument(document);

  const diagnostics = await waitForDiagnostics(uri);
  const messages = diagnostics.map(item => item.message);
  requireCondition(diagnostics.length > 0, 'expected a real schema diagnostic from the pinned providers');
  requireCondition(!messages.some(message => /Problems loading reference|Unable to load schema/i.test(message)), `schema graph failed to load: ${messages.join(' | ')}`);
  requireCondition(messages.some(message => /definitelyNotARealRule|not allowed|additional propert/i.test(message)), `diagnostics did not identify the deliberate invalid property: ${messages.join(' | ')}`);

  const evidence = await vscode.commands.executeCommand('oxceModStudio._inspectBrowserEvidence', uri);
  requireCondition(evidence?.providers?.rulesetTools?.installed === true, 'Studio did not observe Ruleset Tools as installed');
  requireCondition(evidence?.providers?.rulesetTools?.active === true, 'Studio did not observe Ruleset Tools as active');
  requireCondition(evidence?.providers?.rulesetTools?.version === '0.9.45', `Studio reported wrong Ruleset Tools version: ${evidence?.providers?.rulesetTools?.version}`);
  requireCondition(evidence?.providers?.yaml?.installed === true, 'Studio did not observe Red Hat YAML as installed');
  requireCondition(evidence?.providers?.yaml?.active === true, 'Studio did not observe Red Hat YAML as active');
  requireCondition(evidence?.providers?.yaml?.version === '1.24.0', `Studio reported wrong YAML provider version: ${evidence?.providers?.yaml?.version}`);
  requireCondition(evidence?.diagnostics?.some(item => item.evidenceOrigin === 'SCHEMA'), 'Studio did not classify the real YAML/schema diagnostic as SCHEMA evidence');
  requireCondition(evidence?.engine?.evidenceOrigin === 'ENGINE-AUTHORITATIVE', 'engine evidence tier is not explicit');
  requireCondition(evidence?.engine?.state === 'not-run', 'browser provider integration must not imply native OXCE evidence ran');

  console.log(JSON.stringify({
    result: 'pass',
    workspaceScheme: workspaceFolder.uri.scheme,
    rulesetToolsVersion: rulesetTools.packageJSON.version,
    yamlVersion: yaml.packageJSON.version,
    realDiagnosticCount: diagnostics.length,
    studioObservedSchemaEvidence: true,
    engineState: evidence.engine.state
  }, null, 2));
}
