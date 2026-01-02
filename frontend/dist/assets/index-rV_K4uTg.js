(function(){const s=document.createElement("link").relList;if(s&&s.supports&&s.supports("modulepreload"))return;for(const e of document.querySelectorAll('link[rel="modulepreload"]'))o(e);new MutationObserver(e=>{for(const r of e)if(r.type==="childList")for(const c of r.addedNodes)c.tagName==="LINK"&&c.rel==="modulepreload"&&o(c)}).observe(document,{childList:!0,subtree:!0});function n(e){const r={};return e.integrity&&(r.integrity=e.integrity),e.referrerPolicy&&(r.referrerPolicy=e.referrerPolicy),e.crossOrigin==="use-credentials"?r.credentials="include":e.crossOrigin==="anonymous"?r.credentials="omit":r.credentials="same-origin",r}function o(e){if(e.ep)return;e.ep=!0;const r=n(e);fetch(e.href,r)}})();function E(t,s,n){return window.runtime.EventsOnMultiple(t,s,n)}function u(t,s){return E(t,s,-1)}function h(){return window.go.main.PoolCensusApp.LastReport()}function b(t){return window.go.main.PoolCensusApp.StartScan(t)}const x=document.getElementById("app");x.innerHTML=`
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
`;const i=document.getElementById("scanButton"),C=document.getElementById("progressBar"),l=document.getElementById("progressLabel"),a=document.getElementById("scanStatus"),S=document.getElementById("cleanList"),I=document.getElementById("issueList"),m=document.getElementById("summaryText"),v=document.getElementById("cleanCount"),B=document.getElementById("issueCount");function d(){return typeof window!="undefined"&&window.runtime&&window.go&&window.go.main&&window.go.main.PoolCensusApp}function p(t){l.textContent=t,a.textContent=t}function f(){u("scanProgress",t=>{if(!t)return;const s=t.total||0,n=t.current||0,o=s>0?Math.round(n/s*100):0;C.value=o,l.textContent=`Scanning ${t.host}:${t.port} (${n}/${s})`,a.textContent="Scanning pools..."}),u("scanComplete",async t=>{if(t){l.textContent=`Completed with ${t.errorCount} issue(s)`,a.textContent=t.errorCount>0?"Finished with issues":"Scan complete",i.disabled=!1;try{const s=await h();s&&w(s)}catch(s){console.error("failed to refresh report",s)}}}),u("scanError",t=>{const s=t&&t.message?t.message:"Scan failed";p(s),i.disabled=!1})}function g(t,s,n){if(!s||s.length===0){t.innerHTML=`<li class="empty">${n}</li>`;return}t.innerHTML=s.map(o=>{const e=o&&o.Host&&o.Host.Latest?o.Host.Latest:null,r=e&&e.Host?`${e.Host}:${e.Port}`:o&&o.Host&&o.Host.Host?o.Host.Host:"unknown",c=e&&e.Ping!=null?e.Ping:"n/a",y=e&&e.PanelClass==="panel-bad"?"issue":"good",L=e&&e.RewardNote?e.RewardNote:"";return`
        <li>
          <div class="entry-header">
            <span class="badge ${y}">${o.PoolName}</span>
            <span class="entry-ping">${c}</span>
          </div>
          <p class="entry-meta">${r} · ${L}</p>
        </li>
      `}).join("")}function $(t){if(!t){m.textContent="Last scan: no data";return}const s=(t.CleanEntries||[]).length,n=(t.IssueEntries||[]).length;v.textContent=String(s),B.textContent=String(n),m.textContent=`Last scan: ${s+n} pools (${n} issue${n===1?"":"s"})`}function w(t){$(t),g(S,t.CleanEntries,"No clean entries"),g(I,t.IssueEntries,"No issue entries")}async function P(){try{const t=await h();t&&(t.CleanEntries&&t.CleanEntries.length>0||t.IssueEntries&&t.IssueEntries.length>0)&&(w(t),a.textContent="Last scan loaded")}catch(t){console.error("failed to read last report",t)}}P();i.addEventListener("click",async()=>{i.disabled=!0,C.value=0,l.textContent="Starting scan…",a.textContent="Initializing...";try{if(!d()){p("Backend not ready yet (waiting for Wails runtime)"),i.disabled=!1;return}await b(3)}catch(t){l.textContent="Scan failed",a.textContent=`Error: ${t&&t.message||t}`,console.error(t),i.disabled=!1}});(function(){if(d()){f(),a.textContent="Ready";return}const s=Date.now();(function n(){if(d()){f(),a.textContent="Ready";return}if(Date.now()-s>1e4){p("Wails runtime not available. This build may be incompatible with this macOS/WebView.");return}setTimeout(n,50)})()})();
