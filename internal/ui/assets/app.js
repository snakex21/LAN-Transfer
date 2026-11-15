'use strict';

// ---- state -------------------------------------------------------------
const state = {
  info: null,
  devices: [],
  discovered: false,     // first discovery finished
  discoveryError: null,  // message of the last failed discovery
  target: null,          // selected device view (address, name, paired…)
  selected: [],          // local paths
  previewed: '',
  previewInfo: null,     // last loaded preview (re-rendered on language change)
  previewError: null,
  remotePath: '',
  remoteParent: '',
  remoteListing: null,   // last folder listing on the receiver
  remoteError: null,
  destination: '',       // '' = receiver's default folder
  destMode: 'default',
  refreshing: false,
  history: [],
};

const $ = id => document.getElementById(id);

// ---- helpers -------------------------------------------------------------
function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

// ltr shows a path, address or file name: always left-to-right, isolated
// from the surrounding (possibly right-to-left) text.
function ltr(tag, className, text) {
  const node = el(tag, className ? `${className} ltr-text` : 'ltr-text', text);
  node.dir = 'ltr';
  return node;
}

function baseName(path) {
  const parts = path.replace(/[\\/]+$/, '').split(/[\\/]/);
  return parts[parts.length - 1] || path;
}

function toast(message, kind = '') {
  const node = el('div', `toast ${kind}`);
  if (message instanceof Node) node.append(message); else node.textContent = message;
  $('toasts').appendChild(node);
  setTimeout(() => node.remove(), kind === 'error' ? 9000 : 4500);
}

// A failed binding call: keeps the backend message (key + params) so that it
// is shown in the UI language.
class BackendError extends Error {
  constructor(result) {
    super(result.error || '');
    this.msg = result.error_msg || null;
    this.needPairing = !!result.need_pairing;
  }
}

function describe(error) {
  if (error && error.msg) return tm(error.msg);
  return (error && error.message) || String(error);
}

function resultError(result) {
  return result.error_msg ? tm(result.error_msg) : result.error;
}

function statusLabel(status) {
  switch (status) {
    case 'queued': return t('status.queued');
    case 'running': return t('status.running');
    case 'done': return t('status.done');
    case 'failed': return t('status.failed');
    case 'cancelled': return t('status.cancelled');
    case 'interrupted': return t('status.interrupted');
  }
  return status;
}

// ---- receiver info -----------------------------------------------------------
async function loadInfo() {
  state.info = await appInfo();
  renderInfo();
  showCode(state.info.pairing_code);
  showRequireCode(state.info.require_code);
}

function renderInfo() {
  const info = state.info;
  document.title = t('app.windowTitle', { version: info ? info.version : '' }).replace(/[⁨⁩]/g, '');
  if (!info) return;
  $('me').textContent = t('app.subtitle', { name: info.name, port: String(info.port) });
  $('me').title = t('app.receiveDir', { path: info.receive_dir });
  $('fingerprint').textContent = info.fingerprint;
  $('appVersion').textContent = info.version;
  $('appProtocol').textContent = t('footer.protocol', { protocol: String(info.protocol) });
}

function showCode(code) {
  $('pairingCode').textContent = code ? `${code.slice(0, 3)} ${code.slice(3)}` : '––– –––';
}

// showRequireCode reflects the "Require pairing code" setting: a warning
// badge in the header while anyone on the network may send without a code.
function showRequireCode(required) {
  if (typeof required !== 'boolean') return;
  if (state.info) state.info.require_code = required;
  $('noCodeBadge').classList.toggle('hidden', required);
  $('requireCode').checked = required;
}

// ---- devices -------------------------------------------------------------
async function refreshDevices() {
  if (state.refreshing) return;
  state.refreshing = true;
  $('refresh').disabled = true;
  try {
    const result = await discoverDevices();
    if (!result.ok) throw new BackendError(result);
    state.devices = result.devices || [];
    state.discoveryError = null;
    if (state.target) {
      const fresh = state.devices.find(d => d.address === state.target.address);
      if (fresh) state.target = fresh;
    }
  } catch (error) {
    state.discoveryError = error;
  } finally {
    state.refreshing = false;
    state.discovered = true;
    $('refresh').disabled = false;
    renderDevices();
  }
}

