<?php
require_once __DIR__ . '/../config.php';

// Fetch model ports
$modelPortsJson = @file_get_contents(MODELPORTS_ENDPOINT);
$modelPorts = json_decode($modelPortsJson, true);

// Collect unique models for the filter dropdown
$models = [];
if (is_array($modelPorts)) {
    foreach ($modelPorts as $port) {
        if (!empty($port['model']) && !in_array($port['model'], $models)) {
            $models[] = $port['model'];
        }
    }
}

// Get selected model from GET parameter
$selectedModel = isset($_GET['model']) ? $_GET['model'] : '';

// Filter model ports if a model is selected
if ($selectedModel && is_array($modelPorts)) {
    $modelPorts = array_filter($modelPorts, function ($port) use ($selectedModel) {
        // Case-insensitive comparison for robustness
        return isset($port['model']) && strtolower($port['model']) === strtolower($selectedModel);
    });
}

// Filter form with required GET parameters in the URL
?>
<form method="get" action="main.php">
    <input type="hidden" name="actionType" value="get">
    <input type="hidden" name="entity" value="modelport">
    <label for="model">Filter by Model:</label>
    <select name="model" id="model">
        <option value="">-- All Models --</option>
        <?php foreach ($models as $model): ?>
            <option value="<?= htmlspecialchars($model) ?>" <?= ($model === $selectedModel) ? 'selected' : '' ?>>
                <?= htmlspecialchars($model) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Filter</button>
</form>
<?php

// Output the filtered or all model ports
if (is_array($modelPorts) && count($modelPorts) > 0) {
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
    echo '<p><em>Could not fetch model ports or no ports for selected model.</em></p>';
}
?>
