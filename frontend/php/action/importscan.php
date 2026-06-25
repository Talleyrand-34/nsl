<?php
require_once __DIR__ . '/../config.php';

/** api_post_json POSTs a JSON string; returns [httpCode, body, curlError]. */
function api_post_json($url, $jsonBody) {
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
    curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonBody);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
        'Content-Length: ' . strlen($jsonBody),
    ]);
    $body = curl_exec($ch);
    $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $err  = curl_error($ch);
    curl_close($ch);
    return [$code, $body, $err];
}

/** api_method issues a request with an optional JSON body; returns [code, body]. */
function api_method($method, $url, $jsonBody = null) {
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $headers = [];
    if ($jsonBody !== null) {
        curl_setopt($ch, CURLOPT_POSTFIELDS, $jsonBody);
        $headers[] = 'Content-Type: application/json';
        $headers[] = 'Content-Length: ' . strlen($jsonBody);
    }
    curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
    $body = curl_exec($ch);
    $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    return [$code, $body];
}

/** analyze_device POSTs a discovered device to /scan/analyze; returns the plan array or null. */
function analyze_device($device, &$err) {
    list($code, $body) = api_post_json(SCAN_ANALYZE_ENDPOINT, json_encode(['device' => $device]));
    if ($code === 200) {
        return json_decode($body, true);
    }
    $err = 'Analyze failed (HTTP ' . intval($code) . '): ' . $body;
    return null;
}

/** dev_ip returns a discovered device's management IP (its identity in the list). */
function dev_ip($d) {
    return $d['device']['ip'] ?? '';
}

/** import_one analyzes a discovered device and imports it with the default plan
 *  (no manual review). Returns [bool ok, string message]. */
function import_one($device) {
    $err = '';
    $plan = analyze_device($device, $err);
    if ($plan === null) {
        return [false, $err ?: 'analyze failed'];
    }
    $payload = json_encode(['plan' => $plan, 'options' => ['default_zone' => 'Discovered']]);
    list($code, $body, $e2) = api_post_json(SCAN_EXECUTE_ENDPOINT, $payload);
    if ($code === 200) {
        return [true, json_decode($body, true)['message'] ?? 'Imported.'];
    }
    $detail = json_decode($body, true)['message'] ?? ($body ?: $e2);
    return [false, $detail];
}

$scanMessage    = '';
$importMessage  = '';
$profileMessage = '';
$plan           = null;  // analyzed import plan to review/edit
$scanRunId      = '';    // when set, a scan just started: render the live panel
$autoReload     = '';    // extra query params the live panel carries on completion

// The discovered devices and which IPs were already imported persist in the
// session across the analyze/import round-trips, so importing one device (or its
// interfaces) never discards the rest of the scan — no need to re-scan.
$discovered  = $_SESSION['scan_discovered'] ?? [];
$importedIPs = $_SESSION['scan_imported'] ?? [];

// "Clean scan" clears the persisted results.
if (isset($_GET['clear'])) {
    unset($_SESSION['scan_discovered'], $_SESSION['scan_imported']);
    $discovered = [];
    $importedIPs = [];
}

// SNMP form prefill (from a loaded profile, else POST).
$pf = [
    'target'       => $_POST['target'] ?? '',
    'community'    => $_POST['community'] ?? 'public',
    'snmp_version' => $_POST['snmp_version'] ?? '2c',
    'snmp_port'    => $_POST['snmp_port'] ?? '161',
];

// Saved profiles for the dropdowns.
$profiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true);
if (!is_array($profiles)) {
    $profiles = [];
}

