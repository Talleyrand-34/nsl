<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing VLANs for dropdown
$vlansJson = @file_get_contents(VLANS_ENDPOINT);
$vlans = json_decode($vlansJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $vlanDbId = trim($_POST['vlan_db_id'] ?? '');

    if ($vlanDbId !== '') {
        $data = json_encode(['vlanId' => $vlanDbId]);

        $ch = curl_init(VLANS_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "DELETE");
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 204 || $httpCode === 200) {
            $message = "VLAN deleted successfully!";
            // Refresh VLANs list after deletion
            $vlansJson = @file_get_contents(VLANS_ENDPOINT);
            $vlans = json_decode($vlansJson, true);
        } else {
            $message = "Failed to delete VLAN. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a VLAN.";
    }
}
?>

<h2>Delete a VLAN</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($vlans) && count($vlans) > 0): ?>
<form method="post">
    <label for="vlan_db_id">Select VLAN to Delete:</label>
    <select id="vlan_db_id" name="vlan_db_id" required>
        <option value="">-- Select a VLAN --</option>
        <?php foreach ($vlans as $vlan): ?>
            <option value="<?= htmlspecialchars($vlan['id']) ?>">
                VLAN <?= htmlspecialchars($vlan['vlanid']) ?> - <?= htmlspecialchars($vlan['vlanname']) ?> (DB ID: <?= htmlspecialchars($vlan['id']) ?>)
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Delete VLAN</button>
</form>
<?php else: ?>
    <p><em>No VLANs available to delete.</em></p>
<?php endif; ?>
