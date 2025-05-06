
<?php
require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $brand = trim($_POST['brand'] ?? '');

    if ($brand !== '') {
        $data = json_encode(['brand' => $brand]);

        $ch = curl_init(BRANDS_ENDPOINT);
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
            $message = "Brand added successfully!";
        } else {
            $message = "Failed to add brand. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please enter a brand name.";
    }
}
?>

<h2>Add a New Brand</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="brand">Brand name:</label>
    <input type="text" id="brand" name="brand" required>
    <button type="submit">Add Brand</button>
</form>
