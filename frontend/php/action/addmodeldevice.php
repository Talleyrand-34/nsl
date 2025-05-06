
<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch select options
$brands = json_decode(@file_get_contents(BRANDS_ENDPOINT), true) ?: [];
$deviceclasses = json_decode(@file_get_contents(DEVCLASSES_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $model = trim($_POST['model'] ?? '');
    $brand = $_POST['brand'] ?? '';
    $class = $_POST['class'] ?? '';

    if ($model === '') {
        $message = 'Please enter a model name.';
    } elseif ($brand === '') {
        $message = 'Please select a brand.';
    } elseif ($class === '') {
        $message = 'Please select a device class.';
    } else {
        $data = json_encode([
            'model' => $model,
            'brand' => $brand,
            'class' => $class
        ]);

        $ch = curl_init(MODELS_ENDPOINT);
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
            $message = 'Model device added successfully!';
        } else {
            $message = 'Failed to add model device. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Add a New Model Device</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="model">Model name:</label>
    <input type="text" id="model" name="model" required><br><br>

    <label for="brand">Brand:</label>
    <select id="brand" name="brand" required>
        <option value="">-- Select --</option>
        <?php foreach ($brands as $b): ?>
            <option value="<?= htmlspecialchars($b['name']) ?>">
                <?= htmlspecialchars($b['name']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <label for="class">Device Class:</label>
    <select id="class" name="class" required>
        <option value="">-- Select --</option>
        <?php foreach ($deviceclasses as $dc): ?>
            <option value="<?= htmlspecialchars($dc['name']) ?>">
                <?= htmlspecialchars($dc['name']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <button type="submit">Add Model Device</button>
</form>
