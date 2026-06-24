<?php
require_once __DIR__ . '/../config.php';
$message = '';

$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];

$selectedDeviceId = $_POST['device_id'] ?? '';
$selectedModelId = $_POST['model_id'] ?? '';
$showTable = false;
$devicePorts = [];
$targetPorts = [];

if ($_SERVER['REQUEST_METHOD'] === 'POST' && (isset($_POST['load_mapping']) || isset($_POST['do_migrate']))) {
    if ($selectedDeviceId === '' || $selectedModelId === '') {
        $message = 'Please select both a device and a target model.';
    } else {
        // The device's existing ports are the source of truth for the mapping.
        $allDP = json_decode(@file_get_contents(DEVICEPORTS_ENDPOINT), true) ?: [];
        $devicePorts = array_values(array_filter($allDP, fn($p) => ($p['devid'] ?? '') === $selectedDeviceId));

        // Target model's ports (matched by model name, which is what /modelports exposes).
        $modelName = '';
        foreach ($models as $m) {
            if ($m['id'] === $selectedModelId) { $modelName = $m['model']; break; }
        }
        $allMP = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];
        $targetPorts = array_values(array_filter($allMP, fn($mp) => ($mp['model'] ?? '') === $modelName));
        $showTable = true;

        if (isset($_POST['do_migrate'])) {
            $portMap = [];
            foreach (($_POST['map'] ?? []) as $dpId => $mpId) {
                if ($mpId !== '') { $portMap[$dpId] = $mpId; }
            }
            if (empty($targetPorts)) {
                $message = 'The target model has no ports. Add ports to it first (Add → ModelPort).';
            } elseif (count($portMap) !== count($devicePorts)) {
                $message = 'Map every device port to a target model port before migrating.';
            } else {
                $data = json_encode([
                    'device_id' => $selectedDeviceId,
                    'model_id'  => $selectedModelId,
                    'port_map'  => (object) $portMap,
                ]);
                $ch = curl_init(DEVICES_MIGRATE_ENDPOINT);
                curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
                curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
                curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
                curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json', 'Content-Length: ' . strlen($data)]);
                $response = curl_exec($ch);
                $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
                curl_close($ch);
                if ($httpCode >= 200 && $httpCode < 300) {
                    $message = 'Device migrated successfully! Port data (MAC, VLANs, connections) was preserved.';
                    $showTable = false;
                    $devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
                } else {
                    $message = 'Failed to migrate device. Server response: ' . htmlspecialchars((string) $response);
                }
            }
        }
    }
}
?>

<h2>Migrate a Device to Another Model</h2>
<p style="max-width:70ch; color:#555;">
  Re-point a device to a different model and decide how each existing port maps to
  the new model's ports. The device is the source of truth — MAC addresses, VLAN
  configs and connections are preserved; only the model schema changes.
</p>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($devices)): ?>
    <p><em>No devices available to migrate.</em></p>
<?php else: ?>
<form method="post">
    <label for="device_id">Device:</label>
    <select id="device_id" name="device_id" required>
        <option value="">-- Select Device --</option>
        <?php foreach ($devices as $d): ?>
            <option value="<?= htmlspecialchars($d['id']) ?>" <?= ($d['id'] === $selectedDeviceId) ? 'selected' : '' ?>>
                <?= htmlspecialchars($d['label'] ?? '') ?> (current model: <?= htmlspecialchars($d['model'] ?? '') ?>)
            </option>
        <?php endforeach; ?>
    </select>

    <label for="model_id">Target model:</label>
    <select id="model_id" name="model_id" required>
        <option value="">-- Select Model --</option>
        <?php foreach ($models as $m): ?>
            <option value="<?= htmlspecialchars($m['id']) ?>" <?= ($m['id'] === $selectedModelId) ? 'selected' : '' ?>>
                <?= htmlspecialchars($m['model']) ?> (<?= htmlspecialchars($m['brand'] ?? '') ?>)
            </option>
        <?php endforeach; ?>
    </select>

    <button type="submit" name="load_mapping" value="1">Load port mapping</button>

    <?php if ($showTable): ?>
        <h3>Port mapping</h3>
        <?php if (empty($devicePorts)): ?>
            <p><em>This device has no ports to map.</em></p>
        <?php elseif (empty($targetPorts)): ?>
            <p><em>The target model has no ports. Add ports to it first (Add → ModelPort).</em></p>
        <?php else: ?>
            <table border="1" cellpadding="4">
                <tr><th>Existing device port</th><th>MAC</th><th>VLANs</th><th>&rarr; Target model port</th></tr>
                <?php foreach ($devicePorts as $dp): ?>
                    <tr>
                        <td><strong><?= htmlspecialchars($dp['portname'] ?? '') ?></strong></td>
                        <td><?= htmlspecialchars($dp['mac_address'] ?? '') ?></td>
                        <td><?= htmlspecialchars(count($dp['vlan_configs'] ?? [])) ?></td>
                        <td>
                            <select name="map[<?= htmlspecialchars($dp['id']) ?>]" required>
                                <option value="">-- Select target port --</option>
                                <?php foreach ($targetPorts as $tp): ?>
                                    <option value="<?= htmlspecialchars($tp['id']) ?>">
                                        <?= htmlspecialchars($tp['name']) ?>
                                    </option>
                                <?php endforeach; ?>
                            </select>
                        </td>
                    </tr>
                <?php endforeach; ?>
            </table>
            <p><button type="submit" name="do_migrate" value="1">Migrate device</button></p>
        <?php endif; ?>
    <?php endif; ?>
</form>
<?php endif; ?>
