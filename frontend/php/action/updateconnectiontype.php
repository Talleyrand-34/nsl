<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedConnectionTypeId = '';
$selectedConnectionTypeName = '';

// Fetch existing connection types for selection
$connectionTypesJson = @file_get_contents(CONNECTIONTYPES_ENDPOINT);
$connectionTypes = json_decode($connectionTypesJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a connection type selection (not form submission)
    if (isset($_POST['select_connectiontype']) && !empty($_POST['connectiontype_id'])) {
        $selectedConnectionTypeId = trim($_POST['connectiontype_id']);
        // Find the selected connection type to pre-fill the form
        foreach ($connectionTypes as $connectionType) {
            if ($connectionType['id'] == $selectedConnectionTypeId) {
                $selectedConnectionTypeName = $connectionType['name'];
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_connectiontype'])) {
        $connectionTypeId = trim($_POST['connectiontype_id'] ?? '');
        $newConnectionTypeName = trim($_POST['connectiontype_name'] ?? '');
        
        if ($connectionTypeId === '') {
            $message = 'Please select a connection type to update.';
        } elseif ($newConnectionTypeName === '') {
            $message = 'Please enter a new connection type name.';
        } else {
            $data = json_encode([
                'id' => $connectionTypeId,
                'name' => $newConnectionTypeName
            ]);
            $ch = curl_init(CONNECTIONTYPES_ENDPOINT);
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
                $message = 'Connection type updated successfully!';
                // Refresh list
                $connectionTypesJson = @file_get_contents(CONNECTIONTYPES_ENDPOINT);
                $connectionTypes = json_decode($connectionTypesJson, true) ?: [];
                // Clear selection after successful update
                $selectedConnectionTypeId = '';
                $selectedConnectionTypeName = '';
            } else {
                $message = 'Failed to update connection type. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Connection Type</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($connectionTypes)): ?>
    <p><em>No connection types available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="connectiontype_id">Select Connection Type to Update:</label>
        <select id="connectiontype_id" name="connectiontype_id" required>
            <option value="">-- Select Connection Type --</option>
            <?php foreach ($connectionTypes as $connectionType): ?>
                <option value="<?= htmlspecialchars($connectionType['id']) ?>" 
                    <?= ($connectionType['id'] == ($selectedConnectionTypeId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($connectionType['name']) ?> (ID: <?= htmlspecialchars($connectionType['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_connectiontype">Load Connection Type</button>
        <br><br>
        
        <?php if (!empty($selectedConnectionTypeId)): ?>
            <label for="connectiontype_name">New Connection Type Name:</label>
            <input type="text" id="connectiontype_name" name="connectiontype_name" 
                   value="<?= htmlspecialchars($selectedConnectionTypeName) ?>" required><br><br>
            <button type="submit" name="update_connectiontype">Update Connection Type</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
