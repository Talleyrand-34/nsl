<?php
// SPDX-License-Identifier: MIT
// push_actions.php — dispatcher + two-column layout for push.php.
//
// Phase 2: skeleton. Phases 5–6 fill the panels.
require_once __DIR__ . '/push_preview_panel.php';
require_once __DIR__ . '/push_topology_panel.php';

$devices = [];
$raw = @file_get_contents(DEVICES_ENDPOINT);
if ($raw !== false) {
    $decoded = json_decode($raw, true);
    if (is_array($decoded)) {
        $devices = $decoded;
    }
}
?>
<div class="grid">
    <section class="actions">
        <?= push_preview_panel_html($devices) ?>
    </section>
    <section class="results">
        <?= push_topology_panel_html() ?>
    </section>
</div>
<script src="push.js"></script>