// --- Profile: save / delete / load -----------------------------------------
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_create_profile'])) {
    $name = trim($_POST['cp_name'] ?? '');
    if ($name === '') {
        $profileMessage = 'Enter a profile name.';
    } else {
        // An uploaded SSH private key is read and sent as content (encrypted at rest).
        $sshKey = '';
        if (isset($_FILES['cp_ssh_key_file']) && $_FILES['cp_ssh_key_file']['error'] === UPLOAD_ERR_OK) {
            $sshKey = (string) file_get_contents($_FILES['cp_ssh_key_file']['tmp_name']);
        }
        // The profile type combines kind (device|generic) and scan source
        // (snmp|ssh) into one of four choices, e.g. "generic-ssh".
        $cpType   = $_POST['cp_type'] ?? 'generic-ssh';
        $cpKind   = strpos($cpType, 'generic') === 0 ? 'generic' : 'device';
        $cpSource = substr($cpType, -3) === 'ssh' ? 'ssh' : 'snmp';
        $payload = json_encode([
            'name'           => $name,
            'kind'           => $cpKind,
            'host'           => trim($_POST['cp_host'] ?? ''),
            'snmp_community' => trim($_POST['cp_community'] ?? 'public'),
            'snmp_version'   => trim($_POST['cp_version'] ?? '2c'),
            'snmp_port'      => intval($_POST['cp_port'] ?? 161),
            'scan_source'    => $cpSource,
            'device_type'    => trim($_POST['cp_device_type'] ?? ''),
            'ssh_user'       => trim($_POST['cp_ssh_user'] ?? ''),
            'ssh_password'   => $_POST['cp_ssh_password'] ?? '',
            'ssh_key'        => $sshKey,
        ]);
        list($code, $body) = api_method('POST', SCAN_PROFILES_ENDPOINT, $payload);
        $profileMessage = ($code === 201)
            ? "Profile \"$name\" created."
            : 'Create failed: ' . htmlspecialchars($body);
        $profiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];
    }
}
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_delete_profile'])) {
    $name = $_POST['profile_name'] ?? '';
    list($code, $body) = api_method('DELETE', SCAN_PROFILES_ENDPOINT . '?name=' . urlencode($name));
    $profileMessage = ($code === 200) ? "Profile \"$name\" deleted." : 'Delete failed: ' . htmlspecialchars($body);
    $profiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];
}
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_load_profile'])) {
    $name = $_POST['profile_select'] ?? '';
    foreach ($profiles as $p) {
        if (($p['name'] ?? '') === $name) {
            $pf['target']       = $p['host'] ?? '';
            $pf['community']    = $p['snmp_community'] ?? 'public';
            $pf['snmp_version'] = $p['snmp_version'] ?? '2c';
            $pf['snmp_port']    = (string) ($p['snmp_port'] ?? 161);
            $profileMessage     = "Loaded profile \"$name\" into the form.";
            break;
        }
    }
}

// --- Step: run a live scan (method = snmp|ssh; single vs batch inferred) ------
// One unified call: a bare IP is a single host, a CIDR or comma-separated list
// is a batch. SNMP uses community/version/port; SSH uses a device profile's
// stored credentials + device_type (the target overrides the profile's host).
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_scan'])) {
    $method     = ($_POST['scan_method'] ?? 'snmp') === 'ssh' ? 'ssh' : 'snmp';
    $target     = trim($_POST['target'] ?? '');
    $autoImport = !empty($_POST['auto_import']);
    if ($target === '') {
        $scanMessage = 'Please enter a target IP or CIDR.';
    } elseif ($method === 'ssh' && ($_POST['ssh_profile'] ?? '') === '') {
        $scanMessage = 'Select a device profile for the SSH scan.';
    } else {
        $payload = [
            'method'       => $method,
            'target'       => $target,
            'community'    => trim($_POST['community'] ?? 'public'),
            'snmp_version' => trim($_POST['snmp_version'] ?? '2c'),
            'snmp_port'    => intval($_POST['snmp_port'] ?? 161),
            'profile'      => $_POST['ssh_profile'] ?? '',
            'device_type'  => trim($_POST['ssh_device_type'] ?? ''),
        ];
        // The scan now runs async: this returns a scan_id immediately; the live
        // panel below polls /scan/status and reloads with ?scan_id= when done.
        list($code, $body, $err) = api_post_json(SCAN_RUN_ENDPOINT, json_encode($payload));
        $j = json_decode($body, true);
        if (($code === 202 || $code === 200) && !empty($j['scan_id'])) {
            $scanRunId  = $j['scan_id'];
            $autoReload = $autoImport ? 'auto_import=1' : '';
        } else {
            $scanMessage = 'Could not start scan (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
        }
    }
}

