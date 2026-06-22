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

$scanMessage    = '';
$importMessage  = '';
$profileMessage = '';
$discovered     = [];    // list of discovered-device objects
$plan           = null;  // analyzed import plan to review/edit

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
        $payload = json_encode([
            'name'           => $name,
            'host'           => trim($_POST['cp_host'] ?? ''),
            'snmp_community' => trim($_POST['cp_community'] ?? 'public'),
            'snmp_version'   => trim($_POST['cp_version'] ?? '2c'),
            'snmp_port'      => intval($_POST['cp_port'] ?? 161),
            'scan_source'    => $_POST['cp_scan_source'] ?? 'snmp',
            'device_type'    => trim($_POST['cp_device_type'] ?? ''),
            'ssh_user'       => trim($_POST['cp_ssh_user'] ?? ''),
            'ssh_password'   => $_POST['cp_ssh_password'] ?? '',
            'ssh_key'        => $sshKey,
            'passphrase'     => $_POST['cp_passphrase'] ?? '',
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

// --- Step: run a live scan (SNMP host/subnet, or SSH via profile) ------------
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_scan'])) {
    $scanType = $_POST['scan_type'] ?? 'host';

    if ($scanType === 'ssh') {
        $profile    = $_POST['ssh_profile'] ?? '';
        $passphrase = $_POST['passphrase'] ?? '';
        if ($profile === '') {
            $scanMessage = 'Select a profile for the SSH scan.';
        } else {
            list($code, $body, $err) = api_post_json(SCAN_HOST_SSH_ENDPOINT, json_encode([
                'profile'    => $profile,
                'passphrase' => $passphrase,
            ]));
            if ($code === 200) {
                $discovered = [json_decode($body, true)];
            } else {
                $scanMessage = 'SSH scan failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
            }
        }
    } else {
        $target    = trim($_POST['target'] ?? '');
        $community = trim($_POST['community'] ?? 'public');
        $version   = trim($_POST['snmp_version'] ?? '2c');
        $port      = intval($_POST['snmp_port'] ?? 161);
        if ($target === '') {
            $scanMessage = 'Please enter a host IP or subnet.';
        } elseif ($scanType === 'network') {
            list($code, $body, $err) = api_post_json(SCAN_NETWORK_ENDPOINT, json_encode([
                'subnet' => $target, 'community' => $community, 'snmp_version' => $version, 'snmp_port' => $port,
            ]));
            if ($code === 200) {
                $discovered = json_decode($body, true)['devices'] ?? [];
            } else {
                $scanMessage = 'Scan failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
            }
        } else {
            list($code, $body, $err) = api_post_json(SCAN_HOST_ENDPOINT, json_encode([
                'ip' => $target, 'community' => $community, 'snmp_version' => $version, 'snmp_port' => $port,
            ]));
            if ($code === 200) {
                $d = json_decode($body, true);
                $discovered = ($d && isset($d['device'])) ? [$d] : [];
            } else {
                $scanMessage = 'Scan failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
            }
        }
    }

    if (!empty($discovered)) {
        $scanMessage = count($discovered) . ' device(s) discovered.';
        // Single device: go straight to the editable VLAN plan.
        if (count($discovered) === 1) {
            $plan = analyze_device($discovered[0], $scanMessage);
        }
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
        } else {
            $importMessage = 'Import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
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

<h3>Saved scan profiles</h3>
<?php if ($profileMessage): ?>
    <p style="padding:6px; background:#eef; border:1px solid #99c;"><strong><?= htmlspecialchars($profileMessage) ?></strong></p>
<?php endif; ?>
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
        <tr><th>Name</th><th>Host</th><th>Community</th><th>Ver</th><th>SSH pw</th><th>SSH key</th><th></th></tr>
        <?php foreach ($profiles as $p): ?>
            <tr>
                <td><?= htmlspecialchars($p['name'] ?? '') ?></td>
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

<div class="grid2">
    <!-- Live scan box -->
    <div class="box">
        <h3>Live scan</h3>
        <form method="post" action="import.php">
            <label>Scan type:
                <select name="scan_type">
                    <option value="host"    <?= (($_POST['scan_type'] ?? 'host') === 'host') ? 'selected' : '' ?>>SNMP — single host</option>
                    <option value="network" <?= (($_POST['scan_type'] ?? '') === 'network') ? 'selected' : '' ?>>SNMP — subnet</option>
                    <option value="ssh"     <?= (($_POST['scan_type'] ?? '') === 'ssh') ? 'selected' : '' ?>>SSH — via profile</option>
                </select>
            </label>
            <p style="margin:6px 0; color:#555;"><em>SNMP:</em></p>
            <label>Host IP / Subnet (CIDR):
                <input type="text" name="target" value="<?= htmlspecialchars($pf['target']) ?>" placeholder="192.168.1.1 or 192.168.1.0/24">
            </label><br>
            <label>SNMP community: <input type="text" name="community" value="<?= htmlspecialchars($pf['community']) ?>"></label>
            <label>Version:
                <select name="snmp_version">
                    <option value="2c" <?= ($pf['snmp_version'] === '2c') ? 'selected' : '' ?>>2c</option>
                    <option value="1" <?= ($pf['snmp_version'] === '1') ? 'selected' : '' ?>>1</option>
                </select>
            </label>
            <label>Port: <input type="number" name="snmp_port" value="<?= htmlspecialchars($pf['snmp_port']) ?>" style="width:80px;"></label>
            <p style="margin:6px 0; color:#555;"><em>SSH (uses a profile's stored, encrypted credentials):</em></p>
            <label>Profile:
                <select name="ssh_profile">
                    <option value="">— select —</option>
                    <?php foreach ($profiles as $p): ?>
                        <option value="<?= htmlspecialchars($p['name'] ?? '') ?>"><?= htmlspecialchars(($p['name'] ?? '') . ' (' . ($p['host'] ?? '') . ')') ?></option>
                    <?php endforeach; ?>
                </select>
            </label>
            <label>Passphrase: <input type="password" name="passphrase"></label>
            <br>
            <button type="submit" name="do_scan" value="1" style="margin-top:8px;">Scan</button>
        </form>
        <form method="get" action="import.php" style="display:inline; margin:0;">
            <button type="submit">Clean scan</button>
        </form>
        <?php if ($scanMessage): ?>
            <p><em><?= htmlspecialchars($scanMessage) ?></em></p>
        <?php endif; ?>
    </div>

    <!-- Create-profile box -->
    <div class="box">
        <h3>Create profile</h3>
        <form method="post" action="import.php" enctype="multipart/form-data">
            <label>Name: <input type="text" name="cp_name" required></label><br>
            <label>Host / subnet: <input type="text" name="cp_host" placeholder="10.0.0.1"></label><br>
            <label>SNMP community: <input type="text" name="cp_community" value="public"></label>
            <label>Version:
                <select name="cp_version"><option>2c</option><option value="1">1</option></select>
            </label>
            <label>Port: <input type="number" name="cp_port" value="161" style="width:80px;"></label><br>
            <label>Scan source:
                <select name="cp_scan_source"><option value="snmp">snmp</option><option value="ssh">ssh</option></select>
            </label>
            <label>Device type:
                <input type="text" name="cp_device_type" placeholder="opnsense / openwrt / fortinet / cisco" size="20">
            </label><br>
            <p style="margin:6px 0; color:#555;"><em>SSH credentials (optional — password and/or key):</em></p>
            <label>SSH user: <input type="text" name="cp_ssh_user"></label><br>
            <label>SSH password: <input type="password" name="cp_ssh_password"></label><br>
            <label>SSH private key file: <input type="file" name="cp_ssh_key_file"></label><br>
            <label>Passphrase (encrypts the password/key): <input type="password" name="cp_passphrase"></label><br>
            <button type="submit" name="do_create_profile" value="1" style="margin-top:8px;">Create profile</button>
        </form>
        <p style="color:#777; font-size:0.85em;">The SSH password and uploaded private key are stored encrypted (AES-256-GCM); the passphrase is required again to use them in an SSH scan.</p>
    </div>
</div>

<?php
// Multi-device list (subnet scan): pick one to configure & import.
if ($plan === null && !empty($discovered) && count($discovered) > 1): ?>
    <h4>Discovered devices — choose one to configure & import</h4>
    <table border="1" cellpadding="4" cellspacing="0">
        <tr><th>Name</th><th>IP</th><th>Brand</th><th>Model</th><th>Class</th><th></th></tr>
        <?php foreach ($discovered as $d): ?>
            <tr>
                <td><?= htmlspecialchars($d['suggested_name'] ?? '') ?></td>
                <td><?= htmlspecialchars($d['device']['ip'] ?? '') ?></td>
                <td><?= htmlspecialchars($d['brand'] ?? '') ?></td>
                <td><?= htmlspecialchars($d['model'] ?? '') ?></td>
                <td><?= htmlspecialchars($d['device_class'] ?? '') ?></td>
                <td>
                    <form method="post" action="import.php" style="margin:0;">
                        <input type="hidden" name="device_json" value="<?= htmlspecialchars(json_encode($d)) ?>">
                        <button type="submit" name="do_analyze" value="1">Configure &amp; import &rarr;</button>
                    </form>
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
            <?= htmlspecialchars($dev['device_class'] ?? '') ?>)
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
