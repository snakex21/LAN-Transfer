'use strict';

// ---- translations -------------------------------------------------------------
// window.__LT_I18N comes from Go (internal/ui/render.go): the language list
// and the catalogs (locales/ui.<code>.json) of the current language and of
// English, the fallback. Values are strings or, for counts, objects keyed by
// the Intl.PluralRules category. Placeholders: {name} or {name:format} with
// format = size | speed | duration | seconds | percent | number | datetime.
const I18N = (() => {
  const boot = window.__LT_I18N;
  const catalogs = { ...boot.catalogs };
  const languages = boot.languages;
  let lang = boot.language;
  let source = boot.source;
  let detected = boot.detected;
  const formatters = new Map();
  const FSI = '⁨', PDI = '⁩'; // bidi isolates around parameters in RTL text

  const info = code => languages.find(l => l.code === code) || languages[0];
  const locale = () => info(lang).locale;
  const rtl = () => !!info(lang).rtl;

  function cached(kind, options, make) {
    const key = `${kind}|${locale()}|${JSON.stringify(options)}`;
    if (!formatters.has(key)) formatters.set(key, make(locale(), options));
    return formatters.get(key);
  }
  const numberFormat = options => cached('n', options, (l, o) => new Intl.NumberFormat(l, o));

  function formatNumber(value, maxFrac = 2) {
    return numberFormat({ maximumFractionDigits: maxFrac }).format(value);
  }

  function unit(name, value, maxFrac) {
    return numberFormat({ style: 'unit', unit: name, unitDisplay: 'short', maximumFractionDigits: maxFrac }).format(value);
  }

  const SIZE_UNITS = ['byte', 'kilobyte', 'megabyte', 'gigabyte', 'terabyte'];
  function scaled(bytes) {
    let value = bytes, index = 0;
    while (value >= 1024 && index < SIZE_UNITS.length - 1) { value /= 1024; index++; }
    return [value, SIZE_UNITS[index], value >= 10 || index === 0 ? 0 : 1];
  }

  function formatBytes(bytes) {
    if (!Number.isFinite(bytes) || bytes <= 0) return unit('byte', 0, 0);
    const [value, name, frac] = scaled(bytes);
    return unit(name, value, frac);
  }

  function formatSpeed(bps) {
    if (!Number.isFinite(bps) || bps <= 0) return '—';
    const [value, name, frac] = scaled(bps);
    return unit(`${name}-per-second`, value, frac);
  }

  function formatDuration(seconds) {
    if (!Number.isFinite(seconds) || seconds <= 0) return '—';
    if (seconds < 60) return unit('second', Math.ceil(seconds), 0);
    const list = cached('l', {}, l => new Intl.ListFormat(l, { type: 'unit', style: 'narrow' }));
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return list.format([unit('minute', minutes, 0), unit('second', Math.round(seconds % 60), 0)]);
    return list.format([unit('hour', Math.floor(minutes / 60), 0), unit('minute', minutes % 60, 0)]);
  }

  function formatPercent(value) {
    return numberFormat({ style: 'percent', maximumFractionDigits: value < 10 ? 1 : 0 }).format(value / 100);
  }

  function formatDateTime(iso, options = { dateStyle: 'short', timeStyle: 'short' }) {
    const date = new Date(iso);
    if (isNaN(date)) return '';
    return cached('d', options, (l, o) => new Intl.DateTimeFormat(l, o)).format(date);
  }

  function lookup(key) {
    for (const code of [lang, 'en']) {
      const catalog = catalogs[code];
      if (catalog && Object.prototype.hasOwnProperty.call(catalog, key)) return [catalog[key], code];
    }
    return [null, lang];
  }

  function template(key, params) {
    const [value, from] = lookup(key);
    if (value === null) return key;
    if (typeof value === 'string') return value;
    const count = Number(params && params.count);
    const category = new Intl.PluralRules(info(from).locale).select(Number.isFinite(count) ? count : 0);
    return value[category] ?? value.other;
  }

  function formatValue(value, kind) {
    if (value && typeof value === 'object') return tm(value);
    if (typeof value === 'string') {
      if (kind === 'datetime') return formatDateTime(value);
      return value;
    }
    const number = Number(value);
    if (!Number.isFinite(number)) return String(value);
    switch (kind) {
      case 'size': return formatBytes(number);
      case 'speed': return formatSpeed(number);
      case 'duration': return formatDuration(number);
      case 'seconds': return unit('second', number, 1);
      case 'percent': return formatPercent(number);
      default: return formatNumber(number);
    }
  }

  const PLACEHOLDER = /\{(\w+)(?::(\w+))?\}/g;

  // t renders a key as text. Parameters are wrapped in bidi isolates in RTL
  // languages, so paths, IP addresses and names keep their own direction.
  function t(key, params = {}) {
    return template(key, params).replace(PLACEHOLDER, (token, name, kind) => {
      if (!(name in params)) return token;
      const text = formatValue(params[name], kind);
      return rtl() ? FSI + text + PDI : text;
    });
  }

  // tn renders a key into a DocumentFragment; DOM node parameters are
  // inserted as nodes (e.g. a <b> with a device name).
  function tn(key, params = {}) {
    const fragment = document.createDocumentFragment();
    const text = template(key, params);
    let last = 0;
    for (const match of text.matchAll(PLACEHOLDER)) {
      fragment.append(text.slice(last, match.index));
      const [token, name, kind] = match;
      const value = params[name];
      if (value instanceof Node) fragment.append(value);
      else if (name in params) {
        const span = document.createElement('bdi');
        span.textContent = formatValue(value, kind);
        fragment.append(span);
      } else fragment.append(token);
      last = match.index + token.length;
    }
    fragment.append(text.slice(last));
    return fragment;
  }

  // tm renders a backend message {key, params}; an empty key carries
  // untranslated text (OS errors) in params.text.
  function tm(msg) {
    if (!msg) return '';
    if (!msg.key) return String((msg.params && msg.params.text) ?? '');
    return t(msg.key, msg.params || {});
  }

  // apply translates the static markup: data-i18n (text) and
  // data-i18n-title / -placeholder / -aria-label (attributes).
  function apply(root = document) {
    const html = document.documentElement;
    html.lang = lang;
    html.dir = rtl() ? 'rtl' : 'ltr';
    root.querySelectorAll('[data-i18n]').forEach(node => { node.textContent = t(node.dataset.i18n); });
    for (const [data, attr] of [['i18nTitle', 'title'], ['i18nPlaceholder', 'placeholder'], ['i18nAriaLabel', 'aria-label']]) {
      const selector = `[data-${data.replace(/[A-Z]/g, c => '-' + c.toLowerCase())}]`;
      root.querySelectorAll(selector).forEach(node => node.setAttribute(attr, t(node.dataset[data])));
    }
  }

  // update switches to a new bootstrap returned by the setLanguage binding.
  function update(next) {
    Object.assign(catalogs, next.catalogs);
    lang = next.language;
    source = next.source;
    detected = next.detected;
    formatters.clear();
    apply();
  }

  return {
    t, tn, tm, apply, update, formatBytes, formatSpeed, formatDuration, formatPercent, formatNumber, formatDateTime,
    get language() { return lang; },
    get source() { return source; },
    get detected() { return detected; },
    get rtl() { return rtl(); },
    languages,
    info,
  };
})();

const { t, tn, tm } = I18N;
