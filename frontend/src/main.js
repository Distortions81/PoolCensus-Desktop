import './style.css';

import {BrowserOpenURL, EventsOn} from '../wailsjs/runtime/runtime';
import {LastReport, StartScan} from '../wailsjs/go/main/PoolCensusApp';

const app = document.getElementById('app');

app.innerHTML = `
  <div class="page">
    <header class="hero">
      <div>
        <p class="eyebrow">Latency-focused mining telemetry</p>
        <h1>PoolCensus</h1>
        <p class="lead">Scan public stratum endpoints, compare latencies, and spot payout issues without leaving the desktop.</p>
      </div>
      <div class="cta">
        <button class="btn" id="scanButton">Start scan</button>
        <p class="status" id="scanStatus">Idle</p>
      </div>
    </header>

    <section class="progress-card">
      <div class="progress-meta">
        <p id="progressLabel">Waiting to scan...</p>
        <p class="muted" id="progressHint">Press "Start scan" to contact the pools list.</p>
      </div>
      <progress id="progressBar" max="100" value="0"></progress>
    </section>

    <section class="results">
      <article class="panel">
        <header>
          <h2>Clean pools</h2>
          <p class="muted"><span id="cleanCount">0</span> entries</p>
        </header>
        <ul class="list" id="cleanList">
          <li class="empty">No data yet</li>
        </ul>
      </article>
      <article class="panel panel-alt">
        <header>
          <h2>Pools With Issues</h2>
          <p class="muted"><span id="issueCount">0</span> entries</p>
        </header>
        <ul class="list" id="issueList">
          <li class="empty">Issues surface here</li>
        </ul>
      </article>
    </section>

    <section class="summary">
      <p class="muted" id="summaryText">Last scan: none</p>
    </section>
  </div>

  <div class="modal-overlay hidden" id="detailOverlay" aria-hidden="true">
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="detailTitle">
      <header class="modal-header">
        <h2 id="detailTitle">Pool details</h2>
        <button class="btn ghost" id="detailCloseButton" type="button">Close</button>
      </header>
      <div class="modal-body" id="detailBody"></div>
    </div>
  </div>
`;

const scanButton = document.getElementById('scanButton');
const progressBar = document.getElementById('progressBar');
const progressLabel = document.getElementById('progressLabel');
const scanStatus = document.getElementById('scanStatus');
const cleanList = document.getElementById('cleanList');
const issueList = document.getElementById('issueList');
const summaryText = document.getElementById('summaryText');
const cleanCount = document.getElementById('cleanCount');
const issueCount = document.getElementById('issueCount');
const detailOverlay = document.getElementById('detailOverlay');
const detailBody = document.getElementById('detailBody');
const detailTitle = document.getElementById('detailTitle');
const detailCloseButton = document.getElementById('detailCloseButton');

let currentView = null;

function wailsReady() {
  return (
    typeof window !== 'undefined' &&
    window.runtime &&
    window.go &&
    window.go.main &&
    window.go.main.PoolCensusApp
  );
}

function setUiError(message) {
  progressLabel.textContent = message;
  scanStatus.textContent = message;
}

function initWailsBindings() {
  EventsOn('scanProgress', (payload) => {
    if (!payload) {
      return;
    }
    const total = payload.total || 0;
    const current = payload.current || 0;
    const percent = total > 0 ? Math.round((current / total) * 100) : 0;
    progressBar.value = percent;
    progressLabel.textContent = `Scanning ${payload.host}:${payload.port} (${current}/${total})`;
    scanStatus.textContent = 'Scanning pools...';
  });

  EventsOn('scanComplete', async (payload) => {
    if (!payload) {
      return;
    }
    progressLabel.textContent = `Completed with ${payload.errorCount} issue(s)`;
    scanStatus.textContent = payload.errorCount > 0 ? 'Finished with issues' : 'Scan complete';
    scanButton.disabled = false;

    try {
      const view = await LastReport();
      if (view) {
        renderView(view);
      }
    } catch (err) {
      console.error('failed to refresh report', err);
    }
  });

  EventsOn('scanError', (payload) => {
    const message = payload && payload.message ? payload.message : 'Scan failed';
    setUiError(message);
    scanButton.disabled = false;
  });
}

