<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing connections for dropdown
$connectionsJson = @file_get_contents(CONNECTIONS_ENDPOINT);
$connections = json_decode($connectionsJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $id = trim($_POST['id'] ?? '');

    if ($id !== '') {
        $data = json_encode(['id' => $id]);

        $ch = curl_init(CONNECTIONS_ENDPOINT);
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
            $message = "Connection deleted successfully!";
            // Refresh connections list after deletion
            $connectionsJson = @file_get_contents(CONNECTIONS_ENDPOINT);
            $connections = json_decode($connectionsJson, true);
        } else {
            $message = "Failed to delete connection. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a connection.";
    }
}
?>

<h2>Delete a Connection</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($connections) && count($connections) > 0): ?>
<form method="post">
    <label for="id">Select Connection to Delete:</label>
    <select id="id" name="id" required>
        <option value="">-- Select a Connection --</option>
        <?php foreach ($connections as $connection): ?>
            <option value="<?= htmlspecialchars($connection['id']) ?>">
                ID: <?= htmlspecialchars($connection['id']) ?> - <?= htmlspecialchars($connection['fromdevname']) ?> → <?= htmlspecialchars($connection['todevname']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Delete Connection</button>
</form>
<?php else: ?>
    <p><em>No connections available to delete.</em></p>
<?php endif; ?>