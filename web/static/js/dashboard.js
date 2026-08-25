const POLL_INTERVAL_MS = 5000;

let targetsCache = [];
let hostsCache = [];

async function api(path, options = {}) {
  const res = await fetch(path, { headers: { "Content-Type": "application/json" }, ...options });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed (${res.status})`);
  }
  if (res.status === 204) return null;
  return res.json();
}

function fmtTime(iso) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString();
}

function barClass(pct) {
  if (pct >= 90) return "crit";
  if (pct >= 75) return "warn";
  return "";
}

// --- Targets ---

async function loadTargets() {
  targetsCache = await api("/api/targets");
  renderTargets();
  populateRuleTargetRef();
}

function renderTargets() {
  const body = document.getElementById("targetsBody");
  const empty = document.getElementById("targetsEmpty");
  body.innerHTML = "";

  if (targetsCache.length === 0) {
    empty.classList.remove("hidden");
    return;
  }
  empty.classList.add("hidden");

  targetsCache.forEach((t) => {
    const latest = t.latest;
    const statusBadge = latest
      ? `<span class="badge ${latest.success ? "up" : "down"}">${latest.success ? "UP" : "DOWN"}</span>`
      : `<span class="muted">pending…</span>`;
    const errorLine = latest && !latest.success && latest.error
      ? `<div class="muted" style="font-size:11px;margin-top:2px;">${escapeHtml(latest.error)}</div>`
      : "";
    const row = document.createElement("tr");
    row.innerHTML = `
      <td>${escapeHtml(t.name)}</td>
      <td class="muted">${escapeHtml(t.url)}</td>
      <td>${statusBadge}${errorLine}</td>
      <td>${latest ? latest.responseTimeMs + " ms" : "—"}</td>
      <td class="muted">${latest ? fmtTime(latest.checkedAt) : "—"}</td>
      <td><button class="danger" data-id="${t.id}" onclick="deleteTarget(${t.id})">delete</button></td>
    `;
    body.appendChild(row);
  });
}

async function deleteTarget(id) {
  await api(`/api/targets/${id}`, { method: "DELETE" });
  await loadTargets();
}

document.getElementById("targetForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const errorEl = document.getElementById("targetError");
  errorEl.textContent = "";
  try {
    await api("/api/targets", {
      method: "POST",
      body: JSON.stringify({
        name: document.getElementById("targetName").value,
        url: document.getElementById("targetUrl").value,
        intervalSeconds: Number(document.getElementById("targetInterval").value) || 60,
      }),
    });
    document.getElementById("targetForm").reset();
    document.getElementById("targetInterval").value = 60;
    await loadTargets();
  } catch (err) {
    errorEl.textContent = err.message;
  }
});

// --- Hosts ---

async function loadHosts() {
  hostsCache = await api("/api/hosts");
  renderHosts();
  populateRuleTargetRef();
}

function renderHosts() {
  const body = document.getElementById("hostsBody");
  const empty = document.getElementById("hostsEmpty");
  body.innerHTML = "";

  if (hostsCache.length === 0) {
    empty.classList.remove("hidden");
    return;
  }
  empty.classList.add("hidden");

  hostsCache.forEach((h) => {
    const m = h.latest;
    const bar = (pct) =>
      m ? `<div class="bar-track"><div class="bar-fill ${barClass(pct)}" style="width:${Math.min(pct, 100)}%"></div></div>${pct.toFixed(1)}%` : "—";
    const row = document.createElement("tr");
    row.innerHTML = `
      <td>${escapeHtml(h.name)}</td>
      <td>${m ? bar(m.cpuPercent) : "—"}</td>
      <td>${m ? bar(m.memoryPercent) : "—"}</td>
      <td>${m ? bar(m.diskPercent) : "—"}</td>
      <td class="muted">${fmtTime(h.lastSeenAt)}</td>
    `;
    body.appendChild(row);
  });
}

// --- Alert rules ---

function populateRuleTargetRef() {
  const typeSel = document.getElementById("ruleTargetType");
  const refSel = document.getElementById("ruleTargetRef");
  const isTarget = typeSel.value === "target";
  const items = isTarget ? targetsCache : hostsCache;

  refSel.innerHTML = items.map((i) => `<option value="${i.id}">${escapeHtml(i.name)}</option>`).join("");
}
document.getElementById("ruleTargetType").addEventListener("change", populateRuleTargetRef);

async function loadAlertRules() {
  const rules = await api("/api/alert-rules");
  const body = document.getElementById("rulesBody");
  const empty = document.getElementById("rulesEmpty");
  body.innerHTML = "";

  if (rules.length === 0) {
    empty.classList.remove("hidden");
    return;
  }
  empty.classList.add("hidden");

  rules.forEach((r) => {
    const opSymbol = r.operator === "gt" ? ">" : "<";
    const row = document.createElement("tr");
    row.innerHTML = `
      <td>${escapeHtml(r.name)}</td>
      <td class="muted">${r.targetType} #${r.targetRefId}</td>
      <td>${r.metric}</td>
      <td>${opSymbol} ${r.threshold}</td>
      <td><button class="danger" onclick="deleteRule(${r.id})">delete</button></td>
    `;
    body.appendChild(row);
  });
}

async function deleteRule(id) {
  await api(`/api/alert-rules/${id}`, { method: "DELETE" });
  await loadAlertRules();
}

document.getElementById("ruleForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const errorEl = document.getElementById("ruleError");
  errorEl.textContent = "";
  try {
    await api("/api/alert-rules", {
      method: "POST",
      body: JSON.stringify({
        name: document.getElementById("ruleName").value,
        targetType: document.getElementById("ruleTargetType").value,
        targetRefId: Number(document.getElementById("ruleTargetRef").value),
        metric: document.getElementById("ruleMetric").value,
        operator: document.getElementById("ruleOperator").value,
        threshold: Number(document.getElementById("ruleThreshold").value),
        webhookUrl: document.getElementById("ruleWebhook").value,
      }),
    });
    document.getElementById("ruleForm").reset();
    await loadAlertRules();
  } catch (err) {
    errorEl.textContent = err.message;
  }
});

// --- Alert history ---

async function loadAlerts() {
  const alerts = await api("/api/alerts");
  const body = document.getElementById("alertsBody");
  const empty = document.getElementById("alertsEmpty");
  body.innerHTML = "";

  if (alerts.length === 0) {
    empty.classList.remove("hidden");
    return;
  }
  empty.classList.add("hidden");

  alerts.forEach((a) => {
    const row = document.createElement("tr");
    row.innerHTML = `
      <td>${escapeHtml(a.ruleName)}</td>
      <td><span class="badge ${a.status}">${a.status.toUpperCase()}</span></td>
      <td class="muted">${escapeHtml(a.message)}</td>
      <td class="muted">${fmtTime(a.firedAt)}</td>
      <td class="muted">${a.resolvedAt ? fmtTime(a.resolvedAt) : "—"}</td>
    `;
    body.appendChild(row);
  });
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str ?? "";
  return div.innerHTML;
}

async function refreshAll() {
  await Promise.all([loadTargets(), loadHosts(), loadAlertRules(), loadAlerts()]);
}

refreshAll();
setInterval(refreshAll, POLL_INTERVAL_MS);
