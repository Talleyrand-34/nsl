<?php
require_once __DIR__ . '/../config.php';

$devicePortsJson = @file_get_contents(DEVICEPORTS_ENDPOINT);
$devicePorts = json_decode($devicePortsJson, true);
if (!is_array($devicePorts)) {
    $devicePorts = [];
}

// Build the device filter options from the device ports themselves (id => name).
$deviceOptions = [];
foreach ($devicePorts as $dp) {
    $id = $dp['devid'] ?? '';
    if ($id !== '' && !isset($deviceOptions[$id])) {
        $deviceOptions[$id] = $dp['devname'] ?? $id;
    }
}
asort($deviceOptions);

// Active filter (device id) from the query string.
$deviceFilter = $_GET['dp_device'] ?? '';
if ($deviceFilter !== '') {
    $devicePorts = array_values(array_filter($devicePorts, fn($dp) => ($dp['devid'] ?? '') === $deviceFilter));
}
?>

<form method="get" style="margin-bottom:12px;">
    <!-- Preserve the dashboard's action/entity selection. -->
    <input type="hidden" name="actionType" value="<?= htmlspecialchars($_GET['actionType'] ?? 'get') ?>">
    <input type="hidden" name="entity" value="<?= htmlspecialchars($_GET['entity'] ?? 'deviceport') ?>">
    <label for="dp_device">Filter by device:</label>
    <select id="dp_device" name="dp_device" onchange="this.form.submit()">
        <option value="">-- All devices --</option>
        <?php foreach ($deviceOptions as $id => $name): ?>
            <option value="<?= htmlspecialchars($id) ?>" <?= ($deviceFilter === (string) $id) ? 'selected' : '' ?>>
                <?= htmlspecialchars($name) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <noscript><button type="submit">Filter</button></noscript>
</form>

<?php if (count($devicePorts) > 0): ?>
    <ul>
    <?php foreach ($devicePorts as $devicePort): ?>
        <li>
            <strong>Device: <?= htmlspecialchars($devicePort['devname']) ?></strong>
            <ul>
                <li>Device ID: <?= htmlspecialchars($devicePort['devid']) ?></li>
                <li>Model Port ID: <?= htmlspecialchars($devicePort['modelid']) ?></li>
                <li>Port Name: <strong><?= htmlspecialchars($devicePort['portname']) ?></strong></li>
                <li>MAC Address: <?= htmlspecialchars($devicePort['mac_address'] ?? 'N/A') ?></li>
                <li>Position: <?= htmlspecialchars($devicePort['positionx']) ?>, <?= htmlspecialchars($devicePort['positiony']) ?></li>
                <li>
                    <strong>VLAN Configurations:</strong>
                    <?php if (!empty($devicePort['vlan_configs']) && is_array($devicePort['vlan_configs'])): ?>
                        <ul>
                            <?php foreach ($devicePort['vlan_configs'] as $vlanConfig): ?>
                                <li>
                                    VLAN <?= htmlspecialchars($vlanConfig['vlan_number']) ?>
                                    - <?= $vlanConfig['tagged'] ? '<span style="color: blue;">Tagged</span>' : '<span style="color: green;">Untagged</span>' ?>
                                </li>
                            <?php endforeach; ?>
                        </ul>
                    <?php else: ?>
                        <em>No VLANs configured</em>
                    <?php endif; ?>
                </li>
            </ul>
        </li>
    <?php endforeach; ?>
    </ul>
<?php elseif ($deviceFilter !== ''): ?>
    <p><em>No device ports for the selected device.</em></p>
<?php else: ?>
    <p><em>Could not fetch device ports or no device ports found.</em></p>
<?php endif; ?>
