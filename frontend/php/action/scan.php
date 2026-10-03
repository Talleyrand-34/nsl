<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
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
        // Pull profile metadata: the scan method follows the profile's own
        // scan_source (snmp profiles sweep over SNMP; ssh profiles read each
        // attached device over SSH with per-row overrides). Encrypted blobs
        // (ssh_password, ssh_key) aren't returned by /scan/profiles — the
        // backend re-reads them from the store, so no cleartext flows here.
        $method = 'ssh';
        $profiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];
        foreach ($profiles as $p) {
            if (($p['name'] ?? '') === $f['profile']) {
                $profileName = $f['profile'];
                $osType = (string) ($p['os_type'] ?? '');
                $sshUser = (string) ($p['ssh_user'] ?? '');
                $sshKeyFile = (string) ($p['ssh_key_file'] ?? '');
                if ((string) ($p['scan_source'] ?? '') === 'snmp') {
                    $method = 'snmp';
                }
                break;
            }
        }
    } elseif ($source === 'db') {
        // Connection-only path: every device in the store. The connection
        // scan picks up each device's stored profile for creds.
        $method = 'snmp'; // force — device scan not the focus here
        $raw = @file_get_contents(DEVICES_ENDPOINT);
        $decoded = json_decode($raw, true);
        $hosts = [];
        if (is_array($decoded)) {
            foreach ($decoded as $d) {
                $ips = $d['ips'] ?? [];
                if (is_string($ips)) {
                    $ips = $ips === '' ? [] : [$ips];
                }
                if (is_array($ips) && count($ips) > 0) {
                    $ip = trim((string) $ips[0]);
                    if ($ip !== '') {
                        $hosts[] = $ip;
                    }
                }
            }
        }
        if (count($hosts) === 0) {
            $scanMessage = 'No devices in the store yet.';
            return;
        }
        $target = implode(',', $hosts);
        $profileName = '';
    } else {
        // 'target' — free-form.
        $method = $f['method'];
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
        // /scan/run's request struct spells this "timeout" (ScanRunRequest.Timeout);
        // "timeout_sec" is the *internal* RunScanOptions name and was silently
        // dropped on the wire, so every scan ran on the backend default.
        'timeout'      => intval($f['timeout']) ?: 10,
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
        $_SESSION['device_scan_id'] = $j['scan_id'];
        // Stash follow-up info on the session for do_scan_completed.
        $_SESSION['scan_followup'] = $f['also_connections']
            ? [
                'mode'      => $source,
                'profile'   => $profileName,
                'ssh_user'  => $sshUser,
                'collector' => $f['collector'],
                'timeout'   => $f['timeout'],
            ]
            : null;
        scan_log_push('device-scan', $j['scan_id'], 'started',
            $source . ' ' . $method . ' ' . $target, $method . ' ' . $target);
    } else {
        $scanMessage = 'Could not start scan (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
        scan_log_push('device-scan', '', 'failed',
            'HTTP ' . intval($code) . ' ' . ($body ?: $err), $method . ' ' . $target);
    }
}

// scan_error_detail turns a /scan/status reply into one human-readable reason a
// run ended badly. It never returns the bare string "unknown error": when the
// run itself reported nothing, the HTTP code and body are the next best
// evidence, and a 404 means the backend has no such run at all (server
// restarted, or the registry recycled the id).
function scan_error_detail($status, $code, $body) {
    $err = trim((string) ($status['error'] ?? ''));
    if ($err !== '') {
        return $err;
    }
    if ($code === 404) {
        return 'run not found in the backend registry (server restarted, or the run was recycled)';
    }
    if ($code !== 200) {
        $msg = '';
        $j = json_decode((string) $body, true);
        if (is_array($j)) {
            $msg = trim((string) ($j['message'] ?? $j['error'] ?? ''));
        }
        if ($msg === '') {
            $msg = trim((string) $body);
        }
        return 'status poll returned HTTP ' . intval($code) . ($msg !== '' ? ': ' . $msg : '');
    }
    return 'the backend reported state "' . (string) ($status['state'] ?? '') . '" with no error text';
}

