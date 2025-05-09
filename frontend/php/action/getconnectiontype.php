<?php
require_once __DIR__ . '/../config.php';

$contypesJson = @file_get_contents(CONNECTIONTYPES_ENDPOINT);
$contypes = json_decode($contypesJson, true);

if (is_array($contypes)) {
    echo '<ul>';
    foreach ($contypes as $contype) {
        echo '<li>' . htmlspecialchars($contype['name']) . '</li>';
    }
    echo '</ul>';
} else {
    echo '<p><em>Could not fetch contypes.</em></p>';
}
?>
