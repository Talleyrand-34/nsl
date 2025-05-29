<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing connection types for dropdown
$connectionTypesJson = @file_get_contents(CONNECTIONTYPES_ENDPOINT);
$connectionTypes = json_decode($connectionTypesJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $connectionType = trim($_POST['connection_type'] ?? '');

    if ($connectionType !== '') {
        $data = json_encode(['connection_type' => $connectionType]);

        $ch = curl_init(CONNECTIONTYPES_ENDPOINT);
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
            $message = "Connection type deleted successfully!";
            // Refresh connection types list after deletion
            $connectionTypesJson = @file_get_contents(CONNECTIONTYPES_ENDPOINT);
            $connectionTypes = json_decode($connectionTypesJson, true);
        } else {
            $message = "Failed to delete connection type. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a connection type.";
    }
}
?>

<h2>Delete a Connection Type</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($connectionTypes) && count($connectionTypes) > 0): ?>
<form method="post">
    <label for="connection_type">Select Connection Type to Delete:</label>
    <select id="connection_type" name="connection_type" required>
        <option value="">-- Select a Connection Type --</option>
        <?php foreach ($connectionTypes as $connectionType): ?>
            <option value="<?= htmlspecialchars($connectionType['name']) ?>">
                <?= htmlspecialchars($connectionType['name']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Delete Connection Type</button>
</form>
<?php else: ?>
    <p><em>No connection types available to delete.</em></p>
<?php endif; ?>