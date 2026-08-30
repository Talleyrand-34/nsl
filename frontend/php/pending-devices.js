// SPDX-License-Identifier: MIT
// pending-devices.js — browser-side queue of discovered-but-not-yet-imported
// devices.
//
// Storage: two localStorage keys:
//   "nsl.pendingDevices"  = JSON array of discovered-device objects.
//   "nsl.pendingDeleted"  = JSON array of IPs the operator explicitly
//                            deleted. NSL_PENDING_INITIAL on the next page
//                            load must skip these IPs (otherwise Delete
//                            gets undone by the very next reload).
//
// Lifecycle:
//   - A scan completes → dispatcher emits NSL_PENDING_INITIAL → JS init
//     merges into localStorage (de-duped by IP, tombstones skipped).
//   - The operator clicks "Configure & import" → existing flow.
//   - The operator clicks "Import device" → server returns success →
//     page reloads → JS init drops the imported IP from localStorage.
//   - The operator clicks "Delete" → JS adds the IP to tombstones +
//     removes from queue. Tombstone survives page reloads so the
//     server's fresh discovery (NSL_PENDING_INITIAL) is filtered to
//     exclude that IP.

(function () {
  var KEY_QUEUE = 'nsl.pendingDevices';
  var KEY_TOMBSTONES = 'nsl.pendingDeleted';

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

  function readTombstones() {
    var raw = readJSON(KEY_TOMBSTONES);
    var set = {};
    raw.forEach(function (ip) { if (ip) set[ip] = true; });
    return set;
  }

  function addTombstone(ip) {
    if (!ip) return;
    var raw = readJSON(KEY_TOMBSTONES);
    if (raw.indexOf(ip) === -1) raw.push(ip);
    writeJSON(KEY_TOMBSTONES, raw);
  }

  function clearTombstones() {
    writeJSON(KEY_TOMBSTONES, []);
  }

  // Public surface — exposed on window so the page can wire per-row
  // buttons without an extra round-trip through inline handlers.
  var nslPD = {
    KEY_QUEUE: KEY_QUEUE,
    KEY_TOMBSTONES: KEY_TOMBSTONES,
    list: readQueue,
    tombstones: readTombstones,

    // Merge fresh discovery into the queue. Skips:
    //   - IPs already in the queue (dedupe vs. existing local state)
    //   - IPs in the tombstone set (operator explicitly deleted; respect it)
    addAll: function (devices) {
      var existing = readQueue();
      var tombstones = readTombstones();
      var seen = {};
      existing.forEach(function (d) { var ip = ipOf(d); if (ip) seen[ip] = true; });
      devices.forEach(function (d) {
        var ip = ipOf(d);
        if (!ip) return;
        if (seen[ip]) return;
        if (tombstones[ip]) return;
        existing.push(d);
        seen[ip] = true;
      });
      writeQueue(existing);
      return existing;
    },

    add: function (device) { return nslPD.addAll([device]); },

    // Mark device deleted: remove from queue AND tombstone the IP so
    // future server-side discoveries don't re-add it on page reload.
    remove: function (device) {
      var ip = ipOf(device);
      if (!ip) return readQueue();
      nslPD._tombstone(ip);
      var list = readQueue().filter(function (d) { return ipOf(d) !== ip; });
      writeQueue(list);
      return list;
    },

    // Drop a tombstone (re-allows the device to come back on next scan).
    _untombstone: function (ip) {
      if (!ip) return;
      var raw = readJSON(KEY_TOMBSTONES).filter(function (x) { return x !== ip; });
      writeJSON(KEY_TOMBSTONES, raw);
    },

    // Record a tombstone for an IP. Exposed so Delete handlers can call it
    // directly; nslPD.remove also tombstones for convenience.
    _tombstone: addTombstone,

    removeByIp: function (ip) {
      if (!ip) return readQueue();
      var list = readQueue().filter(function (d) { return ipOf(d) !== ip; });
      writeQueue(list);
      return list;
    },

    clear: function () { writeQueue([]); return []; },
    clearTombstones: clearTombstones,

    // syncFromServer merges one server-rendered discovery payload into the
    // queue and repaints the table + heading counts. The auto-init below and
    // the "Reset queue + tombstones" button both go through this, so the
    // merge/render/count logic exists exactly once.
    syncFromServer: function (initial, importedIps) {
      var body = document.querySelector('table.discovered-devices tbody');
      if (!body) return 0;

      var imported = Array.isArray(importedIps) ? importedIps : [];
      imported.forEach(function (ip) { if (ip) nslPD.removeByIp(ip); });

      // Filter the server's fresh discovery against tombstones BEFORE adding.
      // Without this, every page load re-adds devices the operator already
      // deleted.
      var tombstones = readTombstones();
      var list = Array.isArray(initial) ? initial : [];
      var filtered = list.filter(function (d) {
        var ip = ipOf(d);
        return ip && !tombstones[ip];
      });
      if (filtered.length) nslPD.addAll(filtered);

      nslPD.render(body, imported);
      nslPD.updateCounts(body);
      return readQueue().length;
    },

    // updateCounts refreshes the "(N total, M pending)" heading from whatever
    // rows are currently in the table body.
    updateCounts: function (tableBody) {
      var total = 0, pending = 0;
      Array.prototype.forEach.call(tableBody.querySelectorAll('tr'), function (tr) {
        if (tr.children.length !== 7) return; // placeholder row has 1 col-span cell
        total++;
        var status = (tr.children[0].textContent || '').trim();
        if (status === 'pending') pending++;
      });
      var tEl = document.querySelector('.discovered-total');
      var pEl = document.querySelector('.discovered-pending');
      if (tEl) tEl.textContent = total;
      if (pEl) pEl.textContent = pending;
    },

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
            nslPD.updateCounts(tableBody);
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
  //   2. Filter NSL_PENDING_INITIAL against tombstones (so a Delete
  //      survives page reloads).
  //   3. Merge fresh discovery (de-duped by IP, tombstones skipped).
  //   4. Re-render the table from localStorage.
  //   5. Update the heading counts.
  document.addEventListener('DOMContentLoaded', function () {
    nslPD.syncFromServer(window.NSL_PENDING_INITIAL, window.NSL_PENDING_IMPORTED_IPS);
  });
})();