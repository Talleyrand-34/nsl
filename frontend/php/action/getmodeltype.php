<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';

$modeltypesJson = @file_get_contents(MODELTYPES_ENDPOINT);
$modeltypes = json_decode($modeltypesJson, true);

if (is_array($modeltypes)) {
    echo "<ul>";
    foreach ($modeltypes as $modeltype) {
        echo "<li>" . htmlspecialchars($modeltype['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch modeltypes.</em></p>";
}
?>
