<?php
// SPDX-License-Identifier: MIT
// scan.php — unified live scan panel (Phase 1 of docs/plans/SCAN-REFACTOR.md).
//
// Replaces the old scan_devices.php + scan_connections.php pair. The unified
// "Live scan" panel runs a device scan (SNMP / SSH / both) and optionally
// chains a connection-discovery scan after it.
//
// Three target sources:
//   'profile'  — every device attached to a saved profile, with the
//                 profile's OS type and per-row SSH overrides honoured
//                 by the backend.
//   'target'   — free-form IP / CIDR / comma-list (legacy).
//   'db'       — every device already in the store (connection-only path).
//
// One shared credentials block: read-only when Source = Profile (values
// pulled from the profile), editable otherwise. SSH key file uploads,
// OpenSSH config + extra key files all live in the same block.

require_once __DIR__ . '/import_common.php';

// --- form-state helpers -----------------------------------------------------

function scan_form_state() {
    return [
        'source'           => $_POST['scan_source'] ?? 'target',
        'profile'          => $_POST['scan_profile'] ?? '',
        'target'           => $_POST['target'] ?? '',
        'method'           => ($_POST['scan_method'] ?? 'snmp') === 'ssh' ? 'ssh' : 'snmp',
        'community'        => $_POST['community'] ?? 'public',
        'snmp_version'     => $_POST['snmp_version'] ?? '2c',
        'snmp_port'        => $_POST['snmp_port'] ?? '161',
        'ssh_profile'      => $_POST['ssh_profile'] ?? '',
        'ssh_user'         => $_POST['ssh_user'] ?? '',
        'ssh_password'     => $_POST['ssh_password'] ?? '',
        'ssh_key_file'     => $_POST['ssh_key_file'] ?? '',
        'ssh_key'          => $_POST['ssh_key'] ?? '',
        'collector'        => $_POST['collector'] ?? '',
        'timeout'          => $_POST['timeout'] ?? '10',
        'also_connections' => !empty($_POST['also_connections']),
        'auto_import'      => !empty($_POST['auto_import']),
    ];
}

// --- handlers ---------------------------------------------------------------

