<?php
// SPDX-License-Identifier: MIT
// scan_profiles.php — profile CRUD + saved-profiles table + create-profile
// form. Three POSTs feed this: do_create_profile, do_delete_profile,
// do_load_profile. The dispatcher in importscan.php passes $profiles and
// $profileMessage by reference; the handlers mutate them in place.

require_once __DIR__ . '/import_common.php';

// --- handlers ---------------------------------------------------------------

function do_create_profile(&$profileMessage, &$profiles) {
    $name = trim($_POST['cp_name'] ?? '');
    if ($name === '') {
        $profileMessage = 'Enter a profile name.';
        return;
    }
    // Profile-level SSH mode:
    //   '__inline__'   → read cp_ssh_user / cp_ssh_password / cp_ssh_key_file
    //                    and send them as the profile's own credentials.
    //   any other non-empty value → treated as the name of an existing
    //                    generic SSH profile whose credentials the backend
    //                    will copy at create time. The inline fields are
    //                    ignored. The backend resolves the reference and
    //                    persists the resolved (encrypted) credentials, so
    //                    the saved profile carries its own copy and the
    //                    source generic can be deleted later without
    //                    breaking this profile.
    $cpSshMode = trim((string) ($_POST['cp_ssh_profile'] ?? '__inline__'));
    $sshProfileRef = '';
    $sshKey = '';
    $sshUser = '';
    $sshPassword = '';
    if ($cpSshMode === '__inline__' || $cpSshMode === '') {
        if (isset($_FILES['cp_ssh_key_file']) && $_FILES['cp_ssh_key_file']['error'] === UPLOAD_ERR_OK) {
            $sshKey = (string) file_get_contents($_FILES['cp_ssh_key_file']['tmp_name']);
        }
        $sshUser = trim((string) ($_POST['cp_ssh_user'] ?? ''));
        $sshPassword = (string) ($_POST['cp_ssh_password'] ?? '');
    } else {
        $sshProfileRef = $cpSshMode;
    }
    $cpType   = $_POST['cp_type'] ?? 'generic-ssh';
    $cpKind   = strpos($cpType, 'generic') === 0 ? 'generic' : 'device';
    // scan_source is the trailing "ssh" or "snmp"; strip the leading prefix.
    // substr(..., -3) was wrong: "device-snmp" → "smp" (off-by-one).
    $cpSource = preg_replace('/^(device|generic)-/', '', $cpType);
    // Multi-row device list: prefer cp_device_hosts[] over the legacy cp_host.
    // The profile-level "host" field stays set to the first row's host for
    // backwards compatibility with code paths that auto-match by host.
    $deviceHosts = $_POST['cp_device_hosts'] ?? [];
    $deviceProfiles = $_POST['cp_device_ssh_profiles'] ?? [];
    $deviceOsTypes = $_POST['cp_device_os_types'] ?? [];
    if (!is_array($deviceHosts))      $deviceHosts      = [];
    if (!is_array($deviceProfiles))  $deviceProfiles   = [];
    if (!is_array($deviceOsTypes))   $deviceOsTypes    = [];
    $legacyHost = trim($_POST['cp_host'] ?? '');
    $firstHost = '';
    if ($cpKind === 'device') {
        foreach ($deviceHosts as $h) {
            $h = trim((string) $h);
            if ($h !== '') {
                $firstHost = $h;
                break;
            }
        }
        if ($firstHost === '') $firstHost = $legacyHost;
    }
    $payload = json_encode([
        'name'           => $name,
        'kind'           => $cpKind,
        'host'           => $firstHost,
        'snmp_community' => trim($_POST['cp_community'] ?? 'public'),
        'snmp_version'   => trim($_POST['cp_version'] ?? '2c'),
        'snmp_port'      => intval($_POST['cp_port'] ?? 161),
        'scan_source'    => $cpSource,
        'os_type'        => trim($_POST['cp_os_type'] ?? ''),
        'ssh_user'       => $sshUser,
        'ssh_password'   => $sshPassword,
        'ssh_key'        => $sshKey,
    ] + ($sshProfileRef !== '' ? ['ssh_profile_name' => $sshProfileRef] : []));
    list($code, $body) = api_method('POST', SCAN_PROFILES_ENDPOINT, $payload);
    if ($code === 201) {
        $resp = json_decode($body, true) ?: [];
        $overlaps = $resp['overlaps'] ?? [];
        $profileMessage = "Profile \"$name\" created.";
        if (is_array($overlaps) && count($overlaps) > 0) {
            $items = [];
            foreach ($overlaps as $o) {
                $items[] = htmlspecialchars($o['host'] ?? '?') . ' is also in profile "' . htmlspecialchars($o['other_profile'] ?? '?') . '"';
            }
            $profileMessage .= ' Overlaps: ' . implode('; ', $items) . '.';
        }
        // Attach each device row to the profile via the per-row endpoint.
        // Best-effort: report partial failures in the message but don't unwind
        // the profile (it's already created).
        if ($cpKind === 'device') {
            $deviceErrors = [];
            $attached = 0;
            // Per-row fields. Empty rows are skipped silently.
            $inlineUser        = $_POST['cp_device_inline_user']         ?? [];
            $inlinePassword    = $_POST['cp_device_inline_password']     ?? [];
            $inlineKeyFilename = $_POST['cp_device_inline_key_filename'] ?? [];
            $inlineConfigText  = $_POST['cp_device_inline_config_text']  ?? [];
            foreach ($deviceHosts as $i => $h) {
                $h = trim((string) $h);
                if ($h === '') continue;
                $override = isset($deviceProfiles[$i]) ? (string) $deviceProfiles[$i] : '';
                // Per-row OS override wins over the profile's os_type when set.
                $rowOs = trim((string) ($deviceOsTypes[$i] ?? ''));
                $devPayload = ['host' => $h];
                if ($rowOs !== '') {
                    $devPayload['os_type'] = $rowOs;
                }
                if ($override === '__inline__') {
                    // Inline custom: read per-row fields, encrypt the uploaded key.
                    $inlineKey = '';
                    $inlineKeyName = trim((string) ($inlineKeyFilename[$i] ?? ''));
                    if (isset($_FILES['cp_device_inline_key']['error'][$i])
                        && $_FILES['cp_device_inline_key']['error'][$i] === UPLOAD_ERR_OK) {
                        $tmp = $_FILES['cp_device_inline_key']['tmp_name'][$i];
                        $inlineKey = (string) file_get_contents($tmp);
                        if ($inlineKeyName === '') {
                            $inlineKeyName = (string) ($_FILES['cp_device_inline_key']['name'][$i] ?? '');
                        }
                    }
                    $devPayload = [
                        'host'             => $h,
                        'ssh_profile_name' => '', // explicit: no override profile
                        'ssh_user'         => trim((string) ($inlineUser[$i] ?? '')),
                        'ssh_password'     => (string) ($inlinePassword[$i] ?? ''),
                        'ssh_key'          => $inlineKey,
                        'ssh_key_filename' => $inlineKeyName,
                        'ssh_config_text'  => (string) ($inlineConfigText[$i] ?? ''),
                    ];
                } elseif ($override !== '') {
                    // Saved generic profile: reference by name.
                    $devPayload['ssh_profile_name'] = trim($override);
                }
                $payloadJson = json_encode($devPayload);
                list($dcode, $dbody) = api_method(
                    'POST',
                    SCAN_PROFILES_ENDPOINT . '/' . urlencode($name) . '/devices',
                    $payloadJson
                );
                if ($dcode === 201) {
                    $attached++;
                } else {
                    $err = json_decode($dbody, true);
                    $deviceErrors[] = $h . ': ' . htmlspecialchars($err['message'] ?? $dbody);
                }
            }
            $profileMessage .= " Attached {$attached} device" . ($attached === 1 ? '' : 's') . '.';
            if (count($deviceErrors) > 0) {
                $profileMessage .= ' Device errors: ' . implode('; ', $deviceErrors) . '.';
            }
        }
    } else {
        $profileMessage = 'Create failed: ' . htmlspecialchars($body);
    }
    // Refresh the in-memory $profiles list so the saved-profiles table shows the new row.
    $profiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];
}
function do_delete_profile(&$profileMessage, &$profiles) {
    $name = $_POST['profile_name'] ?? '';
    list($code, $body) = api_method('DELETE', SCAN_PROFILES_ENDPOINT . '?name=' . urlencode($name));
    $profileMessage = ($code === 200) ? "Profile \"$name\" deleted." : 'Delete failed: ' . htmlspecialchars($body);
    $profiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];
}

