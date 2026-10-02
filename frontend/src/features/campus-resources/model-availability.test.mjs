import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import ts from 'typescript';

const source = readFileSync(new URL('./data/model-availability.ts', import.meta.url), 'utf8');
const compiled = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2023 },
}).outputText;
const { modelCandidates, modelAvailability, orderModels } = await import(
  `data:text/javascript;base64,${Buffer.from(compiled).toString('base64')}`
);
const now = Date.parse('2026-10-02T12:00:00Z');
const route = (pass, age = 1000) => ({
  model: 'upstream',
  requestModels: ['alias'],
  known: true,
  available: pass,
  lastTestAt: new Date(now - age).toISOString(),
});
const channel = (routes, id = '1') => ({ id, status: 'enabled', models: ['alias'], health: { routes } });
const details = { testTargets: [{ channelID: '1', modelID: 'alias' }] };
test('configured model aliases retain their scoped channel target', () => {
  assert.equal(modelCandidates('public-name', [channel([])], details)[0].modelID, 'alias');
  assert.equal(modelCandidates('public-name', [channel([], '2')], details).length, 0);
  assert.equal(modelAvailability('public-name', [channel([route(true)])], details, now), 'tested');
});
test('one failed channel, untested credential or expired test cannot condemn a model', () => {
  assert.equal(modelAvailability('alias', [channel([route(false)]), channel([], '2')], undefined, now), 'unknown');
  assert.equal(modelAvailability('alias', [channel([route(false), { ...route(false), known: false }])], undefined, now), 'unknown');
  assert.equal(modelAvailability('alias', [channel([route(false, 3600001)])], undefined, now), 'unknown');
  assert.equal(modelAvailability('alias', [channel([{ ...route(false), testInFlight: true }])], undefined, now), 'unknown');
  assert.equal(modelAvailability('alias', [], undefined, now), 'unknown');
});
test('only fresh tests covering every route demote; any passed route is shown', () => {
  assert.equal(modelAvailability('alias', [channel([route(false)])], undefined, now), 'failed');
  assert.equal(modelAvailability('alias', [channel([route(false)]), channel([route(true)], '2')], undefined, now), 'tested');
});
test('newer success evidence prevents demotion; successful volume sorts models', () => {
  const activity = {
    windowStart: new Date(now - 86400000).toISOString(),
    bucketHours: 1,
    models: {
      alias: { successCount: 2, buckets: [...Array(23).fill(0), 2] },
      quiet: { successCount: 0, buckets: Array(24).fill(0) },
      popular: { successCount: 10, buckets: [] },
    },
  };
  assert.equal(modelAvailability('alias', [channel([route(false)])], undefined, now, activity), 'unknown');
  assert.deepEqual(orderModels(['quiet', 'alias', 'popular', 'alias'], activity), ['popular', 'alias', 'quiet']);
});

test('missing target metadata is unknown, and a later success can recover any failed route', () => {
  const targetDetails = {
    testTargets: [
      { channelID: '1', modelID: 'alias' },
      { channelID: '2', modelID: 'alias' },
    ],
  };
  assert.equal(modelAvailability('public', [channel([route(false)])], targetDetails, now), 'unknown');
  const activity = {
    windowStart: new Date(now - 86400000 - 1800000).toISOString(),
    bucketHours: 1,
    models: {
      alias: { successCount: 1, buckets: [...Array(23).fill(0), 1] },
    },
  };
  assert.equal(
    modelAvailability('alias', [channel([route(false, 3500000)]), channel([route(false, 1000)], '2')], undefined, now, activity),
    'unknown'
  );
});
