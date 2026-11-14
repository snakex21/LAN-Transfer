'use strict';

// ---- state -------------------------------------------------------------
const state = {
  info: null,
  devices: [],
  target: null,          // selected device view (address, name, paired…)
  selected: [],          // local paths
  previewed: '',
  remotePath: '',
  remoteParent: '',
  destination: '',       // '' = receiver's default folder
  destMode: 'default',
  refreshing: false,
};

const $ = id => document.getElementById(id);

// ---- helpers -------------------------------------------------------------
function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function formatBytes(bytes) {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes, unit = 0;
  while (value >= 1024 && unit < units.length - 1) { value /= 1024; unit++; }
  return `${value.toFixed(value >= 10 || unit === 0 ? 0 : 1)} ${units[unit]}`;
}

function formatSpeed(bps) {
  return bps > 0 ? `${formatBytes(bps)}/s` : '—';
}

function formatDuration(seconds) {
  if (!Number.isFinite(seconds) || seconds <= 0) return '—';
  if (seconds < 60) return `${Math.ceil(seconds)} s`;
  const m = Math.floor(seconds / 60), s = Math.round(seconds % 60);
  if (m < 60) return `${m} min ${s} s`;
  return `${Math.floor(m / 60)} h ${m % 60} min`;
}

function formatTime(iso) {
  const d = new Date(iso);
  if (isNaN(d)) return '';
  return d.toLocaleString('pl-PL', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
}

function plural(n, one, few, many) {
  if (n === 1) return one;
  const d = n % 10, t = n % 100;
  return d >= 2 && d <= 4 && (t < 12 || t > 14) ? few : many;
}

function baseName(path) {
  const parts = path.replace(/[\\/]+$/, '').split(/[\\/]/);
  return parts[parts.length - 1] || path;
}

function toast(message, kind = '') {
  const node = el('div', `toast ${kind}`, message);
  $('toasts').appendChild(node);
  setTimeout(() => node.remove(), kind === 'error' ? 9000 : 4500);
}

const STATUS_LABEL = {
  queued: 'w kolejce', running: 'trwa', done: 'zakończono', failed: 'błąd',
  cancelled: 'anulowano', interrupted: 'przerwano',
};

// ---- receiver info -----------------------------------------------------------
async function loadInfo() {
  const info = await appInfo();
  state.info = info;
  $('me').textContent = `${info.name} · odbiór na porcie ${info.port}`;
  $('me').title = `Folder odbioru: ${info.receive_dir}`;
  $('fingerprint').textContent = info.fingerprint;
  showCode(info.pairing_code);
}

function showCode(code) {
  $('pairingCode').textContent = code ? `${code.slice(0, 3)} ${code.slice(3)}` : '––– –––';
}

// ---- devices -------------------------------------------------------------
async function refreshDevices() {
  if (state.refreshing) return;
  state.refreshing = true;
  $('refresh').disabled = true;
  try {
    const result = await discoverDevices();
    if (!result.ok) throw new Error(result.error);
    state.devices = result.devices || [];
    if (state.target) {
      const fresh = state.devices.find(d => d.address === state.target.address);
      if (fresh) state.target = fresh;
    }
    renderDevices();
  } catch (error) {
    if (!state.devices.length) {
      $('devices').replaceChildren(el('div', 'empty', `Błąd wykrywania: ${error.message || error}`));
    }
  } finally {
    state.refreshing = false;
    $('refresh').disabled = false;
  }
}

function renderDevices() {
  const box = $('devices');
  const list = [...state.devices];
  if (state.target && !list.some(d => d.address === state.target.address)) list.push(state.target);
  if (!list.length) {
    box.replaceChildren(el('div', 'empty', 'Nie znaleziono innych komputerów. Uruchom LAN Transfer na drugim komputerze w tej samej sieci albo połącz się ręcznie.'));
    return;
  }
  box.replaceChildren(...list.map(device => {
    const button = el('button', 'device');
    button.type = 'button';
    if (state.target && state.target.address === device.address) button.classList.add('active');
    const compatible = device.compatible !== false;
    if (!compatible) button.classList.add('disabled');
    const main = el('div', 'device-main');
    main.append(el('b', '', device.name), el('span', '', `${device.address}${device.short_fingerprint ? ' · ' + device.short_fingerprint : ''}`));
    let badge;
    if (!compatible) badge = el('span', 'badge warn', 'starsza wersja');
    else if (device.paired) badge = el('span', 'badge ok', '✓ sparowano');
    else badge = el('span', 'badge', 'wymaga kodu');
    button.append(el('span', 'device-icon', '🖥'), main, badge);
    button.addEventListener('click', () => {
      if (!compatible) {
        toast(`${device.name} używa starszej wersji LAN Transfer bez szyfrowania. Zaktualizuj program na tym komputerze.`, 'error');
        return;
      }
      selectDevice(device);
    });
    return button;
  }));
}

async function selectDevice(device) {
  if (device.paired) {
    state.target = device;
    onTargetChanged();
    return;
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
  $('destPeer').textContent = state.target ? state.target.name : '';
  state.remotePath = '';
  state.destination = '';
  if (state.destMode === 'custom') loadRemote('');
  updateSummary();
}

// ---- pairing dialog ----------------------------------------------------------
function askForPairing(device) {
  return new Promise(resolve => {
    const dialog = $('pairDialog');
    $('pairPeer').textContent = device.name;
    $('pairFinger').textContent = device.short_fingerprint || 'nieznany (połączenie ręczne)';
    $('pairCode').value = '';
    $('pairError').textContent = '';
    let done = false;
    const finish = value => {
      if (done) return;
      done = true;
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
        $('pairError').textContent = 'Kod ma 6 cyfr.';
        return;
      }
      $('pairSubmit').disabled = true;
      $('pairError').textContent = '';
      try {
        const result = await pairDevice(device.address, code);
        if (!result.ok) throw new Error(result.error);
        if (device.short_fingerprint && result.fingerprint !== device.short_fingerprint) {
          toast(`Uwaga: odcisk certyfikatu (${result.fingerprint}) różni się od ogłaszanego (${device.short_fingerprint}).`, 'error');
        } else {
          toast(`Sparowano z ${result.peer} (certyfikat ${result.fingerprint}).`, 'success');
        }
        finish(true);
      } catch (error) {
        $('pairError').textContent = error.message || String(error);
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
  if (!result.ok) return toast(`Nie można otworzyć okna wyboru: ${result.error}`, 'error');
  addPaths(result.paths || []);
}

async function chooseFolder() {
  const result = await pickLocalFolder();
  if (!result.ok) return toast(`Nie można otworzyć okna wyboru: ${result.error}`, 'error');
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
    toast(`Nie można dodać: ${result.error}`, 'error');
    return;
  }
  $('pathInput').value = '';
  addPaths([result.preview.path]);
}

window.__addDroppedPaths = function (paths) {
  if (!Array.isArray(paths) || !paths.length) return;
  addPaths(paths);
  toast(paths.length === 1 ? 'Dodano przeciągnięty element.' : `Dodano ${paths.length} ${plural(paths.length, 'element', 'elementy', 'elementów')}.`);
};

function renderSelection() {
  const box = $('selection');
  $('selectionHead').classList.toggle('hidden', !state.selected.length);
  $('selectionCount').textContent = `Wybrano ${state.selected.length} ${plural(state.selected.length, 'element', 'elementy', 'elementów')}`;
  box.replaceChildren(...state.selected.map((path, index) => {
    const row = el('div', 'selected-item');
    row.setAttribute('role', 'listitem');
    if (path === state.previewed) row.classList.add('preview-active');
    const text = el('div', 'selected-text');
    text.append(el('b', '', baseName(path)), el('span', '', path));
    const remove = el('button', 'remove-btn', '×');
    remove.title = 'Usuń z listy';
    remove.setAttribute('aria-label', `Usuń ${baseName(path)}`);
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
  renderSelection();
  $('previewPanel').classList.remove('hidden');
  $('previewTitle').textContent = baseName(path);
  $('previewMeta').textContent = path;
  const body = $('previewBody');
  body.replaceChildren(el('div', 'preview-message', 'Wczytywanie podglądu…'));
  try {
    const result = await previewLocalPath(path);
    if (!result.ok) throw new Error(result.error);
    if (state.previewed === path) renderPreview(result.preview);
  } catch (error) {
    if (state.previewed === path) body.replaceChildren(el('div', 'preview-message', `Brak podglądu: ${error.message || error}`));
  }
}

function renderPreview(info) {
  const body = $('previewBody');
  $('previewTitle').textContent = info.name || baseName(info.path || '');
  const meta = [info.kind === 'folder' ? `${info.entry_count || 0} elementów` : formatBytes(info.size || 0)];
  if (info.mime && !['application/octet-stream', 'inode/directory'].includes(info.mime)) meta.push(info.mime);
  if (info.modified) meta.push(`zmieniono ${info.modified}`);
  $('previewMeta').textContent = meta.join(' · ');

  if ((info.kind === 'image' || info.kind === 'pdf') && info.data_url) {
    const node = el(info.kind === 'image' ? 'img' : 'iframe', info.kind === 'image' ? 'preview-image' : 'preview-pdf');
    node.src = info.data_url;
    if (info.kind === 'image') node.alt = info.name || 'Podgląd obrazu'; else node.title = info.name || 'PDF';
    body.replaceChildren(node);
    return;
  }
  if (info.kind === 'text') {
    const nodes = [el('pre', 'preview-text', info.text || '')];
    if (info.truncated) nodes.push(el('div', 'preview-message', 'Pokazano początek pliku.'));
    body.replaceChildren(...nodes);
    return;
  }
  if (info.kind === 'folder') {
    const list = el('div', 'preview-folder-list');
    const entries = info.entries || [];
    if (!entries.length) list.append(el('div', 'preview-message', 'Folder jest pusty.'));
    entries.forEach(name => list.append(el('div', 'preview-folder-item', name)));
    if (info.truncated) list.append(el('div', 'preview-message', `Pokazano ${entries.length} z ${info.entry_count} elementów.`));
    body.replaceChildren(list);
    return;
  }
  const grid = el('div', 'preview-file-info');
  [['Nazwa', info.name], ['Typ', info.mime || 'nieznany'], ['Rozmiar', formatBytes(info.size || 0)], ['Ścieżka', info.path]].forEach(([k, v]) => {
    const row = el('div', 'preview-info-row');
    row.append(el('span', '', k), el('span', '', v || ''));
    grid.append(row);
  });
  if (info.message) grid.append(el('div', 'preview-message', info.message));
  body.replaceChildren(grid);
}

function closePreview() {
  state.previewed = '';
  $('previewPanel').classList.add('hidden');
  $('previewBody').replaceChildren();
  renderSelection();
}

// ---- remote destination --------------------------------------------------------
async function loadRemote(path) {
  if (!state.target) return;
  const entries = $('remoteEntries');
  entries.replaceChildren(el('div', 'muted', 'Wczytywanie folderów…'));
  try {
    const result = await browseRemote(state.target.address, path);
    if (!result.ok) {
      if (result.need_pairing) {
        state.target.paired = false;
        renderDevices();
      }
      throw new Error(result.error);
    }
    const listing = result.listing;
    state.remotePath = listing.path || '';
    state.remoteParent = listing.parent || '';
    state.destination = state.remotePath;
    $('remotePath').textContent = state.remotePath || 'Dyski – wybierz dysk';
    $('remoteUp').disabled = !state.remotePath;
    const items = listing.entries || [];
    if (!items.length) {
      entries.replaceChildren(el('div', 'empty compact', 'Brak podfolderów – pliki trafią tutaj.'));
    } else {
      entries.replaceChildren(...items.map(entry => {
        const row = el('button', 'folder-entry');
        row.type = 'button';
        row.append(el('span', 'file-icon', '📁'), el('span', '', entry.name), el('span', 'muted', '›'));
        row.addEventListener('click', () => loadRemote(entry.path));
        return row;
      }));
    }
  } catch (error) {
    entries.replaceChildren(el('div', 'job-error', `Nie można odczytać folderów: ${error.message || error}`));
  }
  updateSummary();
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
  if (!state.selected.length) summary.textContent = 'Wybierz pliki lub foldery do wysłania.';
  else if (!state.target) summary.textContent = 'Wybierz komputer docelowy.';
  else if (!state.target.paired) summary.textContent = `Sparuj z ${state.target.name}.`;
  else if (customMissing) summary.textContent = 'Wskaż folder na odbiorcy.';
  else {
    const where = state.destination || 'folder odbioru';
    summary.textContent = `${state.selected.length} ${plural(state.selected.length, 'element', 'elementy', 'elementów')} → ${state.target.name} → ${where}`;
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
      throw new Error(result.error);
    }
    toast(`Dodano do kolejki: ${result.ids.length} ${plural(result.ids.length, 'transfer', 'transfery', 'transferów')}.`);
    state.selected = [];
    closePreview();
    renderSelection();
    pollTransfers();
  } catch (error) {
    toast(`Nie udało się rozpocząć wysyłania: ${error.message || error}`, 'error');
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
  const arrow = direction === 'out' ? `→ ${peer}` : `← od ${peer}`;
  title.append(el('b', '', name), el('span', '', `${arrow} · ${STATUS_LABEL[status] || status}`));
  head.append(el('span', `job-dir ${direction}`, direction === 'out' ? '⬆' : '⬇'), title);
  if (direction === 'out') {
    const actions = el('div', 'job-actions');
    if (status === 'running' || status === 'queued') {
      const cancel = el('button', 'btn danger small', 'Anuluj');
      cancel.addEventListener('click', async () => {
        const r = await cancelTransfer(id);
        if (!r.ok) toast(r.error, 'error');
      });
      actions.append(cancel);
    }
    if (status === 'failed' || status === 'cancelled') {
      const retry = el('button', 'btn secondary small', 'Wznów');
      retry.title = 'Wyśle tylko brakującą część';
      retry.addEventListener('click', async () => {
        const r = await retryTransfer(id);
        if (!r.ok) toast(r.error, 'error');
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
  stats.append(el('span', '', compact ? formatBytes(totalBytes) : `${pct.toFixed(pct < 10 ? 1 : 0)}% · ${formatBytes(doneBytes)} z ${formatBytes(totalBytes)}`));
  if (totalFiles > 1 || doneFiles > 1) stats.append(el('span', '', `pliki ${doneFiles}/${totalFiles}`));
  if (status === 'running') {
    stats.append(el('span', '', `prędkość ${formatSpeed(speed)}`));
    stats.append(el('span', '', `pozostało ${formatDuration(eta)}`));
  } else if (status === 'done' && elapsed > 0) {
    stats.append(el('span', '', totalBytes >= 1 << 20 ? `w ${formatDuration(elapsed)} · śr. ${formatSpeed(speed)}` : `w ${formatDuration(elapsed)}`));
  }
  if (current) stats.append(el('span', '', `teraz: ${current}`));
  if (resumed > 0) stats.append(el('span', '', `wznowiono od ${formatBytes(resumed)}`));
  if (status === 'done' && direction === 'out') stats.append(el('span', 'status done', '✓ SHA-256 zgodne'));
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
      row.append(el('span', '', file.rel), mini, el('span', '', formatBytes(file.size)));
      files.append(row);
    });
    card.append(files);
  }
  return card;
}

function renderTransfers(outgoing, incoming) {
  const cards = [];
  incoming.forEach(job => cards.push(jobCard({
    id: job.id, direction: 'in', name: job.name, peer: job.peer || 'nieznany', status: job.error ? 'failed' : 'running',
    error: job.error, totalBytes: job.total_bytes, totalFiles: job.total_files, doneBytes: job.done_bytes,
    doneFiles: job.done_files, speed: job.speed, eta: job.speed > 0 ? (job.total_bytes - job.done_bytes) / job.speed : 0,
    active: [], resumed: 0, current: job.current,
  })));
  outgoing.forEach(job => cards.push(jobCard({
    id: job.id, direction: 'out', name: job.name, peer: job.peer, status: job.status, error: job.error,
    totalBytes: job.total_bytes, totalFiles: job.total_files, doneBytes: job.done_bytes, doneFiles: job.done_files,
    speed: job.speed, eta: job.eta, elapsed: job.elapsed, active: job.active, resumed: job.resumed_bytes,
  })));
  $('transfers').replaceChildren(...(cards.length ? cards : [el('div', 'empty compact', 'Brak aktywnych transferów.')]));
  $('clearFinished').classList.toggle('hidden', !outgoing.some(j => !['running', 'queued'].includes(j.status)));

  outgoing.forEach(job => {
    const key = `${job.id}:${job.status}`;
    if (notified.has(key)) return;
    if (job.status === 'done') toast(`Wysłano „${job.name}” do ${job.peer}.`, 'success');
    if (job.status === 'failed') toast(`Błąd wysyłania „${job.name}”: ${job.error}`, 'error');
    if (['done', 'failed'].includes(job.status)) notified.add(key);
  });
}

function renderHistory(history) {
  if (!history.length) {
    $('history').replaceChildren(el('div', 'empty compact', 'Historia jest pusta.'));
    return;
  }
  $('history').replaceChildren(...history.slice(0, 50).map(entry => {
    const row = el('div', 'history-row');
    const sent = entry.direction === 'sent';
    row.append(
      el('span', `job-dir ${sent ? 'out' : 'in'}`, sent ? '⬆' : '⬇'),
      el('span', '', entry.name),
      el('span', 'h-peer muted', `${sent ? 'do' : 'od'} ${entry.peer || '—'}`),
      el('span', 'h-size muted', formatBytes(entry.bytes)),
      el('span', `status ${entry.status}`, STATUS_LABEL[entry.status] || entry.status),
      el('span', 'h-time muted', formatTime(entry.finished)),
    );
    row.title = [entry.location, entry.error].filter(Boolean).join('\n');
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
      renderHistory(history);
    }
    showCode(result.pairing_code);
  } finally {
    polling = false;
  }
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
  if (!r.ok) toast(r.error, 'error');
});
$('newCode').addEventListener('click', async () => {
  const r = await newPairingCode();
  if (r.ok) {
    showCode(r.pairing_code);
    toast('Wygenerowano nowy kod. Dotychczas sparowani nadawcy muszą sparować się ponownie.');
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

loadInfo();
refreshDevices();
renderSelection();
updateSummary();
pollTransfers();
setInterval(refreshDevices, 5000);
setInterval(pollTransfers, 500);
