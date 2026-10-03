<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing device classes for dropdown
$modelTypesJson = @file_get_contents(MODELTYPES_ENDPOINT);
$modelTypes = json_decode($modelTypesJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $name = trim($_POST['name'] ?? '');

    if ($name !== '') {
        $data = json_encode(['name' => $name, 'cascade' => isset($_POST['cascade'])]);

        $ch = curl_init(MODELTYPES_ENDPOINT);
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
            $message = "Device class deleted successfully!";
            // Refresh device classes list after deletion
            $modelTypesJson = @file_get_contents(MODELTYPES_ENDPOINT);
            $modelTypes = json_decode($modelTypesJson, true);
        } else {
            $message = "Failed to delete device class. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a device class.";
    }
}
?>

<h2>Delete a Device Class</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($modelTypes) && count($modelTypes) > 0): ?>
<form method="post">
    <label for="name">Select Device Class to Delete:</label>
    <select id="name" name="name" required>
        <option value="">-- Select a Device Class --</option>
        <?php foreach ($modelTypes as $modelType): ?>
            <option value="<?= htmlspecialchars($modelType['name']) ?>">
                <?= htmlspecialchars($modelType['name']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <br><br>
    <label><input type="checkbox" name="cascade" value="1"> Delete on cascade (also remove dependents)</label>
    <br><br>
    <button type="submit">Delete Device Class</button>
</form>
<?php else: ?>
    <p><em>No device classes available to delete.</em></p>
<?php endif; ?>