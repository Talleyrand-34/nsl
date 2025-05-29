<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];
$zones = json_decode(@file_get_contents(ZONES_ENDPOINT), true) ?: [];
$proprietaries = json_decode(@file_get_contents(PROPRIETARIES_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $deviceId = trim($_POST['device_id'] ?? '');
    $newDeviceLabel = trim($_POST['device_label'] ?? '');
    $modelId = $_POST['model_id'] ?? '';
    $zoneId = $_POST['zone_id'] ?? '';
    $proprietaryId = $_POST['proprietary_id'] ?? '';

    if ($deviceId === '') {
        $message = 'Please select a device to update.';
    } elseif ($newDeviceLabel === '') {
        $message = 'Please enter a new device label.';
    } elseif ($modelId === '') {
        $message = 'Please select a model.';
    } elseif ($zoneId === '') {
        $message = 'Please select a zone.';
    } elseif ($proprietaryId === '') {
        $message = 'Please select a proprietary.';
    } else {
        $data = json_encode([
            'id' => $deviceId,
            'label' => $newDeviceLabel,
            'model_id' => $modelId,
            'zone_id' => $zoneId,
            'proprietary_id' => $proprietaryId
        ]);

        $ch = curl_init(DEVICES_ENDPOINT);
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
            $message = 'Device updated successfully!';
            // Refresh devices list
            $devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
        } else {
            $message = 'Failed to update device. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Update Device</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($devices)): ?>
    <p><em>No devices available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="device_id">Select Device to Update:</label>
        <select id="device_id" name="device_id" required>
            <option value="">-- Select Device --</option>
            <?php foreach ($devices as $device): ?>
                <option value="<?= htmlspecialchars($device['id']) ?>">
                    <?= htmlspecialchars($device['label']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="device_label">New Device Label:</label>
        <input type="text" id="device_label" name="device_label" required><br><br>

        <label for="model_id">Model:</label>
        <select id="model_id" name="model_id" required>
            <option value="">-- Select Model --</option>
            <?php foreach ($models as $model): ?>
                <option value="<?= htmlspecialchars($model['id']) ?>">
                    <?= htmlspecialchars($model['model']) ?>
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="zone_id">Zone:</label>
        <select id="zone_id" name="zone_id" required>
            <option value="">-- Select Zone --</option>
            <?php foreach ($zones as $zone): ?>
                <option value="<?= htmlspecialchars($zone['id']) ?>">
                    <?= htmlspecialchars($zone['name']) ?>
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="proprietary_id">Proprietary:</label>
        <select id="proprietary_id" name="proprietary_id" required>
            <option value="">-- Select Proprietary --</option>
            <?php foreach ($proprietaries as $proprietary): ?>
                <option value="<?= htmlspecialchars($proprietary['id']) ?>">
                    <?= htmlspecialchars($proprietary['name']) ?>
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <button type="submit">Update Device</button>
    </form>
<?php endif; ?>