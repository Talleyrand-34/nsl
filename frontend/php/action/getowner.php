<?php
require_once __DIR__ . '/../config.php';

$ownersJson = @file_get_contents(OWNERS_ENDPOINT);
$owners = json_decode($ownersJson, true);

if (is_array($owners)) {
    echo "<ul>";
    foreach ($owners as $owner) {
        echo "<li>" . htmlspecialchars($owner['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch owners.</em></p>";
}
?>
