<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing models for dropdown
$modelsJson = @file_get_contents(MODELS_ENDPOINT);
$models = json_decode($modelsJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $id = trim($_POST['id'] ?? '');

    if ($id !== '') {
        $data = json_encode(['model_id' => $id]);

        $ch = curl_init(MODELS_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "DELETE");
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 204) {
            $message = "Model deleted successfully!";
            // Refresh models list after deletion
            $modelsJson = @file_get_contents(MODELS_ENDPOINT);
            $models = json_decode($modelsJson, true);
        } else {
            $message = "Failed to delete model. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a model.";
    }
}
?>

<h2>Delete a Model</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($models) && count($models) > 0): ?>
<form method="post">
    <label for="id">Select Model to Delete:</label>
    <select id="id" name="id" required>
        <option value="">-- Select a Model --</option>
        <?php foreach ($models as $model): ?>
            <option value="<?= htmlspecialchars($model['id']) ?>">
                ID: <?= htmlspecialchars($model['id']) ?> - <?= htmlspecialchars($model['model']) ?> (<?= htmlspecialchars($model['brand']) ?> - <?= htmlspecialchars($model['class_name']) ?>)
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Delete Model</button>
</form>
<?php else: ?>
    <p><em>No models available to delete.</em></p>
<?php endif; ?>