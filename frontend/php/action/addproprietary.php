<?php
require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $proprietary = trim($_POST['proprietary'] ?? '');

    if ($proprietary !== '') {
        $data = json_encode(['name' => $proprietary]);

        $ch = curl_init(PROPRIETARIES_ENDPOINT);
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
            $message = "Proprietary added successfully!";
        } else {
            $message = "Failed to add proprietary. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please enter a proprietary name.";
    }
}
?>

<h2>Add a New Proprietary</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="proprietary">Proprietary name:</label>
    <input type="text" id="proprietary" name="proprietary" required>
    <button type="submit">Add Proprietary</button>
</form>
