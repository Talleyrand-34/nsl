
<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch select options
$zones = json_decode(@file_get_contents(ZONES_ENDPOINT), true) ?: [];
$owners = json_decode(@file_get_contents(OWNERS_ENDPOINT), true) ?: [];
$zonetypes = json_decode(@file_get_contents(ZONETYPES_ENDPOINT), true) ?: [];

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $name = trim($_POST['name'] ?? '');
    $fatherid = $_POST['fatherid'] ?? '';
    $owner = $_POST['owner'] ?? '';
    $location_type = $_POST['location_type'] ?? '';

    if ($name === '') {
        $message = 'Please enter a zone name.';
    } elseif ($owner === '') {
        $message = 'Please select a owner.';
    } else {
        $data = json_encode([
            'name' => $name,
            'father' => '',  // Always empty
            'fatherid' => $fatherid,  // This is the zone id as string, or empty string if none selected
            'owner' => $owner,
            'location_type' => $location_type
        ]);
        // Echo the JSON for debugging
        /* echo '<pre>JSON sent:<br>' . htmlspecialchars($data) . '</pre>'; */
        $ch = curl_init(ZONES_ENDPOINT);
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
            $message = 'Zone added successfully!';
        } else {
            $message = 'Failed to add zone. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    }
}
?>

<h2>Add a New Zone</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="name">Zone name:</label>
    <input type="text" id="name" name="name" required><br><br>

    <label for="fatherid">Parent Zone (optional):</label>
    <select id="fatherid" name="fatherid">
        <option value="">-- None --</option>
        <?php foreach ($zones as $zone): ?>
            <option value="<?= htmlspecialchars($zone['id']) ?>">
                <?= htmlspecialchars($zone['name']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <!-- Owner -->
    <label for="owner">Owner:</label>
    <select id="owner" name="owner" required>
        <option value="">-- Select --</option>
        <?php foreach ($owners as $prop): ?>
            <option value="<?= htmlspecialchars($prop['name']) ?>">
                <?= htmlspecialchars($prop['name']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <!-- Zone Type -->
    <label for="location_type">Zone Type:</label>
    <select id="location_type" name="location_type" required>
        <option value="">-- Select --</option>
        <?php foreach ($zonetypes as $zt): ?>
            <option value="<?= htmlspecialchars($zt['name']) ?>">
                <?= htmlspecialchars($zt['name']) ?>
            </option>
        <?php endforeach; ?>
    </select><br><br>

    <button type="submit">Add Zone</button>
</form>
