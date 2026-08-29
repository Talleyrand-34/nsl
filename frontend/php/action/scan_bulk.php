<?php
// SPDX-License-Identifier: MIT
// scan_bulk.php — bulk scan + bulk import handlers.
//
// Bulk scan:  POST /scan/network with a CIDR + scan params (same shape
//             as /scan/run but optimised for batch). Returns scan_id.
// Bulk import: POST /scan/import with a pre-discovered device list
//             (e.g. pasted from a CSV/JSON export or a previous lab scan).
//
// State carried by reference:
//   $scanRunId   — current batch run id (for the live panel)
//   $scanMessage — human message
//   $importMessage — result of the bulk import
//
// Phase 2 scope: ship the handlers + a basic panel. The dispatcher's
// two-column layout (left = scan options, right = results + upload) is
// laid out in phase 3; for now this file provides the building blocks.

require_once __DIR__ . '/import_common.php';

// --- handlers ---------------------------------------------------------------

function do_bulk_scan(&$scanMessage, &$scanRunId) {
    $subnet    = trim($_POST['bulk_subnet'] ?? '');
    $community = trim($_POST['bulk_community'] ?? 'public');
    $version   = trim($_POST['bulk_snmp_version'] ?? '2c');
    $port      = intval($_POST['bulk_snmp_port'] ?? 161);
    $timeout   = intval($_POST['bulk_timeout'] ?? 10);
    $profile   = trim($_POST['bulk_profile'] ?? '');

    if ($subnet === '') {
        $scanMessage = 'Enter a subnet (CIDR) to scan.';
        return;
    }

    $payload = [
        'subnet'      => $subnet,
        'community'   => $community,
        'snmp_version' => $version,
        'snmp_port'    => $port,
        'timeout_sec' => $timeout,
        'profile'     => $profile,
    ];
    list($code, $body, $err) = api_post_json(SCAN_NETWORK_ENDPOINT, json_encode($payload), 30);
    $j = json_decode($body, true);
    if (($code === 202 || $code === 200) && !empty($j['scan_id'])) {
        $scanRunId = $j['scan_id'];
    } else {
        $scanMessage = 'Could not start bulk scan (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
    }
}

function do_bulk_import(&$importMessage) {
    // Accept either: (a) raw textarea JSON ({"devices":[...]}), or
    // (b) an uploaded file with the same shape.
    $raw = '';
    if (isset($_FILES['bulk_file']) && $_FILES['bulk_file']['error'] === UPLOAD_ERR_OK) {
        $raw = (string) file_get_contents($_FILES['bulk_file']['tmp_name']);
    } else {
        $raw = trim($_POST['bulk_paste'] ?? '');
    }
    if ($raw === '') {
        $importMessage = 'Paste a JSON device list or upload a .json file.';
        return;
    }
    if (json_decode($raw) === null && json_last_error() !== JSON_ERROR_NONE) {
        $importMessage = 'Not valid JSON: ' . json_last_error_msg();
        return;
    }
    // The API expects {"devices":[...],"options":{...}}. Wrap a bare array
    // automatically so a pasted "[dev1, dev2, dev3]" also works.
    $decoded = json_decode($raw, true);
    if (isset($decoded[0]) && is_array($decoded[0])) {
        $body = json_encode(['devices' => $decoded, 'options' => ['default_zone' => 'Discovered']]);
    } else {
        $body = $raw;
    }
    list($code, $resp, $err) = api_post_json(SCAN_IMPORT_ENDPOINT, $body);
    $importMessage = ($code === 200)
        ? (json_decode($resp, true)['message'] ?? 'Bulk import completed.')
        : 'Bulk import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($resp ?: $err);
}

// --- panel ------------------------------------------------------------------

function scan_bulk_panel_html($scanMessage) {
    ob_start();
    ?>
    <div class="box">
        <h3>Bulk scan &amp; bulk import</h3>
        <p style="color:#666; font-size:0.9em;">Run a batch scan over a subnet, or paste a pre-collected device list to import all of it at once.</p>

        <details open>
            <summary><strong>Bulk scan</strong> &mdash; one POST covers every host in a subnet</summary>
            <form method="post" action="import.php">
                <label>Subnet (CIDR): <input type="text" name="bulk_subnet" placeholder="10.0.50.0/24" size="20"></label><br>
                <label>SNMP community: <input type="text" name="bulk_community" value="public" size="20"></label>
                <label>Version:
                    <select name="bulk_snmp_version">
                        <option value="2c">2c</option>
                        <option value="1">1</option>
                    </select>
                </label>
                <label>Port: <input type="number" name="bulk_snmp_port" value="161" style="width:80px;"></label>
                <label>Timeout (s): <input type="number" name="bulk_timeout" value="10" style="width:80px;"></label><br>
                <label>Profile (optional): <input type="text" name="bulk_profile" placeholder="leave blank to auto-match"></label><br>
                <button type="submit" name="do_bulk_scan" value="1">Bulk scan</button>
            </form>
        </details>

        <details>
            <summary><strong>Bulk import</strong> &mdash; paste or upload a JSON device list</summary>
            <form method="post" action="import.php" enctype="multipart/form-data">
                <p style="margin:4px 0;">Paste JSON (array of devices OR {"devices": [...]}):</p>
                <textarea name="bulk_paste" rows="6" style="width:90%; font-family:monospace; font-size:0.85em;"
                          placeholder='[{"device":{"ip":"10.0.0.1","sysname":"R1"}, "brand":"Cisco", "model":"ISR4431"}]'></textarea>
                <p style="margin:6px 0 2px;">Or upload a JSON file:</p>
                <input type="file" name="bulk_file" accept=".json,application/json">
                <p style="margin-top:6px;">
                    <button type="submit" name="do_bulk_import" value="1">Bulk import</button>
                </p>
            </form>
        </details>

        <?php if ($scanMessage): ?>
            <p><em><?= htmlspecialchars($scanMessage) ?></em></p>
        <?php endif; ?>
    </div>
    <?php
    return ob_get_clean();
}