<?php
require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing proprietaries for dropdown
$proprietariesJson = @file_get_contents(PROPRIETARIES_ENDPOINT);
$proprietaries = json_decode($proprietariesJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $proprietary = trim($_POST['proprietary'] ?? '');

    if ($proprietary !== '') {
        $data = json_encode(['proprietary' => $proprietary]);

        $ch = curl_init(PROPRIETARIES_ENDPOINT);
        curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'DELETE');
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, [
            'Content-Type: application/json',
            'Content-Length: ' . strlen($data)
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

        if ($httpCode === 204) {
            $message = 'Proprietary deleted successfully!';
            // Refresh proprietaries list after deletion
            $proprietariesJson = @file_get_contents(PROPRIETARIES_ENDPOINT);
            $proprietaries = json_decode($proprietariesJson, true);
        } else {
            $message = 'Failed to delete proprietary. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = 'Please select a proprietary.';
    }
}
?>

<h2>Delete a Proprietary</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($proprietaries) && count($proprietaries) > 0): ?>
<form method="post">
    <label for="proprietary">Select Proprietary to Delete:</label>
    <select id="proprietary" name="proprietary" required>
        <option value="">-- Select a Proprietary --</option>
        <?php foreach ($proprietaries as $proprietary): ?>
            <option value="<?= htmlspecialchars($proprietary['proprietary']) ?>">
                <?= htmlspecialchars($proprietary['name']) ?>
            </option>
        <?php endforeach; ?>
    </select>
    <button type="submit">Delete Proprietary</button>
</form>
<?php else: ?>
    <p><em>No proprietaries available to delete.</em></p>
<?php endif; ?>
