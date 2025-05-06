
<?php
require_once __DIR__ . '/../config.php';

// Fetch models
$modelsJson = @file_get_contents(MODELS_ENDPOINT);
$models = json_decode($modelsJson, true);

if (is_array($models)) {
    echo '<ul>';
    foreach ($models as $model) {
        echo '<li>';
        echo '<strong>' . htmlspecialchars($model['model']) . '</strong>';
        echo '<ul>';
        echo '<li>Brand: ' . htmlspecialchars($model['brand']) . '</li>';
        echo '<li>Class: ' . htmlspecialchars($model['class']) . '</li>';
        echo '</ul>';
        echo '</li>';
    }
    echo '</ul>';
} else {
    echo '<p><em>Could not fetch models.</em></p>';
}
?>
