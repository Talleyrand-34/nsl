<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedModelTypeId = '';
$selectedModelTypeName = '';

// Fetch existing device classes for selection
$modelTypesJson = @file_get_contents(MODELTYPES_ENDPOINT);
$modelTypes = json_decode($modelTypesJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a device class selection (not form submission)
    if (isset($_POST['select_modeltype']) && !empty($_POST['modeltype_id'])) {
        $selectedModelTypeId = trim($_POST['modeltype_id']);
        // Find the selected device class to pre-fill the form
        foreach ($modelTypes as $modelType) {
            if ($modelType['id'] == $selectedModelTypeId) {
                $selectedModelTypeName = $modelType['name'];
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_modeltype'])) {
        $modelTypeId = trim($_POST['modeltype_id'] ?? '');
        $newModelTypeName = trim($_POST['modeltype_name'] ?? '');
        
        if ($modelTypeId === '') {
            $message = 'Please select a device class to update.';
        } elseif ($newModelTypeName === '') {
            $message = 'Please enter a new device class name.';
        } else {
            $data = json_encode([
                'id' => $modelTypeId,
                'name' => $newModelTypeName
            ]);
            $ch = curl_init(MODELTYPES_ENDPOINT);
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
                $message = 'Device class updated successfully!';
                // Refresh list
                $modelTypesJson = @file_get_contents(MODELTYPES_ENDPOINT);
                $modelTypes = json_decode($modelTypesJson, true) ?: [];
                // Clear selection after successful update
                $selectedModelTypeId = '';
                $selectedModelTypeName = '';
            } else {
                $message = 'Failed to update device class. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Device Class</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($modelTypes)): ?>
    <p><em>No device classes available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="modeltype_id">Select Device Class to Update:</label>
        <select id="modeltype_id" name="modeltype_id" required>
            <option value="">-- Select Device Class --</option>
            <?php foreach ($modelTypes as $modelType): ?>
                <option value="<?= htmlspecialchars($modelType['id']) ?>" 
                    <?= ($modelType['id'] == ($selectedModelTypeId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($modelType['name']) ?> (ID: <?= htmlspecialchars($modelType['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_modeltype">Load Device Class</button>
        <br><br>
        
        <?php if (!empty($selectedModelTypeId)): ?>
            <label for="modeltype_name">New Device Class Name:</label>
            <input type="text" id="modeltype_name" name="modeltype_name" 
                   value="<?= htmlspecialchars($selectedModelTypeName) ?>" required><br><br>
            <button type="submit" name="update_modeltype">Update Device Class</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