function renderDevices() {
  const box = $('devices');
  const list = [...state.devices];
  if (state.target && !list.some(d => d.address === state.target.address)) list.push(state.target);
  if (!list.length && !state.discovered) {
    box.replaceChildren(el('div', 'empty', t('devices.searching')));
    return;
  }
  if (!list.length) {
    box.replaceChildren(el('div', 'empty', state.discoveryError
      ? t('devices.discoveryError', { error: describe(state.discoveryError) })
      : t('devices.none')));
    return;
  }
  box.replaceChildren(...list.map(device => {
    const button = el('button', 'device');
    button.type = 'button';
    if (state.target && state.target.address === device.address) button.classList.add('active');
    const compatible = device.compatible !== false;
    if (!compatible) button.classList.add('disabled');
    const main = el('div', 'device-main');
    const name = el('b', '', device.name);
    name.dir = 'auto';
    main.append(name, ltr('span', '', `${device.address}${device.short_fingerprint ? ' · ' + device.short_fingerprint : ''}`));
    let badge;
    if (!compatible) badge = el('span', 'badge warn', t('devices.badgeOld'));
    else if (device.paired) badge = el('span', 'badge ok', t('devices.badgePaired'));
    else if (device.no_code) badge = el('span', 'badge', t('devices.badgeNoCode'));
    else badge = el('span', 'badge', t('devices.badgeNeedsCode'));
    button.append(el('span', 'device-icon', '🖥'), main, badge);
    button.addEventListener('click', () => {
      if (!compatible) {
        toast(t('devices.oldVersion', { name: device.name }), 'error');
        return;
      }
      selectDevice(device);
    });
    return button;
  }));
}

// connectWithoutCode tries to pair with an empty code. It succeeds when the
// receiver switched its code off; "code required" is not counted as a wrong
// code there. Returns true (connected), false (a code is needed) or throws.
async function connectWithoutCode(device) {
  const result = await pairDevice(device.address, '');
  if (!result.ok) {
    if (result.code_required) return false;
    throw new BackendError(result);
  }
  if (device.short_fingerprint && result.fingerprint !== device.short_fingerprint) {
    toast(t('pair.fingerprintMismatch', { actual: result.fingerprint, announced: device.short_fingerprint }), 'error');
  } else {
    toast(t('pair.connectedNoCode', { peer: result.peer, fingerprint: result.fingerprint }), 'success');
  }
  return true;
}

async function selectDevice(device) {
  if (device.paired) {
    state.target = device;
    onTargetChanged();
    return;
  }
  // Receivers that announce "no code", and manual addresses (unknown), are
  // tried without a code first; the code dialog opens only when needed.
  if (device.no_code || !device.short_fingerprint) {
    try {
      if (await connectWithoutCode(device)) {
        state.target = { ...device, paired: true };
        onTargetChanged();
        refreshDevices();
        return;
      }
    } catch (error) {
      if (device.short_fingerprint) {
        toast(describe(error), 'error');
        return;
      }
    }
  }
  const paired = await askForPairing(device);
  if (!paired) return;
  state.target = { ...device, paired: true };
  onTargetChanged();
  refreshDevices();
}

function onTargetChanged() {
  renderDevices();
  $('remoteHint').classList.toggle('hidden', !!state.target);
  $('destBox').classList.toggle('hidden', !state.target);
  renderDestinationLabel();
  state.remotePath = '';
  state.destination = '';
  if (state.destMode === 'custom') loadRemote('');
  updateSummary();
}

function renderDestinationLabel() {
  const name = el('b', '', state.target ? state.target.name : '');
  name.dir = 'auto';
  $('destDefaultLabel').replaceChildren(tn('dest.default', { name }));
}

// ---- pairing dialog ----------------------------------------------------------
let pairingDevice = null;

