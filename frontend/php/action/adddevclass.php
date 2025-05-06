
<?php
require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $devclass = trim($_POST['devclass'] ?? '');

    if ($devclass !== '') {
        $data = json_encode(['devclass' => $devclass]);

        $ch = curl_init(DEVCLASSES_ENDPOINT);
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
            $message = "devclass added successfully!";
        } else {
            $message = "Failed to add devclass. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please enter a devclass name.";
    }
}
?>

<h2>Add a New devclass</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="devclass">devclass name:</label>
    <input type="text" id="devclass" name="devclass" required>
    <button type="submit">Add devclass</button>
</form>