// --- Step: a started scan finished — fetch its result by id and render --------
if ($_SERVER['REQUEST_METHOD'] === 'GET' && isset($_GET['scan_id'])) {
    list($code, $body) = api_method('GET', SCAN_STATUS_ENDPOINT . '?scan_id=' . urlencode($_GET['scan_id']));
    $st = json_decode($body, true);
    $state = is_array($st) ? ($st['state'] ?? '') : '';
    if ($code === 200 && $state === 'completed') {
        $discovered = $st['result']['devices'] ?? [];
        $_SESSION['scan_discovered'] = $discovered;
        $_SESSION['scan_imported']   = [];
        $importedIPs = [];
        if (!empty($discovered)) {
            $scanMessage = count($discovered) . ' device(s) discovered.';
            if (!empty($_GET['auto_import'])) {
                // Accept results without review: analyze + import every device.
                $ok = 0; $fail = 0; $fails = [];
                foreach ($discovered as $d) {
                    list($success, $m) = import_one($d);
                    if ($success) {
                        $ok++;
                        $importedIPs[] = dev_ip($d);
                    } else {
                        $fail++;
                        $fails[] = (dev_ip($d) ?: '?') . ' — ' . $m;
                    }
                }
                $_SESSION['scan_imported'] = array_values(array_unique($importedIPs));
                $importMessage = "Auto-import: {$ok} imported, {$fail} failed."
                    . ($fails ? ' Failures: ' . implode('; ', $fails) : '');
            } elseif (count($discovered) === 1) {
                $plan = analyze_device($discovered[0], $scanMessage);
            }
        } else {
            $scanMessage = 'No devices discovered (is the SNMP/SSH port open and reachable?).';
        }
    } elseif ($code === 200 && $state === 'failed') {
        $scanMessage = 'Scan failed: ' . htmlspecialchars($st['error'] ?? 'unknown error');
    } elseif ($code === 200 && $state === 'running') {
        // Not finished yet — keep showing the live panel for this id.
        $scanRunId = $_GET['scan_id'];
    } else {
        $scanMessage = 'Scan not found (it may have expired). Run it again.';
    }
}

// --- Step: analyze one chosen device (from a multi-device list) -------------
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_analyze'])) {
    $device = json_decode($_POST['device_json'] ?? 'null', true);
    if ($device) {
        $plan = analyze_device($device, $scanMessage);
    }
}

// --- Step: execute the (edited) import plan ---------------------------------
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_execute'])) {
    $editedPlan = json_decode($_POST['plan_json'] ?? 'null', true);
    if (!$editedPlan) {
        $importMessage = 'Invalid plan.';
    } else {
        $vl = $_POST['vlan'] ?? [];
        $sn = $_POST['subnet'] ?? [];
        // NB: iterate the real array elements by reference (not "$x ?? []", which
        // would iterate a copy and silently drop the edits).
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
        if ($code === 200) {
            $importMessage = json_decode($body, true)['message'] ?? 'Import completed.';
            // Mark this device imported but keep the rest of the scan on screen, so
            // the next device can be imported without re-scanning. $plan stays null
            // so the review form closes and the list (with this one ticked) shows.
            $ip = $editedPlan['device']['device']['ip'] ?? '';
            if ($ip !== '') {
                $importedIPs[] = $ip;
                $_SESSION['scan_imported'] = array_values(array_unique($importedIPs));
            }
        } else {
            // Show the API's human message (e.g. a name clash) rather than raw JSON.
            $detail = json_decode($body, true)['message'] ?? ($body ?: $err);
            $importMessage = 'Import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($detail);
        }
    }
}

// --- Step: import an uploaded scan-result JSON file -------------------------
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_upload'])) {
    if (!isset($_FILES['scanfile']) || $_FILES['scanfile']['error'] !== UPLOAD_ERR_OK) {
        $importMessage = 'Please choose a valid scan-result JSON file.';
    } else {
        $contents = file_get_contents($_FILES['scanfile']['tmp_name']);
        if ($contents === false || trim($contents) === '') {
            $importMessage = 'Uploaded file is empty.';
        } elseif (json_decode($contents) === null && json_last_error() !== JSON_ERROR_NONE) {
            $importMessage = 'Uploaded file is not valid JSON.';
        } else {
            list($code, $body, $err) = api_post_json(SCAN_IMPORT_FILE_ENDPOINT, $contents);
            $importMessage = ($code === 200)
                ? (json_decode($body, true)['message'] ?? 'Import completed.')
                : 'Import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
        }
    }
}
?>

<h2>Import devices</h2>

<?php if ($importMessage): ?>
    <p style="padding:8px; background:#e8f5e9; border:1px solid #a5d6a7;"><strong><?= htmlspecialchars($importMessage) ?></strong></p>
<?php endif; ?>