function renderPairingTexts() {
  const device = pairingDevice;
  if (!device) return;
  const peer = el('span', '', device.name);
  peer.dir = 'auto';
  $('pairTitle').replaceChildren(tn('pair.title', { name: peer }));
  const finger = el('b', '', device.short_fingerprint || t('pair.unknownFingerprint'));
  if (device.short_fingerprint) finger.dir = 'ltr';
  $('pairIntro').replaceChildren(t('pair.intro'), ' ', tn('pair.checkFingerprint', { fingerprint: finger }));
}

function askForPairing(device) {
  return new Promise(resolve => {
    const dialog = $('pairDialog');
    pairingDevice = device;
    renderPairingTexts();
    $('pairCode').value = '';
    $('pairError').textContent = '';
    let done = false;
    const finish = value => {
      if (done) return;
      done = true;
      pairingDevice = null;
      $('pairForm').onsubmit = null;
      $('pairCancel').onclick = null;
      dialog.onclose = null;
      if (dialog.open) dialog.close();
      resolve(value);
    };
    $('pairCancel').onclick = () => finish(false);
    dialog.onclose = () => finish(false);
    $('pairForm').onsubmit = async event => {
      event.preventDefault();
      const code = $('pairCode').value.replace(/\D/g, '');
      if (code.length !== 6) {
        $('pairError').textContent = t('pair.codeDigits');
        return;
      }
      $('pairSubmit').disabled = true;
      $('pairError').textContent = '';
      try {
        const result = await pairDevice(device.address, code);
        if (!result.ok) throw new BackendError(result);
        if (device.short_fingerprint && result.fingerprint !== device.short_fingerprint) {
          toast(t('pair.fingerprintMismatch', { actual: result.fingerprint, announced: device.short_fingerprint }), 'error');
        } else {
          toast(t('pair.success', { peer: result.peer, fingerprint: result.fingerprint }), 'success');
        }
        finish(true);
      } catch (error) {
        $('pairError').textContent = describe(error);
      } finally {
        $('pairSubmit').disabled = false;
      }
    };
    dialog.showModal();
    $('pairCode').focus();
  });
}

function connectManually(event) {
  event.preventDefault();
  let address = $('manualAddress').value.trim();
  if (!address) return;
  if (!/:\d+$/.test(address)) address += `:${state.info ? state.info.port : 45455}`;
  selectDevice({ name: address, address, ip: address.split(':')[0], paired: false, compatible: true, short_fingerprint: '' });
}

// ---- local selection -----------------------------------------------------------
async function chooseFiles() {
  const result = await pickLocalFiles();
  if (!result.ok) return toast(t('pick.dialogFailed', { error: resultError(result) }), 'error');
  addPaths(result.paths || []);
}

async function chooseFolder() {
  const result = await pickLocalFolder();
  if (!result.ok) return toast(t('pick.dialogFailed', { error: resultError(result) }), 'error');
  if (result.path) addPaths([result.path]);
}

function addPaths(paths) {
  for (const path of paths) {
    if (path && !state.selected.includes(path)) state.selected.push(path);
  }
  renderSelection();
  updateSummary();
}

async function addTypedPath(event) {
  event.preventDefault();
  const path = $('pathInput').value.trim().replace(/^"(.*)"$/, '$1');
  if (!path) return;
  const result = await previewLocalPath(path);
  if (!result.ok) {
    toast(t('pick.addFailed', { error: resultError(result) }), 'error');
    return;
  }
  $('pathInput').value = '';
  addPaths([result.preview.path]);
}

window.__addDroppedPaths = function (paths) {
  if (!Array.isArray(paths) || !paths.length) return;
  addPaths(paths);
  toast(t('pick.dropped', { count: paths.length }));
};

