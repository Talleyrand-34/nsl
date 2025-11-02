
<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$connections = json_decode(@file_get_contents(CONNECTIONS_ENDPOINT), true) ?: [];
$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$modelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];

// Helper: get model for a given device id
function getDeviceModel($devices, $deviceId)
{
    foreach ($devices as $dev) {
        if ($dev['id'] == $deviceId) {
            return $dev['model'] ?? null;
        }
    }
    return null;
}

// Helper: get model ports for a given model name
function getModelPortsByModel($modelPorts, $modelName)
{
    $ports = [];
    foreach ($modelPorts as $mp) {
        if ($mp['model'] === $modelName) {
            $ports[] = $mp;
        }
    }
    return $ports;
}

// Get POST values or set defaults
$connectionId = $_POST['connection_id'] ?? '';
$fromDeviceId = $_POST['from_device_id'] ?? '';
$fromModelPortId = $_POST['from_modelport_id'] ?? '';
$toDeviceId = $_POST['to_device_id'] ?? '';
$toModelPortId = $_POST['to_modelport_id'] ?? '';
$allowVLANUnion = isset($_POST['allowVLANUnion']);

// Filter model ports for each device selection
$fromDeviceModel = getDeviceModel($devices, $fromDeviceId);
$toDeviceModel = getDeviceModel($devices, $toDeviceId);

$fromModelPorts = $fromDeviceModel ? getModelPortsByModel($modelPorts, $fromDeviceModel) : [];
$toModelPorts = $toDeviceModel ? getModelPortsByModel($modelPorts, $toDeviceModel) : [];

// Handle form submission
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['submit'])) {
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
            'to_device' => $toDeviceId,
            'to_port' => $toModelPortId,
            'allowVLANUnion' => $allowVLANUnion
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
        <select id="connection_id" name="connection_id" required onchange="this.form.submit()">
            <option value="">-- Select Connection --</option>
            <?php foreach ($connections as $connection): ?>
                <option value="<?= htmlspecialchars($connection['id']) ?>"
                    <?= ($connectionId == $connection['id']) ? 'selected' : '' ?>>
                    ID: <?= htmlspecialchars($connection['id']) ?>
                    (From: <?= htmlspecialchars($connection['fromdevice'] ?? 'N/A') ?>
                    To: <?= htmlspecialchars($connection['todevice'] ?? 'N/A') ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="from_device_id">Source Device:</label>
        <select id="from_device_id" name="from_device_id" required onchange="this.form.submit()">
            <option value="">-- Select Source Device --</option>
            <?php foreach ($devices as $device): ?>
                <option value="<?= htmlspecialchars($device['id']) ?>"
                    <?= ($fromDeviceId == $device['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($device['label'] ?? $device['name'] ?? $device['id']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="from_modelport_id">Source Model Port:</label>
        <select id="from_modelport_id" name="from_modelport_id" required>
            <option value="">-- Select Source Port --</option>
            <?php foreach ($fromModelPorts as $port): ?>
                <option value="<?= htmlspecialchars($port['id']) ?>"
                    <?= ($fromModelPortId == $port['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($port['name']) ?> (<?= htmlspecialchars($port['model']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="to_device_id">Destination Device:</label>
        <select id="to_device_id" name="to_device_id" required onchange="this.form.submit()">
            <option value="">-- Select Destination Device --</option>
            <?php foreach ($devices as $device): ?>
                <option value="<?= htmlspecialchars($device['id']) ?>"
                    <?= ($toDeviceId == $device['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($device['label'] ?? $device['name'] ?? $device['id']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="to_modelport_id">Destination Model Port:</label>
        <select id="to_modelport_id" name="to_modelport_id" required>
            <option value="">-- Select Destination Port --</option>
            <?php foreach ($toModelPorts as $port): ?>
                <option value="<?= htmlspecialchars($port['id']) ?>"
                    <?= ($toModelPortId == $port['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($port['name']) ?> (<?= htmlspecialchars($port['model']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <fieldset>
            <legend>VLAN Validation Options</legend>
            <label>
                <input type="checkbox" name="allowVLANUnion" value="1" <?= $allowVLANUnion ? 'checked' : '' ?>>
                Allow VLAN Union (Allow connection if VLANs have any overlap)
            </label>
            <br><small><em>By default, VLANs must match exactly between ports. Enable this to allow connections if VLANs have any overlap.</em></small>
        </fieldset><br>

        <button type="submit" name="submit">Update Connection</button>
    </form>
<?php endif; ?>