<?php if ($scanRunId): ?>
    <div id="scan-status" class="scan-status">
        <h4><span class="spinner"></span>Scanning…</h4>
        <div class="scan-state" id="scan-state">starting…</div>
        <div class="scan-bar" id="scan-bar" style="display:none;"><div class="scan-bar-fill" id="scan-bar-fill"></div></div>
        <div class="scan-events" id="scan-events"></div>
    </div>
    <script>nslWatchScan(<?= json_encode($scanRunId) ?>, {reloadParams: <?= json_encode($autoReload) ?>});</script>
<?php endif; ?>

<?php if ($profileMessage): ?>
    <p style="padding:6px; background:#eef; border:1px solid #99c;"><strong><?= htmlspecialchars($profileMessage) ?></strong></p>
<?php endif; ?>
<?php ob_start(); // buffer the saved-profiles section (selector + table) ?>
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
<?php $profilesSection = ob_get_clean(); ?>
<?php ob_start(); // buffer the live-scan box (kept visible) ?>
    <!-- Live scan box -->
    <div class="box">
        <h3>Live scan</h3>
        <?php $curMethod = ($_POST['scan_method'] ?? 'snmp') === 'ssh' ? 'ssh' : 'snmp'; ?>
        <form method="post" action="import.php">
            <label>Method:
                <select name="scan_method" id="scan_method" onchange="scanMethodToggle()">
                    <option value="snmp" <?= $curMethod === 'snmp' ? 'selected' : '' ?>>SNMP</option>
                    <option value="ssh"  <?= $curMethod === 'ssh' ? 'selected' : '' ?>>SSH</option>
                </select>
            </label><br>
            <label>Target — IP or CIDR (a bare IP scans one host; a CIDR or comma-separated list scans many):
                <input type="text" name="target" value="<?= htmlspecialchars($pf['target']) ?>" placeholder="10.0.2.245 &nbsp;or&nbsp; 10.0.2.0/26" size="40">
            </label><br>

            <div id="snmp_fields">
                <label>SNMP community: <input type="text" name="community" value="<?= htmlspecialchars($pf['community']) ?>"></label>
                <label>Version:
                    <select name="snmp_version">
                        <option value="2c" <?= ($pf['snmp_version'] === '2c') ? 'selected' : '' ?>>2c</option>
                        <option value="1" <?= ($pf['snmp_version'] === '1') ? 'selected' : '' ?>>1</option>
                    </select>
                </label>
                <label>Port: <input type="number" name="snmp_port" value="<?= htmlspecialchars($pf['snmp_port']) ?>" style="width:80px;"></label>
            </div>

            <div id="ssh_fields">
                <p style="margin:6px 0; color:#555;"><em>SSH reads the device config using a profile's stored credentials. Pick a
                    profile — <strong>device</strong> or <strong>generic</strong> (generic credentials are reusable across many hosts). The target above
                    overrides the profile's host; a CIDR scans every SSH-open host. An <strong>OS / firmware type</strong> is required so the right config
                    parser is used (e.g. openwrt, opnsense — this is the operating system, <strong>not</strong> the hardware model): a device profile
                    supplies its own, or set one below (required for generic profiles).</em></p>
                <label>SSH profile (device or generic):
                    <select name="ssh_profile">
                        <option value="">— select —</option>
                        <?php foreach ($profiles as $p):
                            $pk = ($p['kind'] ?? '') !== '' ? $p['kind'] : 'device';
                            $desc = ($p['name'] ?? '') . ' (' . $pk . (($p['host'] ?? '') !== '' ? ', ' . $p['host'] : '') . ')'; ?>
                            <option value="<?= htmlspecialchars($p['name'] ?? '') ?>" <?= (($_POST['ssh_profile'] ?? '') === ($p['name'] ?? '')) ? 'selected' : '' ?>><?= htmlspecialchars($desc) ?></option>
                        <?php endforeach; ?>
                    </select>
                </label>
                <label>OS / firmware type <small>(operating system for config parsing — not the hardware model; blank = use the profile's, required for generic profiles)</small>:
                    <input type="text" name="ssh_device_type" list="ssh_devtypes" value="<?= htmlspecialchars($_POST['ssh_device_type'] ?? '') ?>" placeholder="openwrt / opnsense / fortinet / cisco">
                    <datalist id="ssh_devtypes"><option value="openwrt"><option value="opnsense"><option value="fortinet"><option value="cisco"></datalist>
                </label>
                <p style="margin:4px 0; color:#777; font-size:0.85em;">The profile's stored SSH secret is decrypted by the credential vault — unlock it from the app bar before scanning.</p>
            </div>

            <label style="display:block; margin-top:6px;">
                <input type="checkbox" name="auto_import" value="1" <?= !empty($_POST['auto_import']) ? 'checked' : '' ?>>
                Auto-import results (accept and import every discovered device without manual VLAN review)
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
        function scanMethodToggle() {
            var ssh = document.getElementById('scan_method').value === 'ssh';
            document.getElementById('snmp_fields').style.display = ssh ? 'none' : '';
            document.getElementById('ssh_fields').style.display  = ssh ? '' : 'none';
        }
        scanMethodToggle();
        </script>
    </div>
<?php $liveScan = ob_get_clean(); ?>
<?php ob_start(); // buffer the create-profile box ?>
    <!-- Create-profile box -->
    <div class="box">
        <h3>Create profile</h3>
        <form method="post" action="import.php" enctype="multipart/form-data">
            <label>Name: <input type="text" name="cp_name" required></label><br>
            <div class="cp-type-buttons">Type:
                <button type="button" class="cp-type-btn" data-type="device-snmp">Device · SNMP</button>
                <button type="button" class="cp-type-btn" data-type="device-ssh">Device · SSH</button>
                <button type="button" class="cp-type-btn" data-type="generic-snmp">Generic · SNMP</button>
                <button type="button" class="cp-type-btn" data-type="generic-ssh">Generic · SSH</button>
            </div>
            <input type="hidden" name="cp_type" id="cp_type" value="generic-ssh">
            <p id="cp_type_hint" style="margin:4px 0; color:#555; font-size:0.85em;"></p>

            <!-- Host: device profiles only (bound to a host). -->
            <div class="cp-grp" data-show="device-snmp device-ssh">
                <label>Host / subnet: <input type="text" name="cp_host" placeholder="10.0.0.1"></label><br>
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
                    <input type="text" name="cp_device_type" placeholder="opnsense / openwrt / fortinet / cisco" size="20">
                </label><br>
                <p style="margin:6px 0; color:#555;"><em>SSH credentials:</em></p>
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
            }
            document.querySelectorAll('.cp-type-btn').forEach(function (b) {
                b.addEventListener('click', function () { cpApplyType(b.dataset.type); });
            });
            cpApplyType('generic-ssh'); // default
        })();
        </script>
    </div>
