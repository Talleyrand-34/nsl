<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$modelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];

// Variables to hold selected model port data
$selectedModelPortId = '';
$selectedPortName = '';
$selectedPositionX = '';
$selectedPositionY = '';
$selectedModelId = '';
$selectedAllowMultiple = false;

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Handle model port selection (Load Model Port button)
    if (isset($_POST['select_modelport']) && !empty($_POST['modelport_id'])) {
        $selectedModelPortId = trim($_POST['modelport_id']);

        // Find the selected model port and pre-fill its data
        foreach ($modelPorts as $modelPort) {
            if ($modelPort['id'] == $selectedModelPortId) {
                $selectedPortName = $modelPort['name'] ?? '';
                $selectedPositionX = $modelPort['positionx'] ?? '';
                $selectedPositionY = $modelPort['positiony'] ?? '';
                $selectedAllowMultiple = $modelPort['allow_multiple_connections'] ?? false;

                // Find model ID by model name
                foreach ($models as $model) {
                    if ($model['model'] === ($modelPort['model'] ?? '')) {
                        $selectedModelId = $model['id'];
                        break;
                    }
                }
                break;
            }
        }
    }
    // Handle actual update submission
    elseif (isset($_POST['update_modelport'])) {
        $modelPortId = trim($_POST['modelport_id'] ?? '');
        $newPortName = trim($_POST['port_name'] ?? '');
        $positionX = trim($_POST['position_x'] ?? '');
        $positionY = trim($_POST['position_y'] ?? '');
        $modelId = $_POST['model_id'] ?? '';
        $allowMultiple = isset($_POST['allow_multiple_connections']);

        if ($modelPortId === '') {
            $message = 'Please select a model port to update.';
        } elseif ($newPortName === '') {
            $message = 'Please enter a port name.';
        } elseif ($positionX === '') {
            $message = 'Please enter position X.';
        } elseif ($positionY === '') {
            $message = 'Please enter position Y.';
        } elseif ($modelId === '') {
            $message = 'Please select a model.';
        } else {
            $data = json_encode([
                'model_port_id' => $modelPortId,
                'new_port_name' => $newPortName,
                'new_position_x' => $positionX,
                'new_position_y' => $positionY,
                'new_model_id' => $modelId,
                'new_allow_multiple_connections' => $allowMultiple
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
                // Clear selection after successful update
                $selectedModelPortId = '';
                $selectedPortName = '';
                $selectedPositionX = '';
                $selectedPositionY = '';
                $selectedModelId = '';
                $selectedAllowMultiple = false;
            } else {
                $message = 'Failed to update model port. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
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
                <option value="<?= htmlspecialchars($modelPort['id']) ?>"
                    <?= ($modelPort['id'] == $selectedModelPortId) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($modelPort['name']) ?> (ID: <?= htmlspecialchars($modelPort['id']) ?>, Model: <?= htmlspecialchars($modelPort['model'] ?? 'N/A') ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_modelport">Load Model Port</button>
        <br><br>

        <?php if (!empty($selectedModelPortId)): ?>
            <label for="port_name">Port Name:</label>
            <input type="text" id="port_name" name="port_name"
                   value="<?= htmlspecialchars($selectedPortName) ?>" required><br><br>

            <label for="position_x">Position X:</label>
            <input type="number" id="position_x" name="position_x"
                   value="<?= htmlspecialchars($selectedPositionX) ?>" required><br><br>

            <label for="position_y">Position Y:</label>
            <input type="number" id="position_y" name="position_y"
                   value="<?= htmlspecialchars($selectedPositionY) ?>" required><br><br>

            <label for="model_id">Model:</label>
            <select id="model_id" name="model_id" required>
                <option value="">-- Select Model --</option>
                <?php foreach ($models as $model): ?>
                    <option value="<?= htmlspecialchars($model['id']) ?>"
                        <?= ($model['id'] == $selectedModelId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($model['model']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="allow_multiple_connections">
                <input type="checkbox" id="allow_multiple_connections" name="allow_multiple_connections"
                       <?= $selectedAllowMultiple ? 'checked' : '' ?>>
                Allow multiple connections to this port
            </label><br><br>

            <button type="submit" name="update_modelport">Update Model Port</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
