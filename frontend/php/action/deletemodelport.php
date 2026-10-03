<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';

// Fetch existing model ports for dropdown
$modelPortsJson = @file_get_contents(MODELPORTS_ENDPOINT);
$modelPorts = json_decode($modelPortsJson, true);

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $id = trim($_POST['id'] ?? '');

    if ($id !== '') {
        $data = json_encode(['model_port_id' => $id, 'cascade' => isset($_POST['cascade'])]);

        $ch = curl_init(MODELPORTS_ENDPOINT);
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
            $message = "Model port deleted successfully!";
            // Refresh model ports list after deletion
            $modelPortsJson = @file_get_contents(MODELPORTS_ENDPOINT);
            $modelPorts = json_decode($modelPortsJson, true);
        } else {
            $message = "Failed to delete model port. Server response: " . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = "Please select a model port.";
    }
}
?>

<h2>Delete a Model Port</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>

<?php if (is_array($modelPorts) && count($modelPorts) > 0): ?>
<form method="post">
    <label for="id">Select Model Port to Delete:</label>
    <select id="id" name="id" required>
        <option value="">-- Select a Model Port --</option>
        <?php foreach ($modelPorts as $modelPort): ?>
            <option value="<?= htmlspecialchars($modelPort['id']) ?>">
                ID: <?= htmlspecialchars($modelPort['id']) ?> - <?= htmlspecialchars($modelPort['name']) ?> (<?= htmlspecialchars($modelPort['model']) ?>)
            </option>
        <?php endforeach; ?>
    </select>
    <br><br>
    <label><input type="checkbox" name="cascade" value="1"> Delete on cascade (also remove dependent device ports, connections)</label>
    <br><br>
    <button type="submit">Delete Model Port</button>
</form>
<?php else: ?>
    <p><em>No model ports available to delete.</em></p>
<?php endif; ?>