// do_scan_completed advances any in-flight async scans and folds their results
// into the session. Returns the scan_id of a scan that is still running (so
// the status panel keeps watching it), or '' when everything has settled.
function do_scan_completed(&$discovered, &$scanMessage, &$plan, &$importedIPs, &$importMessage, &$result) {
    // 1. Device scan.
    $devId = $_SESSION['device_scan_id'] ?? '';
    if ($devId !== '') {
        list($code, $body) = api_method('GET', SCAN_STATUS_ENDPOINT . '?scan_id=' . urlencode($devId));
        $status = ($code === 200) ? (json_decode($body, true) ?: []) : [];
        $state  = $status['state'] ?? '';
        if ($state === 'running' || $state === 'pending') {
            return $devId;
        }
        $_SESSION['device_scan_id'] = '';
        // Log the *reason* a phase ended, not just a device count: a failed run
        // used to be recorded as "0 devices", which hid the backend's error and
        // made every failure look identical.
        if ($state === 'completed') {
            $detail = count($status['result']['devices'] ?? []) . ' devices';
        } elseif ($state === 'failed') {
            $detail = scan_error_detail($status, $code, $body);
        } else {
            $state  = 'failed';
            $detail = scan_error_detail($status, $code, $body);
        }
        scan_log_push('device-scan', $devId, $state, $detail, $status['title'] ?? '');
        if ($state === 'completed') {
            $res = $status['result'] ?? [];
            $discovered = $res['devices'] ?? [];
            $_SESSION['scan_discovered'] = $discovered;
            // Which run this payload came from. The browser-side pending queue
            // is scoped to it, so a new run resets the queue + tombstones
            // instead of filtering the fresh discovery through stale ones.
            $_SESSION['scan_discovered_run'] = $devId;
            // Optional chained connection scan, kicked off exactly once.
            $follow = $_SESSION['scan_followup'] ?? null;
            if (is_array($follow)) {
                $_SESSION['scan_followup'] = null;
                $cpayload = [
                    'community'   => 'public',
                    'collector'   => (string) ($follow['collector'] ?? ''),
                    'timeout_sec' => intval($follow['timeout'] ?? 0) ?: 10,
                    'ssh_user'    => (string) ($follow['ssh_user'] ?? ''),
                ];
                if (($follow['mode'] ?? '') === 'db') {
                    $cpayload['from_db'] = true;
                } else {
                    $cpayload['profiles'] = true;
                    // Only generic profiles are honoured as an SSH fallback;
                    // device profiles are ignored by the backend here.
                    $cpayload['generic_profile'] = (string) ($follow['profile'] ?? '');
                }
                list($cCode, $cBody) = api_post_json(SCAN_CONNECTIONS_ENDPOINT, json_encode($cpayload), 15);
                $cJson = json_decode($cBody, true);
                if (($cCode === 202 || $cCode === 200) && !empty($cJson['scan_id'])) {
                    $_SESSION['connection_scan_id'] = $cJson['scan_id'];
                } else {
                    $cJ = is_array($cJson) ? $cJson : [];
                    $cMsg = trim((string) ($cJ['message'] ?? $cJ['error'] ?? $cBody));
                    if ($cMsg === '') {
                        $cMsg = 'the API returned an empty body';
                    }
                    $scanMessage = 'Connection scan failed to start (HTTP ' . intval($cCode) . '): ' . htmlspecialchars($cMsg);
                    scan_log_push('conn-scan', '', 'failed',
                        'start refused: HTTP ' . intval($cCode) . ' ' . $cMsg, '');
                }
            }
        } else {
            $scanMessage = 'Scan failed: ' . htmlspecialchars($detail);
        }
    }
    $connId = $_SESSION['connection_scan_id'] ?? '';
    if ($connId !== '') {
        list($code, $body) = api_method('GET', SCAN_STATUS_ENDPOINT . '?scan_id=' . urlencode($connId));
        $status = ($code === 200) ? (json_decode($body, true) ?: []) : [];
        $state  = $status['state'] ?? '';
        if ($state === 'running' || $state === 'pending') {
            return $connId;
        }
        $_SESSION['connection_scan_id'] = '';
        $r = $status['result'] ?? [];
        if ($state === 'completed') {
            $detail = count($r['hosts'] ?? []) . ' hosts, ' . count($r['edges'] ?? []) . ' edges';
        } else {
            $state  = 'failed';
            $detail = scan_error_detail($status, $code, $body);
        }
        scan_log_push('conn-scan', $connId, $state, $detail, $status['title'] ?? '');
        if ($state === 'completed') {
            $_SESSION['connections_result'] = $r;
            $result = $r;
        } else {
            $scanMessage = trim($scanMessage . ' Connection scan failed: ' . htmlspecialchars($detail));
        }
    }
    return '';
}