function do_load_profile(&$profileMessage, $profiles, &$pf) {
    $name = $_POST['profile_select'] ?? '';
    foreach ($profiles as $p) {
        if (($p['name'] ?? '') === $name) {
            $pf['target']       = $p['host'] ?? '';
            $pf['community']    = $p['snmp_community'] ?? 'public';
            $pf['snmp_version'] = $p['snmp_version'] ?? '2c';
            $pf['snmp_port']    = (string) ($p['snmp_port'] ?? 161);
            $profileMessage     = "Loaded profile \"$name\" into the form.";
            return;
        }
    }
}

// --- panels -----------------------------------------------------------------

function scan_profiles_saved_panel_html($profiles) {
    ob_start();
    ?>
<form method="post" action="import.php">
    <select name="profile_select">
        <option value="">— select a profile —</option>
        <?php foreach ($profiles as $p): ?>
            <option value="<?= htmlspecialchars($p['name'] ?? '') ?>"><?= htmlspecialchars(($p['name'] ?? '') . ' (' . ($p['host'] ?? '') . ')') ?></option>
        <?php endforeach; ?>
    </select>
    <button type="submit" name="do_load_profile" value="1">Load into form</button>
</form>
<?php if (!empty($profiles)): ?>
    <table border="1" cellpadding="3" cellspacing="0" style="margin-top:6px;">
        <tr><th>Name</th><th>Kind</th><th>Host</th><th>Community</th><th>Ver</th><th>SSH pw</th><th>SSH key</th><th></th></tr>
        <?php foreach ($profiles as $p): ?>
            <tr>
                <td><?= htmlspecialchars($p['name'] ?? '') ?></td>
                <td><?= htmlspecialchars(($p['kind'] ?? '') !== '' ? $p['kind'] : 'device') ?></td>
                <td><?= htmlspecialchars($p['host'] ?? '') ?></td>
                <td><?= htmlspecialchars($p['snmp_community'] ?? '') ?></td>
                <td><?= htmlspecialchars($p['snmp_version'] ?? '') ?></td>
                <td><?= !empty($p['has_ssh_password']) ? 'yes (enc)' : '—' ?></td>
                <td><?= !empty($p['has_ssh_key']) ? 'yes (enc)' : '—' ?></td>
                <td>
                    <form method="post" action="import.php" style="margin:0;"
                          onsubmit="return confirm('Delete profile <?= htmlspecialchars($p['name'] ?? '') ?>?');">
                        <input type="hidden" name="profile_name" value="<?= htmlspecialchars($p['name'] ?? '') ?>">
                        <button type="submit" name="do_delete_profile" value="1">Delete</button>
                    </form>
                </td>
            </tr>
        <?php endforeach; ?>
    </table>
<?php endif; ?>
    <?php
    return ob_get_clean();
}

