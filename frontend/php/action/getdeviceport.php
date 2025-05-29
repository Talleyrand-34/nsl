<?php
require_once __DIR__ . '/../config.php';

$devicePortsJson = @file_get_contents(DEVICEPORTS_ENDPOINT);
$devicePorts = json_decode($devicePortsJson, true);

if (is_array($devicePorts) && count($devicePorts) > 0):
    ?>
    <ul>
    <?php foreach ($devicePorts as $devicePort): ?>
        <li>
            <strong>Device: <?= htmlspecialchars($devicePort['devname']) ?></strong>
            <ul>
                <li>Device ID: <?= htmlspecialchars($devicePort['devid']) ?></li>
                <li>Model Port ID: <?= htmlspecialchars($devicePort['modelid']) ?></li>
                <li>Port Name: <?= htmlspecialchars($devicePort['portname']) ?></li>
                <li>MAC Address: <?= htmlspecialchars($devicePort['mac_address'] ?? '') ?></li>
                <li>Position: <?= htmlspecialchars($devicePort['positionx']) ?>, <?= htmlspecialchars($devicePort['positiony']) ?></li>
            </ul>
        </li>
    <?php endforeach; ?>
    </ul>
<?php else: ?>
    <p><em>Could not fetch device ports or no device ports found.</em></p>
<?php endif; ?>