let target = '';
let targetName = '';
let selectedPaths = [];
let remotePath = '';
let remoteParent = '';
let destination = '';
let previewedPath = '';

async function loadInfo() {
  const info = await appInfo();
  document.getElementById('me').textContent = `${info.name} · odbiór gotowy`;
}

let refreshingDevices = false;

async function refreshDevices() {
  if (refreshingDevices) return;
  refreshingDevices = true;
  const box = document.getElementById('devices');
  try {
    const result = await discoverDevices();
    if (!result.ok) throw new Error(result.error || 'Nieznany błąd wykrywania');
    const devices = result.devices || [];
    box.innerHTML = '';
    if (!devices.length) {
      if (!box.querySelector('.device')) {
        box.innerHTML = '<div class="muted">Nie znaleziono innych urządzeń. Uruchom LAN Transfer na drugim komputerze.</div>';
      }
      return;
    }
    for (const device of devices) {
      const element = document.createElement('button');
      element.className = 'device';
      element.type = 'button';
      element.innerHTML = `<div class="device-main"><b>${escapeHtml(device.name)}</b><span>${escapeHtml(device.ip)}:${device.port}</span></div><div class="device-arrow">›</div>`;
      if (device.ip === target) element.classList.add('active');
      element.addEventListener('click', async () => {
        document.querySelectorAll('.device').forEach(item => item.classList.remove('active'));
        element.classList.add('active');
        target = device.ip;
        targetName = device.name;
        remotePath = '';
        destination = '';
        document.getElementById('remoteHint').classList.add('hidden');
        document.getElementById('remoteBrowser').classList.remove('hidden');
        await loadRemote('');
        updateState();
      });
      box.appendChild(element);
    }
  } catch (error) {
    if (!box.querySelector('.device')) {
      box.innerHTML = `<div class="muted">Błąd wykrywania: ${escapeHtml(String(error))}</div>`;
    }
  } finally {
    refreshingDevices = false;
  }
}

function clearTarget() {
  target = '';
  targetName = '';
  remotePath = '';
  remoteParent = '';
  destination = '';
  document.getElementById('remoteHint').classList.remove('hidden');
  document.getElementById('remoteBrowser').classList.add('hidden');
  updateState();
}

async function chooseFiles() {
  const result = await pickLocalFiles();
  if (!result.ok) return setStatus(`Błąd: ${result.error}`);
  addPaths(result.paths || []);
}

async function chooseFolder() {
  const result = await pickLocalFolder();
  if (!result.ok) return setStatus(`Błąd: ${result.error}`);
  if (result.path) addPaths([result.path]);
}

function addPaths(paths) {
  for (const path of paths) {
    if (path && !selectedPaths.includes(path)) selectedPaths.push(path);
  }
  renderSelection();
  updateState();
}

window.__addDroppedPaths = function(paths) {
  if (!Array.isArray(paths) || !paths.length) return;
  addPaths(paths);
  setStatus(paths.length === 1 ? 'Dodano przeciągnięty element.' : `Dodano ${paths.length} przeciągniętych elementów.`);
};

function renderSelection() {
  const box = document.getElementById('selection');
  if (!selectedPaths.length) {
    box.innerHTML = '<div class="empty">Nic jeszcze nie wybrano.</div>';
    return;
  }
  box.innerHTML = '';
  selectedPaths.forEach((path, index) => {
    const row = document.createElement('div');
    row.className = 'selected-item';
    if (path === previewedPath) row.classList.add('preview-active');
    row.innerHTML = `<div class="file-icon">${pathLooksLikeFolder(path) ? '▣' : '▤'}</div><div class="selected-text"><b>${escapeHtml(baseName(path))}</b><span>${escapeHtml(path)}</span></div><button class="remove-btn" title="Usuń">×</button>`;
    row.addEventListener('click', () => loadPreview(path));
    row.querySelector('.remove-btn').addEventListener('click', event => {
      event.stopPropagation();
      selectedPaths.splice(index, 1);
      if (previewedPath === path) closePreview();
      renderSelection();
      updateState();
    });
    box.appendChild(row);
  });
}

