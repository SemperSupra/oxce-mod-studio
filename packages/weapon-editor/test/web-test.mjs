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
  requireCondition(commands.includes('oxceModStudio._applyWeaponScalar'), 'bounded scalar command not registered');

  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  requireCondition(workspaceFolder, 'expected one web-test workspace folder');
  requireCondition(workspaceFolder.uri.scheme === 'vscode-test-web', `expected vscode-test-web workspace, got ${workspaceFolder.uri.scheme}`);

  const uri = vscode.Uri.joinPath(workspaceFolder.uri, 'weapon.rul');
  const document = await vscode.workspace.openTextDocument(uri);
  await vscode.window.showTextDocument(document);
  const before = document.getText();

  const opened = await vscode.commands.executeCommand('oxceModStudio.openWeaponEditor');
  requireCondition(opened?.weaponId === 'STR_WEB_TEST_WEAPON', `weapon inspector returned unexpected id: ${opened?.weaponId}`);
  requireCondition(opened?.kind === 'item', `weapon inspector must expose canonical kind, got ${opened?.kind}`);
  requireCondition(opened?.evidenceOrigin === 'SOURCE/TEXT', `weapon inspector must label source evidence, got ${opened?.evidenceOrigin}`);

  const applied = await vscode.commands.executeCommand('oxceModStudio._applyWeaponScalar', uri, 'STR_WEB_TEST_WEAPON', 'accuracySnap', 70);
  requireCondition(applied === true, 'bounded source edit was not applied');

  const after = document.getText();
  const expected = before.replace('accuracySnap: 65 # preserve me', 'accuracySnap: 70 # preserve me');
  requireCondition(after === expected, 'VS Code Web edit changed more than the intended scalar source span');
  requireCondition(after.includes('futureField: 123 # must survive'), 'unknown field/comment was not preserved');

  console.log(JSON.stringify({
    result: 'pass',
    extensionId: extension.id,
    extensionVersion: extension.packageJSON.version,
    workspaceScheme: workspaceFolder.uri.scheme,
    weaponId: opened.weaponId,
    kind: opened.kind,
    evidenceOrigin: opened.evidenceOrigin,
    exactSourceSpanEdit: true,
    unknownFieldPreserved: true
  }, null, 2));
}