function scan_profiles_create_panel_html($profiles = [], $osTypes = []) {
    ob_start();
    ?>
    <!-- Create-profile box -->
    <div class="box">
        <h3>Create profile</h3>
        <form id="cp-form" method="post" action="import.php" enctype="multipart/form-data">
            <label>Name: <input type="text" name="cp_name" required></label><br>
            <div class="cp-type-buttons">Type:
                <button type="button" class="cp-type-btn" data-type="device-snmp">Device · SNMP</button>
                <button type="button" class="cp-type-btn" data-type="device-ssh">Device · SSH</button>
                <button type="button" class="cp-type-btn" data-type="generic-snmp">Generic · SNMP</button>
                <button type="button" class="cp-type-btn" data-type="generic-ssh">Generic · SSH</button>
            </div>
            <input type="hidden" name="cp_type" id="cp_type" value="generic-ssh">
            <p id="cp_type_hint" style="margin:4px 0; color:#555; font-size:0.85em;"></p>

            <!-- Devices: device profiles only. The single-host input is replaced
                 with a multi-row device list; each row has its own host + SSH
                 override. Generic profiles (no host) skip this block. -->
            <div class="cp-grp" data-show="device-snmp device-ssh" id="cp_devices_block">
                <p style="margin:6px 0; color:#555;"><em>Devices in this profile</em> &mdash;
                    one row per device. Per-row SSH override options:
                    <strong>(use shared)</strong> = the profile's credentials,
                    <strong>(inline custom)</strong> = write custom SSH credentials on this row only,
                    or pick a saved <em>generic</em> profile from the dropdown.</p>
                <table id="cp_devices_table" border="0" cellpadding="3" cellspacing="0" style="border-collapse:collapse; width:100%;">
                    <thead>
                        <tr style="font-size:0.85em; color:#555;">
                            <th align="left">Host</th>
                            <th align="left">OS</th>
                            <th align="left" class="cp-ssh-col">SSH override</th>
                            <th></th>
                        </tr>
                    </thead>
                    <tbody id="cp_devices_tbody">
                        <tr class="cp-device-row" data-inline="0">
                            <td><input type="text" name="cp_device_hosts[]" placeholder="10.0.0.10"></td>
                            <td>
                                <select name="cp_device_os_types[]" class="cp-row-os">
                                    <option value="">&mdash; (default) &mdash;</option>
                                    <?php foreach ($osTypes as $ot):
                                        $otName = htmlspecialchars((string) ($ot['name'] ?? ''));
                                        if ($otName === '') continue;
                                    ?>
                                        <option value="<?= $otName ?>"><?= $otName ?></option>
                                    <?php endforeach; ?>
                                </select>
                            </td>
                            <td class="cp-ssh-col">
                                <select name="cp_device_ssh_profiles[]" class="cp-row-override">
                                    <option value="">&mdash; (use shared) &mdash;</option>
                                    <?php foreach ($profiles as $p):
                                        // Only GENERIC profiles appear in the SSH override dropdown.
                                        // Device profiles are not credentials-by-design — a generic
                                        // profile is the credential-only one meant for reuse.
                                        $pk = ($p['kind'] ?? '') !== '' ? $p['kind'] : 'device';
                                        $src = ($p['scan_source'] ?? '');
                                        if ($pk !== 'generic' || $src !== 'ssh') continue;
                                        $label = ($p['name'] ?? '');
                                    ?>
                                        <option value="<?= htmlspecialchars($p['name'] ?? '') ?>"><?= htmlspecialchars($label) ?> (generic)</option>
                                    <?php endforeach; ?>
                                    <option value="__inline__">&mdash; (inline custom) &mdash;</option>
                                </select>
                                <div class="cp-inline-fields" style="display:none; margin-top:6px; padding:6px; background:#f6f8fa; border:1px solid #ddd;">
                                    <label>SSH user: <input type="text" name="cp_device_inline_user[]"></label><br>
                                    <label>SSH password: <input type="password" name="cp_device_inline_password[]"></label><br>
                                    <label>SSH private key file: <input type="file" name="cp_device_inline_key[]"></label><br>
                                    <label>SSH key filename (basename for display): <input type="text" name="cp_device_inline_key_filename[]" placeholder="id_ed25519"></label><br>
                                    <label>Inline OpenSSH config <small>(free-form, overrides the profile's; not a secret)</small>:
                                        <textarea name="cp_device_inline_config_text[]" rows="2" style="width:100%;"></textarea>
                                    </label>
                                </div>
                            </td>
                            <td><button type="button" class="cp-device-remove">Remove</button></td>
                        </tr>
                    </tbody>
                </table>
                <button type="button" id="cp_device_add" style="margin-top:6px;">+ Add device</button>
                <p style="color:#777; font-size:0.85em; margin-top:6px;">Tip: for a single-device profile, fill one row. For a sweep, add a row per target.</p>
            </div>
            <!-- SNMP parameters: SNMP profiles only. -->
            <div class="cp-grp" data-show="device-snmp generic-snmp">
                <label>SNMP community: <input type="text" name="cp_community" value="public"></label>
                <label>Version: <select name="cp_version"><option>2c</option><option value="1">1</option></select></label>
                <label>Port: <input type="number" name="cp_port" value="161" style="width:80px;"></label><br>
            </div>
            <!-- SSH config + credentials: SSH profiles only. -->
            <div class="cp-grp" data-show="device-ssh generic-ssh">
                <label>OS / firmware type <small>(operating system for config parsing — not the hardware model; optional for generic)</small>:
                    <select name="cp_os_type">
                        <option value="">&mdash; (use profile default) &mdash;</option>
                        <?php foreach ($osTypes as $ot):
                            $otName = htmlspecialchars((string) ($ot['name'] ?? ''));
                            if ($otName === '') continue;
                        ?>
                            <option value="<?= $otName ?>"><?= $otName ?></option>
                        <?php endforeach; ?>
                    </select>
                </label><br>
                <p style="margin:6px 0; color:#555;"><em>Profile-wide SSH credentials</em> &mdash;
                    pick a saved <em>generic</em> profile from the dropdown to reuse its
                    credentials, or stay on <strong>(inline custom)</strong> to write the
                    profile's credentials here.</p>
                <select name="cp_ssh_profile" id="cp-ssh-profile" style="min-width:240px;">
                    <option value="__inline__" selected>&mdash; (inline custom) &mdash;</option>
                    <?php foreach ($profiles as $p):
                        $pk = ($p['kind'] ?? '') !== '' ? $p['kind'] : 'device';
                        $src = ($p['scan_source'] ?? '');
                        // Reference must point at a generic SSH profile.
                        if ($pk !== 'generic' || $src !== 'ssh') continue;
                        $pName = (string) ($p['name'] ?? '');
                        if ($pName === '') continue;
                    ?>
                        <option value="<?= htmlspecialchars($pName) ?>"><?= htmlspecialchars($pName) ?> (generic)</option>
                    <?php endforeach; ?>
                </select>
                <div id="cp-ssh-inline" style="margin-top:6px; padding:6px; background:#f6f8fa; border:1px solid #ddd;">
                    <label>SSH user: <input type="text" name="cp_ssh_user"></label><br>
                    <label>SSH password: <input type="password" name="cp_ssh_password"></label><br>
                    <label>SSH private key file: <input type="file" name="cp_ssh_key_file"></label><br>
                </div>
            <button type="submit" name="do_create_profile" value="1" style="margin-top:8px;">Create profile</button>
        </form>
        <p style="color:#777; font-size:0.85em;">SSH password and uploaded private key are encrypted by the credential vault (AES-256-GCM); unlock the vault from the app bar before creating an SSH profile, and again whenever an SSH scan uses it.</p>
        <script>
        (function () {
            var hints = {
                'device-snmp':  'Bound to a host, scanned over SNMP.',
                'device-ssh':   'Bound to a host, config read over SSH.',
                'generic-snmp': 'Reusable SNMP settings, not bound to a host.',
                'generic-ssh':  'Reusable SSH credentials, not bound to a host.'
            };
            function cpApplyType(type) {
                document.getElementById('cp_type').value = type;
                document.getElementById('cp_type_hint').textContent = hints[type] || '';
                document.querySelectorAll('.cp-type-btn').forEach(function (b) {
                    b.classList.toggle('active', b.dataset.type === type);
                });
                document.querySelectorAll('.cp-grp').forEach(function (g) {
                    g.style.display = g.dataset.show.split(' ').indexOf(type) >= 0 ? '' : 'none';
                });
                // Hide the SSH override column entirely for SNMP profiles.
                var sshVisible = (type === 'device-ssh');
                document.querySelectorAll('.cp-ssh-col').forEach(function (th) {
                    th.style.display = sshVisible ? '' : 'none';
                });
            }
            function cpUpdateRowOverride(row) {
                var sel = row.querySelector('.cp-row-override');
                var inline = row.querySelector('.cp-inline-fields');
                if (!sel || !inline) return;
                inline.style.display = (sel.value === '__inline__') ? '' : 'none';
            }
            function cpWireRowOverride(row) {
                var sel = row.querySelector('.cp-row-override');
                if (sel) sel.addEventListener('change', function () { cpUpdateRowOverride(row); });
                cpUpdateRowOverride(row);
            }
            function cpMakeRow() {
                var first = document.querySelector('#cp_devices_tbody .cp-device-row');
                if (!first) return null;
                var row = first.cloneNode(true);
                var hostInput = row.querySelector('input[name="cp_device_hosts[]"]');
                if (hostInput) hostInput.value = '';
                var sel = row.querySelector('select[name="cp_device_ssh_profiles[]"]');
                if (sel) sel.selectedIndex = 0;
                // Per-row OS override also resets to "(default)".
                var osSel = row.querySelector('select[name="cp_device_os_types[]"]');
                if (osSel) osSel.selectedIndex = 0;
                // Clear inline fields on the cloned row.
                row.querySelectorAll('input[type="text"], input[type="password"], textarea').forEach(function (inp) {
                    inp.value = '';
                });
                row.querySelectorAll('input[type="file"]').forEach(function (inp) {
                    inp.value = '';
                });
                return row;
            }
            function cpWireRowRemove(row) {
                var btn = row.querySelector('.cp-device-remove');
                if (!btn) return;
                btn.addEventListener('click', function () {
                    var tbody = document.getElementById('cp_devices_tbody');
                    if (tbody.children.length <= 1) {
                        var host = row.querySelector('input[name="cp_device_hosts[]"]');
                        if (host) host.value = '';
                        var sel = row.querySelector('select[name="cp_device_ssh_profiles[]"]');
                        if (sel) {
                            sel.selectedIndex = 0;
                            cpUpdateRowOverride(row);
                        }
                        return;
                    }
                    row.parentNode.removeChild(row);
                });
            }
            document.querySelectorAll('.cp-device-row').forEach(function (row) {
                cpWireRowRemove(row);
                cpWireRowOverride(row);
            });
            var addBtn = document.getElementById('cp_device_add');
            if (addBtn) {
                addBtn.addEventListener('click', function () {
                    var row = cpMakeRow();
                    if (!row) return;
                    cpWireRowRemove(row);
                    cpWireRowOverride(row);
                    document.getElementById('cp_devices_tbody').appendChild(row);
                });
            }
            document.querySelectorAll('.cp-type-btn').forEach(function (b) {
                b.addEventListener('click', function () { cpApplyType(b.dataset.type); });
            });
            // Profile-level SSH override (mirrors per-row). Inline block only
            // shows when the operator picks (inline custom). When the profile
            // points at a saved generic profile, the row-side creds win.
            var profileSshSel = document.getElementById('cp-ssh-profile');
            var profileSshInline = document.getElementById('cp-ssh-inline');
            function cpUpdateProfileSsh() {
                if (!profileSshSel || !profileSshInline) return;
                profileSshInline.style.display = (profileSshSel.value === '__inline__') ? '' : 'none';
            }
            if (profileSshSel) profileSshSel.addEventListener('change', cpUpdateProfileSsh);
            cpUpdateProfileSsh();
            cpApplyType('generic-ssh'); // default
            // Form-level guard: device-* types require at least one host row.
            // Generic types are host-less by design — don't let the browser's
            // hidden-required validation block the submit.
            var form = document.getElementById('cp-form');
            if (form) {
                form.addEventListener('submit', function (ev) {
                    var type = document.getElementById('cp_type').value || '';
                    if (type.indexOf('device-') === 0) {
                        var hosts = form.querySelectorAll('input[name=\"cp_device_hosts[]\"]');
                        var anyFilled = Array.from(hosts).some(function (h) { return (h.value || '').trim() !== ''; });
                        if (!anyFilled) {
                            var errBox = document.getElementById('cp-error');
                            if (!errBox) {
                                errBox = document.createElement('p');
                                errBox.id = 'cp-error';
                                errBox.style.cssText = 'color:#a00; font-size:0.85em; margin:8px 0;';
                                var btn = document.querySelector('button[name=\"do_create_profile\"]');
                                if (btn && btn.parentNode) btn.parentNode.insertBefore(errBox, btn);
                            }
                            errBox.textContent = 'Device profiles need at least one host. Add one or pick a Generic profile.';
                            ev.preventDefault();
                            return false;
                        }
                    }
                });
            }


        })();
        </script>
    </div>
    <?php
    return ob_get_clean();
}
