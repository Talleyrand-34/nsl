<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing devices for dropdown
$devicesJson = @file_get_contents(DEVICES_ENDPOINT);
$devices = json_decode($devicesJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $id = trim($_POST['id'] ?? '');

    if ($id !== '') {
        $data = json_encode(['id' => $id]);

        $ch = curl_init(DEVICES_ENDPOINT);
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
            $message = "Device deleted successfully!";
            // Refresh devices list after deletion
            $devicesJson = @file_get_contents(DEVICES_ENDPOINT);
            $devices = json_decode($devicesJson, true);
        } else {
            $message = "Failed to delete device. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a device.";
    }
}
?>

<h2>Delete a Device</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($devices) && count($devices) > 0): ?>
<form method="post">
    <label for="id">Select Device to Delete:</label>
    <select id="id" name="id" required>
        <option value="">-- Select a Device --</option>
        <?php foreach ($devices as $device): ?>
            <option value="<?= htmlspecialchars($device['id']) ?>">
                ID: <?= htmlspecialchars($device['id']) ?> - <?= htmlspecialchars($device['name']) ?> (<?= htmlspecialchars($device['model']) ?> - <?= htmlspecialchars($device['brand']) ?>)
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Delete Device</button>
</form>
<?php else: ?>
    <p><em>No devices available to delete.</em></p>
<?php endif; ?>