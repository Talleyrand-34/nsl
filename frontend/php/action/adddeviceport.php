<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$modelPorts = json_decode(@file_get_contents(MODELPORTS_ENDPOINT), true) ?: [];
$vlans = json_decode(@file_get_contents(VLANS_ENDPOINT), true) ?: [];

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

$deviceId = $_POST['device_id'] ?? '';
$modelPortId = $_POST['modelport_id'] ?? '';
$macAddress = $_POST['mac_address'] ?? '';

// Filter model ports based on selected device
$deviceModel = getDeviceModel($devices, $deviceId);
$availableModelPorts = $deviceModel ? getModelPortsByModel($modelPorts, $deviceModel) : [];

if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['submit'])) {
    $deviceId = $_POST['device_id'] ?? '';
    $modelPortId = $_POST['modelport_id'] ?? '';
    $macAddress = trim($_POST['mac_address'] ?? '');

    // Build VLAN configs array
    $vlanConfigs = [];
    if (isset($_POST['vlan_numbers']) && is_array($_POST['vlan_numbers'])) {
        foreach ($_POST['vlan_numbers'] as $index => $vlanNumber) {
            if (!empty($vlanNumber)) {
                $tagged = isset($_POST['vlan_tagged'][$index]) && $_POST['vlan_tagged'][$index] === 'true';
                $vlanConfigs[] = [
                    'vlan_number' => $vlanNumber,
                    'tagged' => $tagged
                ];
            }
        }
    }

    if ($deviceId === '') {
        $message = 'Please select a device.';
    } elseif ($modelPortId === '') {
        $message = 'Please select a model port.';
    } else {
        $data = json_encode([
            'deviceid' => $deviceId,
            'modelportid' => $modelPortId,
            'mac_address' => $macAddress,
            'vlan_configs' => $vlanConfigs
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
    <form method="post" id="devicePortForm">
        <label for="device_id">Select Device:</label>
        <select id="device_id" name="device_id" required onchange="this.form.submit()">
            <option value="">-- Select Device --</option>
            <?php foreach ($devices as $device): ?>
                <option value="<?= htmlspecialchars($device['id']) ?>"
                    <?= ($deviceId == $device['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($device['label']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="modelport_id">Select Model Port:</label>
        <select id="modelport_id" name="modelport_id" required>
            <option value="">-- Select Model Port --</option>
            <?php foreach ($availableModelPorts as $modelPort): ?>
                <option value="<?= htmlspecialchars($modelPort['id']) ?>"
                    <?= ($modelPortId == $modelPort['id']) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($modelPort['name']) ?> (<?= htmlspecialchars($modelPort['model']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="mac_address">MAC Address (optional):</label>
        <input type="text" id="mac_address" name="mac_address"
               value="<?= htmlspecialchars($macAddress) ?>"
               placeholder="e.g., AA:BB:CC:DD:EE:FF"
               pattern="[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}"
               title="MAC address format: XX:XX:XX:XX:XX:XX"><br><br>

        <fieldset>
            <legend>VLAN Configuration (Optional)</legend>
            <div id="vlanConfigsContainer">
                <!-- VLAN configs will be added here dynamically -->
            </div>
            <button type="button" onclick="addVlanConfig()">+ Add VLAN</button>
            <br><small><em>Note: Only one untagged VLAN is allowed per port by default (controlled by model port settings)</em></small>
        </fieldset>
        <br>

        <button type="submit" name="submit">Add Device Port</button>
    </form>

    <script>
        let vlanConfigIndex = 0;

        function addVlanConfig() {
            const container = document.getElementById('vlanConfigsContainer');
            const vlanConfigDiv = document.createElement('div');
            vlanConfigDiv.id = 'vlanConfig_' + vlanConfigIndex;
            vlanConfigDiv.style.marginBottom = '10px';
            vlanConfigDiv.style.padding = '10px';
            vlanConfigDiv.style.border = '1px solid #ddd';
            vlanConfigDiv.style.borderRadius = '4px';

            vlanConfigDiv.innerHTML = `
                <label>VLAN Number:</label>
                <select name="vlan_numbers[]" required style="margin-right: 10px;">
                    <option value="">-- Select VLAN --</option>
                    <?php foreach ($vlans as $vlan): ?>
                        <option value="<?= htmlspecialchars($vlan['vlanid']) ?>">
                            VLAN <?= htmlspecialchars($vlan['vlanid']) ?> - <?= htmlspecialchars($vlan['vlanname']) ?>
                        </option>
                    <?php endforeach; ?>
                </select>

                <label>Type:</label>
                <select name="vlan_tagged[]" required style="margin-right: 10px;">
                    <option value="true">Tagged</option>
                    <option value="false">Untagged</option>
                </select>

                <button type="button" onclick="removeVlanConfig(${vlanConfigIndex})">Remove</button>
            `;

            container.appendChild(vlanConfigDiv);
            vlanConfigIndex++;
        }

        function removeVlanConfig(index) {
            const element = document.getElementById('vlanConfig_' + index);
            if (element) {
                element.remove();
            }
        }
    </script>
<?php endif; ?>
