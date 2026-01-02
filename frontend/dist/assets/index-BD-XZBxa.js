(function(){const s=document.createElement("link").relList;if(s&&s.supports&&s.supports("modulepreload"))return;for(const e of document.querySelectorAll('link[rel="modulepreload"]'))o(e);new MutationObserver(e=>{for(const n of e)if(n.type==="childList")for(const a of n.addedNodes)a.tagName==="LINK"&&a.rel==="modulepreload"&&o(a)}).observe(document,{childList:!0,subtree:!0});function r(e){const n={};return e.integrity&&(n.integrity=e.integrity),e.referrerPolicy&&(n.referrerPolicy=e.referrerPolicy),e.crossOrigin==="use-credentials"?n.credentials="include":e.crossOrigin==="anonymous"?n.credentials="omit":n.credentials="same-origin",n}function o(e){if(e.ep)return;e.ep=!0;const n=r(e);fetch(e.href,n)}})();function C(t,s,r){return window.runtime.EventsOnMultiple(t,s,r)}function p(t,s){return C(t,s,-1)}function L(){return window.go.main.PoolCensusApp.LastReport()}function E(t){return window.go.main.PoolCensusApp.StartScan(t)}const y=document.getElementById("app");y.innerHTML=`
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
`;const l=document.getElementById("scanButton"),m=document.getElementById("progressBar"),c=document.getElementById("progressLabel"),i=document.getElementById("scanStatus"),I=document.getElementById("cleanList"),S=document.getElementById("issueList"),u=document.getElementById("summaryText"),b=document.getElementById("cleanCount"),v=document.getElementById("issueCount");p("scanProgress",t=>{if(!t)return;const s=t.Total>0?Math.round(t.Current/t.Total*100):0;m.value=s,c.textContent=`Scanning ${t.Host}:${t.Port} (${t.Current}/${t.Total})`,i.textContent="Scanning pools..."});p("scanComplete",t=>{t&&(c.textContent=`Completed with ${t.errorCount} issue(s)`,i.textContent=t.errorCount>0?"Finished with issues":"Scan complete")});function d(t,s,r){if(!s||s.length===0){t.innerHTML=`<li class="empty">${r}</li>`;return}t.innerHTML=s.map(o=>{const e=o&&o.Host&&o.Host.Latest?o.Host.Latest:null,n=e&&e.Host?`${e.Host}:${e.Port}`:o&&o.Host&&o.Host.Host?o.Host.Host:"unknown",a=e&&e.Ping!=null?e.Ping:"n/a",f=e&&e.PanelClass==="panel-bad"?"issue":"good",h=e&&e.RewardNote?e.RewardNote:"";return`
        <li>
          <div class="entry-header">
            <span class="badge ${f}">${o.PoolName}</span>
            <span class="entry-ping">${a}</span>
          </div>
          <p class="entry-meta">${n} · ${h}</p>
        </li>
      `}).join("")}function x(t){if(!t){u.textContent="Last scan: no data";return}const s=(t.CleanEntries||[]).length,r=(t.IssueEntries||[]).length;b.textContent=String(s),v.textContent=String(r),u.textContent=`Last scan: ${s+r} pools (${r} issue${r===1?"":"s"})`}function g(t){x(t),d(I,t.CleanEntries,"No clean entries"),d(S,t.IssueEntries,"No issue entries")}async function $(){try{const t=await L();t&&(t.CleanEntries&&t.CleanEntries.length>0||t.IssueEntries&&t.IssueEntries.length>0)&&(g(t),i.textContent="Last scan loaded")}catch(t){console.error("failed to read last report",t)}}$();l.addEventListener("click",async()=>{l.disabled=!0,m.value=0,c.textContent="Starting scan…",i.textContent="Initializing...";try{const t=await E(3);t&&g(t)}catch(t){c.textContent="Scan failed",i.textContent=`Error: ${t&&t.message||t}`,console.error(t)}finally{l.disabled=!1}});
