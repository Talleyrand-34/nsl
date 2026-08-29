// SPDX-License-Identifier: MIT
// pending-devices.js — browser-side queue of discovered-but-not-yet-imported
// devices.
//
// Storage: localStorage["nsl.pendingDevices"] = JSON array of
// discovered-device objects. Each entry matches what the server returns
// from /scan/run / /scan/status ({"device":{"ip":...}, "brand":...,
// "model":..., "model_type":..., "suggested_name":...}).
//
// Lifecycle:
//   - A scan completes → dispatcher emits NSL_PENDING_INITIAL = the array
//     → JS init merges into localStorage (de-duped by IP).
//   - The operator clicks "Configure & import" → existing flow.
//   - The operator clicks "Import device" → server returns success →
//     page reloads → JS init drops the imported IP from localStorage.
//   - The operator clicks "Delete" → JS removes the device from localStorage.
//     Gone is gone; the device can come back only via a fresh scan + the
//     operator's fresh decision.

(function () {
  var KEY_QUEUE = 'nsl.pendingDevices';

  function readJSON(key) {
    try {
      var raw = localStorage.getItem(key);
      if (!raw) return [];
      var parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch (e) {
      console.warn(key + ': failed to read, resetting', e);
      return [];
    }
  }

  function writeJSON(key, value) {
    try { localStorage.setItem(key, JSON.stringify(value)); }
    catch (e) { console.warn(key + ': failed to write', e); }
  }

  function ipOf(device) {
    return (device && device.device && device.device.ip) || (device && device.ip) || '';
  }

  function readQueue() { return readJSON(KEY_QUEUE); }
  function writeQueue(list) { writeJSON(KEY_QUEUE, list); }

  // Public surface — exposed on window so the page can wire per-row
  // buttons without an extra round-trip through inline handlers.
  var nslPD = {
    KEY_QUEUE: KEY_QUEUE,
    list: readQueue,

    // Merge fresh discovery into the queue. Skips IPs already in the
    // queue (dedupe vs. existing local state).
    addAll: function (devices) {
      var existing = readQueue();
      var seen = {};
      existing.forEach(function (d) { var ip = ipOf(d); if (ip) seen[ip] = true; });
      devices.forEach(function (d) {
        var ip = ipOf(d);
        if (!ip) return;
        if (seen[ip]) return;
        existing.push(d);
        seen[ip] = true;
      });
      writeQueue(existing);
      return existing;
    },

    add: function (device) { return nslPD.addAll([device]); },

    // Delete = remove from queue. No tombstone, no resurrection: the next
    // scan starts fresh.
    remove: function (device) {
      var ip = ipOf(device);
      if (!ip) return readQueue();
      var list = readQueue().filter(function (d) { return ipOf(d) !== ip; });
      writeQueue(list);
      return list;
    },

    removeByIp: function (ip) {
      if (!ip) return readQueue();
      var list = readQueue().filter(function (d) { return ipOf(d) !== ip; });
      writeQueue(list);
      return list;
    },

    clear: function () { writeQueue([]); return []; },

    render: function (tableBody, importedIps) {
      var list = readQueue();
      var imported = {};
      (importedIps || []).forEach(function (ip) { if (ip) imported[ip] = true; });
      tableBody.innerHTML = '';
      if (list.length === 0) {
        var row = document.createElement('tr');
        var cell = document.createElement('td');
        cell.colSpan = 7;
        cell.style.color = '#777';
        cell.style.textAlign = 'center';
        cell.textContent = 'No pending devices. Run a scan from the left column to populate the queue.';
        row.appendChild(cell);
        tableBody.appendChild(row);
        return list.length;
      }
      list.forEach(function (d) {
        var ip = ipOf(d);
        var done = imported[ip];
        var row = document.createElement('tr');
        if (done) row.style.cssText = 'color:#888; background:#f3f3f3;';
        appendCell(row, done ? '✓ imported' : 'pending');
        appendCell(row, d.suggested_name || '');
        appendCell(row, ip);
        appendCell(row, d.brand || '');
        appendCell(row, d.model || '');
        appendCell(row, d.model_type || '');
        var action = document.createElement('td');
        if (!done) {
          var form = document.createElement('form');
          form.method = 'post';
          form.action = 'import.php';
          form.style.margin = '0';
          form.style.display = 'inline';
          var hidden = document.createElement('input');
          hidden.type = 'hidden';
          hidden.name = 'device_json';
          hidden.value = JSON.stringify(d);
          form.appendChild(hidden);
          var cfgBtn = document.createElement('button');
          cfgBtn.type = 'submit';
          cfgBtn.name = 'do_analyze';
          cfgBtn.value = '1';
          cfgBtn.textContent = 'Configure & import →';
          form.appendChild(cfgBtn);
          action.appendChild(form);

          var delBtn = document.createElement('button');
          delBtn.type = 'button';
          delBtn.style.marginLeft = '6px';
          delBtn.textContent = 'Delete';
          delBtn.addEventListener('click', function () {
            if (!confirm('Remove ' + ip + ' from the pending queue?')) return;
            nslPD.remove(d);
            nslPD.render(tableBody, importedIps);
          });
          action.appendChild(delBtn);
        } else {
          action.textContent = '—';
        }
        row.appendChild(action);
        tableBody.appendChild(row);
      });
      return list.length;
    },
  };

  function appendCell(row, text) {
    var cell = document.createElement('td');
    cell.textContent = text;
    row.appendChild(cell);
  }

  window.nslPD = nslPD;

  // Auto-init: when the page loads:
  //   1. Drop imported devices from the queue.
  //   2. Merge fresh discovery (de-duped by IP).
  //   3. Re-render the table from localStorage.
  //   4. Update the heading counts.
  document.addEventListener('DOMContentLoaded', function () {
    var body = document.querySelector('table.discovered-devices tbody');
    if (!body) return;

    var imported = Array.isArray(window.NSL_PENDING_IMPORTED_IPS) ? window.NSL_PENDING_IMPORTED_IPS : [];
    imported.forEach(function (ip) { if (ip) nslPD.removeByIp(ip); });

    if (Array.isArray(window.NSL_PENDING_INITIAL)) {
      nslPD.addAll(window.NSL_PENDING_INITIAL);
    }

    nslPD.render(body, imported);

    var total = 0, pending = 0;
    Array.prototype.forEach.call(body.querySelectorAll('tr'), function (tr) {
      if (tr.children.length !== 7) return; // placeholder row has 1 col-span cell
      total++;
      var status = (tr.children[0].textContent || '').trim();
      if (status === 'pending') pending++;
    });
    var tEl = document.querySelector('.discovered-total');
    var pEl = document.querySelector('.discovered-pending');
    if (tEl) tEl.textContent = total;
    if (pEl) pEl.textContent = pending;
  });
})();