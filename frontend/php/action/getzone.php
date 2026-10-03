<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';

// Make sure ZONES_ENDPOINT is defined in config.php
// define('ZONES_ENDPOINT', API_BASE_URL . '/zones');

$zonesJson = @file_get_contents(ZONES_ENDPOINT);
$zones = json_decode($zonesJson, true);

if (is_array($zones)) {
    echo "<ul>";
    foreach ($zones as $zone) {
        echo "<li>ID: " . htmlspecialchars($zone['id']);
        echo "<ul>";
        echo "<li>Name: " . htmlspecialchars($zone['name']) . "</li>";
        echo "<li>Father ID: " . htmlspecialchars($zone['fatherid']) . "</li>";
        echo "<li>Father: " . htmlspecialchars($zone['father']) . "</li>";
        echo "<li>Zone Type: " . htmlspecialchars($zone['location_type']) . "</li>";
        echo "<li>Owner: " . htmlspecialchars($zone['owner']) . "</li>";
        echo "</ul>";
        echo "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch zones.</em></p>";
}
?>
