// scan-status.js — live status panel for async scans.
//
// A page that starts an async scan renders a #scan-status panel and calls
// nslWatchScan(scanId, opts). This polls GET <NSL_API>/scan/status?scan_id=&since=
// (~1s) directly against the Go API (CORS is open), appends new granular events,
// updates the progress bar, and on completion/failure reloads the page with
// ?scan_id=<id> so PHP renders the result (or error) with its existing markup.
(function () {
  "use strict";

  function el(id) { return document.getElementById(id); }

  function levelClass(level) {
    switch (level) {
      case "error": return "ev-error";
      case "warn": return "ev-warn";
      case "debug": return "ev-debug";
      default: return "ev-info";
    }
  }

  function fmtFields(f) {
    if (!f) return "";
    var parts = [];
    for (var k in f) {
      if (Object.prototype.hasOwnProperty.call(f, k)) parts.push(k + "=" + f[k]);
    }
    return parts.length ? "  " + parts.join(" ") : "";
  }

  window.nslWatchScan = function (scanId, opts) {
    opts = opts || {};
    var api = window.NSL_API || "";
    var since = 0;
    var stateEl = el("scan-state");
    var barEl = el("scan-bar");
    var fillEl = el("scan-bar-fill");
    var eventsEl = el("scan-events");

    function appendEvents(events) {
      if (!eventsEl || !events) return;
      events.forEach(function (e) {
        var line = document.createElement("div");
        line.className = "scan-ev " + levelClass(e.level);
        line.textContent = "[" + e.level + "] " + e.msg + fmtFields(e.fields);
        eventsEl.appendChild(line);
      });
      eventsEl.scrollTop = eventsEl.scrollHeight;
    }

    function setProgress(done, total) {
      if (!barEl) return;
      if (total > 0) {
        barEl.style.display = "";
        var pct = Math.round((done / total) * 100);
        if (fillEl) fillEl.style.width = pct + "%";
      }
    }

    function done(scanId) {
      var sep = window.location.search ? "" : "";
      var url = window.location.pathname + "?scan_id=" + encodeURIComponent(scanId) +
        (opts.reloadParams ? "&" + opts.reloadParams : "");
      setTimeout(function () { window.location = url; }, 500);
    }

    function poll() {
      fetch(api + "/scan/status?scan_id=" + encodeURIComponent(scanId) + "&since=" + since)
        .then(function (r) { return r.json(); })
        .then(function (st) {
          if (!st || (!st.state && st.error)) {
            if (stateEl) stateEl.textContent = "error: " + (st && st.message ? st.message : "scan not found");
            return;
          }
          if (typeof st.last_seq === "number") since = st.last_seq;
          setProgress(st.done || 0, st.total || 0);
          if (stateEl) {
            stateEl.textContent = st.state +
              (st.total ? "  ·  " + (st.done || 0) + "/" + st.total : "") +
              (st.elapsed ? "  ·  " + st.elapsed : "");
          }
          appendEvents(st.events);
          if (st.state === "completed" || st.state === "failed") {
            if (st.state === "failed" && stateEl) {
              stateEl.textContent = "failed: " + (st.error || "");
            }
            done(scanId);
            return;
          }
          setTimeout(poll, 1000);
        })
        .catch(function () { setTimeout(poll, 2000); });
    }
    poll();
  };
})();
