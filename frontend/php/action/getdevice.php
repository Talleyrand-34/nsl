
<?php
require_once __DIR__ . '/../config.php';

// Fetch devices
$devicesJson = @file_get_contents(DEVICES_ENDPOINT);
$devices = json_decode($devicesJson, true);

if (is_array($devices) && count($devices) > 0):
    ?>
    <ul>
    <?php foreach ($devices as $device): ?>
        <li>
            <strong><?= htmlspecialchars($device['name']) ?></strong>
            <ul>
                <li>ID: <?= htmlspecialchars($device['id']) ?></li>
                <li>Model: <?= htmlspecialchars($device['model']) ?></li>
                <li>Brand: <?= htmlspecialchars($device['brand']) ?></li>
                <li>Zone ID: <?= htmlspecialchars($device['zoneid']) ?></li>
                <li>Zone Name: <?= htmlspecialchars($device['zonename']) ?></li>
                <li>Zone Father: <?= htmlspecialchars($device['zonefathername']) ?></li>
                <li>Proprietary: <?= htmlspecialchars($device['proprietary']) ?></li>
            </ul>
        </li>
    <?php endforeach; ?>
    </ul>
<?php else: ?>
    <p><em>Could not fetch devices or no devices found.</em></p>
<?php endif; ?>