function renderSelection() {
  const box = $('selection');
  $('selectionHead').classList.toggle('hidden', !state.selected.length);
  $('selectionCount').textContent = t('pick.selectedCount', { count: state.selected.length });
  box.replaceChildren(...state.selected.map((path, index) => {
    const row = el('div', 'selected-item');
    row.setAttribute('role', 'listitem');
    if (path === state.previewed) row.classList.add('preview-active');
    const text = el('div', 'selected-text');
    const name = el('b', '', baseName(path));
    name.dir = 'auto';
    text.append(name, ltr('span', '', path));
    const remove = el('button', 'remove-btn', '×');
    remove.title = t('pick.removeHint');
    remove.setAttribute('aria-label', t('pick.removeItem', { name: baseName(path) }));
    remove.addEventListener('click', event => {
      event.stopPropagation();
      state.selected.splice(index, 1);
      if (state.previewed === path) closePreview();
      renderSelection();
      updateSummary();
    });
    row.append(el('span', 'file-icon', /\.[^\\/.]+$/.test(path) ? '📄' : '📁'), text, remove);
    row.addEventListener('click', () => loadPreview(path));
    return row;
  }));
}

// ---- preview -------------------------------------------------------------
async function loadPreview(path) {
  state.previewed = path;
  state.previewInfo = null;
  state.previewError = null;
  renderSelection();
  $('previewPanel').classList.remove('hidden');
  $('previewTitle').textContent = baseName(path);
  $('previewMeta').textContent = path;
  $('previewBody').replaceChildren(el('div', 'preview-message', t('preview.loading')));
  try {
    const result = await previewLocalPath(path);
    if (!result.ok) throw new BackendError(result);
    if (state.previewed === path) {
      state.previewInfo = result.preview;
      renderPreview();
    }
  } catch (error) {
    if (state.previewed === path) {
      state.previewError = error;
      renderPreview();
    }
  }
}

function renderPreview() {
  const body = $('previewBody');
  if (state.previewError) {
    body.replaceChildren(el('div', 'preview-message', t('preview.failed', { error: describe(state.previewError) })));
    return;
  }
  const info = state.previewInfo;
  if (!info) return;
  $('previewTitle').textContent = info.name || baseName(info.path || '');
  const meta = [info.kind === 'folder' ? t('preview.folderItems', { count: info.entry_count || 0 }) : I18N.formatBytes(info.size || 0)];
  if (info.mime && !['application/octet-stream', 'inode/directory'].includes(info.mime)) meta.push(info.mime);
  if (info.modified) meta.push(t('preview.modified', { time: I18N.formatDateTime(info.modified, { dateStyle: 'medium', timeStyle: 'short' }) }));
  $('previewMeta').textContent = meta.join(' · ');

  if ((info.kind === 'image' || info.kind === 'pdf') && info.data_url) {
    const node = el(info.kind === 'image' ? 'img' : 'iframe', info.kind === 'image' ? 'preview-image' : 'preview-pdf');
    node.src = info.data_url;
    if (info.kind === 'image') node.alt = info.name || t('preview.imageAlt'); else node.title = info.name || 'PDF';
    body.replaceChildren(node);
    return;
  }
  if (info.kind === 'text') {
    const text = el('pre', 'preview-text', info.text || '');
    text.dir = 'auto';
    const nodes = [text];
    if (info.truncated) nodes.push(el('div', 'preview-message', t('preview.textTruncated')));
    body.replaceChildren(...nodes);
    return;
  }
  if (info.kind === 'folder') {
    const list = el('div', 'preview-folder-list');
    const entries = info.entries || [];
    if (!entries.length) list.append(el('div', 'preview-message', t('preview.folderEmpty')));
    entries.forEach(name => list.append(ltr('div', 'preview-folder-item', name)));
    if (info.truncated) list.append(el('div', 'preview-message', t('preview.folderTruncated', { shown: entries.length, count: info.entry_count })));
    body.replaceChildren(list);
    return;
  }
  const grid = el('div', 'preview-file-info');
  [
    [t('preview.name'), info.name, false],
    [t('preview.type'), info.mime || t('preview.unknownType'), false],
    [t('preview.size'), I18N.formatBytes(info.size || 0), false],
    [t('preview.path'), info.path, true],
  ].forEach(([k, v, isPath]) => {
    const row = el('div', 'preview-info-row');
    row.append(el('span', '', k), isPath ? ltr('span', '', v || '') : el('span', '', v || ''));
    grid.append(row);
  });
  if (info.message) grid.append(el('div', 'preview-message', tm(info.message)));
  body.replaceChildren(grid);
}

