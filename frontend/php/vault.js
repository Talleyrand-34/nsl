// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

// Credential-vault control for the app bar. The vault is a single server-side
// store: a master passphrase unlocks the data key (held in server memory) so
// stored SSH secrets can be encrypted/decrypted without re-entering it. This
// widget reflects the live status and offers init / unlock / lock, talking to
// the Go API directly (window.NSL_API, CORS open).
(function () {
    function api() { return (window.NSL_API || '').replace(/\/+$/, ''); }

    function els() {
        return {
            box:    document.getElementById('vault-status'),
            action: document.getElementById('vault-action'),
            lock:   document.getElementById('vault-lock'),
        };
    }

    function render(s) {
        var e = els();
        if (!e.box) return;
        window.NSL_VAULT = s;
        if (!s || s.error) {
            e.box.textContent = 'Vault: unavailable';
            e.box.className = 'vault-pill vault-err';
            e.action.style.display = 'none';
            e.lock.style.display = 'none';
            return;
        }
        if (!s.initialized) {
            e.box.textContent = 'Vault: not set up';
            e.box.className = 'vault-pill vault-uninit';
            e.action.textContent = 'Set passphrase';
            e.action.style.display = '';
            e.lock.style.display = 'none';
        } else if (s.unlocked) {
            e.box.textContent = 'Vault: unlocked';
            e.box.className = 'vault-pill vault-open';
            e.action.style.display = 'none';
            e.lock.style.display = '';
        } else {
            e.box.textContent = 'Vault: locked';
            e.box.className = 'vault-pill vault-locked';
            e.action.textContent = 'Unlock';
            e.action.style.display = '';
            e.lock.style.display = 'none';
        }
    }

    function refresh() {
        if (!api()) { render({ error: true }); return; }
        fetch(api() + '/vault/status', { method: 'GET' })
            .then(function (r) { return r.json(); })
            .then(render)
            .catch(function () { render({ error: true }); });
    }

    function post(path, body) {
        return fetch(api() + path, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body || {}),
        }).then(function (r) {
            return r.json().catch(function () { return {}; }).then(function (j) {
                return { ok: r.ok, body: j };
            });
        });
    }

    // The action button initializes a fresh vault or unlocks an existing one,
    // depending on the current status.
    window.nslVaultAction = function () {
        var s = window.NSL_VAULT || {};
        if (!s.initialized) {
            var p1 = prompt('Set a master passphrase for the credential vault:');
            if (!p1) return;
            var p2 = prompt('Confirm the master passphrase:');
            if (p1 !== p2) { alert('Passphrases do not match.'); return; }
            post('/vault/init', { passphrase: p1 }).then(function (res) {
                if (!res.ok) alert('Could not initialize vault: ' + (res.body.message || 'error'));
                refresh();
            });
            return;
        }
        var pass = prompt('Master passphrase to unlock the credential vault:');
        if (!pass) return;
        post('/vault/unlock', { passphrase: pass }).then(function (res) {
            if (!res.ok) alert('Could not unlock vault: ' + (res.body.message || 'wrong passphrase'));
            refresh();
        });
    };

    window.nslVaultLock = function () {
        post('/vault/lock', {}).then(refresh);
    };

    window.nslVaultRefresh = refresh;
    document.addEventListener('DOMContentLoaded', refresh);
})();
