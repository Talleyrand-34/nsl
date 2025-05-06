<?php
require_once __DIR__ . '/../config.php';

$zonetypesJson = @file_get_contents(ZONETYPES_ENDPOINT);
$zonetypes = json_decode($zonetypesJson, true);

if (is_array($zonetypes)) {
    echo "<ul>";
    foreach ($zonetypes as $zonetype) {
        echo "<li>" . htmlspecialchars($zonetype['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch zonetypes.</em></p>";
}
?>