function closePreview() {
  state.previewed = '';
  state.previewInfo = null;
  state.previewError = null;
  $('previewPanel').classList.add('hidden');
  $('previewBody').replaceChildren();
  renderSelection();
}

// ---- remote destination --------------------------------------------------------
async function loadRemote(path) {
  if (!state.target) return;
  $('remoteEntries').replaceChildren(el('div', 'muted', t('dest.loading')));
  try {
    const result = await browseRemote(state.target.address, path);
    if (!result.ok) {
      if (result.need_pairing) {
        state.target.paired = false;
        renderDevices();
      }
      throw new BackendError(result);
    }
    const listing = result.listing;
    state.remoteListing = listing;
    state.remoteError = null;
    state.remotePath = listing.path || '';
    state.remoteParent = listing.parent || '';
    state.destination = state.remotePath;
  } catch (error) {
    state.remoteListing = null;
    state.remoteError = error;
  }
  renderRemote();
  updateSummary();
}

function renderRemote() {
  const entries = $('remoteEntries');
  if (state.remoteError) {
    entries.replaceChildren(el('div', 'job-error', t('dest.readFailed', { error: describe(state.remoteError) })));
    return;
  }
  const listing = state.remoteListing;
  if (!listing) return;
  $('remotePath').textContent = state.remotePath || t('dest.drives');
  $('remotePath').dir = state.remotePath ? 'ltr' : 'auto';
  $('remoteUp').disabled = !state.remotePath;
  const items = listing.entries || [];
  if (!items.length) {
    entries.replaceChildren(el('div', 'empty compact', t('dest.noSubfolders')));
    return;
  }
  entries.replaceChildren(...items.map(entry => {
    const row = el('button', 'folder-entry');
    row.type = 'button';
    row.append(el('span', 'file-icon', '📁'), ltr('span', '', entry.name), el('span', 'muted chevron', '›'));
    row.addEventListener('click', () => loadRemote(entry.path));
    return row;
  }));
}

function onDestModeChanged(event) {
  state.destMode = event.target.value;
  $('remoteBrowser').classList.toggle('hidden', state.destMode !== 'custom');
  if (state.destMode === 'custom') loadRemote(state.remotePath);
  else state.destination = '';
  updateSummary();
}

// ---- sending ---------------------------------------------------------------
function updateSummary() {
  const summary = $('summary');
  const customMissing = state.destMode === 'custom' && !state.destination;
  $('send').disabled = !(state.selected.length && state.target && state.target.paired && !customMissing);
  if (!state.selected.length) summary.textContent = t('summary.pickFiles');
  else if (!state.target) summary.textContent = t('summary.pickTarget');
  else if (!state.target.paired) summary.textContent = t('summary.pairWith', { name: state.target.name });
  else if (customMissing) summary.textContent = t('summary.pickFolder');
  else {
    const where = state.destination || t('summary.defaultFolder');
    summary.textContent = t('summary.ready', { count: state.selected.length, peer: state.target.name, where });
  }
}

async function sendSelected() {
  if ($('send').disabled) return;
  $('send').disabled = true;
  try {
    const dest = state.destMode === 'custom' ? state.destination : '';
    const result = await startTransfer(state.selected, state.target.address, dest);
    if (!result.ok) {
      if (result.need_pairing) {
        state.target.paired = false;
        renderDevices();
      }
      throw new BackendError(result);
    }
    toast(t('send.queued', { count: result.ids.length }));
    state.selected = [];
    closePreview();
    renderSelection();
    pollTransfers();
  } catch (error) {
    toast(t('send.failed', { error: describe(error) }), 'error');
  } finally {
    updateSummary();
  }
}

// ---- transfers + history ------------------------------------------------------
const notified = new Set();

