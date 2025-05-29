<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$connections = json_decode(@file_get_contents(CONNECTIONS_ENDPOINT), true) ?: [];
$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$modelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $connectionId = trim($_POST['connection_id'] ?? '');
    $fromDeviceId = $_POST['from_device_id'] ?? '';
    $fromModelPortId = $_POST['from_modelport_id'] ?? '';
    $fromIPSegment = $_POST['from_ip_segment'] ?? '';
    $toDeviceId = $_POST['to_device_id'] ?? '';
    $toModelPortId = $_POST['to_modelport_id'] ?? '';
    $toIPSegment = $_POST['to_ip_segment'] ?? '';

    if ($connectionId === '') {
        $message = 'Please select a connection to update.';
    } elseif ($fromDeviceId === '') {
        $message = 'Please select a source device.';
    } elseif ($fromModelPortId === '') {
        $message = 'Please select a source model port.';
    } elseif ($toDeviceId === '') {
        $message = 'Please select a destination device.';
    } elseif ($toModelPortId === '') {
        $message = 'Please select a destination model port.';
    } else {
        $data = json_encode([
            'id' => $connectionId,
            'from_device' => $fromDeviceId,
            'from_port' => $fromModelPortId,
            'from_ip_segment' => $fromIPSegment,
            'to_device' => $toDeviceId,
            'to_port' => $toModelPortId,
            'to_ip_segment' => $toIPSegment
        ]);

        $ch = curl_init(CONNECTIONS_ENDPOINT);
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
            $message = 'Connection updated successfully!';
            // Refresh connections list
            $connections = json_decode(@file_get_contents(CONNECTIONS_ENDPOINT), true) ?: [];
        } else {
            $message = 'Failed to update connection. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Update Connection</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($connections)): ?>
    <p><em>No connections available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="connection_id">Select Connection to Update:</label>
        <select id="connection_id" name="connection_id" required>
            <option value="">-- Select Connection --</option>
            <?php foreach ($connections as $connection): ?>
                <option value="<?= htmlspecialchars($connection['id']) ?>">
                    ID: <?= htmlspecialchars($connection['id']) ?> 
                    (From: <?= htmlspecialchars($connection['fromdevice'] ?? 'N/A') ?> 
                    To: <?= htmlspecialchars($connection['todevice'] ?? 'N/A') ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="from_device_id">Source Device:</label>
        <select id="from_device_id" name="from_device_id" required>
            <option value="">-- Select Source Device --</option>
            <?php foreach ($devices as $device): ?>
                <option value="<?= htmlspecialchars($device['id']) ?>">
                    <?= htmlspecialchars($device['label']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="from_modelport_id">Source Model Port:</label>
        <select id="from_modelport_id" name="from_modelport_id" required>
            <option value="">-- Select Source Port --</option>
            <?php foreach ($modelPorts as $port): ?>
                <option value="<?= htmlspecialchars($port['id']) ?>">
                    <?= htmlspecialchars($port['name']) ?> (ID: <?= htmlspecialchars($port['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="to_device_id">Destination Device:</label>
        <select id="to_device_id" name="to_device_id" required>
            <option value="">-- Select Destination Device --</option>
            <?php foreach ($devices as $device): ?>
                <option value="<?= htmlspecialchars($device['id']) ?>">
                    <?= htmlspecialchars($device['label']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="to_modelport_id">Destination Model Port:</label>
        <select id="to_modelport_id" name="to_modelport_id" required>
            <option value="">-- Select Destination Port --</option>
            <?php foreach ($modelPorts as $port): ?>
                <option value="<?= htmlspecialchars($port['id']) ?>">
                    <?= htmlspecialchars($port['name']) ?> (ID: <?= htmlspecialchars($port['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="from_ip_segment">Source IP Segment (optional):</label>
        <input type="text" id="from_ip_segment" name="from_ip_segment" 
               placeholder="e.g., 192.168.1.0/24"><br><br>

        <label for="to_ip_segment">Destination IP Segment (optional):</label>
        <input type="text" id="to_ip_segment" name="to_ip_segment" 
               placeholder="e.g., 10.0.1.0/24"><br><br>

        <button type="submit">Update Connection</button>
    </form>
<?php endif; ?>