async function loadPreview(path) {
  previewedPath = path;
  renderSelection();
  const panel = document.getElementById('previewPanel');
  const body = document.getElementById('previewBody');
  panel.classList.remove('hidden');
  document.getElementById('previewTitle').textContent = baseName(path);
  document.getElementById('previewMeta').textContent = path;
  body.innerHTML = '<div class="preview-message">Wczytywanie podglądu…</div>';

  try {
    const result = await previewLocalPath(path);
    if (!result.ok) throw new Error(result.error || 'Nie można utworzyć podglądu');
    if (previewedPath !== path) return;
    renderPreview(result.preview);
  } catch (error) {
    if (previewedPath !== path) return;
    body.innerHTML = `<div class="preview-message">Błąd podglądu: ${escapeHtml(String(error))}</div>`;
  }
}

function renderPreview(info) {
  const body = document.getElementById('previewBody');
  document.getElementById('previewTitle').textContent = info.name || baseName(info.path || '');
  document.getElementById('previewMeta').textContent = previewMeta(info);

  if (info.kind === 'image' && info.data_url) {
    body.innerHTML = '';
    const image = document.createElement('img');
    image.className = 'preview-image';
    image.src = info.data_url;
    image.alt = info.name || 'Podgląd obrazu';
    body.appendChild(image);
    return;
  }

  if (info.kind === 'pdf' && info.data_url) {
    body.innerHTML = '';
    const frame = document.createElement('iframe');
    frame.className = 'preview-pdf';
    frame.src = info.data_url;
    frame.title = info.name || 'Podgląd PDF';
    body.appendChild(frame);
    return;
  }

  if (info.kind === 'text') {
    body.innerHTML = '';
    const pre = document.createElement('pre');
    pre.className = 'preview-text';
    pre.textContent = info.text || '';
    body.appendChild(pre);
    if (info.truncated) {
      const note = document.createElement('div');
      note.className = 'preview-message';
      note.textContent = 'Pokazano początek pliku.';
      body.appendChild(note);
    }
    return;
  }

  if (info.kind === 'folder') {
    body.innerHTML = '';
    const list = document.createElement('div');
    list.className = 'preview-folder-list';
    if (!(info.entries || []).length) {
      list.innerHTML = '<div class="preview-message">Folder jest pusty.</div>';
    } else {
      for (const name of info.entries || []) {
        const item = document.createElement('div');
        item.className = 'preview-folder-item';
        item.textContent = name;
        list.appendChild(item);
      }
      if (info.truncated) {
        const note = document.createElement('div');
        note.className = 'preview-message';
        note.textContent = `Pokazano pierwsze ${info.entries.length} z ${info.entry_count} elementów.`;
        list.appendChild(note);
      }
    }
    body.appendChild(list);
    return;
  }

  body.innerHTML = fileInfoHtml(info);
}

function previewMeta(info) {
  const parts = [];
  if (info.kind === 'folder') {
    parts.push(`${info.entry_count || 0} elementów`);
  } else {
    parts.push(formatBytes(info.size || 0));
  }
  if (info.mime && info.mime !== 'application/octet-stream' && info.mime !== 'inode/directory') parts.push(info.mime);
  if (info.modified) parts.push(`zmieniono ${info.modified}`);
  return parts.join(' · ');
}

function fileInfoHtml(info) {
  const message = info.message ? `<div class="preview-message">${escapeHtml(info.message)}</div>` : '';
  return `<div class="preview-file-info">
    <div class="preview-info-row"><span>Nazwa</span><span>${escapeHtml(info.name || '')}</span></div>
    <div class="preview-info-row"><span>Typ</span><span>${escapeHtml(info.mime || 'nieznany')}</span></div>
    <div class="preview-info-row"><span>Rozmiar</span><span>${escapeHtml(formatBytes(info.size || 0))}</span></div>
    <div class="preview-info-row"><span>Ścieżka</span><span>${escapeHtml(info.path || '')}</span></div>
    ${message}
  </div>`;
}

function closePreview() {
  previewedPath = '';
  document.getElementById('previewPanel').classList.add('hidden');
  document.getElementById('previewBody').innerHTML = '';
  renderSelection();
}