function jobCard({ id, direction, name, peer, status, error, totalBytes, totalFiles, doneBytes, doneFiles, speed, eta, elapsed, active, resumed, current }) {
  const card = el('div', `job ${status}`);
  const head = el('div', 'job-head');
  const title = el('div', 'job-title');
  const where = direction === 'out' ? t('job.to', { peer }) : t('job.from', { peer });
  const jobName = el('b', '', name);
  jobName.dir = 'auto';
  title.append(jobName, el('span', '', `${where} · ${statusLabel(status)}`));
  head.append(el('span', `job-dir ${direction}`, direction === 'out' ? '⬆' : '⬇'), title);
  if (direction === 'out') {
    const actions = el('div', 'job-actions');
    if (status === 'running' || status === 'queued') {
      const cancel = el('button', 'btn danger small', t('common.cancel'));
      cancel.addEventListener('click', async () => {
        const r = await cancelTransfer(id);
        if (!r.ok) toast(resultError(r), 'error');
      });
      actions.append(cancel);
    }
    if (status === 'failed' || status === 'cancelled') {
      const retry = el('button', 'btn secondary small', t('job.resume'));
      retry.title = t('job.resumeHint');
      retry.addEventListener('click', async () => {
        const r = await retryTransfer(id);
        if (!r.ok) toast(resultError(r), 'error');
        pollTransfers();
      });
      actions.append(retry);
    }
    head.append(actions);
  }
  card.append(head);
  const compact = status === 'done';
  if (compact) card.classList.add('compact');

  const pct = totalBytes > 0 ? Math.min(100, (doneBytes / totalBytes) * 100) : (status === 'done' ? 100 : 0);
  const bar = el('div', 'bar');
  bar.setAttribute('role', 'progressbar');
  bar.setAttribute('aria-valuenow', String(Math.round(pct)));
  bar.setAttribute('aria-valuemin', '0');
  bar.setAttribute('aria-valuemax', '100');
  const fill = el('div');
  fill.style.width = `${pct.toFixed(1)}%`;
  bar.append(fill);
  if (!compact) card.append(bar);

  const stats = el('div', 'job-stats');
  stats.append(el('span', '', compact ? I18N.formatBytes(totalBytes) : t('job.progress', { percent: pct, done: doneBytes, total: totalBytes })));
  if (totalFiles > 1 || doneFiles > 1) stats.append(el('span', '', t('job.files', { done: doneFiles, count: totalFiles })));
  if (status === 'running') {
    stats.append(el('span', '', t('job.speed', { speed })));
    stats.append(el('span', '', t('job.remaining', { eta })));
  } else if (status === 'done' && elapsed > 0) {
    stats.append(el('span', '', totalBytes >= 1 << 20 ? t('job.tookAverage', { elapsed, speed }) : t('job.took', { elapsed })));
  }
  if (current) stats.append(el('span', '', t('job.current', { name: current })));
  if (resumed > 0) stats.append(el('span', '', t('job.resumedFrom', { bytes: resumed })));
  if (status === 'done' && direction === 'out') stats.append(el('span', 'status done', t('job.verified')));
  card.append(stats);
  if (error) card.append(el('div', 'job-error', error));

  if (active && active.length) {
    const files = el('div', 'files');
    active.slice(0, 6).forEach(file => {
      const row = el('div', 'file-row');
      const mini = el('div', 'mini-bar');
      const mfill = el('div');
      mfill.style.width = `${file.size > 0 ? (file.done / file.size) * 100 : 100}%`;
      mini.append(mfill);
      row.append(ltr('span', '', file.rel), mini, el('span', '', I18N.formatBytes(file.size)));
      files.append(row);
    });
    card.append(files);
  }
  return card;
}

// errorText prefers the translatable message over the stored text.
function errorText(item) {
  return item.error_msg ? tm(item.error_msg) : (item.error || '');
}