// --- analyze / execute handlers --------------------------------------------
// Phase 5 follow-up: do_analyze + do_execute were lost when scan_devices.php
// was deleted during the Phase 1 refactor. Restored here so the discovered-
// devices table's "Configure & import" button and the plan-review "Import
// device" button work again.

// analyze_device POSTs a discovered device to /scan/analyze and returns the
// resulting import plan. Used by do_analyze below.
function analyze_device($device, &$err) {
    list($code, $body) = api_post_json(SCAN_ANALYZE_ENDPOINT, json_encode(['device' => $device]));
    if ($code === 200) {
        return json_decode($body, true);
    }
    $err = 'Analyze failed (HTTP ' . intval($code) . '): ' . $body;
    return null;
}

function do_analyze(&$plan, &$scanMessage) {
    $device = json_decode($_POST['device_json'] ?? 'null', true);
    if ($device) {
        $plan = analyze_device($device, $scanMessage);
        $ip = $device['device']['ip'] ?? '';
        scan_log_push('analyze', '', $plan ? 'completed' : 'failed',
            ($plan ? 'plan generated' : 'no plan: ' . $scanMessage) . ' ' . $ip, $ip);
    } else {
        $scanMessage = 'Invalid device payload.';
        scan_log_push('analyze', '', 'failed', 'invalid device payload');
    }
}

function do_execute(&$plan, &$importMessage, &$importedIPs) {
    $editedPlan = json_decode($_POST['plan_json'] ?? 'null', true);
    if (!$editedPlan) {
        $importMessage = 'Invalid plan.';
        return;
    }
    $vl = $_POST['vlan'] ?? [];
    $sn = $_POST['subnet'] ?? [];
    if (isset($editedPlan['interface_plans']) && is_array($editedPlan['interface_plans'])) {
        foreach ($editedPlan['interface_plans'] as $i => &$ipl) {
            if (!isset($ipl['ip_mappings']) || !is_array($ipl['ip_mappings'])) {
                continue;
            }
            foreach ($ipl['ip_mappings'] as $j => &$m) {
                if (isset($vl[$i][$j]) && $vl[$i][$j] !== '') {
                    $m['vlan_number'] = $vl[$i][$j];
                }
                if (isset($sn[$i][$j])) {
                    $m['subnet'] = $sn[$i][$j];
                }
            }
            unset($m);
        }
        unset($ipl);
    }
    $payload = json_encode(['plan' => $editedPlan, 'options' => ['default_zone' => 'Discovered']]);
    list($code, $body, $err) = api_post_json(SCAN_EXECUTE_ENDPOINT, $payload);
    $ip = (string) ($editedPlan['device']['device']['ip'] ?? '');
    if ($code === 200) {
        $importMessage = json_decode($body, true)['message'] ?? 'Import completed.';
        if ($ip !== '') {
            $importedIPs[] = $ip;
            if (!isset($_SESSION['scan_imported'])) {
                $_SESSION['scan_imported'] = [];
            }
            $_SESSION['scan_imported'] = array_values(array_unique(
                array_merge($_SESSION['scan_imported'], $importedIPs)
            ));
        }
        scan_log_push('execute', '', 'completed', 'imported ' . $ip, $ip);
    } else {
        $detail = json_decode($body, true)['message'] ?? ($body ?: $err);
        $importMessage = 'Import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($detail);
        scan_log_push('execute', '', 'failed',
            'HTTP ' . intval($code) . ' ' . $detail, $ip);
    }
}

