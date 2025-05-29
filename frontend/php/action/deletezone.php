<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing zones for dropdown
$zonesJson = @file_get_contents(ZONES_ENDPOINT);
$zones = json_decode($zonesJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $id = trim($_POST['id'] ?? '');

    if ($id !== '') {
        $data = json_encode(['id' => $id]);

        $ch = curl_init(ZONES_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "DELETE");
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 204) {
            $message = "Zone deleted successfully!";
            // Refresh zones list after deletion
            $zonesJson = @file_get_contents(ZONES_ENDPOINT);
            $zones = json_decode($zonesJson, true);
        } else {
            $message = "Failed to delete zone. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a zone.";
    }
}
?>

<h2>Delete a Zone</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($zones) && count($zones) > 0): ?>
<form method="post">
    <label for="id">Select Zone to Delete:</label>
    <select id="id" name="id" required>
        <option value="">-- Select a Zone --</option>
        <?php foreach ($zones as $zone): ?>
            <option value="<?= htmlspecialchars($zone['id']) ?>">
                ID: <?= htmlspecialchars($zone['id']) ?> - <?= htmlspecialchars($zone['name']) ?> (<?= htmlspecialchars($zone['location_type']) ?>)
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Delete Zone</button>
</form>
<?php else: ?>
    <p><em>No zones available to delete.</em></p>
<?php endif; ?>