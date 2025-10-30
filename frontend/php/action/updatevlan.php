<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing VLANs for selection
$vlansJson = @file_get_contents(VLANS_ENDPOINT);
$vlans = json_decode($vlansJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $vlanDbId = trim($_POST['vlan_db_id'] ?? '');
    $newVlanID = trim($_POST['new_vlan_id'] ?? '');
    $newVlanName = trim($_POST['new_vlan_name'] ?? '');

    if ($vlanDbId === '') {
        $message = 'Please select a VLAN to update.';
    } elseif ($newVlanID === '' && $newVlanName === '') {
        $message = 'Please enter at least one field to update (VLAN ID or VLAN Name).';
    } else {
        $data = json_encode([
            'id' => $vlanDbId,
            'vlanID' => $newVlanID,
            'vlanName' => $newVlanName
        ]);

        $ch = curl_init(VLANS_ENDPOINT);
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
            $message = 'VLAN updated successfully!';
            // Refresh VLANs list
            $vlansJson = @file_get_contents(VLANS_ENDPOINT);
            $vlans = json_decode($vlansJson, true) ?: [];
        } else {
            $message = 'Failed to update VLAN. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Update VLAN</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($vlans)): ?>
    <p><em>No VLANs available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="vlan_db_id">Select VLAN to Update:</label>
        <select id="vlan_db_id" name="vlan_db_id" required>
            <option value="">-- Select VLAN --</option>
            <?php foreach ($vlans as $vlan): ?>
                <option value="<?= htmlspecialchars($vlan['id']) ?>">
                    VLAN <?= htmlspecialchars($vlan['vlanid']) ?> - <?= htmlspecialchars($vlan['vlanname']) ?> (DB ID: <?= htmlspecialchars($vlan['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select><br><br>

        <label for="new_vlan_id">New VLAN ID (leave empty to keep current):</label>
        <input type="text" id="new_vlan_id" name="new_vlan_id"><br><br>

        <label for="new_vlan_name">New VLAN Name (leave empty to keep current):</label>
        <input type="text" id="new_vlan_name" name="new_vlan_name"><br><br>

        <button type="submit">Update VLAN</button>
    </form>
<?php endif; ?>
