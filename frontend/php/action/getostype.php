<?php
require_once __DIR__ . '/../config.php';

$ostypesJson = @file_get_contents(OSTYPES_ENDPOINT);
$ostypes = json_decode($ostypesJson, true);

if (is_array($ostypes)) {
    echo "<ul>";
    foreach ($ostypes as $ostype) {
        echo "<li>" . htmlspecialchars($ostype['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch ostypes.</em></p>";
}
?>
