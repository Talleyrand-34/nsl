<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedDevClassId = '';
$selectedDevClassName = '';

// Fetch existing device classes for selection
$devClassesJson = @file_get_contents(DEVCLASSES_ENDPOINT);
$devClasses = json_decode($devClassesJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a device class selection (not form submission)
    if (isset($_POST['select_devclass']) && !empty($_POST['devclass_id'])) {
        $selectedDevClassId = trim($_POST['devclass_id']);
        // Find the selected device class to pre-fill the form
        foreach ($devClasses as $devClass) {
            if ($devClass['id'] == $selectedDevClassId) {
                $selectedDevClassName = $devClass['name'];
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_devclass'])) {
        $devClassId = trim($_POST['devclass_id'] ?? '');
        $newDevClassName = trim($_POST['devclass_name'] ?? '');
        
        if ($devClassId === '') {
            $message = 'Please select a device class to update.';
        } elseif ($newDevClassName === '') {
            $message = 'Please enter a new device class name.';
        } else {
            $data = json_encode([
                'id' => $devClassId,
                'name' => $newDevClassName
            ]);
            $ch = curl_init(DEVCLASSES_ENDPOINT);
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
                $devClassesJson = @file_get_contents(DEVCLASSES_ENDPOINT);
                $devClasses = json_decode($devClassesJson, true) ?: [];
                // Clear selection after successful update
                $selectedDevClassId = '';
                $selectedDevClassName = '';
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

<?php if (empty($devClasses)): ?>
    <p><em>No device classes available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="devclass_id">Select Device Class to Update:</label>
        <select id="devclass_id" name="devclass_id" required>
            <option value="">-- Select Device Class --</option>
            <?php foreach ($devClasses as $devClass): ?>
                <option value="<?= htmlspecialchars($devClass['id']) ?>" 
                    <?= ($devClass['id'] == ($selectedDevClassId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($devClass['name']) ?> (ID: <?= htmlspecialchars($devClass['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_devclass">Load Device Class</button>
        <br><br>
        
        <?php if (!empty($selectedDevClassId)): ?>
            <label for="devclass_name">New Device Class Name:</label>
            <input type="text" id="devclass_name" name="devclass_name" 
                   value="<?= htmlspecialchars($selectedDevClassName) ?>" required><br><br>
            <button type="submit" name="update_devclass">Update Device Class</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
