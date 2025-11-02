<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];
$brands = json_decode(@file_get_contents(BRANDS_ENDPOINT), true) ?: [];
$deviceClasses = json_decode(@file_get_contents(DEVCLASSES_ENDPOINT), true) ?: [];

// Variables to hold selected model data
$selectedModelId = '';
$selectedModelName = '';
$selectedBrandId = '';
$selectedDeviceClassId = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Handle model selection (Load Model button)
    if (isset($_POST['select_model']) && !empty($_POST['model_id'])) {
        $selectedModelId = trim($_POST['model_id']);

        // Find the selected model and pre-fill its data
        foreach ($models as $model) {
            if ($model['id'] == $selectedModelId) {
                $selectedModelName = $model['model'] ?? '';

                // Find brand ID by brand name
                foreach ($brands as $brand) {
                    if ($brand['name'] === ($model['brand'] ?? '')) {
                        $selectedBrandId = $brand['id'];
                        break;
                    }
                }

                // Find device class ID by class name
                foreach ($deviceClasses as $deviceClass) {
                    if ($deviceClass['name'] === ($model['class'] ?? '')) {
                        $selectedDeviceClassId = $deviceClass['id'];
                        break;
                    }
                }
                break;
            }
        }
    }
    // Handle actual update submission
    elseif (isset($_POST['update_model'])) {
        $modelId = trim($_POST['model_id'] ?? '');
        $newModelName = trim($_POST['model_name'] ?? '');
        $brandId = $_POST['brand_id'] ?? '';
        $deviceClassId = $_POST['device_class_id'] ?? '';

        if ($modelId === '') {
            $message = 'Please select a model to update.';
        } elseif ($newModelName === '') {
            $message = 'Please enter a model name.';
        } elseif ($brandId === '') {
            $message = 'Please select a brand.';
        } elseif ($deviceClassId === '') {
            $message = 'Please select a device class.';
        } else {
            $data = json_encode([
                'id' => $modelId,
                'name' => $newModelName,
                'brand_id' => $brandId,
                'device_class_id' => $deviceClassId
            ]);

            $ch = curl_init(MODELS_ENDPOINT);
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
                $message = 'Model updated successfully!';
                // Refresh models list
                $models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];
                // Clear selection after successful update
                $selectedModelId = '';
                $selectedModelName = '';
                $selectedBrandId = '';
                $selectedDeviceClassId = '';
            } else {
                $message = 'Failed to update model. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Model</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($models)): ?>
    <p><em>No models available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="model_id">Select Model to Update:</label>
        <select id="model_id" name="model_id" required>
            <option value="">-- Select Model --</option>
            <?php foreach ($models as $model): ?>
                <option value="<?= htmlspecialchars($model['id']) ?>"
                    <?= ($model['id'] == $selectedModelId) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($model['model']) ?> (ID: <?= htmlspecialchars($model['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_model">Load Model</button>
        <br><br>

        <?php if (!empty($selectedModelId)): ?>
            <label for="model_name">Model Name:</label>
            <input type="text" id="model_name" name="model_name"
                   value="<?= htmlspecialchars($selectedModelName) ?>" required><br><br>

            <label for="brand_id">Brand:</label>
            <select id="brand_id" name="brand_id" required>
                <option value="">-- Select Brand --</option>
                <?php foreach ($brands as $brand): ?>
                    <option value="<?= htmlspecialchars($brand['id']) ?>"
                        <?= ($brand['id'] == $selectedBrandId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($brand['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="device_class_id">Device Class:</label>
            <select id="device_class_id" name="device_class_id" required>
                <option value="">-- Select Device Class --</option>
                <?php foreach ($deviceClasses as $deviceClass): ?>
                    <option value="<?= htmlspecialchars($deviceClass['id']) ?>"
                        <?= ($deviceClass['id'] == $selectedDeviceClassId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($deviceClass['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <button type="submit" name="update_model">Update Model</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