function renderTransfers(outgoing, incoming) {
  const cards = [];
  incoming.forEach(job => cards.push(jobCard({
    id: job.id, direction: 'in', name: job.name, peer: job.peer || t('job.unknownPeer'), status: job.error ? 'failed' : 'running',
    error: errorText(job), totalBytes: job.total_bytes, totalFiles: job.total_files, doneBytes: job.done_bytes,
    doneFiles: job.done_files, speed: job.speed, eta: job.speed > 0 ? (job.total_bytes - job.done_bytes) / job.speed : 0,
    active: [], resumed: 0, current: job.current,
  })));
  outgoing.forEach(job => cards.push(jobCard({
    id: job.id, direction: 'out', name: job.name, peer: job.peer, status: job.status, error: errorText(job),
    totalBytes: job.total_bytes, totalFiles: job.total_files, doneBytes: job.done_bytes, doneFiles: job.done_files,
    speed: job.speed, eta: job.eta, elapsed: job.elapsed, active: job.active, resumed: job.resumed_bytes,
  })));
  $('transfers').replaceChildren(...(cards.length ? cards : [el('div', 'empty compact', t('transfers.empty'))]));
  $('clearFinished').classList.toggle('hidden', !outgoing.some(j => !['running', 'queued'].includes(j.status)));

  outgoing.forEach(job => {
    const key = `${job.id}:${job.status}`;
    if (notified.has(key)) return;
    if (job.status === 'done') toast(t('toast.sent', { name: job.name, peer: job.peer }), 'success');
    if (job.status === 'failed') toast(t('toast.sendFailed', { name: job.name, error: errorText(job) }), 'error');
    if (['done', 'failed'].includes(job.status)) notified.add(key);
  });
}

function renderHistory() {
  const history = state.history;
  if (!history.length) {
    $('history').replaceChildren(el('div', 'empty compact', t('history.empty')));
    return;
  }
  $('history').replaceChildren(...history.slice(0, 50).map(entry => {
    const row = el('div', 'history-row');
    const sent = entry.direction === 'sent';
    const name = el('span', '', entry.name);
    name.dir = 'auto';
    row.append(
      el('span', `job-dir ${sent ? 'out' : 'in'}`, sent ? '⬆' : '⬇'),
      name,
      el('span', 'h-peer muted', sent ? t('history.to', { peer: entry.peer || '—' }) : t('history.from', { peer: entry.peer || '—' })),
      el('span', 'h-size muted', I18N.formatBytes(entry.bytes)),
      el('span', `status ${entry.status}`, statusLabel(entry.status)),
      el('span', 'h-time muted', I18N.formatDateTime(entry.finished)),
    );
    row.title = [entry.location, errorText(entry)].filter(Boolean).join('\n');
    if (entry.location && !sent) {
      row.style.cursor = 'pointer';
      row.addEventListener('click', () => openFolder(entry.location));
    }
    return row;
  }));
}

let polling = false;
let lastHistoryKey = '';
async function pollTransfers() {
  if (polling) return;
  polling = true;
  try {
    const result = await transferState();
    if (!result.ok) return;
    renderTransfers(result.outgoing || [], result.incoming || []);
    const history = result.history || [];
    const key = history.length + ':' + (history[0] ? history[0].finished : '');
    if (key !== lastHistoryKey) {
      lastHistoryKey = key;
      state.history = history;
      renderHistory();
    }
    showCode(result.pairing_code);
    showRequireCode(result.require_code);
  } finally {
    polling = false;
  }
}

// ---- settings: language ---------------------------------------------------------
function renderLanguagePicker() {
  const select = $('languageSelect');
  const detected = I18N.info(I18N.detected);
  const auto = el('option', '', t('settings.auto', { language: detected.name }));
  auto.value = 'auto';
  const options = [auto, ...I18N.languages.map(l => {
    const option = el('option', '', l.name);
    option.value = l.code;
    option.lang = l.code;
    option.dir = l.rtl ? 'rtl' : 'ltr';
    return option;
  })];
  select.replaceChildren(...options);
  select.value = I18N.source === 'manual' ? I18N.language : 'auto';
  $('languageHint').textContent = I18N.source === 'env'
    ? t('settings.envHint', { variable: 'LANTRANSFER_LANG' })
    : t('settings.languageHint');
}

async function changeLanguage() {
  const select = $('languageSelect');
  select.disabled = true;
  try {
    const result = await setLanguage(select.value);
    if (!result.ok) throw new BackendError(result);
    I18N.update(result.i18n);
    renderAll();
  } catch (error) {
    toast(t('settings.saveFailed', { error: describe(error) }), 'error');
  } finally {
    select.disabled = false;
  }
}

