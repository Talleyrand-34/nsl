<?php
require_once __DIR__ . '/../config.php';

$proprietariesJson = @file_get_contents(PROPRIETARIES_ENDPOINT);
$proprietaries = json_decode($proprietariesJson, true);

if (is_array($proprietaries)) {
    echo "<ul>";
    foreach ($proprietaries as $proprietary) {
        echo "<li>" . htmlspecialchars($proprietary['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch proprietaries.</em></p>";
}
?>
