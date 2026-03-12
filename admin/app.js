const API = 'http://localhost:8080';

// ── Etat ──────────────────────────────────────────────────────
let forcedOffline  = false;
let chart          = null;
let chartLabels    = [];
let chartPending   = [];
let chartSynced    = [];
const MAX_POINTS   = 20;

// ── Horloge ───────────────────────────────────────────────────
function updateClock() {
  const el = document.getElementById('footer-clock');
  if (el) el.textContent = new Date().toLocaleString('fr-FR');
}
setInterval(updateClock, 1000);
updateClock();

// ── Admin ─────────────────────────────────────────────────────
function setAdmin(name, role) {
  const parts    = name.trim().split(' ');
  const initials = parts.length >= 2
    ? parts[0][0] + parts[parts.length - 1][0]
    : name.slice(0, 2);
  const avatarEl = document.getElementById('admin-avatar');
  const nameEl   = document.getElementById('admin-name');
  const roleEl   = document.getElementById('admin-role');
  if (avatarEl) avatarEl.textContent = initials.toUpperCase();
  if (nameEl)   nameEl.textContent   = name;
  if (roleEl)   roleEl.textContent   = role;
}

setAdmin('Norberto AMETOSSINA', 'Administrateur Systeme');

// ── Graphique Chart.js ────────────────────────────────────────
function initChart() {
  const ctx = document.getElementById('transactionChart').getContext('2d');
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: chartLabels,
      datasets: [
        {
          label: 'En attente',
          data: chartPending,
          borderColor: '#3b82f6',
          backgroundColor: 'rgba(59,130,246,0.1)',
          borderWidth: 2,
          fill: true,
          tension: 0.4,
          pointBackgroundColor: '#3b82f6',
          pointRadius: 4,
          pointHoverRadius: 6,
        },
        {
          label: 'Synchronisees',
          data: chartSynced,
          borderColor: '#10b981',
          backgroundColor: 'rgba(16,185,129,0.08)',
          borderWidth: 2,
          fill: true,
          tension: 0.4,
          pointBackgroundColor: '#10b981',
          pointRadius: 4,
          pointHoverRadius: 6,
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { mode: 'index', intersect: false },
      plugins: {
        legend: { display: false },
        tooltip: {
          backgroundColor: '#1a2332',
          borderColor: '#2a3a50',
          borderWidth: 1,
          titleColor: '#e2e8f0',
          bodyColor: '#94a3b8',
          titleFont: { family: 'JetBrains Mono', size: 11 },
          bodyFont:  { family: 'JetBrains Mono', size: 11 },
          padding: 10,
        }
      },
      scales: {
        x: {
          grid: { color: 'rgba(42,58,80,0.6)' },
          ticks: {
            color: '#4a6070',
            font: { family: 'JetBrains Mono', size: 10 },
            maxRotation: 0,
          },
          border: { color: '#2a3a50' }
        },
        y: {
          beginAtZero: true,
          grid: { color: 'rgba(42,58,80,0.6)' },
          ticks: {
            color: '#4a6070',
            font: { family: 'JetBrains Mono', size: 10 },
            stepSize: 1,
          },
          border: { color: '#2a3a50' }
        }
      }
    }
  });
}

// Ajouter un point au graphique
function updateChart(pending, synced) {
  const now = new Date().toLocaleTimeString('fr-FR');
  chartLabels.push(now);
  chartPending.push(pending);
  chartSynced.push(synced);

  // Garder seulement MAX_POINTS points
  if (chartLabels.length > MAX_POINTS) {
    chartLabels.shift();
    chartPending.shift();
    chartSynced.shift();
  }

  if (chart) chart.update('none');
}

// ── Toggle ONLINE / OFFLINE ───────────────────────────────────
function toggleMode() {
  forcedOffline = !forcedOffline;
  const btn     = document.getElementById('btn-toggle');
  const pill    = document.getElementById('status-pill');
  const banner  = document.getElementById('status-banner');
  const bannerT = document.getElementById('banner-text');
  const statSt  = document.getElementById('stat-status');
  const statSub = document.getElementById('stat-sub');
  const time    = document.getElementById('banner-time');

  if (forcedOffline) {
    btn.textContent = 'Passer en ONLINE';
    btn.classList.remove('offline-mode');
    pill.className   = 'pill offline';
    pill.textContent = 'OFFLINE';
    banner.className = 'banner banner-offline';
    bannerT.textContent = 'MODE HORS-LIGNE — Transactions stockees dans SQLite local';
    statSt.textContent  = 'OFFLINE';
    statSt.className    = 'stat-value danger';
    statSub.textContent = 'Serveur distant inaccessible';
    showToast('Mode OFFLINE active — transactions mises en file.', 'warning');
  } else {
    btn.textContent = 'Passer en OFFLINE';
    btn.classList.add('offline-mode');
    pill.className   = 'pill online';
    pill.textContent = 'ONLINE';
    banner.className = 'banner banner-online';
    bannerT.textContent = 'MODE EN LIGNE — Transactions transmises directement au serveur';
    statSt.textContent  = 'ONLINE';
    statSt.className    = 'stat-value success';
    statSub.textContent = 'Serveur distant accessible';
    showToast('Mode ONLINE active — synchronisation en cours...', 'success');
    setTimeout(loadAll, 1000);
  }

  time.textContent = new Date().toLocaleTimeString('fr-FR');
}

// ── Charger tout ──────────────────────────────────────────────
async function loadAll() {
  await loadStatus();
  await loadQueue();
  await loadLogs();
}