// --- panels -----------------------------------------------------------------

 function scan_live_panel_html($profiles, $scanMessage, $osTypes = []) {
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

            <fieldset id="scan-creds" style="border:1px solid #ddd; padding:6px 10px; margin:0 0 10px 0;
                                          <?= $curSource==='target'?'':'display:none;' ?>">
                <legend>Credentials <span style="color:#777; font-weight:normal; font-size:0.85em;">
                    (only shown for free-form target)</span></legend>
                <label style="margin-right:14px;">
                    <input type="radio" name="scan_method" value="snmp" <?= $curMethod==='snmp'?'checked':'' ?>>
                    SNMP only
                </label>
                <label>
                    <input type="radio" name="scan_method" value="ssh" <?= $curMethod==='ssh'?'checked':'' ?>>
                    SSH only
                </label>
                <div id="scan-cred-snmp" style="margin-top:6px; <?= $curMethod==='snmp'?'':'display:none;' ?>">
                    <label>SNMP community: <input type="text" name="community" value="<?= htmlspecialchars($f['community']) ?>"></label>
                    <label>Version:
                        <select name="snmp_version">
                            <option value="2c" <?= $f['snmp_version']==='2c'?'selected':'' ?>>2c</option>
                            <option value="1"  <?= $f['snmp_version']==='1'?'selected':'' ?>>1</option>
                        </select>
                    </label>
                    <label>Port: <input type="number" name="snmp_port" value="<?= htmlspecialchars($f['snmp_port']) ?>" style="width:80px;"></label>
                </div>
                <div id="scan-cred-ssh" style="margin-top:6px; <?= $curMethod==='ssh'?'':'display:none;' ?>">
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
                        <select name="ssh_os_type">
                            <option value="">&mdash; none &mdash;</option>
                            <?php foreach ($osTypes as $ot):
                                $otName = (string) ($ot['name'] ?? '');
                                if ($otName === '') continue;
                            ?>
                                <option value="<?= htmlspecialchars($otName) ?>" <?= ($_POST['ssh_os_type'] ?? '') === $otName ? 'selected' : '' ?>><?= htmlspecialchars($otName) ?></option>
                            <?php endforeach; ?>
                        </select>
                    </label>
                    <label>SSH user:
                        <input type="text" name="ssh_user" id="scan-ssh-user" value="<?= htmlspecialchars($f['ssh_user']) ?>">
                    </label><br>
                    <label>SSH password:
                        <input type="password" name="ssh_password" id="scan-ssh-password"
                               value="<?= htmlspecialchars($f['ssh_password']) ?>">
                    </label><br>
                    <label>SSH private key file:
                        <input type="text" name="ssh_key_file" id="scan-ssh-key-file"
                               value="<?= htmlspecialchars($f['ssh_key_file']) ?>"
                               placeholder="/home/.../.ssh/id_ed25519">
                    </label><br>
                    <label>OpenSSH config (upload):
                        <input type="file" name="ssh_config">
                    </label><br>
                    <label>SSH key files referenced by the config (multiple):
                        <input type="file" name="ssh_keys[]" multiple>
                    </label>
                </div>
                <p style="color:#777; font-size:0.85em;">
                    SNMP fields appear for &ldquo;SNMP only&rdquo;; SSH fields appear for &ldquo;SSH only&rdquo;.
                    Profile and DB-only modes read creds from the store and skip this block.
                </p>
            </fieldset>

            <fieldset style="border:1px solid #ddd; padding:6px 10px; margin:0 0 10px 0;">
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
                // Source-specific blocks.
                var sp = document.getElementById('scan-source-profile');
                var st = document.getElementById('scan-source-target');
                var sd = document.getElementById('scan-source-db');
                if (sp) sp.style.display = (source === 'profile') ? '' : 'none';
                if (st) st.style.display = (source === 'target')  ? '' : 'none';
                if (sd) sd.style.display = (source === 'db')     ? '' : 'none';
                // Credentials block only appears for free-form target.
                var creds = document.getElementById('scan-creds');
                var showTargetFields = (source === 'target');
                if (creds) creds.style.display = showTargetFields ? '' : 'none';
                // Inside credentials: swap SNMP vs SSH sub-block based on the
                // chosen method. Both radios live inside the credentials
                // fieldset, so picking one swaps the inner content.
                var snmpFields = document.getElementById('scan-cred-snmp');
                var sshFields  = document.getElementById('scan-cred-ssh');
                if (snmpFields) snmpFields.style.display = (method === 'snmp') ? '' : 'none';
                if (sshFields)  sshFields.style.display  = (method === 'ssh')  ? '' : 'none';
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


// dev_ip resolves the IP of a discovered device. Used by the discovered-
// devices table and the JS pending-devices module.
function dev_ip($d) {
    return $d['device']['ip'] ?? '';
}

function scan_devices_discovered_panel_html($discovered, $importedIPs) {
    $pending = 0;
    if (!empty($discovered)) {
        foreach ($discovered as $d) {
            if (!in_array(dev_ip($d), $importedIPs, true)) $pending++;
        }
    }
    $initialJson = json_encode(array_values($discovered ?: []));
    $importedJson = json_encode($importedIPs);
    // Run id this payload belongs to. When the session has lost it (or the
    // devices came from an upload rather than a run), fall back to a
    // fingerprint of the IP set: stable across reloads of the same result,
    // different as soon as the result is.
    $runId = (string) ($_SESSION['scan_discovered_run'] ?? '');
    if ($runId === '') {
        $ips = array_map('dev_ip', $discovered ?: []);
        sort($ips);
        $runId = $ips ? 'fp-' . substr(sha1(implode(',', $ips)), 0, 12) : '';
    }
    $runJson = json_encode($runId);
    ob_start();
    ?>
    <h4>Discovered devices (<span class="discovered-total"><?= count($discovered) ?></span> total, <span class="discovered-pending"><?= $pending ?></span> pending)</h4>
    <p style="margin:4px 0; color:#777; font-size:0.85em;">
        <button type="button" id="cp-delete-all" style="font-size:0.85em;">Delete all pending</button>
    </p>
    <table border="1" cellpadding="4" cellspacing="0" class="discovered-devices">
        <thead><tr><th>Status</th><th>Name</th><th>IP</th><th>Brand</th><th>Model</th><th>Class</th><th></th></tr></thead>
        <tbody>
            <?php if (empty($discovered)): ?>
                <tr><td colspan="7" style="color:#777; text-align:center;">No pending devices. Run a scan from the Live scan panel to populate the queue.</td></tr>
            <?php else: ?>
                <?php foreach ($discovered as $d): $ip = dev_ip($d); $done = in_array($ip, $importedIPs, true); ?>
                    <tr<?= $done ? ' style="color:#888; background:#f3f3f3;"' : '' ?>>
                        <td><?= $done ? '✓ imported' : 'pending' ?></td>
                        <td><?= htmlspecialchars($d['suggested_name'] ?? '') ?></td>
                        <td><?= htmlspecialchars($ip) ?></td>
                        <td><?= htmlspecialchars($d['brand'] ?? '') ?></td>
                        <td><?= htmlspecialchars($d['model'] ?? '') ?></td>
                        <td><?= htmlspecialchars($d['model_type'] ?? '') ?></td>
                        <td>
                            <?php if (!$done): ?>
                                <form method="post" action="import.php" style="margin:0; display:inline;">
                                    <input type="hidden" name="device_json" value="<?= htmlspecialchars(json_encode($d)) ?>">
                                    <button type="submit" name="do_analyze" value="1">Configure &amp; import &rarr;</button>
                                </form>
                            <?php else: ?>&mdash;<?php endif; ?>
                        </td>
                    </tr>
                <?php endforeach; ?>
            <?php endif; ?>
        </tbody>
    </table>
    <script>
    // Server-rendered rows above are the no-JS fallback. pending-devices.js
    // merges this payload into the browser-side queue on DOMContentLoaded and
    // re-renders the tbody from it; this block only publishes the payload and
    // wires the reset button. Keep the merge/render logic in one place — a
    // second copy here is what previously left this script block unterminated
    // and swallowed the rest of the results column.
    window.NSL_PENDING_INITIAL = <?= $initialJson ?>;
    window.NSL_PENDING_IMPORTED_IPS = <?= $importedJson ?>;
    window.NSL_PENDING_RUN_ID = <?= $runJson ?>;
    (function () {
        function wire() {
            var btn = document.getElementById('cp-delete-all');
            if (!btn || !window.nslPD) return;
            btn.addEventListener('click', function () {
                window.nslPD.removeAllPending(window.NSL_PENDING_IMPORTED_IPS);
            });
        }
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', wire);
        } else {
            wire();
        }
    })();
    </script>
    <?php
    return ob_get_clean();
}

