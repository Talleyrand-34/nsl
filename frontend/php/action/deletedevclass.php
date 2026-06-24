<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing device classes for dropdown
$devClassesJson = @file_get_contents(DEVCLASSES_ENDPOINT);
$devClasses = json_decode($devClassesJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $name = trim($_POST['name'] ?? '');

    if ($name !== '') {
        $data = json_encode(['name' => $name, 'cascade' => isset($_POST['cascade'])]);

        $ch = curl_init(DEVCLASSES_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "DELETE");
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode >= 200 && $httpCode < 300) {
            $message = "Device class deleted successfully!";
            // Refresh device classes list after deletion
            $devClassesJson = @file_get_contents(DEVCLASSES_ENDPOINT);
            $devClasses = json_decode($devClassesJson, true);
        } else {
            $message = "Failed to delete device class. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a device class.";
    }
}
?>

<h2>Delete a Device Class</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($devClasses) && count($devClasses) > 0): ?>
<form method="post">
    <label for="name">Select Device Class to Delete:</label>
    <select id="name" name="name" required>
        <option value="">-- Select a Device Class --</option>
        <?php foreach ($devClasses as $devClass): ?>
            <option value="<?= htmlspecialchars($devClass['name']) ?>">
                <?= htmlspecialchars($devClass['name']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <br><br>
    <label><input type="checkbox" name="cascade" value="1"> Delete on cascade (also remove dependents)</label>
    <br><br>
    <button type="submit">Delete Device Class</button>
</form>
<?php else: ?>
    <p><em>No device classes available to delete.</em></p>
<?php endif; ?>