function formatBytes(bytes) {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  const digits = value >= 10 || unit === 0 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[unit]}`;
}

function pathLooksLikeFolder(path) {
  return !/\.[^\\/.]+$/.test(path);
}

function baseName(path) {
  const clean = path.replace(/[\\/]+$/, '');
  const parts = clean.split(/[\\/]/);
  return parts[parts.length - 1] || clean;
}

async function loadRemote(path) {
  if (!target) return;
  const entries = document.getElementById('remoteEntries');
  entries.innerHTML = '<div class="muted">Wczytywanie folderów…</div>';
  try {
    const result = await browseRemote(target, path);
    if (!result.ok) throw new Error(result.error || 'Nie można odczytać folderów');
    const listing = result.listing;
    remotePath = listing.path || '';
    remoteParent = listing.parent || '';
    document.getElementById('remotePath').textContent = remotePath || 'Dyski';
    document.getElementById('remoteUp').disabled = !remotePath;
    renderRemoteEntries(listing.entries || []);
    if (remotePath) {
      destination = remotePath;
      document.getElementById('destinationLabel').textContent = remotePath;
    } else {
      destination = '';
      document.getElementById('destinationLabel').textContent = 'Wybierz dysk lub folder';
    }
    updateState();
  } catch (error) {
    entries.innerHTML = `<div class="muted">Błąd: ${escapeHtml(String(error))}</div>`;
  }
}

function renderRemoteEntries(entries) {
  const box = document.getElementById('remoteEntries');
  box.innerHTML = '';
  if (!entries.length) {
    box.innerHTML = '<div class="empty compact">Brak podfolderów. Możesz wysłać tutaj.</div>';
    return;
  }
  entries.forEach(entry => {
    const row = document.createElement('button');
    row.className = 'folder-entry';
    row.type = 'button';
    row.innerHTML = `<span class="folder-icon">▰</span><span>${escapeHtml(entry.name)}</span><span class="device-arrow">›</span>`;
    row.addEventListener('click', () => loadRemote(entry.path));
    box.appendChild(row);
  });
}

async function sendSelected() {
  if (!selectedPaths.length || !target || !destination) return;
  const button = document.getElementById('send');
  button.disabled = true;
  setStatus(`Wysyłanie ${selectedPaths.length === 1 ? 'elementu' : selectedPaths.length + ' elementów'}…`);
  try {
    const result = await sendLocalPaths(selectedPaths, target, destination);
    if (!result.ok) throw new Error(result.error || 'Nieznany błąd transferu');
    setStatus('Gotowe. Wszystkie wybrane elementy zostały wysłane.');
  } catch (error) {
    setStatus(`Błąd: ${error}`);
  } finally {
    updateState();
  }
}

function updateState() {
  const ready = selectedPaths.length > 0 && target && destination;
  document.getElementById('send').disabled = !ready;
  const summary = document.getElementById('summary');
  if (!selectedPaths.length) {
    summary.textContent = 'Wybierz pliki lub folder do wysłania.';
  } else if (!target) {
    summary.textContent = 'Teraz wybierz komputer docelowy.';
  } else if (!destination) {
    summary.textContent = `Wybrano ${targetName}. Wskaż folder docelowy.`;
  } else {
    summary.textContent = `${selectedPaths.length} ${selectedPaths.length === 1 ? 'element' : 'elementy'} → ${targetName} → ${destination}`;
  }
}

function setStatus(text) {
  document.getElementById('status').textContent = text;
}

function escapeHtml(value) {
  const element = document.createElement('div');
  element.textContent = value;
  return element.innerHTML;
}

document.getElementById('refresh').addEventListener('click', refreshDevices);
document.getElementById('pickFiles').addEventListener('click', chooseFiles);
document.getElementById('pickFolder').addEventListener('click', chooseFolder);
document.getElementById('clearSelection').addEventListener('click', () => {
  selectedPaths = [];
  closePreview();
  renderSelection();
  updateState();
});
document.getElementById('closePreview').addEventListener('click', closePreview);
document.getElementById('remoteUp').addEventListener('click', () => loadRemote(remoteParent));
document.getElementById('remoteRefresh').addEventListener('click', () => loadRemote(remotePath));
document.getElementById('send').addEventListener('click', sendSelected);

const dropZone = document.getElementById('dropZone');
window.addEventListener('dragover', event => {
  event.preventDefault();
  dropZone.classList.add('drag-active');
});
window.addEventListener('dragleave', event => {
  if (!event.relatedTarget) dropZone.classList.remove('drag-active');
});
window.addEventListener('drop', event => {
  event.preventDefault();
  dropZone.classList.remove('drag-active');

  // Niektóre hosty WebView udostępniają ścieżkę bezpośrednio. Na Windows
  // podstawową drogą pozostaje natywny WM_DROPFILES obsługiwany po stronie Go.
  const paths = Array.from(event.dataTransfer?.files || [])
    .map(file => file.path || '')
    .filter(Boolean);
  if (paths.length) window.__addDroppedPaths(paths);
});

loadInfo();
refreshDevices();
renderSelection();
updateState();
setInterval(refreshDevices, 10000);
