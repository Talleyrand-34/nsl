<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$modelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $deviceId = $_POST['device_id'] ?? '';
    $modelPortId = $_POST['modelport_id'] ?? '';
    $macAddress = trim($_POST['mac_address'] ?? '');

    if ($deviceId === '') {
        $message = 'Please select a device.';
    } elseif ($modelPortId === '') {
        $message = 'Please select a model port.';
    } else {
        $data = json_encode([
            'deviceid' => $deviceId,
            'modelportid' => $modelPortId,
            'mac_address' => $macAddress
        ]);

        $ch = curl_init(DEVICEPORTS_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 201) {
            $message = 'Device port added successfully!';
        } else {
            $message = 'Failed to add device port. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Add Device Port</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($devices)): ?>
    <p><em>No devices available. Please add devices first.</em></p>
<?php elseif (empty($modelPorts)): ?>
    <p><em>No model ports available. Please add model ports first.</em></p>
<?php else: ?>
    <form method="post">
        <label for="device_id">Select Device:</label>
        <select id="device_id" name="device_id" required>
            <option value="">-- Select Device --</option>
            <?php foreach ($devices as $device): ?>
                <option value="<?= htmlspecialchars($device['id']) ?>">
                    <?= htmlspecialchars($device['label']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="modelport_id">Select Model Port:</label>
        <select id="modelport_id" name="modelport_id" required>
            <option value="">-- Select Model Port --</option>
            <?php foreach ($modelPorts as $modelPort): ?>
                <option value="<?= htmlspecialchars($modelPort['id']) ?>">
                    <?= htmlspecialchars($modelPort['name']) ?> (<?= htmlspecialchars($modelPort['model']) ?> - <?= htmlspecialchars($modelPort['brand']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="mac_address">MAC Address (optional):</label>
        <input type="text" id="mac_address" name="mac_address" placeholder="e.g., AA:BB:CC:DD:EE:FF" pattern="[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}" title="MAC address format: XX:XX:XX:XX:XX:XX"><br><br>

        <button type="submit">Add Device Port</button>
    </form>
<?php endif; ?>