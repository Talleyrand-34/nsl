<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedProprietaryId = '';
$selectedProprietaryName = '';

// Fetch existing proprietaries for selection
$proprietariesJson = @file_get_contents(PROPRIETARIES_ENDPOINT);
$proprietaries = json_decode($proprietariesJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a proprietary selection (not form submission)
    if (isset($_POST['select_proprietary']) && !empty($_POST['proprietary_id'])) {
        $selectedProprietaryId = trim($_POST['proprietary_id']);
        // Find the selected proprietary to pre-fill the form
        foreach ($proprietaries as $proprietary) {
            if ($proprietary['id'] == $selectedProprietaryId) {
                $selectedProprietaryName = $proprietary['name'];
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_proprietary'])) {
        $proprietaryId = trim($_POST['proprietary_id'] ?? '');
        $newProprietaryName = trim($_POST['proprietary_name'] ?? '');
        
        if ($proprietaryId === '') {
            $message = 'Please select a proprietary to update.';
        } elseif ($newProprietaryName === '') {
            $message = 'Please enter a new proprietary name.';
        } else {
            $data = json_encode([
                'id' => $proprietaryId,
                'name' => $newProprietaryName
            ]);
            $ch = curl_init(PROPRIETARIES_ENDPOINT);
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
                $message = 'Proprietary updated successfully!';
                // Refresh list
                $proprietariesJson = @file_get_contents(PROPRIETARIES_ENDPOINT);
                $proprietaries = json_decode($proprietariesJson, true) ?: [];
                // Clear selection after successful update
                $selectedProprietaryId = '';
                $selectedProprietaryName = '';
            } else {
                $message = 'Failed to update proprietary. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Proprietary</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($proprietaries)): ?>
    <p><em>No proprietaries available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="proprietary_id">Select Proprietary to Update:</label>
        <select id="proprietary_id" name="proprietary_id" required>
            <option value="">-- Select Proprietary --</option>
            <?php foreach ($proprietaries as $proprietary): ?>
                <option value="<?= htmlspecialchars($proprietary['id']) ?>" 
                    <?= ($proprietary['id'] == ($selectedProprietaryId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($proprietary['name']) ?> (ID: <?= htmlspecialchars($proprietary['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_proprietary">Load Proprietary</button>
        <br><br>
        
        <?php if (!empty($selectedProprietaryId)): ?>
            <label for="proprietary_name">New Proprietary Name:</label>
            <input type="text" id="proprietary_name" name="proprietary_name" 
                   value="<?= htmlspecialchars($selectedProprietaryName) ?>" required><br><br>
            <button type="submit" name="update_proprietary">Update Proprietary</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
