<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';

// Fetch data for form options
$devicePorts = json_decode(@file_get_contents(DEVICEPORTS_ENDPOINT), true) ?: [];
$vlans = json_decode(@file_get_contents(VLANS_ENDPOINT), true) ?: [];

// Extract unique devices from device ports
$devices = [];
foreach ($devicePorts as $devicePort) {
    $devId = $devicePort['devid'];
    if (!isset($devices[$devId])) {
        $devices[$devId] = [
            'devid' => $devId,
            'devname' => $devicePort['devname'] ?? 'N/A'
        ];
    }
}

// Variables to hold selected device port data
$selectedDeviceId = '';
$selectedModelPortId = '';
$selectedDeviceLabel = '';
$selectedPortName = '';
$selectedMacAddress = '';
$selectedVlanConfigs = [];
$filterDeviceId = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Handle device filter
    if (isset($_POST['filter_device'])) {
        $filterDeviceId = $_POST['filter_device_id'] ?? '';
    }
    
    // Handle device port selection (Load Device Port button)
    if (isset($_POST['select_deviceport']) && !empty($_POST['deviceport_key'])) {
        // Preserve the filter
        $filterDeviceId = $_POST['current_filter'] ?? '';
        
        // The deviceport_key format is "deviceid|modelportid"
        $parts = explode('|', $_POST['deviceport_key']);
        if (count($parts) === 2) {
            $selectedDeviceId = $parts[0];
            $selectedModelPortId = $parts[1];

            // Find the selected device port and pre-fill its data
            foreach ($devicePorts as $devicePort) {
                if ($devicePort['devid'] == $selectedDeviceId && $devicePort['modelid'] == $selectedModelPortId) {
                    $selectedDeviceLabel = $devicePort['devname'] ?? '';
                    $selectedPortName = $devicePort['portname'] ?? '';
                    $selectedMacAddress = $devicePort['mac_address'] ?? '';
                    $selectedVlanConfigs = $devicePort['vlan_configs'] ?? [];
                    break;
                }
            }
        }
    }
    // Handle actual update submission
    elseif (isset($_POST['update_deviceport'])) {
        $deviceId = $_POST['device_id'] ?? '';
        $modelPortId = $_POST['modelport_id'] ?? '';
        $macAddress = trim($_POST['mac_address'] ?? '');
        $filterDeviceId = $_POST['current_filter'] ?? '';

        // Collect VLAN configs
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
            $message = 'Please select a device port to update.';
        } elseif ($modelPortId === '') {
            $message = 'Please select a device port to update.';
        } else {
            $data = json_encode([
                'device_id' => $deviceId,
                'model_port_id' => $modelPortId,
                'mac_address' => $macAddress,
                'vlan_configs' => $vlanConfigs
            ]);

            $ch = curl_init(DEVICEPORTS_ENDPOINT);
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
                $message = 'Device port updated successfully!';
                // Refresh device ports list
                $devicePorts = json_decode(@file_get_contents(DEVICEPORTS_ENDPOINT), true) ?: [];
                // Clear selection after successful update
                $selectedDeviceId = '';
                $selectedModelPortId = '';
                $selectedDeviceLabel = '';
                $selectedPortName = '';
                $selectedMacAddress = '';
                $selectedVlanConfigs = [];
            } else {
                $message = 'Failed to update device port. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}

// Filter device ports based on selected device
$filteredDevicePorts = $devicePorts;
if (!empty($filterDeviceId)) {
    $filteredDevicePorts = array_filter($devicePorts, function($port) use ($filterDeviceId) {
        return $port['devid'] == $filterDeviceId;
    });
}
?>

<h2>Update Device Port</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($devicePorts)): ?>
    <p><em>No device ports available to update.</em></p>
