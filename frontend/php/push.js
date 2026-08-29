// push.js — client-side logic for the Push config tab (push.php).
//
// Five verbs map to /api/v1/push/*:
//
//   Preview  → POST /push/preview   {device_id, os, intent, observed}
//   Dry-run  → POST /push/apply     {device_id, os, mode:"dry-run"}
//   Apply    → POST /push/apply     {device_id, os, mode:"apply"}
//   Rollback → POST /push/rollback  {device_id}
//   History  → GET  /push/history?device_id=
//
// The device dropdown drives a separate topology fetch (phase 6 wires
// the right-column D2 render). For now the topology handler returns the
// raw neighbors list and the panel just shows the count.
//
// History auto-refreshes every 5s when a device is selected.
(function () {
  "use strict";

  function el(id) { return document.getElementById(id); }

  function api(path) {
    var base = window.NSL_API || "";
    return base.replace(/\/$/, "") + path;
  }

  function postJSON(path, body) {
    return fetch(api(path), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body || {}),
    }).then(function (r) {
      if (!r.ok) {
        return r.text().then(function (t) {
          throw new Error("HTTP " + r.status + ": " + t);
        });
      }
      return r.json();
    });
  }

  function getJSON(path) {
    return fetch(api(path)).then(function (r) {
      if (!r.ok) {
        return r.text().then(function (t) {
          throw new Error("HTTP " + r.status + ": " + t);
        });
      }
      return r.json();
    });
  }

  function showResult(text) {
    var pre = el("push-result-text");
    if (!pre) return;
    pre.textContent = text || "(no changes)";
    pre.style.display = "";
  }

  function showError(msg) {
    var pre = el("push-result-text");
    if (!pre) return;
    pre.textContent = "ERROR: " + msg;
    pre.style.display = "";
  }

  function clearResult() {
    var pre = el("push-result-text");
    if (pre) { pre.style.display = "none"; pre.textContent = ""; }
  }

  function selectedDevice() {
    var sel = el("push-device");
    return sel ? sel.value : "";
  }

  function selectedOS() {
    // Read OS from the data attr on the option (set by PHP) so we don't
    // need a second round-trip. Falls back to "" when the option was
    // built from a /devices response that lacks the field.
    var sel = el("push-device");
    if (!sel || !sel.selectedOptions[0]) return "";
    return sel.selectedOptions[0].getAttribute("data-os") || "";
  }

  function readIntent() {
    var ta = el("push-intent");
    if (!ta) return {};
    var raw = (ta.value || "").trim();
    if (!raw) return {};
    try { return JSON.parse(raw); }
    catch (e) {
      throw new Error("intent is not valid JSON: " + e.message);
    }
  }

  function renderHistory(runs) {
    var box = el("push-history");
    if (!box) return;
    if (!runs || runs.length === 0) {
      box.innerHTML = '<em style="color:#777;">No runs yet.</em>';
      return;
    }
    var rows = [];
    rows.push('<table style="width:100%; border-collapse:collapse;">');
    rows.push('<thead><tr style="background:#f0f0f0;">' +
              '<th align="left">Time</th>' +
              '<th align="left">Mode</th>' +
              '<th align="left">OS</th>' +
              '<th align="left">Status</th>' +
              '</tr></thead><tbody>');
    runs.forEach(function (r) {
      var when = (r.StartedAt || r.started_at || "");
      var mode = r.Mode || r.mode || "";
      var os = r.OS || r.os || "";
      var status = r.ExitStatus || r.exit_status || (r.Result || r.result || "?");
      rows.push('<tr>' +
                '<td>' + escapeHTML(when) + '</td>' +
                '<td>' + escapeHTML(mode) + '</td>' +
                '<td>' + escapeHTML(os) + '</td>' +
                '<td>' + escapeHTML(status) + '</td>' +
                '</tr>');
    });
    rows.push('</tbody></table>');
    box.innerHTML = rows.join("");
  }

  function escapeHTML(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  // Phase 6: scoped topology. On device change, fetch /push/topology and
  // render a hand-built SVG (center node + N neighbor nodes + edge labels
  // for ports/VLANs) plus the neighbors table. No D2 dependency — keeps
  // the JS testable without a server-side D2 install.
  function renderTopology(topo) {
    var svg = el("push-topology");
    if (!svg) return;
    var neighbors = topo.neighbors || topo.Neighbor || [];
    var center = topo.device_id || topo.DeviceID || "";
    svg.innerHTML = "";
    if (neighbors.length === 0) {
      svg.innerHTML = '<em style="color:#777;">No connections for ' +
        escapeHTML(center) + '.</em>';
      return;
    }
    var w = 360, h = 60 + neighbors.length * 50;
    var s = '<svg viewBox="0 0 ' + w + ' ' + h + '" ' +
      'xmlns="http://www.w3.org/2000/svg" style="width:100%; height:auto;">';
    s += '<defs><style>' +
      '.push-node{fill:#eef;font-family:sans-serif;font-size:12px;stroke:#888;}' +
      '.push-center{fill:#ffd;font-weight:bold;stroke:#a80;}' +
      '.push-edge{stroke:#888;stroke-width:1.5;fill:none;}' +
      '.push-label{font-family:sans-serif;font-size:10px;fill:#444;}' +
      '</style></defs>';
    var cy = 30, cx = w / 2;
    s += '<rect class="push-node push-center" x="' + (cx - 60) + '" y="' + (cy - 15) +
      '" width="120" height="30" rx="4"/>';
    s += '<text class="push-label" x="' + cx + '" y="' + (cy + 5) +
      '" text-anchor="middle">' + escapeHTML(center) + '</text>';
    var r = 90;
    neighbors.forEach(function (n, i) {
      var angle = (Math.PI * 2 * i) / neighbors.length - Math.PI / 2;
      var nx = cx + r * Math.cos(angle);
      var ny = cy + 50 + r * 0.6 * Math.sin(angle);
      s += '<line class="push-edge" x1="' + cx + '" y1="' + (cy + 15) +
        '" x2="' + nx + '" y2="' + ny + '"/>';
      s += '<rect class="push-node" x="' + (nx - 50) + '" y="' + (ny - 12) +
        '" width="100" height="24" rx="3"/>';
      s += '<text class="push-label" x="' + nx + '" y="' + (ny + 4) +
        '" text-anchor="middle">' + escapeHTML(n.neighbor || n.Neighbor || "?") + '</text>';
    });
    s += '</svg>';
    svg.innerHTML = s;
  }

  function renderNeighborsTable(topo) {
    var tbody = el("push-neighbors-body");
    if (!tbody) return;
    var neighbors = topo.neighbors || topo.Neighbor || [];
    if (neighbors.length === 0) {
      tbody.innerHTML = '<tr><td colspan="4" style="color:#777; font-style:italic;">&mdash;</td></tr>';
      return;
    }
    var rows = [];
    neighbors.forEach(function (n) {
      var name = n.neighbor || n.Neighbor || "";
      var os = n.neighbor_os || n.NeighborOS || "?";
      var ports = (n.this_port || n.ThisPort || "?") + " → " + (n.neighbor_port || n.NeighborPort || "?");
      var vlans = (n.vlans || n.Vlans || []).join(", ");
      rows.push('<tr>' +
        '<td>' + escapeHTML(name) + '</td>' +
        '<td>' + escapeHTML(os) + '</td>' +
        '<td>' + escapeHTML(ports) + '</td>' +
        '<td>' + escapeHTML(vlans) + '</td>' +
        '</tr>');
    });
    tbody.innerHTML = rows.join("");
  }

  function refreshTopology() {
    var dev = selectedDevice();
    if (!dev) return Promise.resolve();
    return getJSON("/push/topology?device_id=" + encodeURIComponent(dev))
      .then(function (topo) {
        renderTopology(topo);
        renderNeighborsTable(topo);
      })
      .catch(function (err) {
        var svg = el("push-topology");
        if (svg) svg.innerHTML = '<em style="color:#a00;">topology error: ' +
          escapeHTML(err.message) + '</em>';
      });
  }

  function refreshHistory() {
    var dev = selectedDevice();
    if (!dev) return Promise.resolve();
    return getJSON("/push/history?device_id=" + encodeURIComponent(dev))
      .then(function (resp) {
        renderHistory(resp.runs || resp.Runs || []);
      })
      .catch(function (err) {
        // 404 / empty store is fine; surface other errors softly.
        var box = el("push-history");
        if (box) box.innerHTML = '<em style="color:#a00;">history error: ' + escapeHTML(err.message) + '</em>';
      });
  }

  function doPreview() {
    var dev = selectedDevice();
    var os = selectedOS();
    if (!dev) { showError("Pick a device first."); return; }
    if (!os) { showError("Selected device has no OS set; can't preview."); return; }
    var intent, observed;
    try { intent = readIntent(); }
    catch (e) { showError(e.message); return; }
    clearResult();
    postJSON("/push/preview", {
      device_id: dev, os: os,
      intent: intent, observed: {},
    }).then(function (resp) {
      showResult(resp.patch_text || resp.PatchText || "");
    }).catch(function (err) { showError(err.message); });
  }

  function doApply(mode) {
    var dev = selectedDevice();
    var os = selectedOS();
    if (!dev) { showError("Pick a device first."); return; }
    if (!os) { showError("Selected device has no OS set."); return; }
    clearResult();
    postJSON("/push/apply", { device_id: dev, os: os, mode: mode })
      .then(function (resp) {
        showResult(mode + " → run " + (resp.run_id || resp.RunID || "?") +
                   " (" + (resp.exit_status || resp.ExitStatus || "?") + ")");
        refreshHistory();
      })
      .catch(function (err) { showError(err.message); });
  }

  function doRollback() {
    var dev = selectedDevice();
    if (!dev) { showError("Pick a device first."); return; }
    clearResult();
    postJSON("/push/rollback", { device_id: dev })
      .then(function (resp) {
        showResult("rollback → run " + (resp.run_id || resp.RunID || "?"));
        refreshHistory();
      })
      .catch(function (err) { showError(err.message); });
  }

  function onDeviceChange() {
    var sel = el("push-device");
    var label = el("push-device-os");
    var os = selectedOS();
    if (label) label.textContent = os || "—";
    if (sel) {
      // Pre-fill intent textarea with a stub the operator can edit.
      var ta = el("push-intent");
      if (ta) ta.value = JSON.stringify({ hostname: sel.value, intent_note: "(stub) edit and Apply to push" }, null, 2);
    }
    refreshHistory();
    refreshTopology();
    // Phase 6 hook: fire a custom event the topology panel listens for.
    document.dispatchEvent(new CustomEvent("nsl:push-device-selected", {
      detail: { device_id: selectedDevice(), os: os },
    }));
  }

  function wireUp() {
    var sel = el("push-device");
    if (sel) sel.addEventListener("change", onDeviceChange);
    var pb = el("push-preview"); if (pb) pb.addEventListener("click", doPreview);
    var pa = el("push-apply");   if (pa) pa.addEventListener("click", function () { doApply("apply"); });
    var pd = el("push-dryrun");  if (pd) pd.addEventListener("click", function () { doApply("dry-run"); });
    var pr = el("push-rollback");if (pr) pr.addEventListener("click", doRollback);

    // Initial state.
    onDeviceChange();
    // Auto-refresh history every 5s when a device is selected.
    setInterval(function () {
      if (selectedDevice()) refreshHistory();
    }, 5000);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", wireUp);
  } else {
    wireUp();
  }
})();
