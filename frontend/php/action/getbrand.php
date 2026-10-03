<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';

$brandsJson = @file_get_contents(BRANDS_ENDPOINT);
$brands = json_decode($brandsJson, true);

if (is_array($brands)) {
    echo "<ul>";
    foreach ($brands as $brand) {
        echo "<li>" . htmlspecialchars($brand['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch brands.</em></p>";
}
?>
