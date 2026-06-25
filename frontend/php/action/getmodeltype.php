<?php
require_once __DIR__ . '/../config.php';

$modeltypesJson = @file_get_contents(MODELTYPES_ENDPOINT);
$modeltypes = json_decode($modeltypesJson, true);

if (is_array($modeltypes)) {
    echo "<ul>";
    foreach ($modeltypes as $modeltype) {
        echo "<li>" . htmlspecialchars($modeltype['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch modeltypes.</em></p>";
}
?>
