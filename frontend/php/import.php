<?php
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

require_once __DIR__ . '/config.php';
$pageTitle = 'Import devices — NSL-Graph';
$bodyClass = 'page-import';
include __DIR__ . '/header.php';
?>
<?php include __DIR__ . '/action/importscan.php'; ?>
<?php include __DIR__ . '/footer.php'; ?>
