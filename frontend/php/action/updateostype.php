<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedModelTypeId = '';
$selectedModelTypeName = '';

// Fetch existing device classes for selection
$osTypesJson = @file_get_contents(OSTYPES_ENDPOINT);
$osTypes = json_decode($osTypesJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a device class selection (not form submission)
    if (isset($_POST['select_ostype']) && !empty($_POST['ostype_id'])) {
        $selectedModelTypeId = trim($_POST['ostype_id']);
        // Find the selected device class to pre-fill the form
        foreach ($osTypes as $osType) {
            if ($osType['id'] == $selectedModelTypeId) {
                $selectedModelTypeName = $osType['name'];
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_ostype'])) {
        $osTypeId = trim($_POST['ostype_id'] ?? '');
        $newModelTypeName = trim($_POST['ostype_name'] ?? '');
        
        if ($osTypeId === '') {
            $message = 'Please select a device class to update.';
        } elseif ($newModelTypeName === '') {
            $message = 'Please enter a new device class name.';
        } else {
            $data = json_encode([
                'id' => $osTypeId,
                'name' => $newModelTypeName
            ]);
            $ch = curl_init(OSTYPES_ENDPOINT);
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
                $osTypesJson = @file_get_contents(OSTYPES_ENDPOINT);
                $osTypes = json_decode($osTypesJson, true) ?: [];
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

<?php if (empty($osTypes)): ?>
    <p><em>No device classes available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="ostype_id">Select Device Class to Update:</label>
        <select id="ostype_id" name="ostype_id" required>
            <option value="">-- Select Device Class --</option>
            <?php foreach ($osTypes as $osType): ?>
                <option value="<?= htmlspecialchars($osType['id']) ?>" 
                    <?= ($osType['id'] == ($selectedModelTypeId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($osType['name']) ?> (ID: <?= htmlspecialchars($osType['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_ostype">Load Device Class</button>
        <br><br>
        
        <?php if (!empty($selectedModelTypeId)): ?>
            <label for="ostype_name">New Device Class Name:</label>
            <input type="text" id="ostype_name" name="ostype_name" 
                   value="<?= htmlspecialchars($selectedModelTypeName) ?>" required><br><br>
            <button type="submit" name="update_ostype">Update Device Class</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
