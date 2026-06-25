<?php
require_once __DIR__ . '/../config.php';
$message = '';
$selectedZoneId = '';
$selectedZoneName = '';
$selectedFatherZoneId = '';
$selectedZoneTypeId = '';
$selectedOwnerId = '';

// Fetch data for form options
$zones = json_decode(@file_get_contents(ZONES_ENDPOINT), true) ?: [];
$zoneTypes = json_decode(@file_get_contents(ZONETYPES_ENDPOINT), true) ?: [];
$owners = json_decode(@file_get_contents(OWNERS_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    // Check if this is a zone selection (not form submission)
    if (isset($_POST['select_zone']) && !empty($_POST['zone_id'])) {
        echo $zone_id;
        echo $zones;
        $selectedZoneId = trim($_POST['zone_id']);
        // Find the selected zone to pre-fill the form
        foreach ($zones as $zone) {
            if ($zone['id'] == $selectedZoneId) {
                $selectedZoneName = $zone['name'];
                $selectedFatherZoneId = $zone['father_zone_id'] ?? '';
                $selectedZoneTypeId = $zone['zone_type_id'] ?? '';
                $selectedOwnerId = $zone['owner_id'] ?? '';
                break;
            }
        }
    }
    // This is the actual form submission
    elseif (isset($_POST['update_zone'])) {
        $zoneId = trim($_POST['zone_id'] ?? '');
        $newZoneName = trim($_POST['zone_name'] ?? '');
        $fatherZoneId = $_POST['father_zone_id'] ?? '';
        $zoneTypeId = $_POST['zone_type_id'] ?? '';
        $ownerId = $_POST['owner_id'] ?? '';
        
        if ($zoneId === '') {
            $message = 'Please select a zone to update.';
        } elseif ($newZoneName === '') {
            $message = 'Please enter a new zone name.';
        } elseif ($zoneTypeId === '') {
            $message = 'Please select a zone type.';
        } elseif ($ownerId === '') {
            $message = 'Please select a owner.';
        } else {
            $data = json_encode([
                'id' => $zoneId,
                'name' => $newZoneName,
                'father_zone_id' => $fatherZoneId === '' ? null : $fatherZoneId,
                'zone_type_id' => $zoneTypeId,
                'owner_id' => $ownerId
            ]);
            $ch = curl_init(ZONES_ENDPOINT);
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
                $message = 'Zone updated successfully!';
                // Refresh zones list
                $zones = json_decode(@file_get_contents(ZONES_ENDPOINT), true) ?: [];
                // Clear selection after successful update
                $selectedZoneId = '';
                $selectedZoneName = '';
                $selectedFatherZoneId = '';
                $selectedZoneTypeId = '';
                $selectedOwnerId = '';
            } else {
                $message = 'Failed to update zone. Server response: ' . htmlspecialchars($response);
            }
            curl_close($ch);
        }
    }
}
?>

<h2>Update Zone</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (empty($zones)): ?>
    <p><em>No zones available to update.</em></p>
<?php else: ?>
    <form method="post">
        <label for="zone_id">Select Zone to Update:</label>
        <select id="zone_id" name="zone_id" required>
            <option value="">-- Select Zone --</option>
            <?php foreach ($zones as $zone): ?>
                <option value="<?= htmlspecialchars($zone['id']) ?>" 
                    <?= ($zone['id'] == ($selectedZoneId ?? '')) ? 'selected' : '' ?>>
                    <?= htmlspecialchars($zone['name']) ?> (ID: <?= htmlspecialchars($zone['id']) ?>)
                </option>
            <?php endforeach; ?>
        </select>
        <button type="submit" name="select_zone">Load Zone</button>
        <br><br>
        
        <?php if (!empty($selectedZoneId)): ?>
            <label for="zone_name">New Zone Name:</label>
            <input type="text" id="zone_name" name="zone_name" 
                   value="<?= htmlspecialchars($selectedZoneName) ?>" required><br><br>
            
            <label for="father_zone_id">Parent Zone (optional):</label>
            <select id="father_zone_id" name="father_zone_id">
                <option value="">-- No Parent --</option>
                <?php foreach ($zones as $zone): ?>
                    <option value="<?= htmlspecialchars($zone['id']) ?>" 
                        <?= ($zone['id'] == $selectedFatherZoneId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($zone['name']) ?> (ID: <?= htmlspecialchars($zone['id']) ?>)
                    </option>
                <?php endforeach; ?>
            </select><br><br>
            
            <label for="zone_type_id">Zone Type:</label>
            <select id="zone_type_id" name="zone_type_id" required>
                <option value="">-- Select Zone Type --</option>
                <?php foreach ($zoneTypes as $zoneType): ?>
                    <option value="<?= htmlspecialchars($zoneType['id']) ?>" 
                        <?= ($zoneType['id'] == $selectedZoneTypeId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($zoneType['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>
            
            <label for="owner_id">Owner:</label>
            <select id="owner_id" name="owner_id" required>
                <option value="">-- Select Owner --</option>
                <?php foreach ($owners as $owner): ?>
                    <option value="<?= htmlspecialchars($owner['id']) ?>" 
                        <?= ($owner['id'] == $selectedOwnerId) ? 'selected' : '' ?>>
                        <?= htmlspecialchars($owner['name']) ?>
                    </option>
                <?php endforeach; ?>
            </select><br><br>
            
            <button type="submit" name="update_zone">Update Zone</button>
        <?php endif; ?>
    </form>
<?php endif; ?>
