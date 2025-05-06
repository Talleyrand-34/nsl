<?php
require_once __DIR__ . '/../config.php';

$devclassesJson = @file_get_contents(DEVCLASSES_ENDPOINT);
$devclasses = json_decode($devclassesJson, true);

if (is_array($devclasses)) {
    echo "<ul>";
    foreach ($devclasses as $devclass) {
        echo "<li>" . htmlspecialchars($devclass['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch devclasses.</em></p>";
}
?>
