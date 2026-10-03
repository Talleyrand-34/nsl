
<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';

$message = '';

// Handle the managed/unmanaged VLAN toggle.
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['toggle_device_id'])) {
    $deviceId = (string) $_POST['toggle_device_id'];
    $isUnmanaged = isset($_POST['is_unmanaged']) && $_POST['is_unmanaged'] === '1';

    $data = json_encode([
        'id' => $deviceId,
        'is_unmanaged' => $isUnmanaged,
    ]);

    $ch = curl_init(DEVICES_ENDPOINT);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'PUT');
    curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
        'Content-Length: ' . strlen($data),
    ]);
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    if ($httpCode === 200) {
        $message = 'Device VLAN mode updated to ' . ($isUnmanaged ? 'Unmanaged' : 'Managed') . '.';
    } else {
        $message = 'Failed to update device. Server response: ' . htmlspecialchars((string) $response);
    }
}

// Fetch devices
$devicesJson = @file_get_contents(DEVICES_ENDPOINT);
$devices = json_decode($devicesJson, true);
?>

<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($devices) && count($devices) > 0): ?>
    <ul>
    <?php foreach ($devices as $device): ?>
        <?php $isUnmanaged = !empty($device['is_unmanaged']); ?>
        <li>
            <strong><?= htmlspecialchars($device['name']) ?></strong>
            <ul>
                <li>ID: <?= htmlspecialchars($device['id']) ?></li>
                <li>Model: <?= htmlspecialchars($device['model']) ?></li>
                <li>Name: <strong><?= htmlspecialchars($device['label']) ?></strong></li>
                <li>Brand: <?= htmlspecialchars($device['brand']) ?></li>
                <li>Zone ID: <?= htmlspecialchars($device['zoneid']) ?></li>
                <li>Zone Name: <?= htmlspecialchars($device['zonename']) ?></li>
                <li>Zone Father: <?= htmlspecialchars($device['zonefathername']) ?></li>
                <li>Owner: <?= htmlspecialchars($device['owner']) ?></li>
                <li>
                    VLAN mode: <strong><?= $isUnmanaged ? 'Unmanaged' : 'Managed' ?></strong>
                    <form method="post" style="display:inline; margin-left:8px;">
                        <input type="hidden" name="toggle_device_id" value="<?= htmlspecialchars($device['id']) ?>">
                        <input type="hidden" name="is_unmanaged" value="<?= $isUnmanaged ? '0' : '1' ?>">
                        <button type="submit">Switch to <?= $isUnmanaged ? 'Managed' : 'Unmanaged' ?> VLANs</button>
                    </form>
                </li>
            </ul>
        </li>
    <?php endforeach; ?>
    </ul>
<?php else: ?>
    <p><em>Could not fetch devices or no devices found.</em></p>
<?php endif; ?>
