
<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $modeltype = trim($_POST['modeltype'] ?? '');

    if ($modeltype !== '') {
        $data = json_encode(['name' => $modeltype]);

        $ch = curl_init(MODELTYPES_ENDPOINT);
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
            $message = 'modeltype added successfully!';
        } else {
            $message = 'Failed to add modeltype. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = 'Please enter a modeltype name.';
    }
}
?>

<h2>Add a New modeltype</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="modeltype">modeltype name:</label>
    <input type="text" id="modeltype" name="modeltype" required>
    <button type="submit">Add modeltype</button>
</form>
