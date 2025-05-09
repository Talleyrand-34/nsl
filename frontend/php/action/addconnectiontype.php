
<?php
require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $connection_type = trim($_POST['connection_type'] ?? '');

    if ($connection_type !== '') {
        $data = json_encode(['name' => $connection_type]);

        $ch = curl_init(CONNECTIONTYPES_ENDPOINT);
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
            $message = 'connection_type added successfully!';
        } else {
            $message = 'Failed to add connection_type. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = 'Please enter a connection_type name.';
    }
}
?>

<h2>Add a New connection_type</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="connection_type">connection_type name:</label>
    <input type="text" id="connection_type" name="connection_type" required>
    <button type="submit">Add connection_type</button>
</form>
