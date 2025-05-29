<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$modelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $modelPortId = trim($_POST['modelport_id'] ?? '');
    $newPortName = trim($_POST['port_name'] ?? '');
    $positionX = trim($_POST['position_x'] ?? '');
    $positionY = trim($_POST['position_y'] ?? '');
    $modelId = $_POST['model_id'] ?? '';

    if ($modelPortId === '') {
        $message = 'Please select a model port to update.';
    } elseif ($newPortName === '') {
        $message = 'Please enter a new port name.';
    } elseif ($positionX === '') {
        $message = 'Please enter position X.';
    } elseif ($positionY === '') {
        $message = 'Please enter position Y.';
    } elseif ($modelId === '') {
        $message = 'Please select a model.';
    } else {
        $data = json_encode([
            'id' => $modelPortId,
            'name' => $newPortName,
            'position_x' => $positionX,
            'position_y' => $positionY,
            'model_id' => $modelId
        ]);

        $ch = curl_init(MODELPORTS_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'PUT');
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 200) {
            $message = 'Model port updated successfully!';
            // Refresh model ports list
            $modelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];
        } else {
            $message = 'Failed to update model port. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Update Model Port</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($modelPorts)): ?>
    <p><em>No model ports available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="modelport_id">Select Model Port to Update:</label>
        <select id="modelport_id" name="modelport_id" required>
            <option value="">-- Select Model Port --</option>
            <?php foreach ($modelPorts as $modelPort): ?>
                <option value="<?= htmlspecialchars($modelPort['id']) ?>">
                    <?= htmlspecialchars($modelPort['name']) ?> (ID: <?= htmlspecialchars($modelPort['id']) ?>, Model: <?= htmlspecialchars($modelPort['model'] ?? 'N/A') ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="port_name">New Port Name:</label>
        <input type="text" id="port_name" name="port_name" required><br><br>

        <label for="position_x">Position X:</label>
        <input type="number" id="position_x" name="position_x" required><br><br>

        <label for="position_y">Position Y:</label>
        <input type="number" id="position_y" name="position_y" required><br><br>

        <label for="model_id">Model:</label>
        <select id="model_id" name="model_id" required>
            <option value="">-- Select Model --</option>
            <?php foreach ($models as $model): ?>
                <option value="<?= htmlspecialchars($model['id']) ?>">
                    <?= htmlspecialchars($model['model']) ?>
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <button type="submit">Update Model Port</button>
    </form>
<?php endif; ?>