<?php else: ?>
    <!-- Device Filter Form -->
    <form method="post" style="margin-bottom: 20px; padding: 15px; background-color: #f5f5f5; border-radius: 4px;">
        <fieldset>
            <legend><strong>Filter by Device</strong></legend>
            <label for="filter_device_id">Select Device:</label>
            <select id="filter_device_id" name="filter_device_id">
                <option value="">-- All Devices --</option>
                <?php foreach ($devices as $device): ?>
                    <option value="<?= htmlspecialchars($device['devid']) ?>" 
                        <?= ($device['devid'] == $filterDeviceId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($device['devname']) ?> (ID: <?= htmlspecialchars($device['devid']) ?>)
                    </option>
                <?php endforeach; ?>
            </select>
            <button type="submit" name="filter_device">Apply Filter</button>
            <?php if (!empty($filterDeviceId)): ?>
                <button type="submit" name="filter_device">Clear Filter</button>
            <?php endif; ?>
        </fieldset>
    </form>

    <!-- Device Port Selection and Update Form -->
    <form method="post">
        <input type="hidden" name="current_filter" value="<?= htmlspecialchars($filterDeviceId) ?>">
        
        <label for="deviceport_key">Select Device Port to Update:</label>
        <select id="deviceport_key" name="deviceport_key" required>
            <option value="">-- Select Device Port --</option>
            <?php foreach ($filteredDevicePorts as $devicePort): ?>
                <?php
                    $key = htmlspecialchars($devicePort['devid']) . '|' . htmlspecialchars($devicePort['modelid']);
                    $isSelected = ($devicePort['devid'] == $selectedDeviceId && $devicePort['modelid'] == $selectedModelPortId);
                ?>
                <option value="<?= $key ?>" <?= $isSelected ? 'selected' : '' ?>>
                    <?= htmlspecialchars($devicePort['devname'] ?? 'N/A') ?> - <?= htmlspecialchars($devicePort['portname'] ?? 'N/A') ?>
                    (Device: <?= htmlspecialchars($devicePort['devid']) ?>, Port: <?= htmlspecialchars($devicePort['modelid']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_deviceport">Load Device Port</button>
        
        <?php if (!empty($filterDeviceId) && empty($filteredDevicePorts)): ?>
            <p style="color: #666; font-style: italic;">No ports found for the selected device.</p>
        <?php endif; ?>
        
        <br><br>

        <?php if (!empty($selectedDeviceId) && !empty($selectedModelPortId)): ?>
            <input type="hidden" name="device_id" value="<?= htmlspecialchars($selectedDeviceId) ?>">
            <input type="hidden" name="modelport_id" value="<?= htmlspecialchars($selectedModelPortId) ?>">

            <p><strong>Device:</strong> <?= htmlspecialchars($selectedDeviceLabel) ?> (ID: <?= htmlspecialchars($selectedDeviceId) ?>)</p>
            <p><strong>Port:</strong> <?= htmlspecialchars($selectedPortName) ?> (ID: <?= htmlspecialchars($selectedModelPortId) ?>)</p>

            <label for="mac_address">MAC Address (Optional):</label>
            <input type="text" id="mac_address" name="mac_address"
                   value="<?= htmlspecialchars($selectedMacAddress) ?>"
                   placeholder="e.g., AA:BB:CC:DD:EE:FF"
                   pattern="[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}"
                   title="MAC address format: XX:XX:XX:XX:XX:XX"><br><br>

            <fieldset>
                <legend>VLAN Configurations</legend>
                <div id="vlanConfigsContainer">
                    <?php foreach ($selectedVlanConfigs as $index => $vlanConfig): ?>
                        <div id="vlanConfig_<?= $index ?>" style="margin-bottom: 10px; padding: 10px; border: 1px solid #ddd; border-radius: 4px;">
                            <label>VLAN Number:</label>
                            <select name="vlan_numbers[]" required style="margin-right: 10px;">
                                <option value="">-- Select VLAN --</option>
                                <?php foreach ($vlans as $vlan): ?>
                                    <option value="<?= htmlspecialchars($vlan['vlanid']) ?>"
                                        <?= ($vlan['vlanid'] == $vlanConfig['vlan_number']) ? 'selected' : '' ?>>
                                        VLAN <?= htmlspecialchars($vlan['vlanid']) ?>
                                        <?php if (!empty($vlan['vlanname'])): ?>
                                            - <?= htmlspecialchars($vlan['vlanname']) ?>
                                        <?php endif; ?>
                                    </option>
                                <?php endforeach; ?>
                            </select>

                            <label>Type:</label>
                            <select name="vlan_tagged[]" required style="margin-right: 10px;">
                                <option value="true" <?= ($vlanConfig['tagged'] ?? false) ? 'selected' : '' ?>>Tagged</option>
                                <option value="false" <?= !($vlanConfig['tagged'] ?? false) ? 'selected' : '' ?>>Untagged</option>
                            </select>

                            <button type="button" onclick="removeVlanConfig(<?= $index ?>)">Remove</button>
                        </div>
                    <?php endforeach; ?>
                </div>
                <button type="button" onclick="addVlanConfig()">+ Add VLAN Configuration</button>
                <br><small><em>Note: Only one untagged VLAN is allowed per port by default (controlled by model port settings)</em></small>
            </fieldset>
            <br>

            <button type="submit" name="update_deviceport">Update Device Port</button>
        <?php endif; ?>
    </form>

    <script>
        let vlanConfigIndex = <?= count($selectedVlanConfigs) ?>;

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
                            VLAN <?= htmlspecialchars($vlan['vlanid']) ?>
                            <?php if (!empty($vlan['vlanname'])): ?>
                                - <?= htmlspecialchars($vlan['vlanname']) ?>
                            <?php endif; ?>
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
