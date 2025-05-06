
<?php
require_once __DIR__ . '/../config.php';

// Fetch model ports
$modelPortsJson = @file_get_contents(MODELPORTS_ENDPOINT);
$modelPorts = json_decode($modelPortsJson, true);

if (is_array($modelPorts)) {
    echo '<ul>';
    foreach ($modelPorts as $port) {
        echo '<li>';
        echo '<strong>' . htmlspecialchars($port['name']) . '</strong> (ID: ' . htmlspecialchars($port['id']) . ')';
        echo '<ul>';
        echo '<li>Model: ' . htmlspecialchars($port['model']) . '</li>';
        echo '<li>Brand: ' . htmlspecialchars($port['brand']) . '</li>';
        echo '<li>Position X: ' . htmlspecialchars($port['positionx']) . '</li>';
        echo '<li>Position Y: ' . htmlspecialchars($port['positiony']) . '</li>';
        echo '</ul>';
        echo '</li>';
    }
    echo '</ul>';
} else {
    echo '<p><em>Could not fetch model ports.</em></p>';
}
?>