<?php
$createForm = ob_get_clean();
// The live-scan box (primary action) stays visible; the noisy profile/credential
// blocks are tucked into independent collapsible sections.
$profCount = count($profiles);
?>
<?= $liveScan ?>
<details class="cred-flat"><summary>Saved scan profiles (<?= $profCount ?>)</summary><?= $profilesSection ?></details>
<details class="cred-flat"><summary>Create profile</summary><?= $createForm ?></details>

<?php
// Discovered-devices list (persisted across the import round-trips). Imported
// devices stay listed (ticked) so the rest can be imported without re-scanning.
if (!empty($discovered)):
    $pending = 0;
    foreach ($discovered as $d) {
        if (!in_array(dev_ip($d), $importedIPs, true)) $pending++;
    }
?>
    <h4>Discovered devices (<?= count($discovered) ?> total, <?= $pending ?> pending)</h4>
    <table border="1" cellpadding="4" cellspacing="0">
        <tr><th>Status</th><th>Name</th><th>IP</th><th>Brand</th><th>Model</th><th>Class</th><th></th></tr>
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
                        <form method="post" action="import.php" style="margin:0;">
                            <input type="hidden" name="device_json" value="<?= htmlspecialchars(json_encode($d)) ?>">
                            <button type="submit" name="do_analyze" value="1">Configure &amp; import &rarr;</button>
                        </form>
                    <?php else: ?>&mdash;<?php endif; ?>
                </td>
            </tr>
        <?php endforeach; ?>
    </table>
<?php endif; ?>

<?php
// Editable VLAN plan for a single device.
if ($plan !== null):
    $dev = $plan['device'] ?? [];
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
<?php endif; ?>

<div class="box" style="margin-top:24px;">
    <h3>Upload scan-result JSON</h3>
    <p style="color:#666; font-size:0.9em;">Upload a JSON file produced by <code>nsl-graph scan</code> (same format as <code>scan import &lt;file&gt;</code>).</p>
    <form method="post" action="import.php" enctype="multipart/form-data">
        <input type="file" name="scanfile" accept=".json,application/json" required>
        <button type="submit" name="do_upload" value="1">Upload &amp; import</button>
    </form>
</div>
