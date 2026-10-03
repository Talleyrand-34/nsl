<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];
$brands = json_decode(@file_get_contents(BRANDS_ENDPOINT), true) ?: [];
$modelTypes = json_decode(@file_get_contents(MODELTYPES_ENDPOINT), true) ?: [];
$osTypes = json_decode(@file_get_contents(OSTYPES_ENDPOINT), true) ?: [];

// Variables to hold selected model data
$selectedModelId = '';
$selectedModelName = '';
$selectedBrandId = '';
$selectedModelTypeId = '';
$selectedOsTypeId = '';

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

                // Find model-type ID by name
                foreach ($modelTypes as $modelType) {
                    if ($modelType['name'] === ($model['model_type'] ?? '')) {
                        $selectedModelTypeId = $modelType['id'];
                        break;
                    }
                }

                // Find OS-type ID by name
                foreach ($osTypes as $osType) {
                    if ($osType['name'] === ($model['os_type'] ?? '')) {
                        $selectedOsTypeId = $osType['id'];
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
        $modelTypeId = $_POST['model_type_id'] ?? '';
        $osTypeId = $_POST['os_type_id'] ?? '';

        if ($modelId === '') {
            $message = 'Please select a model to update.';
        } elseif ($newModelName === '') {
            $message = 'Please enter a model name.';
        } elseif ($brandId === '') {
            $message = 'Please select a brand.';
        } elseif ($modelTypeId === '') {
            $message = 'Please select a model type.';
        } else {
            $data = json_encode([
                'id' => $modelId,
                'name' => $newModelName,
                'brand_id' => $brandId,
                'model_type_id' => $modelTypeId,
                'os_type_id' => $osTypeId
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
                $selectedModelTypeId = '';
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

            <label for="model_type_id">Model type:</label>
            <select id="model_type_id" name="model_type_id" required>
                <option value="">-- Select Model Type --</option>
                <?php foreach ($modelTypes as $modelType): ?>
                    <option value="<?= htmlspecialchars($modelType['id']) ?>"
                        <?= ($modelType['id'] == $selectedModelTypeId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($modelType['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="os_type_id">OS type (optional):</label>
            <select id="os_type_id" name="os_type_id">
                <option value="">-- None --</option>
                <?php foreach ($osTypes as $osType): ?>
                    <option value="<?= htmlspecialchars($osType['id']) ?>"
                        <?= ($osType['id'] == $selectedOsTypeId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($osType['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <button type="submit" name="update_model">Update Model</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
