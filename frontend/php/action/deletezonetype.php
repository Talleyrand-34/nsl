<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing zone types for dropdown
$zoneTypesJson = @file_get_contents(ZONETYPES_ENDPOINT);
$zoneTypes = json_decode($zoneTypesJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $locationType = trim($_POST['location_type'] ?? '');

    if ($locationType !== '') {
        $data = json_encode(['name' => $locationType, 'cascade' => isset($_POST['cascade'])]);

        $ch = curl_init(ZONETYPES_ENDPOINT);
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
            $message = "Zone type deleted successfully!";
            // Refresh zone types list after deletion
            $zoneTypesJson = @file_get_contents(ZONETYPES_ENDPOINT);
            $zoneTypes = json_decode($zoneTypesJson, true);
        } else {
            $message = "Failed to delete zone type. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a zone type.";
    }
}
?>

<h2>Delete a Zone Type</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($zoneTypes) && count($zoneTypes) > 0): ?>
<form method="post">
    <label for="location_type">Select Zone Type to Delete:</label>
    <select id="location_type" name="location_type" required>
        <option value="">-- Select a Zone Type --</option>
        <?php foreach ($zoneTypes as $zoneType): ?>
            <option value="<?= htmlspecialchars($zoneType['location_type']) ?>">
                <?= htmlspecialchars($zoneType['location_type']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <br><br>
    <label><input type="checkbox" name="cascade" value="1"> Delete on cascade (also remove dependents)</label>
    <br><br>
    <button type="submit">Delete Zone Type</button>
</form>
<?php else: ?>
    <p><em>No zone types available to delete.</em></p>
<?php endif; ?>