function do_scan(&$scanMessage, &$scanRunId, &$autoReload) {
    $f = scan_form_state();
    $source = $f['source'];
    $method = $f['method'];

    // Resolve the target list + profile-supplied credentials.
    $target = '';
    $profileName = '';
    $osType     = '';
    $sshUser    = '';
    $sshPass    = '';
    $sshKey     = '';
    $sshKeyFile = '';

    if ($source === 'profile') {
        if ($f['profile'] === '') {
            $scanMessage = 'Pick a saved profile to scan.';
            return;
        }
        $rows = [];
        $raw = @file_get_contents(SCAN_PROFILES_ENDPOINT . '/' . rawurlencode($f['profile']) . '/devices');
        $decoded = json_decode($raw, true);
        if (is_array($decoded)) {
            $rows = $decoded;
        }
        $hosts = [];
        foreach ($rows as $r) {
            $h = trim((string) ($r['host'] ?? ''));
            if ($h !== '') {
                $hosts[] = $h;
            }
        }
        if (count($hosts) === 0) {
            $scanMessage = 'Profile "' . htmlspecialchars($f['profile']) . '" has no attached devices.';
            return;
        }
        $target = implode(',', $hosts);
        // Pull profile metadata for OS type + creds (read-only block).
        $profiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];
        foreach ($profiles as $p) {
            if (($p['name'] ?? '') === $f['profile']) {
                $profileName = $f['profile'];
                $osType = (string) ($p['os_type'] ?? '');
                $sshUser = (string) ($p['ssh_user'] ?? '');
                $sshKeyFile = (string) ($p['ssh_key_file'] ?? '');
                // Note: encrypted blobs (ssh_password, ssh_key) aren't returned
                // by /scan/profiles. For SSH method, the backend re-reads them
                // from the store via service.GetScanProfileByName; the PHP
                // layer doesn't need to forward cleartext creds.
                break;
            }
        }
        // Per-row OS overrides: each attached device may carry its own os_type.
        // The backend honours them via perRowOverrideCreds; we just send the
        // comma-list and the profile name.
        if ($method !== 'ssh') {
            $scanMessage = 'Profile-based scans only support SSH (per-row credentials).';
            return;
        }
    } elseif ($source === 'db') {
        // Connection-only path: every device in the store. The connection
        // scan picks up each device's stored profile for creds.
        $raw = @file_get_contents(DEVICES_ENDPOINT);
        $decoded = json_decode($raw, true);
        $hosts = [];
        if (is_array($decoded)) {
            foreach ($decoded as $d) {
                    $ip = trim((string) ($d['ips'][0]));
                if ($ip === '' && !empty($d['ips']) && is_array($d['ips'])) {

                }
                if ($ip !== '') {
                    $hosts[] = $ip;
                }
            }
        }
        if (count($hosts) === 0) {
            $scanMessage = 'No devices in the store yet.';
            return;
        }
        $target = implode(',', $hosts);
        $method = 'snmp'; // force — device scan not the focus here
        $profileName = '';
    } else {
        // 'target' — free-form.
        $target = trim($f['target']);
        if ($target === '') {
            $scanMessage = 'Please enter a target IP or CIDR.';
            return;
        }
        if ($method === 'ssh' && $f['ssh_profile'] === '') {
            // Allow ad-hoc SSH creds too: if the operator typed ssh_user +
            // ssh_password without picking a profile, accept that.
            if ($f['ssh_user'] === '' && $f['ssh_key_file'] === '') {
                $scanMessage = 'Pick a profile or enter SSH credentials.';
                return;
            }
            $profileName = $f['ssh_profile'];
            $sshUser = $f['ssh_user'];
            $sshPass = $f['ssh_password'];
            // Optional uploaded key file.
            if (isset($_FILES['ssh_key_file']) && $_FILES['ssh_key_file']['error'] === UPLOAD_ERR_OK) {
                $sshKey = (string) file_get_contents($_FILES['ssh_key_file']['tmp_name']);
                if ($sshKeyFile === '') {
                    $sshKeyFile = (string) ($_FILES['ssh_key_file']['name'] ?? '');
                }
            }
        } else {
            $profileName = $f['ssh_profile'];
        }
        $osType = trim($_POST['ssh_os_type'] ?? '');
    }

    $payload = [
        'method'       => $method,
        'target'       => $target,
        'community'    => $f['community'],
        'snmp_version' => $f['snmp_version'],
        'snmp_port'    => intval($f['snmp_port']),
        'profile'      => $profileName,
        'os_type'      => $osType,
        'ssh_user'     => $sshUser,
        'ssh_password' => $sshPass,
        'ssh_key'      => $sshKey,
    ];
    list($code, $body, $err) = api_post_json(SCAN_RUN_ENDPOINT, json_encode($payload));
    $j = json_decode($body, true);
    if (($code === 202 || $code === 200) && !empty($j['scan_id'])) {
        $scanRunId  = $j['scan_id'];
        $autoReload = $f['auto_import'] ? 'auto_import=1' : '';
        // Stash follow-up info on the session for do_scan_completed.
        $_SESSION['scan_followup'] = $f['also_connections']
            ? ['mode' => 'profile', 'profile' => $profileName, 'ssh_user' => $sshUser]
            : null;
    } else {
        $scanMessage = 'Could not start scan (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
    }
}

