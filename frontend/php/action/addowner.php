<?php
require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $owner = trim($_POST['owner'] ?? '');

    if ($owner !== '') {
        $data = json_encode(['name' => $owner]);

        $ch = curl_init(OWNERS_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "POST");
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 201) {
            $message = "Owner added successfully!";
        } else {
            $message = "Failed to add owner. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please enter a owner name.";
    }
}
?>

<h2>Add a New Owner</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="owner">Owner name:</label>
    <input type="text" id="owner" name="owner" required>
    <button type="submit">Add Owner</button>
</form>
