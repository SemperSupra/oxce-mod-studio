#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';

const upstream = path.resolve(process.argv[2] ?? '');
if (!process.argv[2] || !fs.existsSync(upstream)) {
  throw new Error('usage: node prepare-upstream.mjs <checked-out-upstream-dir>');
}

const expectedVersion = '0.9.45';
const packagePath = path.join(upstream, 'package.json');
const extensionPath = path.join(upstream, 'src', 'extension.ts');
const contributorPath = path.join(upstream, 'src', 'ruleset-contributor.ts');

const pkg = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
if (pkg.name !== 'ruleset-tools' || pkg.version !== expectedVersion) {
  throw new Error(`unexpected upstream package identity: ${pkg.name}@${pkg.version}; expected ruleset-tools@${expectedVersion}`);
}
if (pkg.main !== './out/extension') {
  throw new Error(`unexpected upstream main entry: ${pkg.main}`);
}

pkg.browser = './dist/web/extension.js';
pkg.capabilities = {
  ...(pkg.capabilities ?? {}),
  virtualWorkspaces: { supported: true }
};
fs.writeFileSync(packagePath, `${JSON.stringify(pkg, null, 4)}\n`);

const extensionOriginal = fs.readFileSync(extensionPath, 'utf8');
const extensionNeedle = 'const ruleset = new RulesetContributor(context.extensionPath);';
if (!extensionOriginal.includes(extensionNeedle)) {
  throw new Error('pinned extension.ts no longer matches expected extensionPath constructor call');
}
const extensionAdapted = extensionOriginal.replace(
  extensionNeedle,
  'const ruleset = new RulesetContributor(context.extensionUri);'
);
fs.writeFileSync(extensionPath, extensionAdapted);

const contributorOriginal = fs.readFileSync(contributorPath, 'utf8');
const expectedContributor = `import * as vscode from "vscode";\nimport { join, extname } from "path";\nimport { SchemaContributor } from "./schema-contributor";\n\nexport class RulesetContributor {\n    private _oxcSchema: string;\n    private _oxceSchema: string;\n\n    public constructor(root: string) {\n        this._oxcSchema = this.getSchema(root, "oxc");\n        this._oxceSchema = this.getSchema(root, "oxce");\n    }\n\n    private getSchema(root: string, version: string): string {\n        const path = join(root, "./schemas/" + version + "/ruleset.json");\n        return vscode.Uri.file(path).toString();\n    }\n\n    public register(schema: SchemaContributor) {\n        schema.registerContributor("openxcom", res => this.request(res), () => "");\n    }\n\n    public request(resource: string): string {\n        if (extname(resource) === ".rul") {\n            const validator = vscode.workspace.getConfiguration().get('openxcom.ruleset.validator');\n            switch (validator) {\n                case "oxc":\n                    return this._oxcSchema;\n                case "oxce":\n                    return this._oxceSchema;\n            }\n        }\n        return "";\n    }\n}`;

if (contributorOriginal.trim() !== expectedContributor.trim()) {
  throw new Error('pinned ruleset-contributor.ts no longer matches the exact reviewed source; refusing fuzzy patch');
}

const contributorAdapted = `import * as vscode from "vscode";\nimport { SchemaContributor } from "./schema-contributor";\n\nexport class RulesetContributor {\n    private _oxcSchema: string;\n    private _oxceSchema: string;\n\n    public constructor(root: vscode.Uri) {\n        this._oxcSchema = this.getSchema(root, "oxc");\n        this._oxceSchema = this.getSchema(root, "oxce");\n    }\n\n    private getSchema(root: vscode.Uri, version: string): string {\n        return vscode.Uri.joinPath(root, "schemas", version, "ruleset.json").toString();\n    }\n\n    public register(schema: SchemaContributor) {\n        schema.registerContributor("openxcom", res => this.request(res), () => "");\n    }\n\n    public request(resource: string): string {\n        let resourcePath: string;\n        try {\n            resourcePath = vscode.Uri.parse(resource).path;\n        } catch {\n            resourcePath = resource;\n        }\n\n        if (resourcePath.toLowerCase().endsWith(".rul")) {\n            const validator = vscode.workspace.getConfiguration().get('openxcom.ruleset.validator');\n            switch (validator) {\n                case "oxc":\n                    return this._oxcSchema;\n                case "oxce":\n                    return this._oxceSchema;\n            }\n        }\n        return "";\n    }\n}`;

fs.writeFileSync(contributorPath, `${contributorAdapted}\n`);

console.log(JSON.stringify({
  upstream,
  package: `${pkg.name}@${pkg.version}`,
  browser: pkg.browser,
  virtualWorkspaces: pkg.capabilities.virtualWorkspaces.supported,
  changes: [
    'context.extensionPath -> context.extensionUri',
    'path.join/Uri.file -> Uri.joinPath',
    'path.extname -> URI-path suffix check',
    'browser entry added',
    'virtual workspace capability declared'
  ]
}, null, 2));