function do_scan_completed(&$discovered, &$scanMessage, &$plan, &$importedIPs, &$importMessage) {
    // Returns true if the scan is still running.
    $scanRunId = $_SESSION['scan_id'] ?? ($_GET['scan_id'] ?? '');
    if ($scanRunId === '') {
        return false;
    }
    list($code, $body) = api_get_json(SCAN_STATUS_ENDPOINT . '?scan_id=' . urlencode($scanRunId));
    if ($code !== 200) {
        return false;
    }
    $status = json_decode($body, true) ?: [];
    $state  = $status['state'] ?? '';
    if ($state === 'running' || $state === 'pending') {
        return true;
    }
    if ($state === 'finished' || $state === 'complete') {
        $res = $status['result'] ?? [];
        $discovered = $res['devices'] ?? [];
        // Optional chained connection scan.
        $follow = $_SESSION['scan_followup'] ?? null;
        if (is_array($follow) && !empty($follow['profile'])) {
            $_SESSION['scan_followup'] = null;
            // Kick off a connection scan on the discovered devices.
            list($cCode, $cBody) = api_post_json(SCAN_CONNECTIONS_ENDPOINT, json_encode([
                'mode'           => 'profiles',
                'community'      => 'public',
                'collector'      => '',
                'timeout_sec'    => 10,
                'ssh_user'       => $follow['ssh_user'] ?? '',
                'generic_profile' => $follow['profile'],
            ]));
            $cJson = json_decode($cBody, true);
            if (($cCode === 202 || $cCode === 200) && !empty($cJson['scan_id'])) {
                $_SESSION['connection_scan_id'] = $cJson['scan_id'];
            }
        }
        // Auto-import path.
        if (!empty($_GET['auto_import']) && count($discovered) > 0) {
            // TODO: existing auto-import path. Kept identical to the old
            // scan_devices flow (the existing inline-import endpoint).
        }
        return false;
    }
    if ($state === 'failed') {
        $scanMessage = 'Scan failed: ' . htmlspecialchars($status['error'] ?? 'unknown error');
        return false;
    }
    return false;
}

// --- panels -----------------------------------------------------------------

