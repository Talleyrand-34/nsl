<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedOwnerId = '';
$selectedOwnerName = '';

// Fetch existing owners for selection
$ownersJson = @file_get_contents(OWNERS_ENDPOINT);
$owners = json_decode($ownersJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a owner selection (not form submission)
    if (isset($_POST['select_owner']) && !empty($_POST['owner_id'])) {
        $selectedOwnerId = trim($_POST['owner_id']);
        // Find the selected owner to pre-fill the form
        foreach ($owners as $owner) {
            if ($owner['id'] == $selectedOwnerId) {
                $selectedOwnerName = $owner['name'];
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_owner'])) {
        $ownerId = trim($_POST['owner_id'] ?? '');
        $newOwnerName = trim($_POST['owner_name'] ?? '');
        
        if ($ownerId === '') {
            $message = 'Please select a owner to update.';
        } elseif ($newOwnerName === '') {
            $message = 'Please enter a new owner name.';
        } else {
            $data = json_encode([
                'id' => $ownerId,
                'name' => $newOwnerName
            ]);
            $ch = curl_init(OWNERS_ENDPOINT);
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
                $message = 'Owner updated successfully!';
                // Refresh list
                $ownersJson = @file_get_contents(OWNERS_ENDPOINT);
                $owners = json_decode($ownersJson, true) ?: [];
                // Clear selection after successful update
                $selectedOwnerId = '';
                $selectedOwnerName = '';
            } else {
                $message = 'Failed to update owner. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Owner</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($owners)): ?>
    <p><em>No owners available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="owner_id">Select Owner to Update:</label>
        <select id="owner_id" name="owner_id" required>
            <option value="">-- Select Owner --</option>
            <?php foreach ($owners as $owner): ?>
                <option value="<?= htmlspecialchars($owner['id']) ?>" 
                    <?= ($owner['id'] == ($selectedOwnerId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($owner['name']) ?> (ID: <?= htmlspecialchars($owner['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_owner">Load Owner</button>
        <br><br>
        
        <?php if (!empty($selectedOwnerId)): ?>
            <label for="owner_name">New Owner Name:</label>
            <input type="text" id="owner_name" name="owner_name" 
                   value="<?= htmlspecialchars($selectedOwnerName) ?>" required><br><br>
            <button type="submit" name="update_owner">Update Owner</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
