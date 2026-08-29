<?php
// SPDX-License-Identifier: MIT
// push.php — third tab in the app: Push config.
//
// Two-column layout, mirroring import.php:
//   .grid > .actions  — left column: operations over a selected device
//                        (device picker, intent editor, [Preview] [Dry-run]
//                         [Apply] [Rollback], audit log below).
//   .grid > .results  — right column: scoped topology preview of the device
//                        + its 1-hop neighbors (D2 inline diagram + table).
//
// Phase 2 lays down the skeleton. Real wiring lands in phases 5 and 6.

require_once __DIR__ . '/config.php';

$pageTitle = 'Push config — NSL-Graph';
$bodyClass = 'page-push';
include __DIR__ . '/header.php';
?>
<?php include __DIR__ . '/action/push_actions.php'; ?>
<?php include __DIR__ . '/footer.php'; ?>
