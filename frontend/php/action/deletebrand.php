<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing brands for dropdown
$brandsJson = @file_get_contents(BRANDS_ENDPOINT);
$brands = json_decode($brandsJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $brand = trim($_POST['brand'] ?? '');
    $cascade = isset($_POST['cascade']);

    if ($brand !== '') {
        $data = json_encode(['name' => $brand, 'cascade' => $cascade]);

        $ch = curl_init(BRANDS_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "DELETE");
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode >= 200 && $httpCode < 300) {
            $message = "Brand deleted successfully!";
            // Refresh brands list after deletion
            $brandsJson = @file_get_contents(BRANDS_ENDPOINT);
            $brands = json_decode($brandsJson, true);
        } else {
            $message = "Failed to delete brand. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a brand.";
    }
}
?>

<h2>Delete a Brand</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($brands) && count($brands) > 0): ?>
<form method="post">
    <label for="brand">Select Brand to Delete:</label>
    <select id="brand" name="brand" required>
        <option value="">-- Select a Brand --</option>
        <?php foreach ($brands as $brand): ?>
            <option value="<?= htmlspecialchars($brand['name']) ?>">
                <?= htmlspecialchars($brand['name']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <br><br>
    <label><input type="checkbox" name="cascade" value="1"> Delete on cascade (also remove dependents)</label>
    <br><br>
    <button type="submit">Delete Brand</button>
</form>
<?php else: ?>
    <p><em>No brands available to delete.</em></p>
<?php endif; ?>