function scan_devices_plan_panel_html($plan) {
    if ($plan === null) {
        return '';
    }
    $dev = $plan['device'] ?? [];
    ob_start();
    ?>
    <h4>Review &amp; import: <?= htmlspecialchars($dev['suggested_name'] ?? '') ?>
        <span style="font-weight:normal; color:#555;">
            (<?= htmlspecialchars($dev['device']['ip'] ?? '') ?> — <?= htmlspecialchars($dev['brand'] ?? '') ?> <?= htmlspecialchars($dev['model'] ?? '') ?>,
            <?= htmlspecialchars($dev['model_type'] ?? '') ?>)
        </span>
    </h4>
    <p style="color:#555;"><?= htmlspecialchars($plan['summary'] ?? '') ?></p>
    <form method="post" action="import.php">
        <input type="hidden" name="plan_json" value="<?= htmlspecialchars(json_encode($plan)) ?>">
        <?php foreach (($plan['interface_plans'] ?? []) as $i => $ipl):
            $iface = $ipl['interface'] ?? [];
            $memberships = [];
            foreach (($iface['vlans'] ?? []) as $vm) {
                $memberships[] = $vm['vlan_number'] . ($vm['tagged'] ? 'T' : 'U');
            }
        ?>
            <fieldset style="margin-top:8px;">
                <legend><strong><?= htmlspecialchars($iface['name'] ?? '') ?></strong>
                    <?php if (!empty($iface['mac'])): ?> · <?= htmlspecialchars($iface['mac']) ?><?php endif; ?>
                    <?php if (!empty($iface['parent'])): ?> · parent <?= htmlspecialchars($iface['parent']) ?><?php endif; ?>
                    <?php if ($memberships): ?> · SNMP VLANs: <?= htmlspecialchars(implode(', ', $memberships)) ?><?php endif; ?>
                </legend>
                <?php if (empty($ipl['ip_mappings'])): ?>
                    <em>No IP addresses on this interface.</em>
                <?php else: ?>
                    <table border="1" cellpadding="3" cellspacing="0">
                        <tr><th>IP</th><th>IP segment (subnet)</th><th>VLAN id</th><th>Confidence / reason</th></tr>
                        <?php foreach ($ipl['ip_mappings'] as $j => $m): ?>
                            <tr>
                                <td><?= htmlspecialchars($m['ip'] ?? '') ?></td>
                                <td><input type="text" name="subnet[<?= $i ?>][<?= $j ?>]" value="<?= htmlspecialchars($m['subnet'] ?? '') ?>" size="18"></td>
                                <td><input type="text" name="vlan[<?= $i ?>][<?= $j ?>]" value="<?= htmlspecialchars($m['vlan_number'] ?? '') ?>" style="width:70px;"></td>
                                <td style="color:#555;"><?= htmlspecialchars(($m['confidence'] ?? '') . ' — ' . ($m['reason'] ?? '')) ?></td>
                            </tr>
                        <?php endforeach; ?>
                    </table>
                <?php endif; ?>
            </fieldset>
        <?php endforeach; ?>
        <button type="submit" name="do_execute" value="1" style="margin-top:10px;">Import device</button>
    </form>
    <?php
    return ob_get_clean();
}

