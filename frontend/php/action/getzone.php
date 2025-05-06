<?php
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
        echo "<li>Location Type: " . htmlspecialchars($zone['location_type']) . "</li>";
        echo "<li>Proprietary: " . htmlspecialchars($zone['proprietary']) . "</li>";
        echo "</ul>";
        echo "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch zones.</em></p>";
}
?>
