import './style.css';

import {EventsOn} from '../wailsjs/runtime/runtime';
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
          <h2>Issue pools</h2>
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

EventsOn('scanProgress', (payload) => {
  if (!payload) {
    return;
  }
  const percent = payload.Total > 0 ? Math.round((payload.Current / payload.Total) * 100) : 0;
  progressBar.value = percent;
  progressLabel.textContent = `Scanning ${payload.Host}:${payload.Port} (${payload.Current}/${payload.Total})`;
  scanStatus.textContent = 'Scanning pools...';
});

EventsOn('scanComplete', (payload) => {
  if (!payload) {
    return;
  }
  progressLabel.textContent = `Completed with ${payload.errorCount} issue(s)`;
  scanStatus.textContent = payload.errorCount > 0 ? 'Finished with issues' : 'Scan complete';
});

function renderEntries(container, entries, emptyMessage) {
  if (!entries || entries.length === 0) {
    container.innerHTML = `<li class="empty">${emptyMessage}</li>`;
    return;
  }
  container.innerHTML = entries
    .map((entry) => {
      const latest = entry?.Host?.Latest;
      const hostLabel = latest?.Host ? `${latest.Host}:${latest.Port}` : entry?.Host?.Host ?? 'unknown';
      const ping = latest?.Ping ?? 'n/a';
      const badge = latest?.PanelClass === 'panel-bad' ? 'issue' : 'good';
      return `
        <li>
          <div class="entry-header">
            <span class="badge ${badge}">${entry.PoolName}</span>
            <span class="entry-ping">${ping}</span>
          </div>
          <p class="entry-meta">${hostLabel} · ${latest?.RewardNote ?? ''}</p>
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
  refreshSummary(view);
  renderEntries(cleanList, view.CleanEntries, 'No clean entries');
  renderEntries(issueList, view.IssueEntries, 'No issue entries');
}

async function loadLastReport() {
  try {
    const lastView = await LastReport();
    if (lastView?.CleanEntries?.length || lastView?.IssueEntries?.length) {
      renderView(lastView);
      scanStatus.textContent = 'Last scan loaded';
    }
  } catch (err) {
    console.error('failed to read last report', err);
  }
}

loadLastReport();

scanButton.addEventListener('click', async () => {
  scanButton.disabled = true;
  progressBar.value = 0;
  progressLabel.textContent = 'Starting scan…';
  scanStatus.textContent = 'Initializing...';

  try {
    const view = await StartScan(3);
    if (view) {
      renderView(view);
    }
  } catch (err) {
    progressLabel.textContent = 'Scan failed';
    scanStatus.textContent = `Error: ${err?.message || err}`;
    console.error(err);
  } finally {
    scanButton.disabled = false;
  }
});