function scan_connections_results_panel_html($result, $scanMessage) {
    if (!is_array($result) || empty($result)) {
        return '';
    }
    $hosts    = $result['hosts'] ?? [];
    $edges    = $result['edges'] ?? [];
    $inter    = $result['intermediaries'] ?? [];
    $dgSvg = '';
    list($dgCode, $dgBody) = api_post_json(SCAN_CONNECTIONS_DIAGRAM_ENDPOINT, json_encode($result), 30);
    if ($dgCode === 200 && $dgBody !== '') {
        $dgSvg = $dgBody;
    }
    ob_start();
    ?>
    <section class="connections-results">
      <p style="color:#555; font-size:0.9em;">
        Hosts discovered: <strong><?= count($hosts) ?></strong>.
        Edges: <strong><?= count($edges) ?></strong><?php if ($inter): ?>.
        Intermediary links: <strong><?= count($inter) ?></strong><?php endif; ?>.
      </p>
      <?php if ($dgSvg !== ''): ?>
        <details open>
          <summary>Topology diagram</summary>
          <div class="resizable-img-container" style="height:480px; border:1px solid #ddd; overflow:auto; margin:8px 0;">
            <?= $dgSvg ?>
          </div>
        </details>
      <?php endif; ?>
      <?php if (!empty($edges)): ?>
        <details open>
          <summary>Discovered edges (<?= count($edges) ?>)</summary>
          <table border="1" cellpadding="3" cellspacing="0">
            <thead><tr><th>Mark</th><th>From</th><th>To</th><th>Via</th></tr></thead>
            <tbody>
              <?php foreach ($edges as $e):
                $mark = $e['confidence'] ?? ($e['remote_resolved'] ? 'confirmed' : 'unresolved');
                $via  = implode(', ', $e['provenance'] ?? []);
              ?>
                <tr>
                  <td><?= htmlspecialchars($mark) ?></td>
                  <td><?= htmlspecialchars($e['from'] ?? '') ?></td>
                  <td><?= htmlspecialchars($e['to'] ?? '') ?></td>
                  <td style="color:#555; font-size:0.85em;"><?= htmlspecialchars($via) ?></td>
                </tr>
              <?php endforeach; ?>
            </tbody>
          </table>
        </details>
      <?php endif; ?>
      <?php if (!empty($inter)): ?>
        <details>
          <summary>Intermediary links (<?= count($inter) ?>)</summary>
          <table border="1" cellpadding="3" cellspacing="0">
            <thead><tr><th>Placeholder</th><th>Seen by</th></tr></thead>
            <tbody>
              <?php foreach ($inter as $in): ?>
                <tr>
                  <td><?= htmlspecialchars($in['device'] ?? '') ?></td>
                  <td><?= htmlspecialchars(implode(', ', $in['seen_by'] ?? [])) ?></td>
                </tr>
              <?php endforeach; ?>
            </tbody>
          </table>
        </details>
      <?php endif; ?>
      <?php if ($scanMessage): ?>
        <p><em><?= $scanMessage ?></em></p>
      <?php endif; ?>
    </section>
    <?php
    return ob_get_clean();
}

