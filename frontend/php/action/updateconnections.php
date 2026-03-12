
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

// Variables to hold selected connection data
$selectedConnectionId = '';
$selectedFromDeviceId = '';
$selectedFromModelPortId = '';
$selectedToDeviceId = '';
$selectedToModelPortId = '';
$selectedAllowVLANUnion = false;

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Handle connection selection (Load Connection button)
    if (isset($_POST['select_connection']) && !empty($_POST['connection_id'])) {
        $selectedConnectionId = trim($_POST['connection_id']);

        // Find the selected connection and pre-fill its data
        foreach ($connections as $connection) {
            if ($connection['id'] == $selectedConnectionId) {
                // Find from device ID by device label/name
                foreach ($devices as $device) {
                    if ($device['label'] === ($connection['fromdevice'] ?? '')) {
                        $selectedFromDeviceId = $device['id'];
                        break;
                    }
                }

                // Find to device ID by device label/name
                foreach ($devices as $device) {
                    if ($device['label'] === ($connection['todevice'] ?? '')) {
                        $selectedToDeviceId = $device['id'];
                        break;
                    }
                }

                // Find from model port ID by port name and device model
                $fromDeviceModel = getDeviceModel($devices, $selectedFromDeviceId);
                if ($fromDeviceModel) {
                    foreach ($modelPorts as $port) {
                        if ($port['name'] === ($connection['frommodel'] ?? '') && $port['model'] === $fromDeviceModel) {
                            $selectedFromModelPortId = $port['id'];
                            break;
                        }
                    }
                }

                // Find to model port ID by port name and device model
                $toDeviceModel = getDeviceModel($devices, $selectedToDeviceId);
                if ($toDeviceModel) {
                    foreach ($modelPorts as $port) {
                        if ($port['name'] === ($connection['tomodel'] ?? '') && $port['model'] === $toDeviceModel) {
                            $selectedToModelPortId = $port['id'];
                            break;
                        }
                    }
                }

                break;
            }
        }
    }
    // Handle form changes to update model port options (when device selection changes)
    elseif (isset($_POST['change_device'])) {
        // Keep the current selections
        $selectedConnectionId = $_POST['connection_id'] ?? '';
        $selectedFromDeviceId = $_POST['from_device_id'] ?? '';
        $selectedFromModelPortId = $_POST['from_modelport_id'] ?? '';
        $selectedToDeviceId = $_POST['to_device_id'] ?? '';
        $selectedToModelPortId = $_POST['to_modelport_id'] ?? '';
        $selectedAllowVLANUnion = isset($_POST['allowVLANUnion']);
    }
    // Handle actual update submission
    elseif (isset($_POST['update_connection'])) {
        $connectionId = $_POST['connection_id'] ?? '';
        $fromDeviceId = $_POST['from_device_id'] ?? '';
        $fromModelPortId = $_POST['from_modelport_id'] ?? '';
        $toDeviceId = $_POST['to_device_id'] ?? '';
        $toModelPortId = $_POST['to_modelport_id'] ?? '';
        $allowVLANUnion = isset($_POST['allowVLANUnion']);

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
                'connection_id' => $connectionId,
                'new_from_device_id' => $fromDeviceId,
                'new_from_model_port_id' => $fromModelPortId,
                'new_to_device_id' => $toDeviceId,
                'new_to_model_port_id' => $toModelPortId,
                'allow_vlan_union' => $allowVLANUnion
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
                // Clear selection after successful update
                $selectedConnectionId = '';
                $selectedFromDeviceId = '';
                $selectedFromModelPortId = '';
                $selectedToDeviceId = '';
                $selectedToModelPortId = '';
                $selectedAllowVLANUnion = false;
            } else {
                $message = 'Failed to update connection. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}

// Filter model ports for each device selection
$fromDeviceModel = getDeviceModel($devices, $selectedFromDeviceId);
$toDeviceModel = getDeviceModel($devices, $selectedToDeviceId);

$fromModelPorts = $fromDeviceModel ? getModelPortsByModel($modelPorts, $fromDeviceModel) : [];
$toModelPorts = $toDeviceModel ? getModelPortsByModel($modelPorts, $toDeviceModel) : [];
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
                <option value="<?= htmlspecialchars($connection['id']) ?>"
                    <?= ($connection['id'] == $selectedConnectionId) ? 'selected' : '' ?>>
                    ID: <?= htmlspecialchars($connection['id']) ?>
                    (From: <?= htmlspecialchars($connection['fromdevice'] ?? 'N/A') ?>
                    To: <?= htmlspecialchars($connection['todevice'] ?? 'N/A') ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_connection">Load Connection</button>
        <br><br>

        <?php if (!empty($selectedConnectionId)): ?>
            <label for="from_device_id">Source Device:</label>
            <select id="from_device_id" name="from_device_id" required onchange="this.form.change_device.click()">
                <option value="">-- Select Source Device --</option>
                <?php foreach ($devices as $device): ?>
                    <option value="<?= htmlspecialchars($device['id']) ?>"
                        <?= ($device['id'] == $selectedFromDeviceId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($device['label'] ?? $device['name'] ?? $device['id']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="from_modelport_id">Source Model Port:</label>
            <select id="from_modelport_id" name="from_modelport_id" required>
                <option value="">-- Select Source Port --</option>
                <?php foreach ($fromModelPorts as $port): ?>
                    <option value="<?= htmlspecialchars($port['id']) ?>"
                        <?= ($port['id'] == $selectedFromModelPortId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($port['name']) ?> (<?= htmlspecialchars($port['model']) ?>)
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="to_device_id">Destination Device:</label>
            <select id="to_device_id" name="to_device_id" required onchange="this.form.change_device.click()">
                <option value="">-- Select Destination Device --</option>
                <?php foreach ($devices as $device): ?>
                    <option value="<?= htmlspecialchars($device['id']) ?>"
                        <?= ($device['id'] == $selectedToDeviceId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($device['label'] ?? $device['name'] ?? $device['id']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="to_modelport_id">Destination Model Port:</label>
            <select id="to_modelport_id" name="to_modelport_id" required>
                <option value="">-- Select Destination Port --</option>
                <?php foreach ($toModelPorts as $port): ?>
                    <option value="<?= htmlspecialchars($port['id']) ?>"
                        <?= ($port['id'] == $selectedToModelPortId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($port['name']) ?> (<?= htmlspecialchars($port['model']) ?>)
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <fieldset>
                <legend>VLAN Validation Options</legend>
                <label>
                    <input type="checkbox" name="allowVLANUnion" value="1" <?= $selectedAllowVLANUnion ? 'checked' : '' ?>>
                    Allow VLAN Union (Allow connection if VLANs have any overlap)
                </label>
                <br><small><em>By default, VLANs must match exactly between ports. Enable this to allow connections if VLANs have any overlap.</em></small>
            </fieldset><br>

            <!-- Hidden button for device change handling -->
            <button type="submit" name="change_device" style="display: none;"></button>

            <button type="submit" name="update_connection">Update Connection</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
