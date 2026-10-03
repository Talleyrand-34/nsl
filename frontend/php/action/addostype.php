
<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';
$message = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $ostype = trim($_POST['ostype'] ?? '');

    if ($ostype !== '') {
        $data = json_encode(['name' => $ostype]);

        $ch = curl_init(OSTYPES_ENDPOINT);
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
            $message = 'ostype added successfully!';
        } else {
            $message = 'Failed to add ostype. Server response: ' . htmlspecialchars($response);
        }
        curl_close($ch);
    } else {
        $message = 'Please enter a ostype name.';
    }
}
?>

<h2>Add a New ostype</h2>
<?php if ($message): ?>
    <p><strong><?= htmlspecialchars($message) ?></strong></p>
<?php endif; ?>
<form method="post">
    <label for="ostype">ostype name:</label>
    <input type="text" id="ostype" name="ostype" required>
    <button type="submit">Add ostype</button>
</form>
