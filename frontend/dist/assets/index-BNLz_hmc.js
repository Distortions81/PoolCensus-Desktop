(function(){const n=document.createElement("link").relList;if(n&&n.supports&&n.supports("modulepreload"))return;for(const t of document.querySelectorAll('link[rel="modulepreload"]'))r(t);new MutationObserver(t=>{for(const o of t)if(o.type==="childList")for(const a of o.addedNodes)a.tagName==="LINK"&&a.rel==="modulepreload"&&r(a)}).observe(document,{childList:!0,subtree:!0});function s(t){const o={};return t.integrity&&(o.integrity=t.integrity),t.referrerPolicy&&(o.referrerPolicy=t.referrerPolicy),t.crossOrigin==="use-credentials"?o.credentials="include":t.crossOrigin==="anonymous"?o.credentials="omit":o.credentials="same-origin",o}function r(t){if(t.ep)return;t.ep=!0;const o=s(t);fetch(t.href,o)}})();function C(e,n,s){return window.runtime.EventsOnMultiple(e,n,s)}function g(e,n){return C(e,n,-1)}function L(){return window.go.main.PoolCensusApp.LastReport()}function E(e){return window.go.main.PoolCensusApp.StartScan(e)}const y=document.getElementById("app");y.innerHTML=`
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
`;const l=document.getElementById("scanButton"),f=document.getElementById("progressBar"),c=document.getElementById("progressLabel"),i=document.getElementById("scanStatus"),I=document.getElementById("cleanList"),S=document.getElementById("issueList"),p=document.getElementById("summaryText"),b=document.getElementById("cleanCount"),v=document.getElementById("issueCount");g("scanProgress",e=>{if(!e)return;const n=e.Total>0?Math.round(e.Current/e.Total*100):0;f.value=n,c.textContent=`Scanning ${e.Host}:${e.Port} (${e.Current}/${e.Total})`,i.textContent="Scanning pools..."});g("scanComplete",e=>{e&&(c.textContent=`Completed with ${e.errorCount} issue(s)`,i.textContent=e.errorCount>0?"Finished with issues":"Scan complete")});function m(e,n,s){if(!n||n.length===0){e.innerHTML=`<li class="empty">${s}</li>`;return}e.innerHTML=n.map(r=>{var u,d;const t=(u=r==null?void 0:r.Host)==null?void 0:u.Latest,o=t!=null&&t.Host?`${t.Host}:${t.Port}`:((d=r==null?void 0:r.Host)==null?void 0:d.Host)??"unknown",a=(t==null?void 0:t.Ping)??"n/a";return`
        <li>
          <div class="entry-header">
            <span class="badge ${(t==null?void 0:t.PanelClass)==="panel-bad"?"issue":"good"}">${r.PoolName}</span>
            <span class="entry-ping">${a}</span>
          </div>
          <p class="entry-meta">${o} · ${(t==null?void 0:t.RewardNote)??""}</p>
        </li>
      `}).join("")}function x(e){if(!e){p.textContent="Last scan: no data";return}const n=(e.CleanEntries||[]).length,s=(e.IssueEntries||[]).length;b.textContent=String(n),v.textContent=String(s),p.textContent=`Last scan: ${n+s} pools (${s} issue${s===1?"":"s"})`}function h(e){x(e),m(I,e.CleanEntries,"No clean entries"),m(S,e.IssueEntries,"No issue entries")}async function $(){var e,n;try{const s=await L();((e=s==null?void 0:s.CleanEntries)!=null&&e.length||(n=s==null?void 0:s.IssueEntries)!=null&&n.length)&&(h(s),i.textContent="Last scan loaded")}catch(s){console.error("failed to read last report",s)}}$();l.addEventListener("click",async()=>{l.disabled=!0,f.value=0,c.textContent="Starting scan…",i.textContent="Initializing...";try{const e=await E(3);e&&h(e)}catch(e){c.textContent="Scan failed",i.textContent=`Error: ${(e==null?void 0:e.message)||e}`,console.error(e)}finally{l.disabled=!1}});
