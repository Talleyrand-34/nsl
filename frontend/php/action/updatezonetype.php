<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';
$selectedZoneTypeId = '';
$selectedZoneTypeName = '';

// Fetch existing zone types for selection
$zoneTypesJson = @file_get_contents(ZONETYPES_ENDPOINT);
$zoneTypes = json_decode($zoneTypesJson, true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a zone type selection (not form submission)
    if (isset($_POST['select_zonetype']) && !empty($_POST['zonetype_id'])) {
        $selectedZoneTypeId = trim($_POST['zonetype_id']);
        // Find the selected zone type to pre-fill the form
        foreach ($zoneTypes as $zoneType) {
            if ($zoneType['id'] == $selectedZoneTypeId) {
                $selectedZoneTypeName = $zoneType['name'];
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_zonetype'])) {
        $zoneTypeId = trim($_POST['zonetype_id'] ?? '');
        $newZoneTypeName = trim($_POST['zonetype_name'] ?? '');
        
        if ($zoneTypeId === '') {
            $message = 'Please select a zone type to update.';
        } elseif ($newZoneTypeName === '') {
            $message = 'Please enter a new zone type name.';
        } else {
            $data = json_encode([
                'id' => $zoneTypeId,
                'name' => $newZoneTypeName
            ]);
            $ch = curl_init(ZONETYPES_ENDPOINT);
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
                $message = 'Zone type updated successfully!';
                // Refresh list
                $zoneTypesJson = @file_get_contents(ZONETYPES_ENDPOINT);
                $zoneTypes = json_decode($zoneTypesJson, true) ?: [];
                // Clear selection after successful update
                $selectedZoneTypeId = '';
                $selectedZoneTypeName = '';
            } else {
                $message = 'Failed to update zone type. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Zone Type</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($zoneTypes)): ?>
    <p><em>No zone types available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="zonetype_id">Select Zone Type to Update:</label>
        <select id="zonetype_id" name="zonetype_id" required>
            <option value="">-- Select Zone Type --</option>
            <?php foreach ($zoneTypes as $zoneType): ?>
                <option value="<?= htmlspecialchars($zoneType['id']) ?>" 
                    <?= ($zoneType['id'] == ($selectedZoneTypeId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($zoneType['name']) ?> (ID: <?= htmlspecialchars($zoneType['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_zonetype">Load Zone Type</button>
        <br><br>
        
        <?php if (!empty($selectedZoneTypeId)): ?>
            <label for="zonetype_name">New Zone Type Name:</label>
            <input type="text" id="zonetype_name" name="zonetype_name" 
                   value="<?= htmlspecialchars($selectedZoneTypeName) ?>" required><br><br>
            <button type="submit" name="update_zonetype">Update Zone Type</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