function scan_live_panel_html($profiles, $scanMessage) {
    $f = scan_form_state();
    $curSource = $f['source'];
    $curMethod = $f['method'];
    ob_start();
    ?>
    <div class="box">
        <h3>Live scan</h3>
        <p style="color:#555; font-size:0.85em;">
            One panel for both device and connection discovery. Pick a target
            source, choose a method, and optionally chain a connection scan
            onto the device scan.
        </p>
        <form id="scan-form" method="post" action="import.php" enctype="multipart/form-data">
            <fieldset style="border:1px solid #ddd; padding:6px 10px; margin:0 0 10px 0;">
                <legend>Target source</legend>
                <label>
                    <input type="radio" name="scan_source" value="profile" <?= $curSource==='profile'?'checked':'' ?>
                           onchange="scanSourceToggle()">
                    Profile + attached devices
                </label>
                <label style="margin-left:14px;">
                    <input type="radio" name="scan_source" value="target" <?= $curSource==='target'?'checked':'' ?>
                           onchange="scanSourceToggle()">
                    Free-form target
                </label>
                <label style="margin-left:14px;">
                    <input type="radio" name="scan_source" value="db" <?= $curSource==='db'?'checked':'' ?>
                           onchange="scanSourceToggle()">
                    DB devices only (connection scan)
                </label>
                <div id="scan-source-profile" style="margin-top:6px; <?= $curSource==='profile'?'':'display:none;' ?>">
                    <label>Profile:
                        <select name="scan_profile">
                            <option value="">&mdash; select &mdash;</option>
                            <?php foreach ($profiles as $p):
                                $pName = (string) ($p['name'] ?? '');
                                $pKind = (string) ($p['kind'] ?? 'device');
                                if ($pName === '') continue;
                                $pOs = (string) ($p['os_type'] ?? '');
                                $pHosts = isset($p['host']) && $p['host'] !== '' ? ' · ' . htmlspecialchars($p['host']) : '';
                                $label = htmlspecialchars($pName . ' (' . $pKind . ($pOs !== '' ? ', ' . $pOs : '') . ')' . $pHosts);
                            ?>
                                <option value="<?= htmlspecialchars($pName) ?>" <?= $f['profile']===$pName?'selected':'' ?>><?= $label ?></option>
                            <?php endforeach; ?>
                        </select>
                    </label>
                    <p style="margin:4px 0; color:#555; font-size:0.85em;">
                        Each attached device row is scanned with the profile's stored SSH creds
                        and its per-row OS override (if any). The backend reads encrypted blobs
                        from the store — no cleartext flows through this form.
                    </p>
                </div>
                <div id="scan-source-target" style="margin-top:6px; <?= $curSource==='target'?'':'display:none;' ?>">
                    <label>Target — IP, CIDR, or comma-separated list:
                        <input type="text" name="target" value="<?= htmlspecialchars($f['target']) ?>"
                               placeholder="10.0.2.245 or 10.0.2.0/26" size="40">
                    </label>
                </div>
                <div id="scan-source-db" style="margin-top:6px; <?= $curSource==='db'?'':'display:none;' ?>">
                    <p style="color:#555; font-size:0.85em;">
                        Scans every device currently in the store for LLDP/CDP/FDB edges.
                        Each device's stored SSH profile supplies credentials.
                    </p>
                </div>
            </fieldset>

            <fieldset style="border:1px solid #ddd; padding:6px 10px; margin:0 0 10px 0;">
                <legend>Method</legend>
                <label>
                    <input type="radio" name="scan_method" value="snmp" <?= $curMethod==='snmp'?'checked':'' ?>>
                    SNMP only
                </label>
                <label style="margin-left:14px;">
                    <input type="radio" name="scan_method" value="ssh" <?= $curMethod==='ssh'?'checked':'' ?>>
                    SSH only
                </label>
                <div id="scan-snmp-fields" style="margin-top:6px; <?= $curMethod==='snmp'?'':'display:none;' ?>">
                    <label>SNMP community: <input type="text" name="community" value="<?= htmlspecialchars($f['community']) ?>"></label>
                    <label>Version:
                        <select name="snmp_version">
                            <option value="2c" <?= $f['snmp_version']==='2c'?'selected':'' ?>>2c</option>
                            <option value="1"  <?= $f['snmp_version']==='1'?'selected':'' ?>>1</option>
                        </select>
                    </label>
                    <label>Port: <input type="number" name="snmp_port" value="<?= htmlspecialchars($f['snmp_port']) ?>" style="width:80px;"></label>
                </div>
                <div id="scan-ssh-fields" style="margin-top:6px; <?= $curMethod==='ssh'?'':'display:none;' ?>">
                    <label>SSH profile (device or generic):
                        <select name="ssh_profile">
                            <option value="">&mdash; none — ad-hoc creds below &mdash;</option>
                            <?php foreach ($profiles as $p):
                                $pName = (string) ($p['name'] ?? '');
                                if ($pName === '') continue;
                            ?>
                                <option value="<?= htmlspecialchars($pName) ?>" <?= $f['ssh_profile']===$pName?'selected':'' ?>><?= htmlspecialchars($pName) ?></option>
                            <?php endforeach; ?>
                        </select>
                    </label>
                    <label>OS / firmware type (required for generic profiles):
                        <input type="text" name="ssh_os_type" value="<?= htmlspecialchars($_POST['ssh_os_type'] ?? '') ?>"
                               placeholder="openwrt / opnsense / fortinet">
                    </label>
                </div>
            </fieldset>

            <fieldset id="scan-creds" style="border:1px solid #ddd; padding:6px 10px; margin:0 0 10px 0;">
                <legend>Credentials <span style="color:#777; font-weight:normal; font-size:0.85em;">
                    (read-only when source = Profile)</span></legend>
                <label>SSH user:
                    <input type="text" name="ssh_user" id="scan-ssh-user" value="<?= htmlspecialchars($f['ssh_user']) ?>"
                           <?= $curSource==='profile'?'disabled':'' ?>>
                </label><br>
                <label>SSH password:
                    <input type="password" name="ssh_password" id="scan-ssh-password"
                           value="<?= htmlspecialchars($f['ssh_password']) ?>"
                           <?= $curSource==='profile'?'disabled':'' ?>>
                </label><br>
                <label>SSH private key file:
                    <input type="text" name="ssh_key_file" id="scan-ssh-key-file"
                           value="<?= htmlspecialchars($f['ssh_key_file']) ?>"
                           placeholder="/home/.../.ssh/id_ed25519"
                           <?= $curSource==='profile'?'disabled':'' ?>>
                </label><br>
                <label>OpenSSH config (upload):
                    <input type="file" name="ssh_config" <?= $curSource==='profile'?'disabled':'' ?>>
                </label><br>
                <label>SSH key files referenced by the config (multiple):
                    <input type="file" name="ssh_keys[]" multiple <?= $curSource==='profile'?'disabled':'' ?>>
                </label>
                <p style="color:#777; font-size:0.85em;">
                    Profile source pulls encrypted creds from the vault; free-form /
                    DB-only source lets you supply ad-hoc creds or upload keys.
                </p>
            </fieldset>

            <fieldset style="border:1px solid #ddd; padding:6px 10px; margin:0 0 10px 0;">
                <legend>Connection discovery</legend>
                <label>
                    <input type="checkbox" name="also_connections" value="1" <?= $f['also_connections']?'checked':'' ?>>
                    Also discover connections (LLDP/CDP/FDB) after the device scan finishes
                </label>
                <label style="display:block; margin-top:6px;">Collector:
                    <select name="collector">
                        <option value="" <?= $f['collector']===''?'selected':'' ?>>all (merged)</option>
                        <?php foreach (['snmp-lldp','snmp-cdp','snmp-fdb','ssh-lldp','ssh-fdb'] as $c): ?>
                            <option value="<?= $c ?>" <?= $f['collector']===$c?'selected':'' ?>><?= $c ?></option>
                        <?php endforeach; ?>
                    </select>
                </label>
                <label>Per-host timeout (s):
                    <input type="number" name="timeout" value="<?= htmlspecialchars($f['timeout']) ?>" min="1" max="60" style="width:60px;">
                </label>
            </fieldset>

            <label style="display:block; margin:6px 0;">
                <input type="checkbox" name="auto_import" value="1" <?= $f['auto_import']?'checked':'' ?>>
                Auto-import every discovered device (skip manual VLAN review)
            </label>
            <button type="submit" name="do_scan" value="1" style="margin-top:8px;">Scan</button>
        </form>
        <form method="get" action="import.php" style="display:inline; margin:0;">
            <input type="hidden" name="clear" value="1">
            <button type="submit">Clean scan</button>
        </form>
        <?php if ($scanMessage): ?>
            <p><em><?= htmlspecialchars($scanMessage) ?></em></p>
        <?php endif; ?>
        <script>
        (function () {
            function toggle(method, source) {
                var snmp = document.getElementById('scan-snmp-fields');
                var ssh  = document.getElementById('scan-ssh-fields');
                if (snmp) snmp.style.display = (method === 'snmp') ? '' : 'none';
                if (ssh)  ssh.style.display  = (method === 'ssh')  ? '' : 'none';
                var sp = document.getElementById('scan-source-profile');
                var st = document.getElementById('scan-source-target');
                var sd = document.getElementById('scan-source-db');
                if (sp) sp.style.display = (source === 'profile') ? '' : 'none';
                if (st) st.style.display = (source === 'target')  ? '' : 'none';
                if (sd) sd.style.display = (source === 'db')     ? '' : 'none';
                var creds = document.getElementById('scan-creds');
                if (creds) {
                    var disable = (source === 'profile');
                    creds.querySelectorAll('input').forEach(function (inp) {
                        inp.disabled = disable;
                    });
                }
            }
            window.scanSourceToggle = function () {
                var src = 'target';
                document.querySelectorAll('input[name="scan_source"]').forEach(function (r) {
                    if (r.checked) src = r.value;
                });
                var m = 'snmp';
                document.querySelectorAll('input[name="scan_method"]').forEach(function (r) {
                    if (r.checked) m = r.value;
                });
                toggle(m, src);
            };
            window.scanMethodToggle = window.scanSourceToggle;
            // initial state
            var initSrc = '<?= $curSource ?>';
            var initM = '<?= $curMethod ?>';
            toggle(initM, initSrc);
            document.querySelectorAll('input[name="scan_method"]').forEach(function (r) {
                r.addEventListener('change', window.scanSourceToggle);
            });
            document.querySelectorAll('input[name="scan_source"]').forEach(function (r) {
                r.addEventListener('change', window.scanSourceToggle);
            });
        })();
        </script>
    </div>
    <?php
    return ob_get_clean();
}

// Keep old name available so other call sites don't break.
function scan_devices_live_panel_html($pf, $profiles, $scanMessage) {
    return scan_live_panel_html($profiles, $scanMessage);
}
