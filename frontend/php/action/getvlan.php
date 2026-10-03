<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';

$vlansJson = @file_get_contents(VLANS_ENDPOINT);
$vlans = json_decode($vlansJson, true);

?>

<h2>VLANs</h2>

<?php if (is_array($vlans) && count($vlans) > 0): ?>
<table border="1" cellpadding="8" cellspacing="0">
    <thead>
        <tr>
            <th>Database ID</th>
            <th>VLAN ID</th>
            <th>VLAN Name</th>
        </tr>
    </thead>
    <tbody>
        <?php foreach ($vlans as $vlan): ?>
        <tr>
            <td><?= htmlspecialchars($vlan['id']) ?></td>
            <td><?= htmlspecialchars($vlan['vlanid']) ?></td>
            <td><?= htmlspecialchars($vlan['vlanname']) ?></td>
        </tr>
        <?php endforeach; ?>
    </tbody>
</table>
<?php else: ?>
    <p><em>No VLANs found or unable to fetch data.</em></p>
<?php endif; ?>
