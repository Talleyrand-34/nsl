<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedBrandId = '';
$selectedBrandName = '';

// Fetch existing brands for selection
$brandsJson = @file_get_contents(BRANDS_ENDPOINT);
$brands = json_decode($brandsJson, true) ?: [];

// DEBUG: Let's see what we're actually getting
echo "<!-- DEBUG: Raw JSON: " . htmlspecialchars($brandsJson ?: 'NULL/FALSE') . " -->";
echo "<!-- DEBUG: Decoded array count: " . count($brands) . " -->";
if (!empty($brands)) {
    echo "<!-- DEBUG: First brand: " . htmlspecialchars(print_r($brands[0], true)) . " -->";
}

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a brand selection (not form submission)
    if (isset($_POST['select_brand']) && !empty($_POST['brand_id'])) {
        $selectedBrandId = trim($_POST['brand_id']);
        // Find the selected brand to pre-fill the form
        foreach ($brands as $brand) {
            if ($brand['id'] == $selectedBrandId) {
                $selectedBrandName = $brand['name'];
                break;
            }
        }
    } 
    // This is the actual form submission
    elseif (isset($_POST['update_brand'])) {
        $brandId = trim($_POST['brand_id'] ?? '');
        $newBrandName = trim($_POST['brand_name'] ?? '');
        
        if ($brandId === '') {
            $message = 'Please select a brand to update.';
        } elseif ($newBrandName === '') {
            $message = 'Please enter a new brand name.';
        } else {
            $data = json_encode([
                'id' => $brandId,
                'name' => $newBrandName
            ]);
            $ch = curl_init(BRANDS_ENDPOINT);
            curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'PUT');
            curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
            curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
            curl_setopt($ch, CURLOPT_HTTPHEADER, [
                'Content-Type: application/json',
                'Content-Length: ' . strlen($data)
            ]);
            $response = curl_exec($ch);
            $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
            if ($httpCode === 200) {
                $message = 'Brand updated successfully!';
                // Refresh brands list
                $brandsJson = @file_get_contents(BRANDS_ENDPOINT);
                $brands = json_decode($brandsJson, true) ?: [];
                // Clear selection after successful update
                $selectedBrandId = '';
                $selectedBrandName = '';
            } else {
                $message = 'Failed to update brand. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Brand</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($brands)): ?>
    <p><em>No brands available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="brand_id">Select Brand to Update:</label>
        <select id="brand_id" name="brand_id" required>
            <option value="">-- Select Brand --</option>
            <?php foreach ($brands as $brand): ?>
                <option value="<?= htmlspecialchars($brand['id']) ?>" 
                    <?= ($brand['id'] == ($selectedBrandId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($brand['name']) ?> (ID: <?= htmlspecialchars($brand['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_brand">Load Brand</button>
        <br><br>
        
        <?php if (!empty($selectedBrandId)): ?>
            <label for="brand_name">New Brand Name:</label>
            <input type="text" id="brand_name" name="brand_name" 
                   value="<?= htmlspecialchars($selectedBrandName) ?>" required><br><br>
            <button type="submit" name="update_brand">Update Brand</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
