<?php
// SPDX-License-Identifier: MIT
// scan_devices.php — action handlers + panels for the device-scan flow.
//
// Handlers:
//   do_scan       — async POST to /scan/run, returns scan_id, kicks off
//                    nslWatchScan polling.
//   do_analyze    — single-device analyze (POST /scan/analyze).
//   do_execute    — commit the (possibly edited) import plan (POST /scan/execute).
//
// Panels:
//   scan_devices_live_panel_html       — the scan form (method + target +
//                                          snmp/ssh fields + auto-import).
//   scan_devices_discovered_panel_html — discovered-devices table with
//                                          per-row "Configure & import" button.
//   scan_devices_plan_panel_html       — editable VLAN plan + "Import device"
//                                          submit.

// --- helpers ----------------------------------------------------------------

// dev_ip returns the discovered device's management IP (its identity in the
// list). The API returns the discovered device as {"device":{"ip":...},...}.
function dev_ip($d) {
    return $d['device']['ip'] ?? '';
}

// analyze_device POSTs a discovered device to /scan/analyze; returns the plan
// array or null. Writes the human error message into $err when it fails so
// the caller can surface it next to the plan UI.
function analyze_device($device, &$err) {
    list($code, $body) = api_post_json(SCAN_ANALYZE_ENDPOINT, json_encode(['device' => $device]));
    if ($code === 200) {
        return json_decode($body, true);
    }
    $err = 'Analyze failed (HTTP ' . intval($code) . '): ' . $body;
    return null;
}

// import_one analyzes + executes the default import plan for a discovered
// device. Used by do_scan when --auto-import is set so the operator can
// walk away while every discovered device gets imported without review.
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
    return [false, $detail];
}

// State carried across handlers: $scanRunId, $scanMessage, $discovered,
// $importedIPs, $plan, $importMessage, $pf, $autoReload. The dispatcher
// declares these once; the handlers mutate them in place.

require_once __DIR__ . '/import_common.php';

// --- handlers ---------------------------------------------------------------

function do_scan(&$scanMessage, &$scanRunId, &$autoReload) {
    $method     = ($_POST['scan_method'] ?? 'snmp') === 'ssh' ? 'ssh' : 'snmp';
    $target     = trim($_POST['target'] ?? '');
    $autoImport = !empty($_POST['auto_import']);
    if ($target === '') {
        $scanMessage = 'Please enter a target IP or CIDR.';
        return;
    }
    if ($method === 'ssh' && ($_POST['ssh_profile'] ?? '') === '') {
        $scanMessage = 'Select a device profile for the SSH scan.';
        return;
    }
    $payload = [
        'method'       => $method,
        'target'       => $target,
        'community'    => trim($_POST['community'] ?? 'public'),
        'snmp_version' => trim($_POST['snmp_version'] ?? '2c'),
        'snmp_port'    => intval($_POST['snmp_port'] ?? 161),
        'profile'      => $_POST['ssh_profile'] ?? '',
        'os_type'      => trim($_POST['ssh_os_type'] ?? ''),
    ];
    list($code, $body, $err) = api_post_json(SCAN_RUN_ENDPOINT, json_encode($payload));
    $j = json_decode($body, true);
    if (($code === 202 || $code === 200) && !empty($j['scan_id'])) {
        $scanRunId  = $j['scan_id'];
        $autoReload = $autoImport ? 'auto_import=1' : '';
    } else {
        $scanMessage = 'Could not start scan (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
    }
}

function do_scan_completed(&$discovered, &$scanMessage, &$plan, &$importedIPs, &$importMessage) {
    // Returns true if the scan is still running (dispatcher should re-arm
    // the live watcher); false if completed, failed, expired, or missing
    // (dispatcher should NOT re-arm the watcher, otherwise the page
    // re-emits the live panel on every reload and nslWatchScan fires
    // its own reload on completion → infinite loop).
    if (!isset($_GET['scan_id'])) {
        return false;
    }
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
        return false;
    }
    if ($code === 200 && $state === 'failed') {
        $scanMessage = 'Scan failed: ' . htmlspecialchars($st['error'] ?? 'unknown error');
        return false;
    }
    if ($code === 200 && $state === 'running') {
        return true;
    }
    $scanMessage = 'Scan not found (it may have expired). Run it again.';
    return false;
}

