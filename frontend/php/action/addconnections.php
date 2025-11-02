
<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch devices and model ports
$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$modelports = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];

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
function getModelPortsByModel($modelports, $modelName)
{
    $ports = [];
    foreach ($modelports as $mp) {
        if ($mp['model'] === $modelName) {
            $ports[] = $mp;
        }
    }
    return $ports;
}

$fromDevice = $_POST['fromDevice'] ?? '';
$fromModelPort = $_POST['fromModelPort'] ?? '';
$toDevice = $_POST['toDevice'] ?? '';
$toModelPort = $_POST['toModelPort'] ?? '';
$allowVLANUnion = isset($_POST['allowVLANUnion']);

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['submit'])) {
    if ($fromDevice === '' || $fromModelPort === '' || $toDevice === '' || $toModelPort === '') {
        $message = 'Please select all connection parameters.';
    } else {
        $data = json_encode([
            'fromDevice' => $fromDevice,
            'fromModelPort' => $fromModelPort,
            'toDevice' => $toDevice,
            'toModelPort' => $toModelPort,
            'allowVLANUnion' => $allowVLANUnion
        ]);

        // Debug: echo the JSON being sent
        echo '<pre>JSON sent:<br>' . htmlspecialchars($data) . '</pre>';

        $ch = curl_init(CONNECTIONS_ENDPOINT);
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
            $message = 'Connection created successfully!';
        } else {
            $message = 'Failed to create connection. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}

// Filter model ports for each device selection
$fromDeviceModel = getDeviceModel($devices, $fromDevice);
$toDeviceModel = getDeviceModel($devices, $toDevice);

$fromModelPorts = $fromDeviceModel ? getModelPortsByModel($modelports, $fromDeviceModel) : [];
$toModelPorts = $toDeviceModel ? getModelPortsByModel($modelports, $toDeviceModel) : [];
?>

<h2>Add a New Connection</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<form method="post">
    <fieldset>
        <legend>From</legend>
        <label for="fromDevice">Device:</label>
        <select id="fromDevice" name="fromDevice" required onchange="this.form.submit()">
            <option value="">-- Select Device --</option>
            <?php foreach ($devices as $dev): ?>
                <option value="<?= htmlspecialchars($dev['id']) ?>"
                    <?= ($fromDevice == $dev['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($dev['label'] ?? $dev['name'] ?? $dev['id']) ?>
                </option>
            <?php endforeach; ?>
        </select>

        <label for="fromModelPort">Model Port:</label>
        <select id="fromModelPort" name="fromModelPort" required>
            <option value="">-- Select Port --</option>
            <?php foreach ($fromModelPorts as $mp): ?>
                <option value="<?= htmlspecialchars($mp['id']) ?>"
                    <?= ($fromModelPort == $mp['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($mp['name']) ?> (<?= htmlspecialchars($mp['model']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
    </fieldset>
    <br>
    <fieldset>
        <legend>To</legend>
        <label for="toDevice">Device:</label>
        <select id="toDevice" name="toDevice" required onchange="this.form.submit()">
            <option value="">-- Select Device --</option>
            <?php foreach ($devices as $dev): ?>
                <option value="<?= htmlspecialchars($dev['id']) ?>"
                    <?= ($toDevice == $dev['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($dev['label'] ?? $dev['name'] ?? $dev['id']) ?>
                </option>
            <?php endforeach; ?>
        </select>

        <label for="toModelPort">Model Port:</label>
        <select id="toModelPort" name="toModelPort" required>
            <option value="">-- Select Port --</option>
            <?php foreach ($toModelPorts as $mp): ?>
                <option value="<?= htmlspecialchars($mp['id']) ?>"
                    <?= ($toModelPort == $mp['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($mp['name']) ?> (<?= htmlspecialchars($mp['model']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
    </fieldset>
    <br>
    <fieldset>
        <legend>VLAN Validation Options</legend>
        <label>
            <input type="checkbox" name="allowVLANUnion" value="1" <?= $allowVLANUnion ? 'checked' : '' ?>>
            Allow VLAN Union (Allow connection if VLANs have any overlap)
        </label>
        <br><small><em>By default, VLANs must match exactly between ports. Enable this to allow connections if VLANs have any overlap.</em></small>
    </fieldset>
    <br>
    <button type="submit" name="submit">Add Connection</button>
</form>