// renderAll redraws every text after a language change.
function renderAll() {
  I18N.apply();
  renderInfo();
  renderDevices();
  renderSelection();
  if (state.previewed) renderPreview();
  if (state.target) renderDestinationLabel();
  if (state.destMode === 'custom') renderRemote();
  renderPairingTexts();
  renderLanguagePicker();
  updateSummary();
  renderHistory();
  pollTransfers();
}

// ---- wiring ----------------------------------------------------------------
$('refresh').addEventListener('click', refreshDevices);
$('pickFiles').addEventListener('click', chooseFiles);
$('pickFolder').addEventListener('click', chooseFolder);
$('clearSelection').addEventListener('click', () => {
  state.selected = [];
  closePreview();
  renderSelection();
  updateSummary();
});
$('closePreview').addEventListener('click', closePreview);
$('remoteUp').addEventListener('click', () => loadRemote(state.remoteParent));
$('remoteRefresh').addEventListener('click', () => loadRemote(state.remotePath));
document.querySelectorAll('input[name=destMode]').forEach(input => input.addEventListener('change', onDestModeChanged));
$('send').addEventListener('click', sendSelected);
$('manualForm').addEventListener('submit', connectManually);
$('pathForm').addEventListener('submit', addTypedPath);
$('clearFinished').addEventListener('click', async () => { await clearFinished(); pollTransfers(); });
$('clearHistory').addEventListener('click', async () => { await clearHistory(); lastHistoryKey = ''; pollTransfers(); });
$('openReceive').addEventListener('click', async () => {
  const r = await openFolder(state.info ? state.info.receive_dir : '');
  if (!r.ok) toast(resultError(r), 'error');
});
$('newCode').addEventListener('click', async () => {
  const r = await newPairingCode();
  if (r.ok) {
    showCode(r.pairing_code);
    toast(t('receive.newCodeDone'));
  }
});
$('openSettings').addEventListener('click', () => {
  renderLanguagePicker();
  $('settingsDialog').showModal();
  $('languageSelect').focus();
});
$('languageSelect').addEventListener('change', changeLanguage);
$('requireCode').addEventListener('change', async () => {
  const box = $('requireCode');
  box.disabled = true;
  try {
    const result = await setRequireCode(box.checked);
    if (!result.ok) throw new BackendError(result);
    showRequireCode(result.require_code);
  } catch (error) {
    toast(describe(error), 'error');
    showRequireCode(!box.checked);
  } finally {
    box.disabled = false;
  }
});

const dropZone = $('dropZone');
window.addEventListener('dragover', event => { event.preventDefault(); dropZone.classList.add('drag-active'); });
window.addEventListener('dragleave', event => { if (!event.relatedTarget) dropZone.classList.remove('drag-active'); });
window.addEventListener('drop', event => {
  event.preventDefault();
  dropZone.classList.remove('drag-active');
  // Windows delivers full paths through the native drop handler in Go;
  // some WebView hosts also expose them here.
  const paths = Array.from(event.dataTransfer?.files || []).map(file => file.path || '').filter(Boolean);
  if (paths.length) window.__addDroppedPaths(paths);
});

// ---- firewall (1.x rules → 2.x rules, only on user action) --------------------
async function checkFirewall() {
  const r = await firewallStatus();
  $('firewallBanner').classList.toggle('hidden', !(r.ok && r.firewall.legacy && !r.firewall.configured));
}
$('firewallLater').addEventListener('click', () => $('firewallBanner').classList.add('hidden'));
$('firewallUpdate').addEventListener('click', async () => {
  const r = await updateFirewall();
  if (!r.ok) return toast(resultError(r), 'error');
  toast(t('firewall.confirmUac'));
  setTimeout(checkFirewall, 8000);
});

I18N.apply();
renderInfo();
renderDevices();
renderHistory();
renderTransfers([], []);
document.body.classList.remove('i18n-pending');
loadInfo();
checkFirewall();
refreshDevices();
renderSelection();
updateSummary();
pollTransfers();
setInterval(refreshDevices, 5000);
setInterval(pollTransfers, 500);