// ── Statut ────────────────────────────────────────────────────
async function loadStatus() {
  if (forcedOffline) return;
  try {
    const res  = await fetch(`${API}/status`);
    const data = await res.json();
    const pill    = document.getElementById('status-pill');
    const banner  = document.getElementById('status-banner');
    const bannerT = document.getElementById('banner-text');
    const statSt  = document.getElementById('stat-status');
    const statSub = document.getElementById('stat-sub');
    const statPen = document.getElementById('stat-pending');
    const time    = document.getElementById('banner-time');

    if (data.status === 'ONLINE') {
      pill.className    = 'pill online';
      pill.textContent  = 'ONLINE';
      banner.className  = 'banner banner-online';
      bannerT.textContent = 'MODE EN LIGNE — Transactions transmises directement au serveur';
      statSt.textContent  = 'ONLINE';
      statSt.className    = 'stat-value success';
      statSub.textContent = 'Serveur distant accessible';
    } else {
      pill.className    = 'pill offline';
      pill.textContent  = 'OFFLINE';
      banner.className  = 'banner banner-offline';
      bannerT.textContent = 'MODE HORS-LIGNE — Transactions stockees dans SQLite local';
      statSt.textContent  = 'OFFLINE';
      statSt.className    = 'stat-value danger';
      statSub.textContent = 'Serveur distant inaccessible';
    }

    time.textContent = new Date().toLocaleTimeString('fr-FR');
    if (statPen) statPen.textContent = data.pending ?? 0;

  } catch {
    document.getElementById('stat-status').textContent = 'Erreur';
    document.getElementById('stat-sub').textContent    = 'Middleware inaccessible';
  }
}

// ── File FIFO ─────────────────────────────────────────────────
async function loadQueue() {
  try {
    const res   = await fetch(`${API}/queue`);
    const data  = await res.json();
    const tbody = document.getElementById('queue-body');
    const count = document.getElementById('fifo-count');

    if (!data || data.length === 0) {
      tbody.innerHTML = `<tr>
        <td colspan="6" class="empty">
          Aucune transaction en attente — file FIFO vide.
        </td>
      </tr>`;
      if (count) count.textContent = '0 en attente';
      document.getElementById('stat-synced').textContent = '0';
      updateChart(0, 0);
      return;
    }

    const pending = data.filter(t => t.Status === 'pending').length;
    const synced  = data.filter(t => t.Status === 'synced').length;
    if (count) count.textContent = `${pending} en attente`;

    document.getElementById('stat-synced').textContent = synced;
    document.getElementById('stat-pending').textContent = pending;

    // Mettre a jour le graphique
    updateChart(pending, synced);

    tbody.innerHTML = data.map(t => {
      const badge = t.Status === 'synced'
        ? `<span class="badge badge-success">Synchronisee</span>`
        : `<span class="badge badge-warning">En attente</span>`;
      const body = t.Body
        ? (t.Body.length > 45 ? t.Body.slice(0, 45) + '...' : t.Body)
        : '—';
      return `<tr>
        <td style="font-family:var(--mono);color:var(--text-dim);
            font-size:.8rem">#${t.ID}</td>
        <td><span class="method method-${t.Method}">${t.Method}</span></td>
        <td style="font-family:var(--mono);font-size:.78rem;
            color:var(--blue)">${t.Endpoint}</td>
        <td style="font-family:var(--mono);font-size:.72rem;
            color:var(--text-soft)">${body}</td>
        <td style="font-size:.78rem;color:var(--text-soft)">
            ${formatDate(t.CreatedAt)}</td>
        <td>${badge}</td>
      </tr>`;
    }).join('');

  } catch {
    document.getElementById('queue-body').innerHTML =
      `<tr><td colspan="6" class="empty">Erreur de chargement.</td></tr>`;
  }
}

// ── Logs ──────────────────────────────────────────────────────
async function loadLogs() {
  try {
    const res  = await fetch(`${API}/logs`);
    const data = await res.json();
    const container = document.getElementById('logs-container');

    if (!data || data.length === 0) {
      container.innerHTML = `<div class="empty">Aucun evenement enregistre.</div>`;
      return;
    }

    container.innerHTML = data.map(l => `
      <div class="log-entry">
        <span class="log-msg">${l.message}</span>
        <span class="log-date">${formatDate(l.date)}</span>
      </div>`).join('');

    const last = document.getElementById('stat-last-sync');
    if (last && data[0]) last.textContent = formatDate(data[0].date);

  } catch {
    document.getElementById('logs-container').innerHTML =
      `<div class="empty">Erreur de chargement des logs.</div>`;
  }
}

// ── Synchronisation manuelle ──────────────────────────────────
async function triggerSync() {
  const btn = document.getElementById('btn-sync');
  btn.textContent = 'Synchronisation en cours...';
  btn.disabled = true;
  try {
    const res  = await fetch(`${API}/sync`, { method: 'POST' });
    const data = await res.json();
    showToast(data.message || 'Synchronisation lancee.', 'success');
  } catch {
    showToast('Erreur — verifiez que le middleware tourne sur :8080', 'error');
  } finally {
    btn.textContent = 'Synchroniser maintenant';
    btn.disabled = false;
    setTimeout(loadAll, 1500);
  }
}

// ── Toast ─────────────────────────────────────────────────────
function showToast(message, type = 'info') {
  const stack = document.getElementById('toast-stack');
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.textContent = message;
  stack.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity   = '0';
    toast.style.transform = 'translateX(20px)';
    toast.style.transition = 'all .3s';
    setTimeout(() => toast.remove(), 300);
  }, 4000);
}

// ── Date ──────────────────────────────────────────────────────
function formatDate(dateStr) {
  if (!dateStr) return '—';
  try { return new Date(dateStr).toLocaleString('fr-FR'); }
  catch { return dateStr; }
}

// ── Lancement ─────────────────────────────────────────────────
initChart();
loadAll();
setInterval(loadAll, 5000);