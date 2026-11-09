import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import ts from 'typescript';

const source = readFileSync(new URL('./merge.ts', import.meta.url), 'utf8');
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022 } });
const { mergeChannelSettingsForUpdate } = await import(`data:text/javascript;base64,${Buffer.from(compiled.outputText).toString('base64')}`);

test('create/duplicate submission retains selected users when adding proxy settings', () => {
  const usageAccess = { mode: 'selected', userIDs: [2, 3] };
  const serialized = JSON.parse(JSON.stringify(mergeChannelSettingsForUpdate({ usageAccess }, { proxy: { type: 'disabled' } })));
  assert.deepEqual(serialized.usageAccess, usageAccess);
});

test('unrelated settings dialogs retain restrictions and an explicit edit can restore public access', () => {
  const existing = { usageAccess: { mode: 'selected', userIDs: [2] } };
  assert.deepEqual(mergeChannelSettingsForUpdate(existing, { extraModelPrefix: 'alias' }).usageAccess, existing.usageAccess);
  assert.deepEqual(mergeChannelSettingsForUpdate(existing, { usageAccess: { mode: 'all', userIDs: [] } }).usageAccess, { mode: 'all', userIDs: [] });
  assert.equal(mergeChannelSettingsForUpdate(undefined, {}).usageAccess, undefined);
});
