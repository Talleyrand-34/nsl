<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once 'config.php';

echo "<h1>Plugin API Test</h1>";

echo "<p><strong>API Base URL:</strong> " . htmlspecialchars(API_BASE_URL) . "</p>";
echo "<p><strong>Plugins Endpoint:</strong> " . htmlspecialchars(PLUGINS_ENDPOINT) . "</p>";

echo "<h2>Testing API Connection</h2>";

$ch = curl_init(PLUGINS_ENDPOINT);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_TIMEOUT, 5);
$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$error = curl_error($ch);
curl_close($ch);

echo "<p><strong>HTTP Status Code:</strong> " . $httpCode . "</p>";

if ($error) {
    echo "<p style='color: red;'><strong>cURL Error:</strong> " . htmlspecialchars($error) . "</p>";
}

if ($httpCode === 200) {
    echo "<p style='color: green;'><strong>Success!</strong> API is reachable.</p>";
    echo "<h3>Response:</h3>";
    echo "<pre>" . htmlspecialchars($response) . "</pre>";

    $plugins = json_decode($response, true);
    if ($plugins) {
        echo "<h3>Parsed Plugins:</h3>";
        echo "<ul>";
        foreach ($plugins as $plugin) {
            echo "<li>";
            echo "<strong>" . htmlspecialchars($plugin['name']) . "</strong> ";
            echo "(" . htmlspecialchars($plugin['id']) . ")";
            if ($plugin['is_active']) {
                echo " <span style='color: green;'>✓ ACTIVE</span>";
            }
            echo "<br><em>" . htmlspecialchars($plugin['description']) . "</em>";
            echo "</li>";
        }
        echo "</ul>";
    }
} else {
    echo "<p style='color: red;'><strong>Failed!</strong> Could not reach API.</p>";
    echo "<p><strong>Response:</strong> " . htmlspecialchars($response) . "</p>";
}

echo "<h2>JavaScript Test</h2>";
echo "<button onclick='testFetch()'>Test Fetch from Browser</button>";
echo "<div id='fetchResult' style='margin-top: 10px;'></div>";
?>

<script>
function testFetch() {
    const resultDiv = document.getElementById('fetchResult');
    resultDiv.innerHTML = '<p>Testing fetch...</p>';

    const url = '<?= PLUGINS_ENDPOINT ?>';
    console.log('Fetching from:', url);

    fetch(url)
        .then(response => {
            console.log('Response:', response);
            resultDiv.innerHTML += '<p>Status: ' + response.status + '</p>';
            return response.json();
        })
        .then(data => {
            console.log('Data:', data);
            resultDiv.innerHTML += '<p style="color: green;">Success! Got ' + data.length + ' plugins.</p>';
            resultDiv.innerHTML += '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
        })
        .catch(error => {
            console.error('Error:', error);
            resultDiv.innerHTML += '<p style="color: red;">Error: ' + error.message + '</p>';
        });
}
</script>