// Phase 5: combined results panel — wraps both discovered devices + connection
// edges in one card with two collapsible sections. Falls back gracefully when
// either side has no data.
function scan_results_panel_html($discovered, $importedIPs, $plan, $connResult, $scanMessage) {
    $hasDevices = !empty($discovered) || $plan !== null;
    $hasConn    = is_array($connResult) && !empty($connResult);
    if (!$hasDevices && !$hasConn) {
        return '';
    }
    ob_start();
    ?>
    <section class="accordion-card" data-accordion="results">
      <h3 class="accordion-header"><button type="button" class="accordion-toggle" aria-expanded="true">Scan results</button></h3>
      <div class="accordion-body">
        <details open class="result-section">
          <summary><strong>Discovered devices</strong></summary>
          <?php if ($hasDevices): ?>
            <?= scan_devices_discovered_panel_html($discovered, $importedIPs) ?>
            <?= scan_devices_plan_panel_html($plan) ?>
          <?php else: ?>
            <p style="color:#777; font-style:italic;">No device scan results yet.</p>
          <?php endif; ?>
        </details>
        <details open class="result-section">
          <summary><strong>Connection edges</strong></summary>
          <?php if ($hasConn): ?>
            <?= scan_connections_results_panel_html($connResult, $scanMessage) ?>
          <?php else: ?>
            <p style="color:#777; font-style:italic;">No connection scan results yet. Tick &ldquo;Also discover connections&rdquo; in the Live scan panel to chain a connection scan.</p>
          <?php endif; ?>
        </details>
      </div>
    </section>
    <?php
    return ob_get_clean();
}

