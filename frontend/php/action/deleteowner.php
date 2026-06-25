<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing owners for dropdown
$ownersJson = @file_get_contents(OWNERS_ENDPOINT);
$owners = json_decode($ownersJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $owner = trim($_POST['owner'] ?? '');

    if ($owner !== '') {
        $data = json_encode(['name' => $owner, 'cascade' => isset($_POST['cascade'])]);

        $ch = curl_init(OWNERS_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'DELETE');
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode >= 200 && $httpCode < 300) {
            $message = 'Owner deleted successfully!';
            // Refresh owners list after deletion
            $ownersJson = @file_get_contents(OWNERS_ENDPOINT);
            $owners = json_decode($ownersJson, true);
        } else {
            $message = 'Failed to delete owner. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = 'Please select a owner.';
    }
}
?>

<h2>Delete a Owner</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($owners) && count($owners) > 0): ?>
<form method="post">
    <label for="owner">Select Owner to Delete:</label>
    <select id="owner" name="owner" required>
        <option value="">-- Select a Owner --</option>
        <?php foreach ($owners as $owner): ?>
            <option value="<?= htmlspecialchars($owner['owner']) ?>">
                <?= htmlspecialchars($owner['name']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <br><br>
    <label><input type="checkbox" name="cascade" value="1"> Delete on cascade (also remove dependents)</label>
    <br><br>
    <button type="submit">Delete Owner</button>
</form>
<?php else: ?>
    <p><em>No owners available to delete.</em></p>
<?php endif; ?>
