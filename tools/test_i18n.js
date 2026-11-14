// Tests internal/ui/assets/i18n.js (the translation runtime of the window)
// with the real catalogs in Node.js, which has the same Intl (ICU/CLDR) as
// WebView2.
//
// Run: node tools/test_i18n.js   (also part of CI)
'use strict';
const fs = require('fs');
const path = require('path');
const vm = require('vm');
const assert = require('assert');

const root = path.resolve(__dirname, '..');
const goSrc = fs.readFileSync(path.join(root, 'internal/i18n/i18n.go'), 'utf8');
const languages = [...goSrc.matchAll(/\{"([a-z]{2})", "([^"]+)", "([^"]+)", (true|false)\}/g)]
  .map(m => ({ code: m[1], name: m[2], locale: m[3], rtl: m[4] === 'true' }));
assert.strictEqual(languages.length, 30, 'language list in i18n.go');
const catalog = code => JSON.parse(fs.readFileSync(path.join(root, `locales/ui.${code}.json`), 'utf8'));
const runtime = fs.readFileSync(path.join(root, 'internal/ui/assets/i18n.js'), 'utf8');

// load evaluates i18n.js for one language and returns its I18N object.
function load(code) {
  const window = { __LT_I18N: { language: code, source: 'manual', detected: 'en', languages, catalogs: { en: catalog('en'), [code]: catalog(code) } } };
  const context = vm.createContext({ window, Intl, Node: class {}, document: {} });
  vm.runInContext(runtime + '\n;globalThis.I18N = I18N;', context);
  return context.I18N;
}

let checks = 0;
function check(actual, expected, what) {
  assert.strictEqual(actual, expected, what);
  checks++;
}
const plain = s => s.replace(/[\u2068\u2069]/g, '');

// Polish plural forms and number formats.
const pl = load('pl');
check(pl.t('pick.selectedCount', { count: 1 }), 'Wybrano 1 element', 'pl one');
check(pl.t('pick.selectedCount', { count: 3 }), 'Wybrano 3 elementy', 'pl few');
check(pl.t('pick.selectedCount', { count: 12 }), 'Wybrano 12 elementów', 'pl many (12)');
check(pl.t('pick.selectedCount', { count: 22 }), 'Wybrano 22 elementy', 'pl few (22)');
check(pl.formatBytes(1536), '1,5 kB', 'pl size');
check(pl.t('err.withPath', { path: 'a.bin', error: { key: 'err.diskFull' } }), 'a.bin: zabrakło miejsca na dysku odbiorcy', 'nested message');
check(pl.tm({ key: '', params: { text: 'Access is denied.' } }), 'Access is denied.', 'raw text message');
check(pl.tm({ key: 'err.notEnoughSpace', params: { need: 3221225472, free: 1610612736 } }),
  'za mało miejsca na dysku odbiorcy (potrzeba 3 GB, wolne 1,5 GB)', 'sizes in a backend message');

// English fallback for a missing key, key itself when missing everywhere.
const en = load('en');
check(en.t('no.such.key'), 'no.such.key', 'missing key');
check(en.t('send.queued', { count: 1 }), 'Queued 1 transfer.', 'en one');
check(en.formatDuration(125), '2 min 5 sec', 'en duration');

// Arabic: RTL, six plural categories, parameters isolated.
const ar = load('ar');
check(ar.rtl, true, 'ar is RTL');
const arTwo = ar.t('pick.selectedCount', { count: 2 });
const arMany = ar.t('pick.selectedCount', { count: 11 });
assert.notStrictEqual(arTwo, arMany, 'ar dual and 11 differ');
assert.ok(ar.t('app.subtitle', { name: 'LAPTOP', port: '45455' }).includes('\u2068LAPTOP\u2069'), 'ar isolates parameters');
checks += 2;

// Every language: every key renders without leftover placeholders, every
// plural key for many counts, and sizes use the language's own format.
for (const { code } of languages) {
  const i18n = load(code);
  const keys = Object.keys(catalog(code));
  const params = { count: 2, version: 'v2.1.0', name: 'N', peer: 'P', path: 'C:\\x', error: 'E', port: '45455', fingerprint: 'AAAA-BBBB',
    actual: 'A', announced: 'B', expected: 'C', language: 'L', variable: 'V', protocol: '2', percent: 42.5, done: 1, total: 2,
    speed: 1048576, eta: 65, elapsed: 3, bytes: 1024, time: '2026-11-14T12:00:00Z', where: 'W', shown: 1, max: 255, have: 10,
    need: 10, free: 5, before: 1, after: 2, status: '500', command: 'x', usage: 'U', address: 'A', code: '123456', size: 10,
    seconds: 1.5, doneFiles: '1', totalFiles: '2', text: 'T', status2: '' };
  for (const key of keys) {
    for (const count of [0, 1, 2, 3, 5, 11, 21, 22, 100, 101, 1000000]) {
      const text = plain(i18n.t(key, { ...params, count }));
      assert.ok(!/\{[A-Za-z]/.test(text), `${code}: ${key} (count ${count}) leaves a placeholder: ${text}`);
      assert.ok(text.trim().length > 0, `${code}: ${key} is empty`);
      checks++;
      if (typeof catalog(code)[key] === 'string') break;
    }
  }
  assert.ok(/\d/.test(i18n.formatBytes(1536)), `${code}: size has digits`);
  checks++;
}

console.log(`i18n.js OK: ${checks} checks in ${languages.length} languages`);
