<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$devices = json_decode(@file_get_contents(DEVICES_ENDPOINT), true) ?: [];
$models = json_decode(@file_get_contents(MODELS_ENDPOINT), true) ?: [];
$zones = json_decode(@file_get_contents(ZONES_ENDPOINT), true) ?: [];
$proprietaries = json_decode(@file_get_contents(PROPRIETARIES_ENDPOINT), true) ?: [];
$scanProfiles = json_decode(@file_get_contents(SCAN_PROFILES_ENDPOINT), true) ?: [];

// Variables to hold selected device data
$selectedDeviceId = '';
$selectedDeviceLabel = '';
$selectedModelId = '';
$selectedZoneId = '';
$selectedProprietaryId = '';
$selectedIPs = [];
$selectedIsUnmanaged = false;
$selectedProfile = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Handle device selection (Load Device button)
    if (isset($_POST['select_device']) && !empty($_POST['device_id'])) {
        $selectedDeviceId = trim($_POST['device_id']);

        // Find the selected device and pre-fill its data
        foreach ($devices as $device) {
            if ($device['id'] == $selectedDeviceId) {
                $selectedDeviceLabel = $device['label'] ?? '';
                $selectedIPs = $device['ips'] ?? [];
                $selectedIsUnmanaged = !empty($device['is_unmanaged']);
                $selectedProfile = $device['profile'] ?? '';

                // Find model ID by model name
                foreach ($models as $model) {
                    if ($model['model'] === ($device['model'] ?? '')) {
                        $selectedModelId = $model['id'];
                        break;
                    }
                }

                // Get zone ID (already in device data)
                $selectedZoneId = $device['zoneid'] ?? '';

                // Find proprietary ID by name
                foreach ($proprietaries as $prop) {
                    if ($prop['name'] === ($device['proprietary'] ?? '')) {
                        $selectedProprietaryId = $prop['id'];
                        break;
                    }
                }
                break;
            }
        }
    }
    // Handle actual update submission
    elseif (isset($_POST['update_device'])) {
        $deviceId = trim($_POST['device_id'] ?? '');
        $newDeviceLabel = trim($_POST['device_label'] ?? '');
        $modelId = $_POST['model_id'] ?? '';
        $zoneId = $_POST['zone_id'] ?? '';
        $proprietaryId = $_POST['proprietary_id'] ?? '';
        $isUnmanaged = isset($_POST['is_unmanaged']);
        $profile = $_POST['profile'] ?? '';
        // Preserve state if validation below fails and re-renders the form.
        $selectedIsUnmanaged = $isUnmanaged;
        $selectedProfile = $profile;

        // Collect IPs from dynamic inputs
        $ips = [];
        if (isset($_POST['ips']) && is_array($_POST['ips'])) {
            foreach ($_POST['ips'] as $ip) {
                $ip = trim($ip);
                if (!empty($ip)) {
                    $ips[] = $ip;
                }
            }
        }

        if ($deviceId === '') {
            $message = 'Please select a device to update.';
        } elseif ($newDeviceLabel === '') {
            $message = 'Please enter a device label.';
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
                'proprietary_id' => $proprietaryId,
                'is_unmanaged' => $isUnmanaged,
                'profile' => $profile,
                'ips' => $ips
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
                // Clear selection after successful update
                $selectedDeviceId = '';
                $selectedDeviceLabel = '';
                $selectedModelId = '';
                $selectedZoneId = '';
                $selectedProprietaryId = '';
                $selectedIPs = [];
                $selectedIsUnmanaged = false;
            } else {
                $message = 'Failed to update device. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
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
                <option value="<?= htmlspecialchars($device['id']) ?>"
                    <?= ($device['id'] == $selectedDeviceId) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($device['label']) ?> (ID: <?= htmlspecialchars($device['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_device">Load Device</button>
        <br><br>

        <?php if (!empty($selectedDeviceId)): ?>
            <label for="device_label">Device Label:</label>
            <input type="text" id="device_label" name="device_label"
                   value="<?= htmlspecialchars($selectedDeviceLabel) ?>" required><br><br>

            <label for="model_id">Model:</label>
            <select id="model_id" name="model_id" required>
                <option value="">-- Select Model --</option>
                <?php foreach ($models as $model): ?>
                    <option value="<?= htmlspecialchars($model['id']) ?>"
                        <?= ($model['id'] == $selectedModelId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($model['model']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="zone_id">Zone:</label>
            <select id="zone_id" name="zone_id" required>
                <option value="">-- Select Zone --</option>
                <?php foreach ($zones as $zone): ?>
                    <option value="<?= htmlspecialchars($zone['id']) ?>"
                        <?= ($zone['id'] == $selectedZoneId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($zone['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label for="proprietary_id">Proprietary:</label>
            <select id="proprietary_id" name="proprietary_id" required>
                <option value="">-- Select Proprietary --</option>
                <?php foreach ($proprietaries as $proprietary): ?>
                    <option value="<?= htmlspecialchars($proprietary['id']) ?>"
                        <?= ($proprietary['id'] == $selectedProprietaryId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($proprietary['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <label>
                <input type="checkbox" name="is_unmanaged" value="1" <?= $selectedIsUnmanaged ? 'checked' : '' ?>>
                Unmanaged VLANs (switch replicates all VLANs through all ports)
            </label><br><br>

            <label for="profile">Scan profile:</label>
            <select id="profile" name="profile">
                <option value="">-- None --</option>
                <?php foreach ($scanProfiles as $prof): $pk = ($prof['kind'] ?? '') !== '' ? $prof['kind'] : 'device'; ?>
                    <option value="<?= htmlspecialchars($prof['name'] ?? '') ?>" <?= (($prof['name'] ?? '') === $selectedProfile) ? 'selected' : '' ?>>
                        <?= htmlspecialchars(($prof['name'] ?? '') . ' (' . $pk . (($prof['host'] ?? '') !== '' ? ', ' . $prof['host'] : '') . ')') ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>

            <fieldset>
                <legend>IP Addresses</legend>
                <div id="ipsContainer">
                    <?php foreach ($selectedIPs as $index => $ip): ?>
                        <div id="ip_<?= $index ?>" style="margin-bottom: 5px;">
                            <input type="text" name="ips[]"
                                   value="<?= htmlspecialchars($ip) ?>"
                                   placeholder="e.g., 192.168.1.100"
                                   pattern="^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$"
                                   title="IPv4 format: xxx.xxx.xxx.xxx"
                                   style="width: 200px; margin-right: 5px;">
                            <button type="button" onclick="removeIPField(<?= $index ?>)">Remove</button>
                        </div>
                    <?php endforeach; ?>
                </div>
                <button type="button" onclick="addIPField()">+ Add IP Address</button>
            </fieldset>
            <br>

            <button type="submit" name="update_device">Update Device</button>
        <?php endif; ?>
    </form>

    <script>
        let ipIndex = <?= count($selectedIPs) ?>;

        function addIPField() {
            const container = document.getElementById('ipsContainer');
            const ipDiv = document.createElement('div');
            ipDiv.id = 'ip_' + ipIndex;
            ipDiv.style.marginBottom = '5px';

            ipDiv.innerHTML = `
                <input type="text" name="ips[]" placeholder="e.g., 192.168.1.100"
                       pattern="^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$"
                       title="IPv4 format: xxx.xxx.xxx.xxx"
                       style="width: 200px; margin-right: 5px;">
                <button type="button" onclick="removeIPField(${ipIndex})">Remove</button>
            `;

            container.appendChild(ipDiv);
            ipIndex++;
        }

        function removeIPField(index) {
            const element = document.getElementById('ip_' + index);
            if (element) {
                element.remove();
            }
        }
    </script>
<?php endif; ?>

<hr>
<?php include __DIR__ . '/migratedevice.php'; ?>
