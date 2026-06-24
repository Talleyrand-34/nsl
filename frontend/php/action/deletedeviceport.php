<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing device ports for dropdown
$devicePortsJson = @file_get_contents(DEVICEPORTS_ENDPOINT);
$devicePorts = json_decode($devicePortsJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $deviceId = $_POST['device_id'] ?? '';
    $modelPortId = $_POST['model_port_id'] ?? '';

    if ($deviceId !== '' && $modelPortId !== '') {
        $data = json_encode([
            'device_id' => $deviceId,
            'model_port_id' => $modelPortId,
            'cascade' => isset($_POST['cascade'])
        ]);

        $ch = curl_init(DEVICEPORTS_ENDPOINT);
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
            $message = "Device port deleted successfully!";
            // Refresh device ports list after deletion
            $devicePortsJson = @file_get_contents(DEVICEPORTS_ENDPOINT);
            $devicePorts = json_decode($devicePortsJson, true);
        } else {
            $message = "Failed to delete device port. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a device port.";
    }
}
?>

<h2>Delete a Device Port</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($devicePorts) && count($devicePorts) > 0): ?>
<form method="post">
    <label for="device_port">Select Device Port to Delete:</label>
    <select id="device_port" name="device_port" required onchange="setDevicePortIds(this.value)">
        <option value="">-- Select a Device Port --</option>
        <?php foreach ($devicePorts as $devicePort): ?>
            <option value="<?= htmlspecialchars($devicePort['devid']) . '|' . htmlspecialchars($devicePort['modelid']) ?>">
                <?= htmlspecialchars($devicePort['devname']) ?> - <?= htmlspecialchars($devicePort['portname']) ?>
                <?php if (!empty($devicePort['mac_address'])): ?>
                    (MAC: <?= htmlspecialchars($devicePort['mac_address']) ?>)
                <?php endif; ?>
            </option>
        <?php endforeach; ?>
    </select>
    
    <input type="hidden" id="device_id" name="device_id">
    <input type="hidden" id="model_port_id" name="model_port_id">
    <br><br>
    <label><input type="checkbox" name="cascade" value="1"> Delete on cascade (also remove dependent connections)</label>
    <br><br>
    <button type="submit">Delete Device Port</button>
</form>

<script>
function setDevicePortIds(value) {
    if (value) {
        const [deviceId, modelPortId] = value.split('|');
        document.getElementById('device_id').value = deviceId;
        document.getElementById('model_port_id').value = modelPortId;
    } else {
        document.getElementById('device_id').value = '';
        document.getElementById('model_port_id').value = '';
    }
}
</script>

<?php else: ?>
    <p><em>No device ports available to delete.</em></p>
<?php endif; ?>