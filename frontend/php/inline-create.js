// Inline recursive "create object" for the Dashboard add/update forms.
//
// Every composed field (a <select> that references another entity) gets a
// "+ New" button. Clicking it opens a modal with a create-form for the
// referenced entity; if that entity has its own composed fields (e.g. a Model
// needs a Brand and a Device Class) those fields get their own "+ New" buttons
// that stack another modal on top — recursively. On save the object is created
// via the Go API (CORS is open, base in window.NSL_API) and the originating
// dropdown is refreshed with the new value pre-selected, so the user never
// leaves the form they were filling in.
//
// This is purely additive: the existing server-side form submission is
// untouched; the create endpoints return only {message, name} (no id), so after
// a create we re-fetch the entity's list endpoint to refresh the dropdown.
(function () {
    'use strict';

    function apiBase() { return (window.NSL_API || '').replace(/\/+$/, ''); }

    // ---- Entity creation registry -------------------------------------------
    // path:      list (GET) + create (POST) endpoint, relative to the API base.
    // fields:    create-form schema. type ∈ text|number|checkbox|ref|iplist.
    //            a `ref` field references another entity (sub-select + "+ New").
    //            json: the JSON key sent to the API for that field.
    // labelProp / labelFn: how a list row is shown in a dropdown.
    // matchField / matchProp: the create field whose value identifies the new
    //            row, and the list property to compare it against (so the new
    //            row can be found and selected after a create). numeric: compare
    //            as numbers (VLAN ids).
    // extra:     constant keys always added to the payload.
    // transform: last-chance payload tweak (gets a refLabel(fieldName) helper
    //            returning the selected option's text of a ref field).
    var ENTITIES = {
        brand: {
            title: 'Brand', path: '/brands', labelProp: 'name',
            matchField: 'name', matchProp: 'name',
            fields: [{ name: 'name', json: 'name', label: 'Name', type: 'text', required: true }],
        },
        devclass: {
            title: 'Device Class', path: '/deviceclasses', labelProp: 'name',
            matchField: 'name', matchProp: 'name',
            fields: [{ name: 'name', json: 'name', label: 'Name', type: 'text', required: true }],
        },
        owner: {
            title: 'Owner', path: '/owners', labelProp: 'name',
            matchField: 'name', matchProp: 'name',
            fields: [{ name: 'name', json: 'name', label: 'Name', type: 'text', required: true }],
        },
        zonetype: {
            title: 'Zone Type', path: '/zonetypes', labelProp: 'name',
            matchField: 'name', matchProp: 'name',
            fields: [{ name: 'name', json: 'name', label: 'Name', type: 'text', required: true }],
        },
        connectiontype: {
            title: 'Connection Type', path: '/connectiontypes', labelProp: 'name',
            matchField: 'name', matchProp: 'name',
            fields: [{ name: 'name', json: 'name', label: 'Name', type: 'text', required: true }],
        },
        vlan: {
            title: 'VLAN', path: '/vlans',
            labelFn: function (o) { return 'VLAN ' + o.vlanid + (o.vlanname ? ' - ' + o.vlanname : ''); },
            matchField: 'vlan_id', matchProp: 'vlanid', numeric: true,
            fields: [
                { name: 'vlan_id', json: 'vlan_id', label: 'VLAN id', type: 'number', required: true },
                { name: 'vlan_name', json: 'vlan_name', label: 'Name (optional)', type: 'text' },
            ],
        },
        zone: {
            title: 'Zone', path: '/zones', labelProp: 'name',
            matchField: 'name', matchProp: 'name', extra: { father: '' },
            fields: [
                { name: 'name', json: 'name', label: 'Zone name', type: 'text', required: true },
                { name: 'fatherid', json: 'fatherid', label: 'Parent zone (optional)', type: 'ref', ref: 'zone', valueProp: 'id' },
                { name: 'owner', json: 'owner', label: 'Owner', type: 'ref', ref: 'owner', valueProp: 'name', required: true },
                { name: 'location_type', json: 'location_type', label: 'Zone type', type: 'ref', ref: 'zonetype', valueProp: 'name', required: true },
            ],
        },
        model: {
            title: 'Model', path: '/models', labelProp: 'model',
            matchField: 'model', matchProp: 'model',
            fields: [
                { name: 'model', json: 'model_name', label: 'Model name', type: 'text', required: true },
                { name: 'brand', json: 'brand_name', label: 'Brand', type: 'ref', ref: 'brand', valueProp: 'name', required: true },
                { name: 'class', json: 'device_class_name', label: 'Device class', type: 'ref', ref: 'devclass', valueProp: 'name', required: true },
            ],
        },
        device: {
            // The API create endpoint takes model_name/zone_id/zone_name and does
            // not accept IPs (added later via device ports), so this modal omits
            // them. zone_name is filled from the chosen zone option's text.
            title: 'Device', path: '/devices', labelProp: 'label',
            matchField: 'label', matchProp: 'label',
            transform: function (p, refLabel) { p.zone_name = refLabel('zoneid') || ''; },
            fields: [
                { name: 'label', json: 'label', label: 'Device label', type: 'text', required: true },
                { name: 'model', json: 'model_name', label: 'Model', type: 'ref', ref: 'model', valueProp: 'model', required: true },
                { name: 'zoneid', json: 'zone_id', label: 'Zone', type: 'ref', ref: 'zone', valueProp: 'id', required: true },
                { name: 'owner', json: 'owner', label: 'Owner', type: 'ref', ref: 'owner', valueProp: 'name', required: true },
            ],
        },
        modelport: {
            title: 'Model Port', path: '/modelports', labelProp: 'name',
            matchField: 'name', matchProp: 'name',
            fields: [
                { name: 'name', json: 'port_name', label: 'Port name', type: 'text', required: true },
                { name: 'posx', json: 'position_x', label: 'Position X', type: 'number' },
                { name: 'posy', json: 'position_y', label: 'Position Y', type: 'number' },
                { name: 'modelName', json: 'model_name', label: 'Model', type: 'ref', ref: 'model', valueProp: 'model', required: true },
                { name: 'allow_multiple_connections', json: 'allow_multiple_connections', label: 'Allow multiple connections', type: 'checkbox' },
            ],
        },
    };

    // ---- Per-action FK maps --------------------------------------------------
    // Keyed by main.php's $selectedAction (actionType + entity), so the same
    // <select> name can mean different things on different forms. Each entry:
    // selectName -> { e: entity, v: option-value property, filterBy?: sibling
    // device <select> name (model-filtered model-port lists) }. Record-picker
    // and filter selects are deliberately omitted (no "+ New" there).
    var FK_MAPS = {
        addzone: { fatherid: { e: 'zone', v: 'id' }, owner: { e: 'owner', v: 'name' }, location_type: { e: 'zonetype', v: 'name' } },
        addmodeldevice: { brand: { e: 'brand', v: 'name' }, 'class': { e: 'devclass', v: 'name' } },
        adddevice: { model: { e: 'model', v: 'model' }, zoneid: { e: 'zone', v: 'id' }, owner: { e: 'owner', v: 'name' } },
        addmodelport: { modelName: { e: 'model', v: 'model' } },
        adddeviceport: { device_id: { e: 'device', v: 'id' }, modelport_id: { e: 'modelport', v: 'id', filterBy: 'device_id' }, 'vlan_numbers[]': { e: 'vlan', v: 'vlanid' } },
        addconnections: {
            fromDevice: { e: 'device', v: 'id' }, toDevice: { e: 'device', v: 'id' },
            fromModelPort: { e: 'modelport', v: 'id', filterBy: 'fromDevice' }, toModelPort: { e: 'modelport', v: 'id', filterBy: 'toDevice' },
            connectionType: { e: 'connectiontype', v: 'name' },
        },
        updatezone: { father_zone_id: { e: 'zone', v: 'id' }, zone_type_id: { e: 'zonetype', v: 'id' }, owner_id: { e: 'owner', v: 'id' } },
        updatemodeldevice: { brand_id: { e: 'brand', v: 'id' }, device_class_id: { e: 'devclass', v: 'id' } },
        updatedevice: { model_id: { e: 'model', v: 'id' }, zone_id: { e: 'zone', v: 'id' }, owner_id: { e: 'owner', v: 'id' } },
        updatemodelport: { model_id: { e: 'model', v: 'id' } },
        updatedeviceport: { 'vlan_numbers[]': { e: 'vlan', v: 'vlanid' } },
        updateconnections: {
            from_device_id: { e: 'device', v: 'id' }, to_device_id: { e: 'device', v: 'id' },
            from_modelport_id: { e: 'modelport', v: 'id', filterBy: 'from_device_id' }, to_modelport_id: { e: 'modelport', v: 'id', filterBy: 'to_device_id' },
            connection_type: { e: 'connectiontype', v: 'name' },
        },
    };

    // ---- small helpers -------------------------------------------------------
    function el(tag, attrs, text) {
        var e = document.createElement(tag);
        if (attrs) Object.keys(attrs).forEach(function (k) { e.setAttribute(k, attrs[k]); });
        if (text != null) e.textContent = text;
        return e;
    }
    function rowLabel(entityKey, row) {
        var ent = ENTITIES[entityKey];
        return ent.labelFn ? ent.labelFn(row) : String(row[ent.labelProp]);
    }
    function getJSON(path) {
        return fetch(apiBase() + path, { method: 'GET' }).then(function (r) {
            if (!r.ok) throw new Error('list ' + path + ' failed (' + r.status + ')');
            return r.json();
        }).then(function (j) { return Array.isArray(j) ? j : []; });
    }
    function postJSON(path, body) {
        return fetch(apiBase() + path, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        }).then(function (r) {
            return r.json().catch(function () { return {}; }).then(function (j) { return { ok: r.ok, body: j }; });
        });
    }
    // Find the row a just-created entity produced, by its human identity.
    function findCreated(entityKey, rows, identity) {
        var ent = ENTITIES[entityKey];
        for (var i = 0; i < rows.length; i++) {
            var v = rows[i][ent.matchProp];
            if (ent.numeric ? Number(v) === Number(identity) : String(v) === String(identity)) return rows[i];
        }
        return null;
    }

    // ---- Auto-enhancer: inject "+ New" buttons -------------------------------
    function enhanceSelect(select, map) {
        var fk = map[select.name];
        if (!fk || select.dataset.nslEnhanced) return;
        select.dataset.nslEnhanced = '1';
        var btn = el('button', { type: 'button', 'class': 'inline-create-btn', title: 'Create a new ' + ENTITIES[fk.e].title }, '+ New');
        btn.addEventListener('click', function () {
            openModal(fk.e, function (identity) { refreshDashboardSelect(select, fk, identity); });
        });
        select.parentNode.insertBefore(btn, select.nextSibling);
    }
    function enhanceAll(root, map) {
        var sels = root.querySelectorAll ? root.querySelectorAll('select') : [];
        for (var i = 0; i < sels.length; i++) enhanceSelect(sels[i], map);
    }

    // ---- Refreshing a dropdown after a create --------------------------------
    // Re-fetch the entity list (model-filtered for model-port selects), rebuild
    // the options and select the newly created row.
    function refreshDashboardSelect(select, fk, identity) {
        listForFK(fk, select).then(function (rows) {
            rebuildOptions(select, fk.e, fk.v, rows);
            var row = findCreated(fk.e, rows, identity);
            if (row) select.value = row[fk.v];
        }).catch(function (e) { console.error('inline-create refresh failed', e); });
    }
    function listForFK(fk, select) {
        if (!fk.filterBy) return getJSON(ENTITIES[fk.e].path);
        // Model-filtered model-port list: read the sibling device <select>, map
        // the chosen device to its model, then keep only that model's ports.
        var form = select.form;
        var devSel = form ? form.querySelector('[name="' + fk.filterBy + '"]') : null;
        var devId = devSel ? devSel.value : '';
        if (!devId) return getJSON(ENTITIES.modelport.path);
        return getJSON(ENTITIES.device.path).then(function (devs) {
            var model = '';
            for (var i = 0; i < devs.length; i++) { if (String(devs[i].id) === String(devId)) { model = devs[i].model; break; } }
            return getJSON(ENTITIES.modelport.path).then(function (ports) {
                return model ? ports.filter(function (p) { return p.model === model; }) : ports;
            });
        });
    }
    function rebuildOptions(select, entityKey, valueProp, rows) {
        var current = select.value;
        // Preserve a leading placeholder (empty value), drop the rest.
        var placeholder = (select.options.length && select.options[0].value === '') ? select.options[0].textContent : null;
        select.innerHTML = '';
        if (placeholder !== null) select.appendChild(el('option', { value: '' }, placeholder));
        rows.forEach(function (row) {
            select.appendChild(el('option', { value: row[valueProp] }, rowLabel(entityKey, row)));
        });
        select.value = current; // keep prior selection if still present
    }

    // ---- Modal stack ---------------------------------------------------------
    var modalStack = [];
    function openModal(entityKey, onCreated) {
        var ent = ENTITIES[entityKey];
        var overlay = el('div', { 'class': 'nsl-modal-overlay' });
        var modal = el('div', { 'class': 'nsl-modal' });
        overlay.appendChild(modal);
        modal.appendChild(el('h4', null, 'New ' + ent.title));
        var errBox = el('div', { 'class': 'nsl-modal-error', style: 'display:none' });
        modal.appendChild(errBox);

        var form = el('form');
        var inputs = {}; // field name -> { field, getValue, getRefLabel? }
        ent.fields.forEach(function (field) {
            var wrap = el('div', { 'class': 'nsl-field' });
            wrap.appendChild(el('label', null, field.label + (field.required ? ' *' : '')));
            if (field.type === 'ref') {
                var refSel = el('select');
                refSel.appendChild(el('option', { value: '' }, '-- Select --'));
                getJSON(ENTITIES[field.ref].path).then(function (rows) {
                    rows.forEach(function (row) { refSel.appendChild(el('option', { value: row[field.valueProp] }, rowLabel(field.ref, row))); });
                });
                var refBtn = el('button', { type: 'button', 'class': 'inline-create-btn', title: 'Create a new ' + ENTITIES[field.ref].title }, '+ New');
                (function (sel, f) {
                    refBtn.addEventListener('click', function () {
                        openModal(f.ref, function (identity) {
                            getJSON(ENTITIES[f.ref].path).then(function (rows) {
                                // rebuild + select the newly created ref row
                                var cur = sel.value;
                                sel.innerHTML = '';
                                sel.appendChild(el('option', { value: '' }, '-- Select --'));
                                rows.forEach(function (row) { sel.appendChild(el('option', { value: row[f.valueProp] }, rowLabel(f.ref, row))); });
                                sel.value = cur;
                                var row = findCreated(f.ref, rows, identity);
                                if (row) sel.value = row[f.valueProp];
                            });
                        });
                    });
                })(refSel, field);
                wrap.appendChild(refSel);
                wrap.appendChild(refBtn);
                inputs[field.name] = {
                    getValue: function () { return refSel.value; },
                    getRefLabel: function () { return refSel.options[refSel.selectedIndex] ? refSel.options[refSel.selectedIndex].textContent : ''; },
                };
            } else if (field.type === 'checkbox') {
                var cb = el('input', { type: 'checkbox' });
                wrap.appendChild(cb);
                inputs[field.name] = { getValue: function () { return cb.checked; } };
            } else {
                var inp = el('input', { type: field.type === 'number' ? 'number' : 'text' });
                wrap.appendChild(inp);
                inputs[field.name] = { getValue: function () { return inp.value; } };
            }
            form.appendChild(wrap);
        });

        var actions = el('div', { 'class': 'nsl-modal-actions' });
        var cancel = el('button', { type: 'button' }, 'Cancel');
        var save = el('button', { type: 'submit' }, 'Save');
        cancel.addEventListener('click', function () { closeModal(overlay); });
        actions.appendChild(cancel);
        actions.appendChild(save);
        form.appendChild(actions);

        form.addEventListener('submit', function (ev) {
            ev.preventDefault();
            submitModal(entityKey, inputs, errBox, save, function (identity) {
                closeModal(overlay);
                onCreated(identity);
            });
        });
        modal.appendChild(form);

        // Clicking the dimmed backdrop (not the dialog) cancels.
        overlay.addEventListener('mousedown', function (ev) { if (ev.target === overlay) closeModal(overlay); });
        document.body.appendChild(overlay);
        modalStack.push(overlay);
        var firstInput = form.querySelector('input, select');
        if (firstInput) firstInput.focus();
    }
    function closeModal(overlay) {
        var i = modalStack.indexOf(overlay);
        if (i >= 0) modalStack.splice(i, 1);
        if (overlay.parentNode) overlay.parentNode.removeChild(overlay);
    }

    function submitModal(entityKey, inputs, errBox, saveBtn, onDone) {
        var ent = ENTITIES[entityKey];
        var payload = {};
        var refLabels = {};
        for (var k = 0; k < ent.fields.length; k++) {
            var field = ent.fields[k];
            var raw = inputs[field.name].getValue();
            if (field.required && (raw === '' || raw == null)) {
                showErr(errBox, field.label + ' is required.');
                return;
            }
            if (inputs[field.name].getRefLabel) refLabels[field.name] = inputs[field.name].getRefLabel();
            if (field.type === 'iplist') {
                payload[field.json] = String(raw || '').split(/[\s,]+/).map(function (s) { return s.trim(); }).filter(Boolean);
            } else if (field.type === 'checkbox') {
                payload[field.json] = !!raw;
            } else {
                payload[field.json] = raw;
            }
        }
        if (ent.extra) Object.keys(ent.extra).forEach(function (k) { payload[k] = ent.extra[k]; });
        if (ent.transform) ent.transform(payload, function (name) { return refLabels[name]; });

        var identity = inputs[ent.matchField].getValue();
        saveBtn.disabled = true;
        showErr(errBox, '');
        postJSON(ent.path, payload).then(function (res) {
            saveBtn.disabled = false;
            if (!res.ok) { showErr(errBox, (res.body && res.body.message) || 'Create failed.'); return; }
            onDone(identity);
        }).catch(function (e) { saveBtn.disabled = false; showErr(errBox, String(e)); });
    }
    function showErr(box, msg) {
        box.textContent = msg || '';
        box.style.display = msg ? '' : 'none';
    }

    // ---- init ----------------------------------------------------------------
    function init() {
        var dash = window.NSL_DASH || {};
        var map = FK_MAPS[dash.action];
        if (!map) return; // not an add/update form with composed fields
        var container = document.querySelector('.actions') || document.body;
        enhanceAll(container, map);
        // Forms add VLAN/IP rows (hence selects) dynamically — enhance those too.
        var obs = new MutationObserver(function (muts) {
            muts.forEach(function (m) {
                for (var i = 0; i < m.addedNodes.length; i++) {
                    var n = m.addedNodes[i];
                    if (n.nodeType !== 1) continue;
                    if (n.tagName === 'SELECT') enhanceSelect(n, map);
                    else enhanceAll(n, map);
                }
            });
        });
        obs.observe(container, { childList: true, subtree: true });
    }
    document.addEventListener('DOMContentLoaded', init);
})();