function do_analyze(&$plan, &$scanMessage) {
    $device = json_decode($_POST['device_json'] ?? 'null', true);
    if ($device) {
        $plan = analyze_device($device, $scanMessage);
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
        $ip = $editedPlan['device']['device']['ip'] ?? '';
        if ($ip !== '') {
            $importedIPs[] = $ip;
            $_SESSION['scan_imported'] = array_values(array_unique($importedIPs));
        }
    } else {
        $detail = json_decode($body, true)['message'] ?? ($body ?: $err);
        $importMessage = 'Import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($detail);
    }
}

// --- panels -----------------------------------------------------------------

function scan_devices_live_panel_html($pf, $profiles, $scanMessage) {
    ob_start();
    ?>
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
                    <input type="text" name="ssh_os_type" list="ssh_devtypes" value="<?= htmlspecialchars($_POST['ssh_os_type'] ?? '') ?>" placeholder="openwrt / opnsense / fortinet">
                    <datalist id="ssh_devtypes"><option value="openwrt"><option value="opnsense"><option value="fortinet"></datalist>
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
    <?php
    return ob_get_clean();
}

function scan_devices_discovered_panel_html($discovered, $importedIPs) {
    $pending = 0;
    if (!empty($discovered)) {
        foreach ($discovered as $d) {
            if (!in_array(dev_ip($d), $importedIPs, true)) $pending++;
        }
    }
    // The discovered-devices <tbody> is replaced client-side by
    // pending-devices.js from localStorage. Operators without JS still see
    // the SSR'd PHP render (this page load only).
    $initialJson = json_encode(array_values($discovered ?: []));
    $importedJson = json_encode($importedIPs);
    ob_start();
    ?>

    <h4>Discovered devices (<span class="discovered-total"><?= count($discovered) ?></span> total, <span class="discovered-pending"><?= $pending ?></span> pending)</h4>
    <table border="1" cellpadding="4" cellspacing="0" class="discovered-devices">
        <thead><tr><th>Status</th><th>Name</th><th>IP</th><th>Brand</th><th>Model</th><th>Class</th><th></th></tr></thead>
        <tbody>
            <?php if (empty($discovered)): ?>
                <tr><td colspan="7" style="color:#777; text-align:center;">No pending devices. Run a scan from the left column to populate the queue.</td></tr>
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
    // Hand the initial server-side discovery to the localStorage module so
 // the operator doesn't lose devices when the page reloads between
 // scans. nslPD.addAll dedupes by IP.
    window.NSL_PENDING_INITIAL = <?= $initialJson ?>;
    window.NSL_PENDING_IMPORTED_IPS = <?= $importedJson ?>;
    (function () {
        var init = function () {
            if (window.nslPD && Array.isArray(window.NSL_PENDING_INITIAL)) {
                window.nslPD.addAll(window.NSL_PENDING_INITIAL);
                var body = document.querySelector('table.discovered-devices tbody');
                if (body) window.nslPD.render(body, window.NSL_PENDING_IMPORTED_IPS || []);
                var total = 0, pending = 0;
                document.querySelectorAll('table.discovered-devices tbody tr').forEach(function (tr) {
                    var cells = tr.children;
                    if (cells.length === 1) return; // placeholder row
                    total++;
                    var status = (cells[0].textContent || '').trim();
                    if (status === 'pending') pending++;
                });
                var tEl = document.querySelector('.discovered-total');
                var pEl = document.querySelector('.discovered-pending');
                if (tEl) tEl.textContent = total;
                if (pEl) pEl.textContent = pending;
            }
        };
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', init);
        } else {
            init();
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