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
<?php else: ?>
    <p><em>Could not fetch device ports or no device ports found.</em></p>
<?php endif; ?>
