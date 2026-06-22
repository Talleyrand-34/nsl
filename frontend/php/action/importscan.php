<?php
require_once __DIR__ . '/../config.php';

/**
 * api_post_json POSTs a JSON string to $url and returns
 * [httpCode, responseBody, curlError].
 */
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

$scanMessage   = '';
$importMessage = '';
$discovered    = [];   // list of discovered-device objects (assoc arrays)

// --- Step: run a live SNMP scan ---------------------------------------------
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_scan'])) {
    $scanType  = $_POST['scan_type'] ?? 'host';
    $target    = trim($_POST['target'] ?? '');
    $community = trim($_POST['community'] ?? 'public');
    $version   = trim($_POST['snmp_version'] ?? '2c');
    $port      = intval($_POST['snmp_port'] ?? 161);

    if ($target === '') {
        $scanMessage = 'Please enter a host IP or subnet.';
    } else {
        if ($scanType === 'network') {
            $payload = json_encode([
                'subnet'       => $target,
                'community'    => $community,
                'snmp_version' => $version,
                'snmp_port'    => $port,
            ]);
            list($code, $body, $err) = api_post_json(SCAN_NETWORK_ENDPOINT, $payload);
        } else {
            $payload = json_encode([
                'ip'           => $target,
                'community'    => $community,
                'snmp_version' => $version,
                'snmp_port'    => $port,
            ]);
            list($code, $body, $err) = api_post_json(SCAN_HOST_ENDPOINT, $payload);
        }

        if ($body === false || $body === '') {
            $scanMessage = 'Scan request failed: ' . htmlspecialchars($err);
        } elseif ($code !== 200) {
            $scanMessage = 'Scan failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body);
        } else {
            $data = json_decode($body, true);
            if ($scanType === 'network') {
                $discovered = $data['devices'] ?? [];
            } else {
                // /scan/host returns a single discovered device object.
                $discovered = ($data && isset($data['device'])) ? [$data] : [];
            }
            $scanMessage = count($discovered) . ' device(s) discovered.';
        }
    }
}

// --- Step: import selected scanned devices ----------------------------------
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['do_import'])) {
    $all      = json_decode($_POST['devices_json'] ?? '[]', true) ?: [];
    $selected = $_POST['selected'] ?? [];
    $subset   = [];
    foreach ($selected as $idx) {
        if (isset($all[intval($idx)])) {
            $subset[] = $all[intval($idx)];
        }
    }

    if (empty($subset)) {
        $importMessage = 'No devices selected for import.';
    } else {
        $payload = json_encode([
            'devices' => $subset,
            'options' => [
                'auto_import'         => true,
                'create_zones'        => true,
                'default_zone'        => 'Discovered',
                'skip_existing'       => true,
                'vlan_accuracy_level' => 2,
            ],
        ]);
        list($code, $body, $err) = api_post_json(SCAN_IMPORT_ENDPOINT, $payload);
        if ($code === 200) {
            $resp = json_decode($body, true);
            $importMessage = $resp['message'] ?? 'Import completed.';
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
            if ($code === 200) {
                $resp = json_decode($body, true);
                $importMessage = $resp['message'] ?? 'Import completed.';
            } else {
                $importMessage = 'Import failed (HTTP ' . intval($code) . '): ' . htmlspecialchars($body ?: $err);
            }
        }
    }
}
?>

<h2>Import devices</h2>

<?php if ($importMessage): ?>
    <p style="padding:8px; background:#e8f5e9; border:1px solid #a5d6a7;"><strong><?= htmlspecialchars($importMessage) ?></strong></p>
<?php endif; ?>

<h3>1. Live SNMP scan</h3>
<form method="post" action="import.php">
    <label>Scan type:
        <select name="scan_type">
            <option value="host" <?= (($_POST['scan_type'] ?? 'host') === 'host') ? 'selected' : '' ?>>Single host</option>
            <option value="network" <?= (($_POST['scan_type'] ?? '') === 'network') ? 'selected' : '' ?>>Subnet</option>
        </select>
    </label><br>
    <label>Host IP / Subnet (CIDR):
        <input type="text" name="target" value="<?= htmlspecialchars($_POST['target'] ?? '') ?>" placeholder="192.168.1.1 or 192.168.1.0/24" required>
    </label><br>
    <label>SNMP community:
        <input type="text" name="community" value="<?= htmlspecialchars($_POST['community'] ?? 'public') ?>">
    </label><br>
    <label>SNMP version:
        <select name="snmp_version">
            <option value="2c" <?= (($_POST['snmp_version'] ?? '2c') === '2c') ? 'selected' : '' ?>>2c</option>
            <option value="1" <?= (($_POST['snmp_version'] ?? '') === '1') ? 'selected' : '' ?>>1</option>
        </select>
    </label>
    <label>Port:
        <input type="number" name="snmp_port" value="<?= htmlspecialchars($_POST['snmp_port'] ?? '161') ?>" style="width:80px;">
    </label><br>
    <button type="submit" name="do_scan" value="1">Scan</button>
</form>

<?php if ($scanMessage): ?>
    <p><em><?= htmlspecialchars($scanMessage) ?></em></p>
<?php endif; ?>

<?php if (!empty($discovered)): ?>
    <form method="post" action="import.php">
        <input type="hidden" name="devices_json" value="<?= htmlspecialchars(json_encode($discovered)) ?>">
        <table border="1" cellpadding="4" cellspacing="0" style="margin-top:8px;">
            <tr>
                <th></th><th>Name</th><th>IP</th><th>Brand</th><th>Model</th><th>Class</th><th>Zone</th>
            </tr>
            <?php foreach ($discovered as $i => $d): ?>
                <tr>
                    <td><input type="checkbox" name="selected[]" value="<?= $i ?>" checked></td>
                    <td><?= htmlspecialchars($d['suggested_name'] ?? '') ?></td>
                    <td><?= htmlspecialchars($d['device']['ip'] ?? '') ?></td>
                    <td><?= htmlspecialchars($d['brand'] ?? '') ?></td>
                    <td><?= htmlspecialchars($d['model'] ?? '') ?></td>
                    <td><?= htmlspecialchars($d['device_class'] ?? '') ?></td>
                    <td><?= htmlspecialchars($d['suggested_zone'] ?? '') ?></td>
                </tr>
            <?php endforeach; ?>
        </table>
        <button type="submit" name="do_import" value="1" style="margin-top:8px;">Import selected</button>
    </form>
<?php endif; ?>

<hr>

<h3>2. Upload scan-result JSON</h3>
<p style="color:#666; font-size:0.9em;">Upload a JSON file produced by <code>nsl-graph scan</code> (same format as <code>scan import &lt;file&gt;</code>).</p>
<form method="post" action="import.php" enctype="multipart/form-data">
    <input type="file" name="scanfile" accept=".json,application/json" required>
    <button type="submit" name="do_upload" value="1">Upload &amp; import</button>
</form>