function escapeHtml(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function openDetails(hostEntry) {
  if (!hostEntry || !hostEntry.Host || !hostEntry.Host.Latest) {
    return;
  }

  const latest = hostEntry.Host.Latest;
  const poolName = hostEntry.PoolName || 'Pool details';
  const address = latest.Host ? `${latest.Host}:${latest.Port}` : 'unknown';
  const rewardNote = latest.RewardNote || '';

  const details = [
    `<div class="detail-row"><span class="detail-key">Endpoint</span><span class="detail-value mono">${escapeHtml(address)}</span></div>`,
    `<div class="detail-row"><span class="detail-key">Ping</span><span class="detail-value mono">${escapeHtml(latest.Ping || 'n/a')}</span></div>`,
    `<div class="detail-row"><span class="detail-key">Total payout</span><span class="detail-value mono">${escapeHtml(
      typeof latest.TotalPayout === 'number' ? `${latest.TotalPayout.toFixed(8)} BTC` : 'n/a',
    )}</span></div>`,
  ];

  if (latest.TLS) {
    details.unshift(
      `<div class="detail-row"><span class="detail-key">TLS</span><span class="detail-value"><span class="badge tls">TLS</span></span></div>`,
    );
  }

  if (typeof latest.WorkerPercent === 'number' && latest.WorkerPercent > 0) {
    details.push(
      `<div class="detail-row"><span class="detail-key">Worker share</span><span class="detail-value mono">${escapeHtml(
        `${latest.WorkerPercent.toFixed(2)}%`,
      )}</span></div>`,
    );
  }

  if (rewardNote) {
    details.push(
      `<div class="detail-row"><span class="detail-key">Note</span><span class="detail-value">${escapeHtml(
        rewardNote,
      )}</span></div>`,
    );
  }

  if (!latest.Connected) {
    details.push(
      `<div class="detail-row"><span class="detail-key">Status</span><span class="detail-value">Disconnected</span></div>`,
    );
    if (latest.Error) {
      details.push(
        `<div class="detail-row"><span class="detail-key">Last error</span><span class="detail-value mono">${escapeHtml(
          latest.Error,
        )}</span></div>`,
      );
    }
  }

  const issues = Array.isArray(latest.Issues) ? latest.Issues : [];
  if (issues.length > 0) {
    details.push(
      `<div class="detail-section">
        <div class="detail-key">Issues</div>
        <ul class="detail-list">
          ${issues
            .map((issue) => `<li>${escapeHtml(issue && issue.Message ? issue.Message : 'Issue')}</li>`)
            .join('')}
        </ul>
      </div>`,
    );
  }

  const scanUrl = latest.ScanURL && latest.ScanURL !== '#' ? latest.ScanURL : '';
  if (scanUrl) {
    details.push(
      `<div class="detail-actions"><button class="btn ghost" id="openDetailUrl" type="button">Open details</button></div>`,
    );
  }

  detailTitle.textContent = poolName;
  detailBody.innerHTML = details.join('');

  if (scanUrl) {
    const openButton = document.getElementById('openDetailUrl');
    if (openButton) {
      openButton.addEventListener('click', () => BrowserOpenURL(scanUrl), {once: true});
    }
  }

  detailOverlay.classList.remove('hidden');
  detailOverlay.setAttribute('aria-hidden', 'false');
}

function closeDetails() {
  detailOverlay.classList.add('hidden');
  detailOverlay.setAttribute('aria-hidden', 'true');
  detailBody.innerHTML = '';
}

function attachEntryHandlers(container, listKey) {
  container.addEventListener('click', (event) => {
    const item = event.target.closest('li[data-entry-index]');
    if (!item || !container.contains(item)) {
      return;
    }
    const index = Number(item.getAttribute('data-entry-index'));
    const entries = currentView ? currentView[listKey] : null;
    if (!entries || Number.isNaN(index) || index < 0 || index >= entries.length) {
      return;
    }
    openDetails(entries[index]);
  });

  container.addEventListener('keydown', (event) => {
    if (event.key !== 'Enter' && event.key !== ' ') {
      return;
    }
    const item = event.target.closest('li[data-entry-index]');
    if (!item || !container.contains(item)) {
      return;
    }
    event.preventDefault();
    item.click();
  });
}

function renderEntries(container, entries, emptyMessage, listKey) {
  if (!entries || entries.length === 0) {
    container.innerHTML = `<li class="empty">${emptyMessage}</li>`;
    return;
  }
  container.innerHTML = entries
    .map((entry, index) => {
      const latest = entry && entry.Host && entry.Host.Latest ? entry.Host.Latest : null;
      const hostLabel =
        latest && latest.Host
          ? `${latest.Host}:${latest.Port}`
          : entry && entry.Host && entry.Host.Host
            ? entry.Host.Host
            : 'unknown';
      const ping = latest && latest.Ping != null ? latest.Ping : 'n/a';
      const badge = latest && latest.PanelClass === 'panel-bad' ? 'issue' : 'good';
      const rewardNote = latest && latest.RewardNote ? latest.RewardNote : '';
      const metaParts = [hostLabel];
      if (rewardNote) {
        metaParts.push(rewardNote);
      }
      const tlsTag = latest && latest.TLS ? '<span class="badge tls">TLS</span>' : '';
      return `
        <li role="button" tabindex="0" data-entry-index="${index}" data-entry-kind="${listKey}">
          <div class="entry-header">
            <span class="badge ${badge}">${entry.PoolName}</span>
            <span class="entry-tags">
              ${tlsTag}
              <span class="entry-ping">${ping}</span>
            </span>
          </div>
          <p class="entry-meta">${metaParts.join(' · ')}</p>
        </li>
      `;
    })
    .join('');
}

function refreshSummary(view) {
  if (!view) {
    summaryText.textContent = 'Last scan: no data';
    return;
  }
  const totalClean = (view.CleanEntries || []).length;
  const totalIssues = (view.IssueEntries || []).length;
  cleanCount.textContent = String(totalClean);
  issueCount.textContent = String(totalIssues);
  summaryText.textContent = `Last scan: ${totalClean + totalIssues} pools (${totalIssues} issue${totalIssues === 1 ? '' : 's'})`;
}

function renderView(view) {
  currentView = view;
  refreshSummary(view);
  renderEntries(cleanList, view.CleanEntries, 'No clean entries', 'CleanEntries');
  renderEntries(issueList, view.IssueEntries, 'No issue entries', 'IssueEntries');
}

async function loadLastReport() {
  try {
    const lastView = await LastReport();
    if (
      lastView &&
      ((lastView.CleanEntries && lastView.CleanEntries.length > 0) ||
        (lastView.IssueEntries && lastView.IssueEntries.length > 0))
    ) {
      renderView(lastView);
      scanStatus.textContent = 'Last scan loaded';
    }
  } catch (err) {
    console.error('failed to read last report', err);
  }
}

loadLastReport();

attachEntryHandlers(cleanList, 'CleanEntries');
attachEntryHandlers(issueList, 'IssueEntries');

detailCloseButton.addEventListener('click', closeDetails);
detailOverlay.addEventListener('click', (event) => {
  if (event.target === detailOverlay) {
    closeDetails();
  }
});
document.addEventListener('keydown', (event) => {
  if (event.key === 'Escape' && !detailOverlay.classList.contains('hidden')) {
    closeDetails();
  }
});

scanButton.addEventListener('click', async () => {
  scanButton.disabled = true;
  progressBar.value = 0;
  progressLabel.textContent = 'Starting scan…';
  scanStatus.textContent = 'Initializing...';

  try {
    if (!wailsReady()) {
      setUiError('Backend not ready yet (waiting for Wails runtime)');
      scanButton.disabled = false;
      return;
    }
    await StartScan(3);
  } catch (err) {
    progressLabel.textContent = 'Scan failed';
    scanStatus.textContent = `Error: ${(err && err.message) || err}`;
    console.error(err);
    scanButton.disabled = false;
  }
});

// On older macOS WebViews, Wails may inject the runtime after the module executes.
// Wait briefly for the runtime before registering EventsOn handlers.
(function waitForWails() {
  if (wailsReady()) {
    initWailsBindings();
    scanStatus.textContent = 'Ready';
    return;
  }

  const start = Date.now();
  (function poll() {
    if (wailsReady()) {
      initWailsBindings();
      scanStatus.textContent = 'Ready';
      return;
    }
    if (Date.now() - start > 10000) {
      setUiError('Wails runtime not available. This build may be incompatible with this macOS/WebView.');
      return;
    }
    setTimeout(poll, 50);
  })();
})();
