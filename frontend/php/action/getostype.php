<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/../config.php';

$ostypesJson = @file_get_contents(OSTYPES_ENDPOINT);
$ostypes = json_decode($ostypesJson, true);

if (is_array($ostypes)) {
    echo "<ul>";
    foreach ($ostypes as $ostype) {
        echo "<li>" . htmlspecialchars($ostype['name']) . "</li>";
    }
    echo "</ul>";
} else {
    echo "<p><em>Could not fetch ostypes.</em></p>";
}
?>
